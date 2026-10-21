package crypto

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

const coinGeckoURL = "https://api.coingecko.com/api/v3/coins/markets"

// CoinGecko provides aggregated crypto market data when the exchange API is
// unavailable from the hosting region.
type CoinGecko struct {
	client  *http.Client
	symbols []string
	apiKey  string
}

func NewCoinGecko(client *http.Client, configured, apiKey string) *CoinGecko {
	return &CoinGecko{client: client, symbols: parseSymbols(configured), apiKey: apiKey}
}

type marketCoin struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Price     float64   `json:"current_price"`
	Change    float64   `json:"price_change_24h"`
	ChangePct float64   `json:"price_change_percentage_24h"`
	Volume    float64   `json:"total_volume"`
	High      float64   `json:"high_24h"`
	Low       float64   `json:"low_24h"`
	UpdatedAt time.Time `json:"last_updated"`
}

var coinGeckoIDs = map[string]string{
	"BTC": "bitcoin", "ETH": "ethereum", "BNB": "binancecoin", "XRP": "ripple",
	"SOL": "solana", "DOGE": "dogecoin", "TRX": "tron", "ADA": "cardano",
	"LINK": "chainlink", "AVAX": "avalanche-2", "LTC": "litecoin", "DOT": "polkadot",
	"BCH": "bitcoin-cash", "ETC": "ethereum-classic", "XLM": "stellar", "ATOM": "cosmos",
	"UNI": "uniswap", "NEAR": "near", "FIL": "filecoin", "SHIB": "shiba-inu",
}

func (c *CoinGecko) Quotes(ctx context.Context) ([]domain.Quote, error) {
	ids := make([]string, 0, len(c.symbols))
	for _, symbol := range c.symbols {
		base := cryptoBaseSymbol(symbol)
		id, ok := coinGeckoIDs[base]
		if !ok {
			return nil, fmt.Errorf("CoinGecko não tem mapeamento configurado para %s", symbol)
		}
		ids = append(ids, id)
	}

	query := url.Values{}
	query.Set("vs_currency", "usd")
	query.Set("ids", strings.Join(ids, ","))
	query.Set("price_change_percentage", "24h")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, coinGeckoURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição CoinGecko: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		request.Header.Set("x-cg-demo-api-key", c.apiKey)
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar CoinGecko: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinGecko respondeu HTTP %d", response.StatusCode)
	}

	var rows []marketCoin
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decodificar CoinGecko: %w", err)
	}
	quotes := make([]domain.Quote, 0, len(rows))
	for _, row := range rows {
		quotes = append(quotes, domain.Quote{
			Symbol: strings.ToUpper(row.Symbol), Name: row.Name, Market: "crypto", Currency: "USD",
			Price: row.Price, Change: row.Change, ChangePct: row.ChangePct,
			Volume: row.Volume, High: row.High, Low: row.Low, UpdatedAt: row.UpdatedAt.UTC(),
		})
	}
	return quotes, nil
}

func cryptoBaseSymbol(symbol string) string {
	for _, suffix := range []string{"USDT", "USDC", "BUSD", "USD", "BTC", "ETH"} {
		if strings.HasSuffix(symbol, suffix) {
			return strings.TrimSuffix(symbol, suffix)
		}
	}
	return symbol
}
