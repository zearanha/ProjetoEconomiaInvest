package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Hub distributes the latest quote updates to connected browser clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
	latest  map[string][]byte
}

type client struct {
	conn *websocket.Conn
	send chan []byte
}

// QuoteUpdate is the JSON envelope sent to WebSocket subscribers.
type QuoteUpdate struct {
	Type      string    `json:"type"`
	Market    string    `json:"market"`
	Data      any       `json:"data"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*client]struct{}), latest: make(map[string][]byte)}
}

// Publish sends an update to every subscriber. A slow client receives the
// newest update instead of accumulating stale quote snapshots.
func (h *Hub) Publish(update QuoteUpdate) error {
	payload, err := json.Marshal(update)
	if err != nil {
		return err
	}

	h.mu.Lock()
	h.latest[update.Market] = payload
	for subscriber := range h.clients {
		enqueue(subscriber.send, payload)
	}
	h.mu.Unlock()
	return nil
}

func enqueue(queue chan []byte, payload []byte) {
	select {
	case queue <- payload:
	default:
		select {
		case <-queue:
		default:
		}
		select {
		case queue <- payload:
		default:
		}
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx := conn.CloseRead(context.Background())
	subscriber := &client{conn: conn, send: make(chan []byte, 8)}
	h.mu.Lock()
	h.clients[subscriber] = struct{}{}
	for _, payload := range h.latest {
		subscriber.send <- payload
	}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, subscriber)
		h.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case payload := <-subscriber.send:
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Write(writeCtx, websocket.MessageText, payload)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
