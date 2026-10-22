package brasilapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"painel-mercado/internal/domain"
)

const ratesURL = "https://brasilapi.com.br/api/taxas/v1"

type Rates struct{ client *http.Client }

func NewRates(client *http.Client) *Rates { return &Rates{client: client} }

type tax struct {
	Name  string  `json:"nome"`
	Value float64 `json:"valor"`
}

func (p *Rates) Rates(ctx context.Context) ([]domain.FixedIncomeRate, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, ratesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição BrasilAPI: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar BrasilAPI: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("BrasilAPI respondeu HTTP %d", response.StatusCode)
	}
	var rows []tax
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decodificar taxas BrasilAPI: %w", err)
	}
	values := make(map[string]float64, len(rows))
	for _, row := range rows {
		values[strings.ToLower(strings.TrimSpace(row.Name))] = row.Value
	}
	definitions := []struct {
		name, indicator, unit string
	}{
		{"selic", "SELIC", "% a.a."},
		{"cdi", "CDI", "% a.a."},
		{"ipca", "IPCA", "% em 12 meses"},
	}
	now := time.Now().UTC()
	rates := make([]domain.FixedIncomeRate, 0, len(definitions))
	for _, item := range definitions {
		value, ok := values[item.name]
		if !ok {
			return nil, fmt.Errorf("BrasilAPI não retornou %s", item.indicator)
		}
		rates = append(rates, domain.FixedIncomeRate{
			Indicator: item.indicator, Value: value, Unit: item.unit,
			Date: now, Source: "BrasilAPI (data da consulta)",
		})
	}
	return rates, nil
}
