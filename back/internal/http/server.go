package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"painel-mercado/internal/service"
	"painel-mercado/internal/ws"
)

type Server struct {
	quotes         *service.Quotes
	fixedIncome    *service.FixedIncome
	news           *service.News
	updates        *ws.Hub
	logger         *slog.Logger
	frontendOrigin string
}

func NewServer(quotes *service.Quotes, fixedIncome *service.FixedIncome, news *service.News, updates *ws.Hub, logger *slog.Logger, frontendOrigin string) http.Handler {
	s := &Server{quotes: quotes, fixedIncome: fixedIncome, news: news, updates: updates, logger: logger, frontendOrigin: strings.TrimRight(frontendOrigin, "/")}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/quotes", s.getQuotes)
	mux.HandleFunc("GET /api/quotes/history", s.getQuoteHistory)
	mux.HandleFunc("GET /api/fixed-income/rates", s.getFixedIncomeRates)
	mux.HandleFunc("POST /api/fixed-income/simulate", s.simulateFixedIncome)
	mux.HandleFunc("GET /api/news", s.getNews)
	mux.Handle("GET /ws", updates)
	return s.logging(recoverer(s.cors(mux)))
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.frontendOrigin == "" {
			next.ServeHTTP(w, r)
			return
		}
		origin := strings.TrimRight(r.Header.Get("Origin"), "/")
		if !isAllowedOrigin(origin, s.frontendOrigin) {
			if r.Method == http.MethodOptions || origin != "" {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "origem não permitida"})
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(origin, allowed string) bool {
	if origin == allowed {
		return true
	}
	if !strings.HasPrefix(allowed, ".") {
		return false
	}
	parsed, err := url.Parse(origin)
	return err == nil && parsed.Scheme == "https" && strings.HasSuffix(parsed.Hostname(), allowed)
}

func (s *Server) getQuoteHistory(w http.ResponseWriter, r *http.Request) {
	symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
	period := r.URL.Query().Get("range")
	if symbol == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "symbol é obrigatório"})
		return
	}
	if period == "" {
		period = "1d"
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	points, err := s.quotes.History(ctx, symbol, period)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "range deve ser 1d, 1w ou 1m"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"symbol": symbol, "range": period, "data": points})
}

func (s *Server) getNews(w http.ResponseWriter, r *http.Request) {
	page := 1
	if rawPage := r.URL.Query().Get("page"); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err != nil || parsed < 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "page deve ser um inteiro maior que zero"})
			return
		}
		page = parsed
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result, err := s.news.List(ctx, service.NewsFilter{Category: r.URL.Query().Get("category"), Query: r.URL.Query().Get("q"), Page: page})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "category deve ser brasil, global, crypto, fixed-income ou stocks"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) getFixedIncomeRates(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	rates, err := s.fixedIncome.Rates(ctx)
	if err != nil {
		s.logger.Error("falha ao buscar indicadores de renda fixa", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "indicadores indisponíveis no momento"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": rates})
}

func (s *Server) simulateFixedIncome(w http.ResponseWriter, r *http.Request) {
	var input service.SimulationInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "corpo JSON inválido"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	result, err := s.fixedIncome.Simulate(ctx, input)
	if errors.Is(err, service.ErrInvalidSimulation) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valor, prazo ou percentual do CDI fora do intervalo permitido"})
		return
	}
	if err != nil {
		s.logger.Error("falha ao simular renda fixa", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "não foi possível obter a taxa CDI para a simulação"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getQuotes(w http.ResponseWriter, r *http.Request) {
	market := strings.ToLower(r.URL.Query().Get("market"))
	if market != "crypto" && market != "fx" && market != "stocks" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "market deve ser crypto, fx ou stocks"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	quotes, err := s.quotes.Get(ctx, market)
	if err != nil {
		s.logger.Error("falha ao buscar cotações", "market", market, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "provedor de cotações indisponível"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"market": market, "data": quotes})
}

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro interno"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
