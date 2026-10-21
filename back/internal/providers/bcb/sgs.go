package bcb

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"painel-mercado/internal/domain"
)

const baseURL = "https://api.bcb.gov.br/dados/serie/bcdata.sgs."
const recentObservations = 10

type SGS struct{ client *http.Client }

func NewSGS(client *http.Client) *SGS { return &SGS{client: client} }

type observation struct {
	Date  string `json:"data"`
	Value string `json:"valor"`
}

func (s *SGS) Rates(ctx context.Context) ([]domain.FixedIncomeRate, error) {
	series := []struct {
		code, name, unit string
		annualizeCDI     bool
	}{
		{"432", "SELIC", "% a.a.", false},
		{"12", "CDI", "% a.a. (anualizado)", true},
		{"433", "IPCA", "% ao mês", false},
	}
	rates := make([]domain.FixedIncomeRate, 0, len(series))
	for _, item := range series {
		rate, err := s.getLatest(ctx, item.code)
		if err != nil {
			return nil, fmt.Errorf("buscar %s no SGS: %w", item.name, err)
		}
		if item.annualizeCDI {
			daily := 1 + rate.Value/100
			rate.Value = (math.Pow(daily, 252) - 1) * 100
		}
		rate.Indicator, rate.Unit, rate.Source = item.name, item.unit, "Banco Central do Brasil (SGS)"
		rates = append(rates, rate)
	}
	return rates, nil
}

func (s *SGS) getLatest(ctx context.Context, code string) (domain.FixedIncomeRate, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s%s/dados/ultimos/%d?formato=json", baseURL, code, recentObservations), nil)
	if err != nil {
		return domain.FixedIncomeRate{}, fmt.Errorf("criar requisição: %w", err)
	}
	response, err := s.client.Do(request)
	if err != nil {
		return domain.FixedIncomeRate{}, fmt.Errorf("consultar API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return domain.FixedIncomeRate{}, fmt.Errorf("API respondeu HTTP %d", response.StatusCode)
	}
	var rows []observation
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return domain.FixedIncomeRate{}, fmt.Errorf("decodificar resposta: %w", err)
	}
	if len(rows) == 0 {
		return domain.FixedIncomeRate{}, fmt.Errorf("série %s sem observações", code)
	}
	value, err := strconv.ParseFloat(rows[len(rows)-1].Value, 64)
	if err != nil {
		return domain.FixedIncomeRate{}, fmt.Errorf("converter valor %q: %w", rows[len(rows)-1].Value, err)
	}
	date, err := time.Parse("02/01/2006", rows[len(rows)-1].Date)
	if err != nil {
		return domain.FixedIncomeRate{}, fmt.Errorf("converter data %q: %w", rows[len(rows)-1].Date, err)
	}
	return domain.FixedIncomeRate{Value: value, Date: date}, nil
}
