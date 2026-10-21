package rss

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"painel-mercado/internal/domain"
)

type Feed struct {
	Name     string
	URL      string
	Category string
}

type Provider struct {
	client *http.Client
	feeds  []Feed
}

func NewProvider(client *http.Client) *Provider {
	return &Provider{client: client, feeds: []Feed{
		{Name: "Agência Brasil", URL: "https://agenciabrasil.ebc.com.br/rss/economia/feed.xml", Category: "brasil"},
		{Name: "Le Monde", URL: "https://www.lemonde.fr/en/economy/rss_full.xml", Category: "global"},
	}}
}

type rssDocument struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

var tags = regexp.MustCompile(`<[^>]*>`)

func (p *Provider) Fetch(ctx context.Context) ([]domain.News, error) {
	all := make([]domain.News, 0)
	var feedErrors []error
	for _, feed := range p.feeds {
		items, err := p.fetchFeed(ctx, feed)
		if err != nil {
			feedErrors = append(feedErrors, fmt.Errorf("feed %s: %w", feed.Name, err))
			continue
		}
		all = append(all, items...)
	}
	if len(all) == 0 && len(feedErrors) > 0 {
		return nil, errors.Join(feedErrors...)
	}
	return all, nil
}

func (p *Provider) fetchFeed(ctx context.Context, feed Feed) ([]domain.News, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição: %w", err)
	}
	request.Header.Set("User-Agent", "PainelDeMercado/1.0 (+RSS news reader)")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar RSS: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS respondeu HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("ler RSS: %w", err)
	}
	var document rssDocument
	if err := xml.Unmarshal(body, &document); err != nil {
		return nil, fmt.Errorf("interpretar RSS: %w", err)
	}
	items := make([]domain.News, 0, len(document.Channel.Items))
	for _, item := range document.Channel.Items {
		link := strings.TrimSpace(item.Link)
		if link == "" {
			link = strings.TrimSpace(item.GUID)
		}
		if link == "" || strings.TrimSpace(item.Title) == "" {
			continue
		}
		publishedAt, _ := parseDate(item.PubDate)
		items = append(items, domain.News{
			ID: link, Title: strings.TrimSpace(item.Title), Source: feed.Name,
			PublishedAt: publishedAt, Summary: shortSummary(item.Description, 280),
			URL: link, Category: classify(feed.Category, item.Title+" "+item.Description),
		})
	}
	return items, nil
}

func parseDate(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RFC3339} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("data RSS inválida: %q", value)
}

func shortSummary(value string, maxRunes int) string {
	plain := strings.Join(strings.Fields(html.UnescapeString(tags.ReplaceAllString(value, " "))), " ")
	runes := []rune(plain)
	if len(runes) <= maxRunes {
		return plain
	}
	return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
}

func classify(defaultCategory, content string) string {
	value := strings.ToLower(content)
	containsAny := func(words ...string) bool {
		for _, word := range words {
			if strings.Contains(value, word) {
				return true
			}
		}
		return false
	}
	switch {
	case containsAny("cripto", "crypto", "bitcoin", "ethereum", "criptomoeda"):
		return "crypto"
	case containsAny("renda fixa", "cdb", "lci", "lca", "selic", "cdi", "ipca", "bonds", "fixed income"):
		return "fixed-income"
	case containsAny("ibovespa", "ações", "bolsa", "stocks", "shares", "nasdaq", "s&p 500", "b3"):
		return "stocks"
	default:
		return defaultCategory
	}
}
