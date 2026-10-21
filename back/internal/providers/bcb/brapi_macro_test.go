package bcb

import (
	"context"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"
)

func TestBrapiMacroRates(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got := request.URL.Query().Get("symbols"); got != "selic,cdi,ipca" {
			t.Errorf("symbols = %q, want selic,cdi,ipca", got)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", got)
		}
		body := `{"results":[{"series":{"slug":"selic"},"latest":{"date":"2026-10-15","value":15}},{"series":{"slug":"cdi"},"latest":{"date":"2026-10-15","value":0.05}},{"series":{"slug":"ipca"},"latest":{"date":"2026-10-01","value":0.4}}]}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	rates, err := NewBrapiMacro(client, "test-token").Rates(context.Background())
	if err != nil {
		t.Fatalf("Rates() error = %v", err)
	}
	if len(rates) != 3 {
		t.Fatalf("got %d rates, want 3", len(rates))
	}
	wantCDI := (math.Pow(1.0005, 252) - 1) * 100
	if rates[0].Indicator != "SELIC" || rates[0].Value != 15 || rates[0].Unit != "% a.a." {
		t.Errorf("unexpected SELIC rate: %+v", rates[0])
	}
	if rates[1].Indicator != "CDI" || math.Abs(rates[1].Value-wantCDI) > 1e-9 {
		t.Errorf("unexpected CDI rate: %+v", rates[1])
	}
	if rates[2].Indicator != "IPCA" || rates[2].Value != 0.4 || rates[2].Unit != "% ao mês" {
		t.Errorf("unexpected IPCA rate: %+v", rates[2])
	}
}
