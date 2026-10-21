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

type Binance struct {
	client  *http.Client
	symbols []string
}

var defaultSymbols = []string{
	"BTCUSDT", "ETHUSDT", "BNBUSDT", "XRPUSDT", "SOLUSDT", "DOGEUSDT", "TRXUSDT",
	"ADAUSDT", "LINKUSDT", "AVAXUSDT", "LTCUSDT", "DOTUSDT", "BCHUSDT", "ETCUSDT",
	"XLMUSDT", "ATOMUSDT", "UNIUSDT", "NEARUSDT", "FILUSDT", "SHIBUSDT",
}

func NewBinance(client *http.Client, configured string) *Binance {
	return &Binance{client: client, symbols: parseSymbols(configured)}
}

func parseSymbols(value string) []string {
	seen := make(map[string]struct{})
	symbols := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		symbol := strings.ToUpper(strings.TrimSpace(part))
		if symbol == "" {
			continue
		}
		if _, exists := seen[symbol]; exists {
			continue
		}
		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)
	}
	if len(symbols) == 0 {
		return append([]string(nil), defaultSymbols...)
	}
	return symbols
}

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
	encodedSymbols, err := json.Marshal(b.symbols)
	if err != nil {
		return nil, fmt.Errorf("codificar lista de símbolos Binance: %w", err)
	}
	params.Set("symbols", string(encodedSymbols))
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
		name := cryptoNames[base]
		if name == "" {
			name = base
		}
		quotes = append(quotes, domain.Quote{Symbol: base, Name: name, Market: "crypto", Currency: "USDT", Price: price, Change: change, ChangePct: pct, Volume: volume, High: high, Low: low, UpdatedAt: time.UnixMilli(row.CloseTime).UTC()})
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

var cryptoNames = map[string]string{
	"BTC": "Bitcoin", "ETH": "Ethereum", "BNB": "BNB", "XRP": "XRP", "SOL": "Solana",
	"DOGE": "Dogecoin", "TRX": "TRON", "ADA": "Cardano", "LINK": "Chainlink", "AVAX": "Avalanche",
	"LTC": "Litecoin", "DOT": "Polkadot", "BCH": "Bitcoin Cash", "ETC": "Ethereum Classic",
	"XLM": "Stellar", "ATOM": "Cosmos", "UNI": "Uniswap", "NEAR": "NEAR Protocol",
	"FIL": "Filecoin", "SHIB": "Shiba Inu",
}
