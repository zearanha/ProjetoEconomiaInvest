package stocks

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestQuotesMapsBrapiResponse(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.URL.Query().Get("symbols"); got != "PETR4,VALE3" {
			t.Errorf("symbols = %q, want PETR4,VALE3", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", got)
		}
		body := `{"results":[{"symbol":"PETR4","data":{"shortName":"PETROBRAS PN","currency":"BRL","regularMarketPrice":38.5,"regularMarketChange":0.3,"regularMarketChangePercent":0.78,"regularMarketVolume":45678901,"regularMarketTime":"2026-06-14T17:08:02.000Z"}}]}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	provider := NewBrapi(client, "test-token", " petr4, vale3 ")
	quotes, err := provider.Quotes(context.Background())
	if err != nil {
		t.Fatalf("Quotes() error = %v", err)
	}
	if len(quotes) != 1 {
		t.Fatalf("got %d quotes, want 1", len(quotes))
	}
	quote := quotes[0]
	if quote.Symbol != "PETR4" || quote.Name != "PETROBRAS PN" || quote.Market != "stocks" || quote.Currency != "BRL" {
		t.Errorf("unexpected quote identity: %+v", quote)
	}
	if quote.Price != 38.5 || quote.Change != 0.3 || quote.ChangePct != 0.78 || quote.Volume != 45678901 {
		t.Errorf("unexpected quote values: %+v", quote)
	}
	wantTime := time.Date(2026, 6, 14, 17, 8, 2, 0, time.UTC)
	if !quote.UpdatedAt.Equal(wantTime) {
		t.Errorf("UpdatedAt = %s, want %s", quote.UpdatedAt, wantTime)
	}
}
