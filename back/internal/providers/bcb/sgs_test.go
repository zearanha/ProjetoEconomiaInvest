package bcb

import (
	"context"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestRatesReadsAndAnnualizesSGSSeries(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.URL.Path; !strings.HasSuffix(got, "/dados/ultimos/10") {
			t.Errorf("request path = %q, want a 10-observation window", got)
		}
		var value string
		switch {
		case strings.Contains(request.URL.Path, ".432/"):
			value = "15"
		case strings.Contains(request.URL.Path, ".12/"):
			value = "0.05"
		case strings.Contains(request.URL.Path, ".433/"):
			value = "0.4"
		default:
			t.Errorf("unexpected SGS URL: %s", request.URL)
			value = "0"
		}
		body := `[{"data":"15/10/2026","valor":"` + value + `"}]`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}

	rates, err := NewSGS(client).Rates(context.Background())
	if err != nil {
		t.Fatalf("Rates() error = %v", err)
	}
	if len(rates) != 3 {
		t.Fatalf("got %d rates, want 3", len(rates))
	}
	if rates[0].Indicator != "SELIC" || rates[0].Value != 15 || rates[0].Unit != "% a.a." {
		t.Errorf("unexpected SELIC rate: %+v", rates[0])
	}
	wantCDI := (math.Pow(1.0005, 252) - 1) * 100
	if rates[1].Indicator != "CDI" || math.Abs(rates[1].Value-wantCDI) > 1e-9 {
		t.Errorf("annualized CDI = %v, want %v", rates[1].Value, wantCDI)
	}
	if rates[2].Indicator != "IPCA" || rates[2].Value != 0.4 || rates[2].Unit != "% ao mês" {
		t.Errorf("unexpected IPCA rate: %+v", rates[2])
	}
	wantDate := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	if !rates[0].Date.Equal(wantDate) {
		t.Errorf("date = %s, want %s", rates[0].Date, wantDate)
	}
}
