package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"painel-mercado/internal/domain"
	"painel-mercado/internal/providers"
	"painel-mercado/internal/repository"
)

type Quotes struct {
	crypto, fx, stocks providers.Quotes
	history            repository.QuoteHistory
	cache              map[string]*quoteCache
}

const quoteCacheTTL = time.Minute
const cryptoQuoteCacheTTL = 5 * time.Minute
const fxQuoteCacheTTL = time.Hour

type quoteCache struct {
	mu          sync.Mutex
	items       []domain.Quote
	nextAttempt time.Time
	lastErr     error
}

func NewQuotes(crypto, fx, stocks providers.Quotes) *Quotes {
	return &Quotes{
		crypto: crypto, fx: fx, stocks: stocks,
		cache: map[string]*quoteCache{"crypto": {}, "fx": {}, "stocks": {}},
	}
}

func (s *Quotes) SetHistory(history repository.QuoteHistory) { s.history = history }

func (s *Quotes) History(ctx context.Context, symbol, period string) ([]domain.QuotePoint, error) {
	durations := map[string]time.Duration{"1d": 24 * time.Hour, "1w": 7 * 24 * time.Hour, "1m": 30 * 24 * time.Hour}
	duration, ok := durations[period]
	if !ok {
		return nil, fmt.Errorf("período inválido: %s", period)
	}
	if s.history == nil {
		return []domain.QuotePoint{}, nil
	}
	return s.history.List(ctx, symbol, time.Now().UTC().Add(-duration))
}

func (s *Quotes) Get(ctx context.Context, market string) ([]domain.Quote, error) {
	var provider providers.Quotes
	switch market {
	case "crypto":
		provider = s.crypto
	case "fx":
		provider = s.fx
	case "stocks":
		provider = s.stocks
	default:
		return nil, fmt.Errorf("mercado inválido: %s", market)
	}
	cache := s.cache[market]
	cache.mu.Lock()
	defer cache.mu.Unlock()
	now := time.Now()
	if now.Before(cache.nextAttempt) {
		if len(cache.items) > 0 {
			return cloneQuotes(cache.items), nil
		}
		return nil, cache.lastErr
	}

	items, err := provider.Quotes(ctx)
	if err != nil {
		cache.nextAttempt = now.Add(cacheTTL(market))
		cache.lastErr = err
		if len(cache.items) > 0 {
			slog.Warn("provedor indisponível; servindo última cotação em cache", "market", market, "error", err)
			return cloneQuotes(cache.items), nil
		}
		return nil, err
	}
	cache.items = cloneQuotes(items)
	cache.nextAttempt = now.Add(cacheTTL(market))
	cache.lastErr = nil
	if s.history != nil {
		saveCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.history.Save(saveCtx, items)
	}
	return cloneQuotes(items), nil
}

func cloneQuotes(quotes []domain.Quote) []domain.Quote {
	return append([]domain.Quote(nil), quotes...)
}

func cacheTTL(market string) time.Duration {
	switch market {
	case "crypto":
		return cryptoQuoteCacheTTL
	case "fx":
		return fxQuoteCacheTTL
	}
	return quoteCacheTTL
}
