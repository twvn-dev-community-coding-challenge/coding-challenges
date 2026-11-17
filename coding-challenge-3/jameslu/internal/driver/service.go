package driver

import (
	"context"
	"log/slog"

	"ridedispatch/internal/geo"
	"ridedispatch/internal/platform/clock"
)

// Service is the Driver module's API. Other modules change drivers only
// through these methods.
type Service struct {
	repo  Repository
	clock clock.Clock
	log   *slog.Logger
}

func NewService(repo Repository, clk clock.Clock, log *slog.Logger) *Service {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Service{repo: repo, clock: clk, log: log}
}

// Candidate is what matching needs to know about a driver.
type Candidate struct {
	ID       string
	Location geo.Location
}

// Register adds an OFFLINE driver. It is used for seed data only;
// registration flows are out of scope.
func (s *Service) Register(ctx context.Context, id string, loc geo.Location) error {
	return s.repo.Add(ctx, New(id, loc, s.clock.Now()))
}

func (s *Service) Get(ctx context.Context, id string) (Driver, error) {
	return s.repo.Get(ctx, id)
}

// SetAvailability is the availability API (OFFLINE <-> AVAILABLE).
func (s *Service) SetAvailability(ctx context.Context, id string, target Status) (Driver, error) {
	d, changed, err := s.repo.Update(ctx, id, func(d *Driver) (bool, error) {
		return d.SetAvailability(target, s.clock.Now())
	})
	if err == nil && changed {
		s.log.Info("driver.availability_changed", "driver_id", id, "status", d.Status())
	}
	return d, err
}

func (s *Service) UpdateLocation(ctx context.Context, id string, loc geo.Location) (Driver, error) {
	if _, err := geo.NewLocation(loc.Latitude, loc.Longitude); err != nil {
		return Driver{}, err
	}
	d, _, err := s.repo.Update(ctx, id, func(d *Driver) (bool, error) {
		d.UpdateLocation(loc, s.clock.Now())
		return true, nil
	})
	if err == nil {
		s.log.Debug("driver.location_updated", "driver_id", id, "latitude", loc.Latitude, "longitude", loc.Longitude)
	}
	return d, err
}

// Claim, MarkBusy and Release are the steps the Dispatch saga uses. Each is
// one atomic, idempotent step; changed reports whether this call moved the
// driver.

func (s *Service) Claim(ctx context.Context, id, rideID string) (bool, error) {
	_, changed, err := s.repo.Update(ctx, id, func(d *Driver) (bool, error) { return d.Claim(rideID, s.clock.Now()) })
	return changed, err
}

func (s *Service) MarkBusy(ctx context.Context, id, rideID string) (bool, error) {
	_, changed, err := s.repo.Update(ctx, id, func(d *Driver) (bool, error) { return d.MarkBusy(rideID, s.clock.Now()) })
	return changed, err
}

func (s *Service) Release(ctx context.Context, id, rideID string) (bool, error) {
	_, changed, err := s.repo.Update(ctx, id, func(d *Driver) (bool, error) { return d.Release(rideID, s.clock.Now()) })
	return changed, err
}

// AvailableNear returns AVAILABLE drivers that may be within radiusKm of p.
// The in-memory implementation has no spatial index, so it returns every
// AVAILABLE driver and leaves exact filtering to the matcher. A geo-indexed
// implementation would return only drivers in nearby cells.
func (s *Service) AvailableNear(ctx context.Context, _ geo.Location, _ float64) []Candidate {
	drivers := s.repo.Available(ctx)
	out := make([]Candidate, len(drivers))
	for i, d := range drivers {
		out[i] = Candidate{ID: d.ID(), Location: d.Location()}
	}
	return out
}
