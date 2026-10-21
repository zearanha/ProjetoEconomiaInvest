package repository

import (
	"context"
	"sync"
	"time"

	"painel-mercado/internal/domain"
)

// QuoteHistory stores periodic quote samples for charting.
type QuoteHistory interface {
	Save(context.Context, []domain.Quote) error
	List(context.Context, string, time.Time) ([]domain.QuotePoint, error)
}

type MemoryQuoteHistory struct {
	mu    sync.RWMutex
	items map[string][]domain.QuotePoint
}

func NewMemoryQuoteHistory() *MemoryQuoteHistory {
	return &MemoryQuoteHistory{items: make(map[string][]domain.QuotePoint)}
}

func (r *MemoryQuoteHistory) Save(_ context.Context, quotes []domain.Quote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, quote := range quotes {
		points := r.items[quote.Symbol]
		if len(points) > 0 && points[len(points)-1].Time.Truncate(time.Minute).Equal(quote.UpdatedAt.Truncate(time.Minute)) {
			continue
		}
		r.items[quote.Symbol] = append(points, domain.QuotePoint{Time: quote.UpdatedAt, Price: quote.Price})
	}
	return nil
}

func (r *MemoryQuoteHistory) List(_ context.Context, symbol string, since time.Time) ([]domain.QuotePoint, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	points := r.items[symbol]
	result := make([]domain.QuotePoint, 0, len(points))
	for _, point := range points {
		if !point.Time.Before(since) {
			result = append(result, point)
		}
	}
	return result, nil
}
