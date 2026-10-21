package jobs

import (
	"context"
	"log/slog"
	"time"

	"painel-mercado/internal/service"
)

func RunNews(ctx context.Context, interval time.Duration, news *service.News, logger *slog.Logger) {
	refresh := func() {
		jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		count, err := news.Refresh(jobCtx)
		if err != nil {
			logger.Error("falha ao atualizar feed de notícias", "error", err)
			return
		}
		logger.Info("feed de notícias atualizado", "items", count)
	}
	refresh()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refresh()
		}
	}
}
