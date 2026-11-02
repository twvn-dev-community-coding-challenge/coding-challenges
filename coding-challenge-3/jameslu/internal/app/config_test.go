package app_test

import (
	"testing"
	"time"

	"ridedispatch/internal/app"
)

func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestConfigDefaults(t *testing.T) {
	cfg, err := app.ConfigFromEnv(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	want := app.Config{Port: "8080", RadiusKm: 2, OfferTimeout: 30 * time.Second, Workers: 4, LogLevel: "info"}
	if cfg != want {
		t.Fatalf("defaults = %+v, want %+v", cfg, want)
	}
}

func TestConfigFromEnvironment(t *testing.T) {
	cfg, err := app.ConfigFromEnv(env(map[string]string{
		"PORT": "9090", "DISPATCH_RADIUS_KM": "3.5", "OFFER_TIMEOUT": "10s", "DISPATCH_WORKERS": "8", "LOG_LEVEL": "debug",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := app.Config{Port: "9090", RadiusKm: 3.5, OfferTimeout: 10 * time.Second, Workers: 8, LogLevel: "debug"}
	if cfg != want {
		t.Fatalf("config = %+v, want %+v", cfg, want)
	}
}

func TestConfigRejectsInvalidValues(t *testing.T) {
	for k, v := range map[string]string{
		"DISPATCH_RADIUS_KM": "-1",
		"OFFER_TIMEOUT":      "soon",
		"DISPATCH_WORKERS":   "0",
		"LOG_LEVEL":          "loud",
	} {
		if _, err := app.ConfigFromEnv(env(map[string]string{k: v})); err == nil {
			t.Fatalf("%s=%s accepted", k, v)
		}
	}
}
