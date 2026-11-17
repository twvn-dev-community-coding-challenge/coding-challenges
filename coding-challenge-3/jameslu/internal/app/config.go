package app

import (
	"fmt"
	"strconv"
	"time"
)

type Config struct {
	Port         string
	RadiusKm     float64
	OfferTimeout time.Duration
	Workers      int
	LogLevel     string // debug, info, warn, error
}

// ConfigFromEnv reads PORT, DISPATCH_RADIUS_KM, OFFER_TIMEOUT,
// DISPATCH_WORKERS and LOG_LEVEL, falling back to the defaults.
func ConfigFromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{Port: "8080", RadiusKm: 2, OfferTimeout: 30 * time.Second, Workers: 4, LogLevel: "info"}

	if v := getenv("PORT"); v != "" {
		cfg.Port = v
	}
	if v := getenv("DISPATCH_RADIUS_KM"); v != "" {
		km, err := strconv.ParseFloat(v, 64)
		if err != nil || km <= 0 {
			return Config{}, fmt.Errorf("DISPATCH_RADIUS_KM must be a positive number, got %q", v)
		}
		cfg.RadiusKm = km
	}
	if v := getenv("OFFER_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("OFFER_TIMEOUT must be a positive duration such as 30s, got %q", v)
		}
		cfg.OfferTimeout = d
	}
	if v := getenv("DISPATCH_WORKERS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("DISPATCH_WORKERS must be a positive integer, got %q", v)
		}
		cfg.Workers = n
	}
	if v := getenv("LOG_LEVEL"); v != "" {
		switch v {
		case "debug", "info", "warn", "error":
			cfg.LogLevel = v
		default:
			return Config{}, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error, got %q", v)
		}
	}
	return cfg, nil
}
