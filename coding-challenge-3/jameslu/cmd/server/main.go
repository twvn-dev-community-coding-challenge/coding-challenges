// Command server runs the ride-hailing dispatch service.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ridedispatch/internal/app"
	"ridedispatch/internal/platform/clock"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server.failed", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := app.ConfigFromEnv(os.Getenv)
	if err != nil {
		return err
	}
	var level slog.Level
	_ = level.UnmarshalText([]byte(cfg.LogLevel))
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	a, err := app.New(cfg, clock.Real{}, log)
	if err != nil {
		return err
	}
	a.Start()
	defer a.Stop()

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: a.Handler, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() { errs <- srv.ListenAndServe() }()
	log.Info("server.started", "port", cfg.Port, "radius_km", cfg.RadiusKm,
		"offer_timeout", cfg.OfferTimeout.String(), "workers", cfg.Workers)

	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("server.stopping")
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			return err
		}
	}
	return nil
}
