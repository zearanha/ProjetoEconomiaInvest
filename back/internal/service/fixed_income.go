package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"painel-mercado/internal/domain"
	"painel-mercado/internal/providers"
)

var ErrInvalidSimulation = errors.New("parâmetros de simulação inválidos")

type SimulationInput struct {
	Amount     float64 `json:"amount"`
	TermDays   int     `json:"termDays"`
	CDBPercent float64 `json:"cdbPercentCDI"`
	LCIPercent float64 `json:"lciPercentCDI"`
	LCAPercent float64 `json:"lcaPercentCDI"`
}

type InvestmentResult struct {
	Product          string  `json:"product"`
	GrossProfit      float64 `json:"grossProfit"`
	Tax              float64 `json:"tax"`
	NetProfit        float64 `json:"netProfit"`
	NetAmount        float64 `json:"netAmount"`
	IncomeTaxPercent float64 `json:"incomeTaxPercent"`
}

type Simulation struct {
	CDIAnnualPercent      float64            `json:"cdiAnnualPercent"`
	EstimatedBusinessDays int                `json:"estimatedBusinessDays"`
	Results               []InvestmentResult `json:"results"`
	Assumption            string             `json:"assumption"`
}

type FixedIncome struct{ rates providers.Rates }

func NewFixedIncome(rates providers.Rates) *FixedIncome { return &FixedIncome{rates: rates} }

func (s *FixedIncome) Rates(ctx context.Context) ([]domain.FixedIncomeRate, error) {
	return s.rates.Rates(ctx)
}

func (s *FixedIncome) Simulate(ctx context.Context, input SimulationInput) (Simulation, error) {
	if input.Amount <= 0 || input.Amount > 1_000_000_000 || input.TermDays < 1 || input.TermDays > 36500 ||
		input.CDBPercent <= 0 || input.CDBPercent > 500 || input.LCIPercent <= 0 || input.LCIPercent > 500 ||
		input.LCAPercent <= 0 || input.LCAPercent > 500 {
		return Simulation{}, ErrInvalidSimulation
	}
	rates, err := s.rates.Rates(ctx)
	if err != nil {
		return Simulation{}, fmt.Errorf("obter CDI para simulação: %w", err)
	}
	var cdiAnnual float64
	for _, rate := range rates {
		if rate.Indicator == "CDI" {
			cdiAnnual = rate.Value
			break
		}
	}
	if cdiAnnual <= 0 {
		return Simulation{}, errors.New("taxa CDI indisponível para simulação")
	}
	businessDays := int(math.Round(float64(input.TermDays) * 252 / 365))
	dailyCDI := math.Pow(1+cdiAnnual/100, 1.0/252) - 1
	results := []InvestmentResult{
		calculate("CDB", input.Amount, businessDays, dailyCDI, input.CDBPercent, input.TermDays),
		calculate("LCI", input.Amount, businessDays, dailyCDI, input.LCIPercent, input.TermDays),
		calculate("LCA", input.Amount, businessDays, dailyCDI, input.LCAPercent, input.TermDays),
	}
	return Simulation{
		CDIAnnualPercent: cdiAnnual, EstimatedBusinessDays: businessDays, Results: results,
		Assumption: "Projeção com a taxa CDI atual mantida constante; dias úteis estimados pela proporção 252/365, sem feriados.",
	}, nil
}

func calculate(product string, amount float64, days int, dailyCDI, percentCDI float64, termDays int) InvestmentResult {
	factor := math.Pow(1+dailyCDI*percentCDI/100, float64(days))
	grossProfit := amount * (factor - 1)
	taxPercent := 0.0
	if product == "CDB" {
		switch {
		case termDays <= 180:
			taxPercent = 22.5
		case termDays <= 360:
			taxPercent = 20
		case termDays <= 720:
			taxPercent = 17.5
		default:
			taxPercent = 15
		}
	}
	tax := grossProfit * taxPercent / 100
	netProfit := grossProfit - tax
	return InvestmentResult{Product: product, GrossProfit: roundCent(grossProfit), Tax: roundCent(tax), NetProfit: roundCent(netProfit), NetAmount: roundCent(amount + netProfit), IncomeTaxPercent: taxPercent}
}

func roundCent(value float64) float64 { return math.Round(value*100) / 100 }
