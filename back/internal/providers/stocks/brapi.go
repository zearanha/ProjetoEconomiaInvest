package stocks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"painel-mercado/internal/domain"
)

const baseURL = "https://brapi.dev/api/v2/stocks/quote"
const publicListURL = "https://brapi.dev/api/quote/list"
const publicListLimit = "50"

type Brapi struct {
	client  *http.Client
	token   string
	symbols []string
}

func NewBrapi(client *http.Client, token, symbols string) *Brapi {
	return &Brapi{client: client, token: token, symbols: splitSymbols(symbols)}
}

type response struct {
	Results []result `json:"results"`
}

type result struct {
	Symbol          string `json:"symbol"`
	RequestedSymbol string `json:"requestedSymbol"`
	Data            struct {
		ShortName              string  `json:"shortName"`
		Currency               string  `json:"currency"`
		RegularMarketPrice     float64 `json:"regularMarketPrice"`
		RegularMarketChange    float64 `json:"regularMarketChange"`
		RegularMarketChangePct float64 `json:"regularMarketChangePercent"`
		RegularMarketVolume    float64 `json:"regularMarketVolume"`
		RegularMarketDayHigh   float64 `json:"regularMarketDayHigh"`
		RegularMarketDayLow    float64 `json:"regularMarketDayLow"`
		RegularMarketTime      string  `json:"regularMarketTime"`
	} `json:"data"`
}

func (b *Brapi) Quotes(ctx context.Context) ([]domain.Quote, error) {
	if b.token == "" || len(b.symbols) == 0 {
		return b.publicQuotes(ctx)
	}

	query := url.Values{}
	query.Set("symbols", strings.Join(b.symbols, ","))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição brapi: %w", err)
	}
	if b.token != "" {
		request.Header.Set("Authorization", "Bearer "+b.token)
	}
	resp, err := b.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar brapi: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brapi respondeu HTTP %d", resp.StatusCode)
	}
	var body response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decodificar resposta brapi: %w", err)
	}
	quotes := make([]domain.Quote, 0, len(body.Results))
	for _, row := range body.Results {
		symbol := row.Symbol
		if symbol == "" {
			symbol = row.RequestedSymbol
		}
		updatedAt, _ := time.Parse(time.RFC3339, row.Data.RegularMarketTime)
		quotes = append(quotes, domain.Quote{
			Symbol: symbol, Name: row.Data.ShortName, Market: "stocks",
			Currency: row.Data.Currency, Price: row.Data.RegularMarketPrice,
			Change: row.Data.RegularMarketChange, ChangePct: row.Data.RegularMarketChangePct,
			Volume: row.Data.RegularMarketVolume, High: row.Data.RegularMarketDayHigh,
			Low: row.Data.RegularMarketDayLow, UpdatedAt: updatedAt.UTC(),
		})
	}
	return quotes, nil
}

type publicQuoteListResponse struct {
	Stocks      []publicStock `json:"stocks"`
	RequestedAt string        `json:"requestedAt"`
}

type publicStock struct {
	Symbol string  `json:"stock"`
	Name   string  `json:"name"`
	Price  float64 `json:"close"`
	Change float64 `json:"change"`
	Volume float64 `json:"volume"`
}

func (b *Brapi) publicQuotes(ctx context.Context) ([]domain.Quote, error) {
	query := url.Values{}
	query.Set("limit", publicListLimit)
	query.Set("sortBy", "volume")
	query.Set("sortOrder", "desc")
	query.Set("subType", "stock")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, publicListURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisiÃ§Ã£o pÃºblica brapi: %w", err)
	}
	response, err := b.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar lista pÃºblica brapi: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lista pÃºblica brapi respondeu HTTP %d", response.StatusCode)
	}
	var body publicQuoteListResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decodificar lista pÃºblica brapi: %w", err)
	}
	updatedAt, _ := time.Parse(time.RFC3339Nano, body.RequestedAt)
	quotes := make([]domain.Quote, 0, len(body.Stocks))
	for _, stock := range body.Stocks {
		if len(b.symbols) > 0 && !containsSymbol(b.symbols, stock.Symbol) {
			continue
		}
		previousClose := stock.Price - stock.Change
		changePct := 0.0
		if previousClose != 0 {
			changePct = stock.Change / previousClose * 100
		}
		quotes = append(quotes, domain.Quote{
			Symbol: stock.Symbol, Name: stock.Name, Market: "stocks", Currency: "BRL",
			Price: stock.Price, Change: stock.Change, ChangePct: changePct,
			Volume: stock.Volume, UpdatedAt: updatedAt.UTC(),
		})
	}
	return quotes, nil
}

func splitSymbols(value string) []string {
	parts := strings.Split(value, ",")
	symbols := make([]string, 0, len(parts))
	seen := make(map[string]struct{})
	for _, part := range parts {
		if symbol := strings.ToUpper(strings.TrimSpace(part)); symbol != "" {
			if _, exists := seen[symbol]; exists {
				continue
			}
			seen[symbol] = struct{}{}
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

func containsSymbol(symbols []string, symbol string) bool {
	for _, candidate := range symbols {
		if candidate == symbol {
			return true
		}
	}
	return false
}
