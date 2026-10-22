package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"painel-mercado/internal/domain"
)

const baseURL = "https://economia.awesomeapi.com.br/json/last/"

type AwesomeAPI struct{ client *http.Client }

func NewAwesomeAPI(client *http.Client) *AwesomeAPI { return &AwesomeAPI{client: client} }

type currency struct {
	Code      string `json:"code"`
	CodeIn    string `json:"codein"`
	Name      string `json:"name"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	VarBid    string `json:"varBid"`
	PctChange string `json:"pctChange"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Timestamp string `json:"timestamp"`
}

func (a *AwesomeAPI) Quotes(ctx context.Context) ([]domain.Quote, error) {
	pairs := make([]string, 0, len(currencies))
	for _, currency := range currencies {
		pairs = append(pairs, currency.code+"-BRL")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+strings.Join(pairs, ","), nil)
	if err != nil {
		return nil, fmt.Errorf("criar requisição AwesomeAPI: %w", err)
	}
	response, err := a.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("consultar AwesomeAPI: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AwesomeAPI respondeu HTTP %d", response.StatusCode)
	}
	var rows map[string]currency
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("decodificar AwesomeAPI: %w", err)
	}
	quotes := make([]domain.Quote, 0, len(rows))
	for _, row := range rows {
		price, e1 := parse(row.Bid)
		change, e2 := parse(row.VarBid)
		pct, e3 := parse(row.PctChange)
		high, e4 := parse(row.High)
		low, e5 := parse(row.Low)
		if err := firstError(e1, e2, e3, e4, e5); err != nil {
			return nil, fmt.Errorf("converter par %s-%s: %w", row.Code, row.CodeIn, err)
		}
		unixSeconds, _ := strconv.ParseInt(row.Timestamp, 10, 64)
		stamp := time.Unix(unixSeconds, 0)
		quotes = append(quotes, domain.Quote{Symbol: row.Code + "-" + row.CodeIn, Name: strings.TrimSpace(row.Name), Market: "fx", Currency: row.CodeIn, Price: price, Change: change, ChangePct: pct, High: high, Low: low, UpdatedAt: stamp.UTC()})
	}
	return quotes, nil
}

func parse(value string) (float64, error) { return strconv.ParseFloat(value, 64) }
func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
