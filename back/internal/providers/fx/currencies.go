package fx

// currencies is the set of fiat currencies shown against the Brazilian real.
var currencies = []struct{ code, name string }{
	{"USD", "Dólar americano"},
	{"EUR", "Euro"},
	{"GBP", "Libra esterlina"},
	{"JPY", "Iene japonês"},
	{"CAD", "Dólar canadense"},
	{"AUD", "Dólar australiano"},
	{"CHF", "Franco suíço"},
	{"CNY", "Yuan chinês"},
	{"HKD", "Dólar de Hong Kong"},
	{"INR", "Rupia indiana"},
	{"KRW", "Won sul-coreano"},
	{"MXN", "Peso mexicano"},
	{"NOK", "Coroa norueguesa"},
	{"NZD", "Dólar neozelandês"},
	{"SEK", "Coroa sueca"},
	{"SGD", "Dólar de Singapura"},
	{"ZAR", "Rand sul-africano"},
	{"TRY", "Lira turca"},
	{"PLN", "Zloty polonês"},
	{"DKK", "Coroa dinamarquesa"},
}

func currencyCodes() []string {
	codes := make([]string, len(currencies))
	for i, currency := range currencies {
		codes[i] = currency.code
	}
	return codes
}

func currencyName(code string) string {
	for _, currency := range currencies {
		if currency.code == code {
			return currency.name
		}
	}
	return code
}
