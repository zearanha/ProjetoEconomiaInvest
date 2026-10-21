package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"painel-mercado/internal/domain"
)

const coinLoreAssetsURL = "https://api.coinlore.net/api/assets/"
const coinLoreTickerURL = "https://api.coinlore.net/api/ticker/"

// CoinLore supplies the configured crypto quotes using its public asset list
// and batched ticker endpoint.
type CoinLore struct {
	client  *http.Client
	symbols []string
	mu      sync.Mutex
	ids     map[string]string
}

func NewCoinLore(client *http.Client, configured string) *CoinLore {
	return &CoinLore{client: client, symbols: parseSymbols(configured)}
}

type coinLoreAsset struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Rank   int    `json:"rank"`
}

type coinLoreTicker struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	PriceUSD  string  `json:"price_usd"`
	ChangePct string  `json:"percent_change_24h"`
	Volume24  float64 `json:"volume24"`
}

type coinLoreAssetResponse struct {
	Data []coinLoreAsset `json:"data"`
}

func (c *CoinLore) Quotes(ctx context.Context) ([]domain.Quote, error) {
	c.mu.Lock()
	if c.ids == nil {
		ids, err := c.loadIDs(ctx)
		if err != nil {
			c.mu.Unlock()
			return nil, err
		}
		c.ids = ids
	}
	idsBySymbol := c.ids
	c.mu.Unlock()

	ids := make([]string, 0, len(c.symbols))
	for _, symbol := range c.symbols {
		id, ok := idsBySymbol[cryptoBaseSymbol(symbol)]
		if !ok {
			return nil, fmt.Errorf("CoinLore não encontrou o ativo %s", symbol)
		}
		ids = append(ids, id)
	}

	endpoint, _ := url.Parse(coinLoreTickerURL)
	query := endpoint.Query()
	query.Set("id", strings.Join(ids, ","))
	endpoint.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição CoinLore: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar CoinLore: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinLore respondeu HTTP %d", response.StatusCode)
	}
	var rows []coinLoreTicker
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decodificar cotações CoinLore: %w", err)
	}

	bySymbol := make(map[string]coinLoreTicker, len(rows))
	for _, row := range rows {
		bySymbol[strings.ToUpper(row.Symbol)] = row
	}
	quotes := make([]domain.Quote, 0, len(c.symbols))
	for _, configuredSymbol := range c.symbols {
		base := cryptoBaseSymbol(configuredSymbol)
		row, ok := bySymbol[base]
		if !ok {
			return nil, fmt.Errorf("CoinLore não retornou cotação para %s", configuredSymbol)
		}
		price, err := strconv.ParseFloat(row.PriceUSD, 64)
		if err != nil {
			return nil, fmt.Errorf("preço inválido da CoinLore para %s", configuredSymbol)
		}
		changePct, _ := strconv.ParseFloat(row.ChangePct, 64)
		change := 0.0
		if changePct > -100 {
			change = price * changePct / (100 + changePct)
		}
		quotes = append(quotes, domain.Quote{
			Symbol: strings.ToUpper(configuredSymbol), Name: row.Name, Market: "crypto", Currency: "USD",
			Price: price, Change: change, ChangePct: changePct,
			Volume: row.Volume24, UpdatedAt: time.Now().UTC(),
		})
	}
	return quotes, nil
}

func (c *CoinLore) loadIDs(ctx context.Context) (map[string]string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, coinLoreAssetsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição da lista CoinLore: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar ativos CoinLore: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CoinLore respondeu HTTP %d ao listar ativos", response.StatusCode)
	}
	var result coinLoreAssetResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decodificar ativos CoinLore: %w", err)
	}
	ids := make(map[string]string, len(result.Data))
	ranks := make(map[string]int, len(result.Data))
	for _, asset := range result.Data {
		symbol := strings.ToUpper(asset.Symbol)
		rank := asset.Rank
		if symbol == "" || asset.ID == "" {
			continue
		}
		if previous, exists := ranks[symbol]; !exists || (rank > 0 && rank < previous) || previous == 0 {
			ids[symbol] = asset.ID
			ranks[symbol] = rank
		}
	}
	return ids, nil
}
