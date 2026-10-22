package stocks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"painel-mercado/internal/domain"
)

const yahooIbovespaURL = "https://query1.finance.yahoo.com/v8/finance/chart/%5EBVSP?range=5d&interval=1d"

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string  `json:"currency"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
				RegularMarketTime  int64   `json:"regularMarketTime"`
				DayHigh            float64 `json:"regularMarketDayHigh"`
				DayLow             float64 `json:"regularMarketDayLow"`
			} `json:"meta"`
		} `json:"result"`
		Error *struct {
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

func yahooIbovespa(ctx context.Context, client *http.Client) (domain.Quote, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, yahooIbovespaURL, nil)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("criar requisição do Ibovespa: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MercadoPainel/1.0)")
	response, err := client.Do(request)
	if err != nil {
		return domain.Quote{}, fmt.Errorf("consultar Ibovespa: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return domain.Quote{}, fmt.Errorf("Yahoo Finance respondeu HTTP %d", response.StatusCode)
	}
	var body yahooChartResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return domain.Quote{}, fmt.Errorf("decodificar cotação do Ibovespa: %w", err)
	}
	if body.Chart.Error != nil {
		return domain.Quote{}, fmt.Errorf("Yahoo Finance: %s", body.Chart.Error.Description)
	}
	if len(body.Chart.Result) == 0 || body.Chart.Result[0].Meta.RegularMarketPrice <= 0 {
		return domain.Quote{}, fmt.Errorf("Yahoo Finance não retornou cotação do Ibovespa")
	}
	meta := body.Chart.Result[0].Meta
	previous := meta.PreviousClose
	if previous <= 0 {
		previous = meta.ChartPreviousClose
	}
	change, changePct := 0.0, 0.0
	if previous > 0 {
		change = meta.RegularMarketPrice - previous
		changePct = change / previous * 100
	}
	updatedAt := time.Now().UTC()
	if meta.RegularMarketTime > 0 {
		updatedAt = time.Unix(meta.RegularMarketTime, 0).UTC()
	}
	currency := meta.Currency
	if currency == "" {
		currency = "BRL"
	}
	return domain.Quote{
		Symbol: "^BVSP", Name: "Ibovespa", Market: "stocks", Currency: currency,
		Price: meta.RegularMarketPrice, Change: change, ChangePct: changePct,
		High: meta.DayHigh, Low: meta.DayLow, UpdatedAt: updatedAt,
	}, nil
}
