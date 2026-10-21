package domain

import "time"

type Quote struct {
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Market    string    `json:"market"`
	Currency  string    `json:"currency"`
	Price     float64   `json:"price"`
	Change    float64   `json:"change"`
	ChangePct float64   `json:"changePct"`
	Volume    float64   `json:"volume,omitempty"`
	High      float64   `json:"high,omitempty"`
	Low       float64   `json:"low,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type QuotePoint struct {
	Time  time.Time `json:"time"`
	Price float64   `json:"price"`
}

type FixedIncomeRate struct {
	Indicator string    `json:"indicator"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Date      time.Time `json:"date"`
	Source    string    `json:"source"`
}
