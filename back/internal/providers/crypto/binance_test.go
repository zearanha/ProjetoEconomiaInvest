package crypto

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewBinanceUsesExpandedDefaults(t *testing.T) {
	provider := NewBinance(&http.Client{}, "")
	if len(provider.symbols) < 15 {
		t.Fatalf("got %d default markets, want at least 15", len(provider.symbols))
	}
	for _, expected := range []string{"BTCUSDT", "XRPUSDT", "DOGEUSDT", "LINKUSDT", "SHIBUSDT"} {
		if !contains(provider.symbols, expected) {
			t.Errorf("default symbols do not include %s", expected)
		}
	}
}

func TestNewBinanceNormalizesConfiguredSymbols(t *testing.T) {
	provider := NewBinance(&http.Client{}, " btcusdt, ETHUSDT,btcusdt,, solusdt ")
	want := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	if !reflect.DeepEqual(provider.symbols, want) {
		t.Fatalf("symbols = %#v, want %#v", provider.symbols, want)
	}
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
