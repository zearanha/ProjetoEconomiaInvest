package rss

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

func TestFetchParsesAndShortensFeedItem(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.Header.Get("User-Agent"); got == "" {
			t.Error("RSS request is missing a User-Agent")
		}
		body := `<?xml version="1.0"?><rss version="2.0"><channel><item><title>Bitcoin sobe</title><link>https://news.example/bitcoin</link><description>&lt;p&gt;Resumo &lt;b&gt;curto&lt;/b&gt; do mercado.&lt;/p&gt;</description><pubDate>Sun, 18 Oct 2026 12:00:00 +0000</pubDate></item></channel></rss>`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	provider := &Provider{client: client, feeds: []Feed{{Name: "Fonte", URL: "https://news.example/feed.xml", Category: "global"}}}
	items, err := provider.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	item := items[0]
	if item.Title != "Bitcoin sobe" || item.Source != "Fonte" || item.Category != "crypto" || item.URL != "https://news.example/bitcoin" {
		t.Errorf("unexpected item: %+v", item)
	}
	if item.Summary != "Resumo curto do mercado." {
		t.Errorf("summary = %q", item.Summary)
	}
	want := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
	if !item.PublishedAt.Equal(want) {
		t.Errorf("PublishedAt = %s, want %s", item.PublishedAt, want)
	}
}
