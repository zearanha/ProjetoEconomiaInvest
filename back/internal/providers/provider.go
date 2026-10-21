package providers

import (
	"context"
	"painel-mercado/internal/domain"
)

type Quotes interface {
	Quotes(context.Context) ([]domain.Quote, error)
}

type Rates interface {
	Rates(context.Context) ([]domain.FixedIncomeRate, error)
}

type News interface {
	Fetch(context.Context) ([]domain.News, error)
}
