package bcb

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"painel-mercado/internal/domain"
)

const brapiMacroURL = "https://brapi.dev/api/v2/macro/latest"

type BrapiMacro struct {
	client *http.Client
	token  string
}

func NewBrapiMacro(client *http.Client, token string) *BrapiMacro {
	return &BrapiMacro{client: client, token: token}
}

type macroResponse struct {
	Results []struct {
		Series struct {
			Slug string `json:"slug"`
			Name string `json:"name"`
			Unit string `json:"unit"`
		} `json:"series"`
		Latest struct {
			Date  string  `json:"date"`
			Value float64 `json:"value"`
		} `json:"latest"`
	} `json:"results"`
}

func (b *BrapiMacro) Rates(ctx context.Context) ([]domain.FixedIncomeRate, error) {
	query := url.Values{}
	query.Set("symbols", "selic,cdi,ipca")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, brapiMacroURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição brapi macro: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+b.token)
	response, err := b.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar brapi macro: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("brapi macro respondeu HTTP %d", response.StatusCode)
	}
	var body macroResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decodificar resposta brapi macro: %w", err)
	}
	rates := make([]domain.FixedIncomeRate, 0, len(body.Results))
	seen := make(map[string]bool, len(body.Results))
	for _, item := range body.Results {
		name := map[string]string{"selic": "SELIC", "cdi": "CDI", "ipca": "IPCA"}[item.Series.Slug]
		if name == "" {
			continue
		}
		date, err := time.Parse("2006-01-02", item.Latest.Date)
		if err != nil {
			return nil, fmt.Errorf("converter data da série %s: %w", item.Series.Slug, err)
		}
		value, unit := item.Latest.Value, "% a.a."
		if item.Series.Slug == "cdi" {
			value = (math.Pow(1+value/100, 252) - 1) * 100
			unit = "% a.a. (anualizado)"
		} else if item.Series.Slug == "ipca" {
			unit = "% ao mês"
		}
		rates = append(rates, domain.FixedIncomeRate{Indicator: name, Value: value, Unit: unit, Date: date, Source: "brapi.dev"})
		seen[item.Series.Slug] = true
	}
	for _, slug := range []string{"selic", "cdi", "ipca"} {
		if !seen[slug] {
			return nil, fmt.Errorf("brapi macro não retornou a série %s", strings.ToUpper(slug))
		}
	}
	return rates, nil
}
