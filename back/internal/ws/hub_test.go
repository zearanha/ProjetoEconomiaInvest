package ws

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func TestHandlerSendsLatestQuoteSnapshot(t *testing.T) {
	hub := NewHub()
	if err := hub.Publish(QuoteUpdate{Type: "quotes", Market: "crypto", Data: []string{"BTC"}, UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("publish quote update: %v", err)
	}

	server := httptest.NewServer(hub)
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial WebSocket: %v", err)
	}
	defer conn.CloseNow()

	var got QuoteUpdate
	if err := wsjson.Read(ctx, conn, &got); err != nil {
		t.Fatalf("read quote update: %v", err)
	}
	if got.Type != "quotes" || got.Market != "crypto" {
		t.Fatalf("unexpected update: %#v", got)
	}
}
