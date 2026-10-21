package repository

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"painel-mercado/internal/domain"
)

//go:embed migrations/002_create_quote_history.sql
var quoteHistorySchema string

type PostgresQuoteHistory struct{ pool *pgxpool.Pool }

func NewPostgresQuoteHistory(ctx context.Context, databaseURL string) (*PostgresQuoteHistory, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("criar pool de histórico: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectar ao PostgreSQL para histórico: %w", err)
	}
	if _, err := pool.Exec(ctx, quoteHistorySchema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("aplicar schema de histórico: %w", err)
	}
	return &PostgresQuoteHistory{pool: pool}, nil
}

func (r *PostgresQuoteHistory) Close() { r.pool.Close() }

func (r *PostgresQuoteHistory) Save(ctx context.Context, quotes []domain.Quote) error {
	for _, quote := range quotes {
		_, err := r.pool.Exec(ctx, `INSERT INTO quote_history (symbol, market, sampled_at, price, currency) VALUES ($1,$2,date_trunc('minute',$3),$4,$5) ON CONFLICT (symbol,sampled_at) DO UPDATE SET price=EXCLUDED.price, market=EXCLUDED.market, currency=EXCLUDED.currency`, quote.Symbol, quote.Market, quote.UpdatedAt, quote.Price, quote.Currency)
		if err != nil {
			return fmt.Errorf("salvar histórico de %s: %w", quote.Symbol, err)
		}
	}
	return nil
}

func (r *PostgresQuoteHistory) List(ctx context.Context, symbol string, since time.Time) ([]domain.QuotePoint, error) {
	rows, err := r.pool.Query(ctx, `SELECT sampled_at, price FROM quote_history WHERE symbol=$1 AND sampled_at >= $2 ORDER BY sampled_at`, symbol, since)
	if err != nil {
		return nil, fmt.Errorf("consultar histórico: %w", err)
	}
	defer rows.Close()
	points := make([]domain.QuotePoint, 0)
	for rows.Next() {
		var point domain.QuotePoint
		if err := rows.Scan(&point.Time, &point.Price); err != nil {
			return nil, fmt.Errorf("ler histórico: %w", err)
		}
		points = append(points, point)
	}
	return points, rows.Err()
}
