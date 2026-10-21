package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"painel-mercado/internal/domain"
)

const baseURL = "https://api.binance.com/api/v3/ticker/24hr"

type Binance struct{ client *http.Client }

func NewBinance(client *http.Client) *Binance { return &Binance{client: client} }

type ticker struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	Volume             string `json:"volume"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	CloseTime          int64  `json:"closeTime"`
}

func (b *Binance) Quotes(ctx context.Context) ([]domain.Quote, error) {
	params := url.Values{}
	params.Set("symbols", `["BTCUSDT","ETHUSDT","SOLUSDT"]`)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição Binance: %w", err)
	}
	response, err := b.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar Binance: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Binance respondeu HTTP %d", response.StatusCode)
	}
	var rows []ticker
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decodificar Binance: %w", err)
	}
	quotes := make([]domain.Quote, 0, len(rows))
	for _, row := range rows {
		price, e1 := number(row.LastPrice)
		change, e2 := number(row.PriceChange)
		pct, e3 := number(row.PriceChangePercent)
		volume, e4 := number(row.Volume)
		high, e5 := number(row.HighPrice)
		low, e6 := number(row.LowPrice)
		if err := firstError(e1, e2, e3, e4, e5, e6); err != nil {
			return nil, fmt.Errorf("converter ticker %s: %w", row.Symbol, err)
		}
		base := strings.TrimSuffix(row.Symbol, "USDT")
		quotes = append(quotes, domain.Quote{Symbol: base, Name: base, Market: "crypto", Currency: "USDT", Price: price, Change: change, ChangePct: pct, Volume: volume, High: high, Low: low, UpdatedAt: time.UnixMilli(row.CloseTime).UTC()})
	}
	return quotes, nil
}

func number(value string) (float64, error) { return strconv.ParseFloat(value, 64) }
func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
