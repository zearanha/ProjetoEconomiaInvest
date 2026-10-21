package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"painel-mercado/internal/config"
	httpapi "painel-mercado/internal/http"
	"painel-mercado/internal/jobs"
	"painel-mercado/internal/providers"
	bcbprovider "painel-mercado/internal/providers/bcb"
	cryptoprovider "painel-mercado/internal/providers/crypto"
	fxprovider "painel-mercado/internal/providers/fx"
	rssprovider "painel-mercado/internal/providers/rss"
	stockprovider "painel-mercado/internal/providers/stocks"
	"painel-mercado/internal/repository"
	"painel-mercado/internal/service"
	"painel-mercado/internal/ws"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	client := &http.Client{Timeout: cfg.ProviderTimeout}
	quotes := service.NewQuotes(
		cryptoprovider.NewBinance(client, cfg.CryptoSymbols),
		fxprovider.NewAwesomeAPI(client),
		stockprovider.NewBrapi(client, cfg.BrapiToken, cfg.StockSymbols),
	)
	var quoteHistory repository.QuoteHistory = repository.NewMemoryQuoteHistory()
	var rateProvider providers.Rates
	switch strings.ToLower(cfg.MacroProvider) {
	case "brapi":
		if cfg.BrapiToken == "" {
			logger.Error("MACRO_PROVIDER=brapi exige BRAPI_TOKEN; usando SGS do Banco Central")
			rateProvider = bcbprovider.NewSGS(client)
		} else {
			rateProvider = bcbprovider.NewBrapiMacro(client, cfg.BrapiToken)
		}
	default:
		rateProvider = bcbprovider.NewSGS(client)
	}
	fixedIncome := service.NewFixedIncome(rateProvider)
	var newsRepo repository.News = repository.NewMemoryNews()
	if cfg.DatabaseURL != "" {
		dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		postgresRepo, err := repository.NewPostgresNews(dbCtx, cfg.DatabaseURL)
		cancel()
		if err != nil {
			logger.Error("falha ao iniciar persistência PostgreSQL", "error", err)
			os.Exit(1)
		}
		defer postgresRepo.Close()
		newsRepo = postgresRepo
		historyCtx, historyCancel := context.WithTimeout(context.Background(), 10*time.Second)
		postgresHistory, historyErr := repository.NewPostgresQuoteHistory(historyCtx, cfg.DatabaseURL)
		historyCancel()
		if historyErr != nil {
			logger.Error("falha ao iniciar histórico de cotações", "error", historyErr)
			os.Exit(1)
		}
		defer postgresHistory.Close()
		quoteHistory = postgresHistory
		logger.Info("histórico persistente de cotações PostgreSQL ativado")
		logger.Info("persistência de notícias PostgreSQL ativada")
	} else {
		logger.Warn("DATABASE_URL não configurada; notícias ficarão em memória")
	}
	quotes.SetHistory(quoteHistory)
	newsService := service.NewNews(rssprovider.NewProvider(client), newsRepo)
	quoteHub := ws.NewHub()
	server := &http.Server{
		Addr: cfg.HTTPAddr, Handler: httpapi.NewServer(quotes, fixedIncome, newsService, quoteHub, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go jobs.RunNews(ctx, cfg.NewsRefreshInterval, newsService, logger)
	go jobs.RunQuoteUpdates(ctx, cfg.QuotesRefreshInterval, quotes, quoteHub, logger)
	go func() {
		logger.Info("servidor iniciado", "addr", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("falha no servidor", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("falha ao encerrar servidor", "error", err)
	}
}
