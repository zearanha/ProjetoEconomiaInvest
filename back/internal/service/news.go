package service

import (
	"context"
	"fmt"
	"strings"

	"painel-mercado/internal/domain"
	"painel-mercado/internal/providers"
	"painel-mercado/internal/repository"
)

const newsPageSize = 20

type NewsFilter struct {
	Category string
	Query    string
	Page     int
}

type NewsPage struct {
	Data     []domain.News `json:"data"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
	Total    int           `json:"total"`
}

type News struct {
	provider providers.News
	repo     repository.News
}

func NewNews(provider providers.News, repo repository.News) *News {
	return &News{provider: provider, repo: repo}
}

func (s *News) Refresh(ctx context.Context) (int, error) {
	items, err := s.provider.Fetch(ctx)
	if err != nil {
		return 0, err
	}
	if err := s.repo.Upsert(ctx, items); err != nil {
		return 0, fmt.Errorf("guardar notícias: %w", err)
	}
	return len(items), nil
}

func (s *News) List(ctx context.Context, filter NewsFilter) (NewsPage, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	category := strings.ToLower(strings.TrimSpace(filter.Category))
	if category != "" && category != "brasil" && category != "global" && category != "crypto" && category != "fixed-income" && category != "stocks" {
		return NewsPage{}, fmt.Errorf("categoria inválida: %s", category)
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	items, err := s.repo.All(ctx)
	if err != nil {
		return NewsPage{}, fmt.Errorf("listar notícias: %w", err)
	}
	filtered := make([]domain.News, 0, len(items))
	for _, item := range items {
		if category != "" && item.Category != category {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Title+" "+item.Summary+" "+item.Source), query) {
			continue
		}
		filtered = append(filtered, item)
	}
	start := (filter.Page - 1) * newsPageSize
	if start >= len(filtered) {
		filtered = []domain.News{}
	} else {
		end := start + newsPageSize
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}
	return NewsPage{Data: filtered, Page: filter.Page, PageSize: newsPageSize, Total: countMatches(items, category, query)}, nil
}

func countMatches(items []domain.News, category, query string) int {
	count := 0
	for _, item := range items {
		if category != "" && item.Category != category {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Title+" "+item.Summary+" "+item.Source), query) {
			continue
		}
		count++
	}
	return count
}
