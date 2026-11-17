package ride

import (
	"context"
	"fmt"
	"log/slog"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/platform/clock"
	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/platform/ids"
)

// PassengerDirectory answers whether a passenger record exists.
type PassengerDirectory interface {
	Exists(ctx context.Context, id string) bool
}

type Deps struct {
	Repo        Repository
	Passengers  PassengerDirectory
	Idempotency *IdempotencyStore
	Events      events.Publisher
	Clock       clock.Clock
	Logger      *slog.Logger
}

// Service is the Ride module's API. Other modules change rides only through
// these methods.
type Service struct {
	repo       Repository
	passengers PassengerDirectory
	idem       *IdempotencyStore
	events     events.Publisher
	clock      clock.Clock
	log        *slog.Logger
}

func NewService(d Deps) *Service {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	return &Service{
		repo:       d.Repo,
		passengers: d.Passengers,
		idem:       d.Idempotency,
		events:     d.Events,
		clock:      d.Clock,
		log:        d.Logger,
	}
}

type CreateRequest struct {
	PassengerID string       `json:"passengerId"`
	Pickup      geo.Location `json:"pickup"`
	Destination geo.Location `json:"destination"`
}

// Create stores a new REQUESTED ride and publishes RideCreated. Repeating a
// request with the same idempotency key returns the original ride with
// replayed=true and creates nothing.
func (s *Service) Create(ctx context.Context, key string, req CreateRequest) (Ride, bool, error) {
	if key == "" {
		return Ride{}, false, fmt.Errorf("%w: send an Idempotency-Key header", apperr.ErrIdempotencyKeyMissing)
	}
	if err := s.validate(ctx, req); err != nil {
		return Ride{}, false, err
	}

	original, replayed, err := s.idem.Reserve(key, RequestHash(req))
	if err != nil {
		return Ride{}, false, err
	}
	if replayed {
		s.log.Info("ride.create_replayed", "ride_id", original.ID(), "idempotency_key", key)
		return original, true, nil
	}

	r := New(ids.New("R"), req.PassengerID, req.Pickup, req.Destination, s.clock.Now())
	if err := s.repo.Create(ctx, r); err != nil {
		s.idem.Abandon(key)
		return Ride{}, false, err
	}
	s.idem.Complete(key, r)

	created := events.New(ctx, events.RideCreated, r.ID(), "", s.clock.Now())
	s.log.Info("ride.created", "ride_id", r.ID(), "passenger_id", r.PassengerID(),
		"event_id", created.ID, "causation_id", created.CausationID)
	s.events.Publish(ctx, created)
	return r, false, nil
}

func (s *Service) validate(ctx context.Context, req CreateRequest) error {
	if req.PassengerID == "" {
		return fmt.Errorf("%w: passengerId is required", apperr.ErrValidation)
	}
	for name, loc := range map[string]geo.Location{"pickup": req.Pickup, "destination": req.Destination} {
		if _, err := geo.NewLocation(loc.Latitude, loc.Longitude); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	if !s.passengers.Exists(ctx, req.PassengerID) {
		return fmt.Errorf("%w: passenger %s", apperr.ErrNotFound, req.PassengerID)
	}
	return nil
}

func (s *Service) Get(ctx context.Context, id string) (Ride, error) {
	return s.repo.Get(ctx, id)
}

// The transition methods below are the commands the Dispatch saga uses. Each
// is one atomic, idempotent step; changed reports whether this call moved
// the ride.

func (s *Service) StartMatching(ctx context.Context, id string) (Ride, bool, error) {
	return s.repo.Update(ctx, id, func(r *Ride) (bool, error) { return r.StartMatching(s.clock.Now()) })
}

func (s *Service) Assign(ctx context.Context, id, driverID string) (Ride, bool, error) {
	return s.repo.Update(ctx, id, func(r *Ride) (bool, error) { return r.Assign(driverID, s.clock.Now()) })
}

func (s *Service) ReleaseAssignment(ctx context.Context, id, driverID string) (Ride, bool, error) {
	return s.repo.Update(ctx, id, func(r *Ride) (bool, error) { return r.ReleaseAssignment(driverID, s.clock.Now()) })
}

func (s *Service) Accept(ctx context.Context, id, driverID string) (Ride, bool, error) {
	return s.repo.Update(ctx, id, func(r *Ride) (bool, error) { return r.Accept(driverID, s.clock.Now()) })
}

func (s *Service) Cancel(ctx context.Context, id string) (Ride, bool, error) {
	return s.repo.Update(ctx, id, func(r *Ride) (bool, error) { return r.Cancel(s.clock.Now()) })
}

func (s *Service) Complete(ctx context.Context, id string) (Ride, bool, error) {
	return s.repo.Update(ctx, id, func(r *Ride) (bool, error) { return r.Complete(s.clock.Now()) })
}
