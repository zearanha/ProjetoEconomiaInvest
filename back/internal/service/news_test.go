package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"painel-mercado/internal/domain"
	"painel-mercado/internal/repository"
)

type newsProviderStub struct{ items []domain.News }

func (p newsProviderStub) Fetch(context.Context) ([]domain.News, error) { return p.items, nil }

func TestNewsRefreshDeduplicatesAndFilters(t *testing.T) {
	item := domain.News{ID: "https://news.example/a", Title: "Market update", Source: "Test", Category: "global", PublishedAt: time.Now()}
	items := []domain.News{item, item}
	service := NewNews(newsProviderStub{items: items}, repository.NewMemoryNews())
	if count, err := service.Refresh(context.Background()); err != nil || count != 2 {
		t.Fatalf("Refresh() = %d, %v; want 2, nil", count, err)
	}
	page, err := service.List(context.Background(), NewsFilter{Category: "GLOBAL", Query: "market", Page: 1})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if page.Total != 1 || len(page.Data) != 1 || page.Data[0].ID != item.ID {
		t.Errorf("unexpected filtered page: %+v", page)
	}
}

func TestNewsPagination(t *testing.T) {
	items := make([]domain.News, 25)
	for i := range items {
		items[i] = domain.News{ID: fmt.Sprintf("%d", i), Title: "News", Category: "brasil", PublishedAt: time.Now().Add(time.Duration(i) * time.Minute)}
	}
	service := NewNews(newsProviderStub{items: items}, repository.NewMemoryNews())
	if _, err := service.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	page, err := service.List(context.Background(), NewsFilter{Page: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(page.Data) != 5 || page.Total != 25 {
		t.Errorf("page 2 has %d items, total %d; want 5, 25", len(page.Data), page.Total)
	}
}
