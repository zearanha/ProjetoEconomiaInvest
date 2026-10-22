package fx

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

const frankfurterURL = "https://api.frankfurter.dev/v2/rates"

type Frankfurter struct{ client *http.Client }

func NewFrankfurter(client *http.Client) *Frankfurter { return &Frankfurter{client: client} }

type frankfurterRate struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

func (f *Frankfurter) Quotes(ctx context.Context) ([]domain.Quote, error) {
	query := url.Values{}
	query.Set("base", "BRL")
	query.Set("quotes", strings.Join(currencyCodes(), ","))
	query.Set("from", time.Now().AddDate(0, 0, -14).Format("2006-01-02"))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, frankfurterURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição Frankfurter: %w", err)
	}
	response, err := f.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar Frankfurter: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Frankfurter respondeu HTTP %d", response.StatusCode)
	}
	var rows []frankfurterRate
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decodificar Frankfurter: %w", err)
	}
	type dailyRates struct{ previous, latest frankfurterRate }
	byCurrency := make(map[string]dailyRates)
	for _, row := range rows {
		if row.Rate <= 0 {
			continue
		}
		daily := byCurrency[row.Quote]
		if row.Date > daily.latest.Date {
			daily.previous = daily.latest
			daily.latest = row
			byCurrency[row.Quote] = daily
		}
	}
	quotes := make([]domain.Quote, 0, len(byCurrency))
	for _, currency := range currencies {
		daily := byCurrency[currency.code]
		if daily.latest.Rate <= 0 {
			continue
		}
		updatedAt, _ := time.Parse("2006-01-02", daily.latest.Date)
		price := 1 / daily.latest.Rate // Frankfurter returns foreign currency per BRL.
		change, changePct := 0.0, 0.0
		if daily.previous.Rate > 0 {
			previousPrice := 1 / daily.previous.Rate
			change = price - previousPrice
			changePct = change / previousPrice * 100
		}
		quotes = append(quotes, domain.Quote{
			Symbol: currency.code + "-BRL", Name: currency.name, Market: "fx", Currency: "BRL",
			Price: price, Change: change, ChangePct: changePct, UpdatedAt: updatedAt.UTC(),
		})
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("Frankfurter não retornou taxas")
	}
	return quotes, nil
}
