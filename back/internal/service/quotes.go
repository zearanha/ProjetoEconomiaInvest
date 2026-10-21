package service

import (
	"context"
	"fmt"
	"time"

	"painel-mercado/internal/domain"
	"painel-mercado/internal/providers"
	"painel-mercado/internal/repository"
)

type Quotes struct {
	crypto, fx, stocks providers.Quotes
	history            repository.QuoteHistory
}

func NewQuotes(crypto, fx, stocks providers.Quotes) *Quotes {
	return &Quotes{crypto: crypto, fx: fx, stocks: stocks}
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
	switch market {
	case "crypto":
		return s.record(ctx, s.crypto)
	case "fx":
		return s.record(ctx, s.fx)
	case "stocks":
		return s.record(ctx, s.stocks)
	default:
		return nil, fmt.Errorf("mercado inválido: %s", market)
	}
}

func (s *Quotes) record(ctx context.Context, provider providers.Quotes) ([]domain.Quote, error) {
	items, err := provider.Quotes(ctx)
	if err != nil {
		return nil, err
	}
	if s.history != nil {
		saveCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.history.Save(saveCtx, items)
	}
	return items, nil
}
