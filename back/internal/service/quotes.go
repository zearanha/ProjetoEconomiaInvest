package service

import (
	"context"
	"fmt"

	"painel-mercado/internal/domain"
	"painel-mercado/internal/providers"
)

type Quotes struct{ crypto, fx, stocks providers.Quotes }

func NewQuotes(crypto, fx, stocks providers.Quotes) *Quotes {
	return &Quotes{crypto: crypto, fx: fx, stocks: stocks}
}

func (s *Quotes) Get(ctx context.Context, market string) ([]domain.Quote, error) {
	switch market {
	case "crypto":
		return s.crypto.Quotes(ctx)
	case "fx":
		return s.fx.Quotes(ctx)
	case "stocks":
		return s.stocks.Quotes(ctx)
	default:
		return nil, fmt.Errorf("mercado inválido: %s", market)
	}
}
