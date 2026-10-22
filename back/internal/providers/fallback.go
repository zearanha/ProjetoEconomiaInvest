package providers

import (
	"context"
	"fmt"
	"log/slog"

	"painel-mercado/internal/domain"
)

// Fallback tries a secondary market data source when the primary one fails.
type Fallback struct {
	Primary   Quotes
	Secondary Quotes
	Market    string
}

// RatesFallback tries a secondary indicator source when the primary fails.
type RatesFallback struct {
	Primary   Rates
	Secondary Rates
}

func (p RatesFallback) Rates(ctx context.Context) ([]domain.FixedIncomeRate, error) {
	rates, primaryErr := p.Primary.Rates(ctx)
	if primaryErr == nil {
		return rates, nil
	}

	slog.Warn("fonte primária de indicadores indisponível; tentando alternativa", "error", primaryErr)
	rates, secondaryErr := p.Secondary.Rates(ctx)
	if secondaryErr != nil {
		return nil, fmt.Errorf("fonte primária: %v; fonte alternativa: %w", primaryErr, secondaryErr)
	}
	return rates, nil
}

func (p Fallback) Quotes(ctx context.Context) ([]domain.Quote, error) {
	quotes, primaryErr := p.Primary.Quotes(ctx)
	if primaryErr == nil {
		return quotes, nil
	}

	slog.Warn("fonte primária de cotações indisponível; tentando alternativa", "market", p.Market, "error", primaryErr)
	quotes, secondaryErr := p.Secondary.Quotes(ctx)
	if secondaryErr != nil {
		return nil, fmt.Errorf("fonte primária: %v; fonte alternativa: %w", primaryErr, secondaryErr)
	}
	return quotes, nil
}
