package repository

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"painel-mercado/internal/domain"
)

//go:embed migrations/001_create_news.sql
var newsSchema string

type PostgresNews struct{ pool *pgxpool.Pool }

func NewPostgresNews(ctx context.Context, databaseURL string) (*PostgresNews, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("criar pool PostgreSQL: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectar ao PostgreSQL: %w", err)
	}
	if _, err := pool.Exec(ctx, newsSchema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("aplicar schema de notícias: %w", err)
	}
	return &PostgresNews{pool: pool}, nil
}

func (r *PostgresNews) Close() { r.pool.Close() }

func (r *PostgresNews) Upsert(ctx context.Context, items []domain.News) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback(ctx)
	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO news (id, title, source, published_at, summary, url, category)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				source = EXCLUDED.source,
				published_at = EXCLUDED.published_at,
				summary = EXCLUDED.summary,
				url = EXCLUDED.url,
				category = EXCLUDED.category`,
			item.ID, item.Title, item.Source, item.PublishedAt, item.Summary, item.URL, item.Category)
		if err != nil {
			return fmt.Errorf("salvar notícia %q: %w", item.ID, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("confirmar transação: %w", err)
	}
	return nil
}

func (r *PostgresNews) All(ctx context.Context) ([]domain.News, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, source, published_at, summary, url, category
		FROM news ORDER BY published_at DESC LIMIT 2000`)
	if err != nil {
		return nil, fmt.Errorf("consultar notícias: %w", err)
	}
	defer rows.Close()
	items := make([]domain.News, 0)
	for rows.Next() {
		var item domain.News
		if err := rows.Scan(&item.ID, &item.Title, &item.Source, &item.PublishedAt, &item.Summary, &item.URL, &item.Category); err != nil {
			return nil, fmt.Errorf("ler notícia: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar notícias: %w", err)
	}
	return items, nil
}
