package jobs

import (
	"context"
	"log/slog"
	"time"

	"painel-mercado/internal/service"
	"painel-mercado/internal/ws"
)

var quoteMarkets = []string{"crypto", "fx", "stocks"}

// RunQuoteUpdates polls each configured quote provider and publishes snapshots
// to active WebSocket clients. Provider calls are bounded by a timeout.
func RunQuoteUpdates(ctx context.Context, interval time.Duration, quotes *service.Quotes, hub *ws.Hub, logger *slog.Logger) {
	refresh := func() {
		for _, market := range quoteMarkets {
			if ctx.Err() != nil {
				return
			}
			requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			items, err := quotes.Get(requestCtx, market)
			cancel()
			if err != nil {
				logger.Warn("falha ao atualizar cotações para WebSocket", "market", market, "error", err)
				continue
			}
			if err := hub.Publish(ws.QuoteUpdate{Type: "quotes", Market: market, Data: items, UpdatedAt: time.Now().UTC()}); err != nil {
				logger.Error("falha ao publicar atualização WebSocket", "market", market, "error", err)
			}
		}
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
