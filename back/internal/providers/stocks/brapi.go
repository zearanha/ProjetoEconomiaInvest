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

func splitSymbols(value string) []string {
	parts := strings.Split(value, ",")
	symbols := make([]string, 0, len(parts))
	for _, part := range parts {
		if symbol := strings.ToUpper(strings.TrimSpace(part)); symbol != "" {
			symbols = append(symbols, symbol)
		}
	}
	if len(symbols) == 0 {
		return []string{"PETR4", "VALE3", "ITUB4", "MGLU3"}
	}
	return symbols
}
