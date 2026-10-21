package service

import (
	"context"
	"testing"

	"painel-mercado/internal/domain"
)

type fixedRatesStub struct{ rates []domain.FixedIncomeRate }

func (s fixedRatesStub) Rates(context.Context) ([]domain.FixedIncomeRate, error) { return s.rates, nil }

func TestSimulateComparesNetReturns(t *testing.T) {
	service := NewFixedIncome(fixedRatesStub{rates: []domain.FixedIncomeRate{{Indicator: "CDI", Value: 10}}})
	result, err := service.Simulate(context.Background(), SimulationInput{
		Amount: 10000, TermDays: 365, CDBPercent: 100, LCIPercent: 90, LCAPercent: 90,
	})
	if err != nil {
		t.Fatalf("Simulate() error = %v", err)
	}
	if result.EstimatedBusinessDays != 252 {
		t.Errorf("business days = %d, want 252", result.EstimatedBusinessDays)
	}
	if len(result.Results) != 3 {
		t.Fatalf("got %d products, want 3", len(result.Results))
	}
	cdb, lci := result.Results[0], result.Results[1]
	if cdb.IncomeTaxPercent != 17.5 || cdb.Tax <= 0 || cdb.NetAmount <= 10000 {
		t.Errorf("unexpected CDB result: %+v", cdb)
	}
	if lci.IncomeTaxPercent != 0 || lci.Tax != 0 || lci.NetProfit <= 0 {
		t.Errorf("unexpected LCI result: %+v", lci)
	}
}

func TestCDBTaxBands(t *testing.T) {
	cases := []struct {
		days int
		want float64
	}{{180, 22.5}, {181, 20}, {360, 20}, {361, 17.5}, {720, 17.5}, {721, 15}}
	for _, tc := range cases {
		got := calculate("CDB", 1000, 10, 0.001, 100, tc.days)
		if got.IncomeTaxPercent != tc.want {
			t.Errorf("term %d: tax = %v, want %v", tc.days, got.IncomeTaxPercent, tc.want)
		}
	}
}
