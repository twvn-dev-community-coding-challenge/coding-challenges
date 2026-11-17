// Package app wires the modules together: in-memory adapters, the event bus,
// the HTTP API and the seed data. cmd/server and the API tests both use it,
// so the tests exercise exactly what runs.
package app

import (
	"context"
	"log/slog"
	"net/http"

	"ridedispatch/internal/dispatch"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/httpapi"
	"ridedispatch/internal/passenger"
	"ridedispatch/internal/platform/bus"
	"ridedispatch/internal/platform/clock"
	"ridedispatch/internal/ride"
	"ridedispatch/internal/timeline"
)

type App struct {
	Handler  http.Handler
	Bus      *bus.Bus
	Rides    *ride.Service
	Drivers  *driver.Service
	Dispatch *dispatch.Service
	Timeline *timeline.Store
}

// SeedPickup is the brief's example pickup; seed drivers are placed around it.
var SeedPickup = geo.MustLocation(10.7769, 106.7009)

// SeedPassengers and SeedDrivers stand in for registration, which is out of
// scope. Distances are due north of SeedPickup and match the brief's matching
// example (D001 0.8 km, D002 1.5 km, D003 3.2 km).
var (
	SeedPassengers = []string{"P001", "P002", "P003"}
	SeedDrivers    = []struct {
		ID        string
		Km        float64
		Available bool
	}{
		{"D001", 0.8, true},
		{"D002", 1.5, true},
		{"D003", 3.2, true},
		{"D004", 0.3, false},
	}
)

func New(cfg Config, clk clock.Clock, log *slog.Logger) (*App, error) {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	b := bus.New(bus.Options{Partitions: cfg.Workers, Logger: log})

	rides := ride.NewService(ride.Deps{
		Repo:        ride.NewMemoryRepository(),
		Passengers:  passenger.NewDirectory(SeedPassengers...),
		Idempotency: ride.NewIdempotencyStore(),
		Events:      b,
		Clock:       clk,
		Logger:      log,
	})
	drivers := driver.NewService(driver.NewMemoryRepository(), clk, log)
	dispatcher := dispatch.NewService(dispatch.Deps{
		Rides:   rides,
		Drivers: drivers,
		Offers:  dispatch.NewOfferStore(),
		Events:  b,
		Clock:   clk,
		Logger:  log,
		Config:  dispatch.Config{RadiusKm: cfg.RadiusKm, OfferTimeout: cfg.OfferTimeout},
	})
	tl := timeline.NewStore()

	b.Subscribe("dispatch", dispatch.SubscribedEvents, dispatcher.HandleEvent)
	b.Subscribe("timeline", nil, tl.Append)

	if err := seedDrivers(context.Background(), drivers); err != nil {
		return nil, err
	}

	return &App{
		Handler:  httpapi.New(httpapi.Deps{Rides: rides, Drivers: drivers, Dispatch: dispatcher, Timeline: tl, Logger: log}),
		Bus:      b,
		Rides:    rides,
		Drivers:  drivers,
		Dispatch: dispatcher,
		Timeline: tl,
	}, nil
}

func (a *App) Start() { a.Bus.Start() }
func (a *App) Stop()  { a.Bus.Stop() }

func seedDrivers(ctx context.Context, drivers *driver.Service) error {
	for _, d := range SeedDrivers {
		if err := drivers.Register(ctx, d.ID, geo.OffsetNorthKm(SeedPickup, d.Km)); err != nil {
			return err
		}
		if d.Available {
			if _, err := drivers.SetAvailability(ctx, d.ID, driver.Available); err != nil {
				return err
			}
		}
	}
	return nil
}
