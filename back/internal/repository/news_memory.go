package repository

import (
	"context"
	"sort"
	"sync"

	"painel-mercado/internal/domain"
)

type News interface {
	Upsert(context.Context, []domain.News) error
	All(context.Context) ([]domain.News, error)
}

type MemoryNews struct {
	mu    sync.RWMutex
	items map[string]domain.News
}

func NewMemoryNews() *MemoryNews { return &MemoryNews{items: make(map[string]domain.News)} }

func (r *MemoryNews) Upsert(_ context.Context, items []domain.News) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range items {
		if item.ID != "" {
			r.items[item.ID] = item
		}
	}
	if len(r.items) > 2000 {
		all := make([]domain.News, 0, len(r.items))
		for _, item := range r.items {
			all = append(all, item)
		}
		sort.Slice(all, func(i, j int) bool { return all[i].PublishedAt.After(all[j].PublishedAt) })
		r.items = make(map[string]domain.News, 2000)
		for _, item := range all[:2000] {
			r.items[item.ID] = item
		}
	}
	return nil
}

func (r *MemoryNews) All(_ context.Context) ([]domain.News, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domain.News, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].PublishedAt.After(items[j].PublishedAt) })
	return items, nil
}
