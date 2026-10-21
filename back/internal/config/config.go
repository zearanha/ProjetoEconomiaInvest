package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPAddr              string
	ProviderTimeout       time.Duration
	NewsRefreshInterval   time.Duration
	QuotesRefreshInterval time.Duration
	BrapiToken            string
	MacroProvider         string
	StockSymbols          string
	CryptoSymbols         string
	DatabaseURL           string
}

func Load() Config {
	return Config{
		HTTPAddr:              envOr("HTTP_ADDR", ":8080"),
		ProviderTimeout:       durationOr("PROVIDER_TIMEOUT", 8*time.Second),
		NewsRefreshInterval:   durationOr("NEWS_REFRESH_INTERVAL", 5*time.Minute),
		QuotesRefreshInterval: durationOr("QUOTES_REFRESH_INTERVAL", 30*time.Second),
		BrapiToken:            os.Getenv("BRAPI_TOKEN"),
		MacroProvider:         envOr("MACRO_PROVIDER", "bcb"),
		StockSymbols:          os.Getenv("STOCK_SYMBOLS"),
		CryptoSymbols:         os.Getenv("CRYPTO_SYMBOLS"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationOr(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
