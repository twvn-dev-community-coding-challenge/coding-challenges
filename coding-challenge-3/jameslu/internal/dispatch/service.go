package dispatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/platform/clock"
	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/platform/ids"
	"ridedispatch/internal/platform/trace"
	"ridedispatch/internal/ride"
)

const DefaultOfferTimeout = 30 * time.Second

// Rides is what Dispatch needs from the Ride module. Each method is one
// atomic, idempotent transition; changed reports whether the call moved the ride.
type Rides interface {
	Get(ctx context.Context, id string) (ride.Ride, error)
	StartMatching(ctx context.Context, id string) (ride.Ride, bool, error)
	Assign(ctx context.Context, id, driverID string) (ride.Ride, bool, error)
	ReleaseAssignment(ctx context.Context, id, driverID string) (ride.Ride, bool, error)
	Accept(ctx context.Context, id, driverID string) (ride.Ride, bool, error)
	Cancel(ctx context.Context, id string) (ride.Ride, bool, error)
	Complete(ctx context.Context, id string) (ride.Ride, bool, error)
}

// Drivers is what Dispatch needs from the Driver module.
type Drivers interface {
	DriverLocator
	Claim(ctx context.Context, id, rideID string) (bool, error)
	MarkBusy(ctx context.Context, id, rideID string) (bool, error)
	Release(ctx context.Context, id, rideID string) (bool, error)
}

type Config struct {
	RadiusKm     float64       // default 2
	OfferTimeout time.Duration // default 30s
}

type Deps struct {
	Rides   Rides
	Drivers Drivers
	Offers  *OfferStore
	Events  events.Publisher
	Clock   clock.Clock
	Logger  *slog.Logger
	Config  Config
}

// SubscribedEvents are the events that (re)start matching for a ride.
var SubscribedEvents = []string{events.RideCreated, events.OfferRejected, events.OfferExpired}

// Service orchestrates the offer saga (docs/architecture.md §4). Every
// command is a sequence of single-aggregate steps; each step is idempotent,
// so replaying a command is safe. Once an offer exists, the ride is the
// arbiter: competing commands race on a ride transition and only the winner
// touches the driver and the offer.
type Service struct {
	rides   Rides
	drivers Drivers
	offers  *OfferStore
	events  events.Publisher
	clock   clock.Clock
	log     *slog.Logger
	matcher Matcher
	timeout time.Duration

	mu     sync.Mutex
	timers map[offerKey]clock.Timer
}

func NewService(d Deps) *Service {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	if d.Config.OfferTimeout <= 0 {
		d.Config.OfferTimeout = DefaultOfferTimeout
	}
	return &Service{
		rides:   d.Rides,
		drivers: d.Drivers,
		offers:  d.Offers,
		events:  d.Events,
		clock:   d.Clock,
		log:     d.Logger,
		matcher: Matcher{Locator: d.Drivers, RadiusKm: d.Config.RadiusKm},
		timeout: d.Config.OfferTimeout,
		timers:  map[offerKey]clock.Timer{},
	}
}

// HandleEvent is the bus handler for SubscribedEvents.
func (s *Service) HandleEvent(ctx context.Context, e events.Event) error {
	switch e.Type {
	case events.RideCreated, events.OfferRejected, events.OfferExpired:
		return s.Match(ctx, e.RideID)
	}
	return nil
}

// Match offers the ride to the nearest eligible driver it can claim. It does
// nothing if the ride is not waiting for a driver, so duplicate events are
// harmless.
func (s *Service) Match(ctx context.Context, rideID string) error {
	log := s.logger(ctx, rideID, "")
	r, err := s.rides.Get(ctx, rideID)
	if errors.Is(err, apperr.ErrNotFound) {
		log.Warn("dispatch.ride_not_found")
		return nil
	}
	if err != nil {
		return err
	}
	if r.Status() == ride.Requested {
		if r, _, err = s.rides.StartMatching(ctx, rideID); err != nil {
			return err
		}
	}
	if r.Status() != ride.Matching || r.AssignedDriverID() != "" {
		log.Info("dispatch.matching_skipped", "ride_status", r.Status())
		return nil
	}

	excluded := s.offers.Excluded(ctx, rideID)
	candidates := s.matcher.Candidates(ctx, r.Pickup(), excluded)
	log.Info("dispatch.matching_started", "candidates", len(candidates), "excluded", len(excluded))

	for _, c := range candidates {
		done, err := s.tryOffer(ctx, r.ID(), c)
		if done || err != nil {
			return err
		}
	}

	exhausted := events.New(ctx, events.MatchingExhausted, rideID, "", s.clock.Now())
	log.Info("dispatch.matching_exhausted", "event_id", exhausted.ID)
	s.events.Publish(ctx, exhausted)
	return nil
}

// tryOffer runs the offer saga for one candidate: claim the driver, assign the
// ride, record the offer. done=false means the claim was lost and the next
// candidate should be tried.
func (s *Service) tryOffer(ctx context.Context, rideID string, c Candidate) (bool, error) {
	log := s.logger(ctx, rideID, c.DriverID)

	// Step 1: claim the contended resource first.
	if _, err := s.drivers.Claim(ctx, c.DriverID, rideID); err != nil {
		if errors.Is(err, apperr.ErrConflict) {
			log.Info("dispatch.claim_lost", "reason", err.Error())
			return false, nil
		}
		return true, err
	}

	// Step 2: assign the ride. If the ride moved on (cancelled, or assigned
	// by a concurrent worker), compensate by releasing the driver.
	_, assigned, err := s.rides.Assign(ctx, rideID, c.DriverID)
	if err != nil {
		if _, rerr := s.drivers.Release(ctx, c.DriverID, rideID); rerr != nil {
			return true, rerr
		}
		log.Info("dispatch.claim_compensated", "reason", err.Error())
		if errors.Is(err, apperr.ErrConflict) || errors.Is(err, apperr.ErrInvalidTransition) {
			return true, nil
		}
		return true, err
	}
	if !assigned {
		// A concurrent worker already assigned this driver to this ride.
		return true, nil
	}

	// Step 3: record the offer and arm its timeout.
	now := s.clock.Now()
	o := NewOffer(ids.New("O"), rideID, c.DriverID, now, now.Add(s.timeout))
	if err := s.offers.Create(ctx, o); err != nil {
		// An invariant was violated: undo the saga so nothing is left half-done.
		_, _, _ = s.rides.ReleaseAssignment(ctx, rideID, c.DriverID)
		_, _ = s.drivers.Release(ctx, c.DriverID, rideID)
		return true, fmt.Errorf("create offer: %w", err)
	}

	// A cancel that ran between steps 2 and 3 could not see this offer. The
	// cancel transitions the ride before it looks for the offer, so either it
	// found the offer, or this read sees the cancelled ride and closes it.
	if current, err := s.rides.Get(ctx, rideID); err == nil && current.Status().Terminal() {
		log.Info("dispatch.offer_superseded", "offer_id", o.ID(), "ride_status", current.Status())
		return true, s.reconcile(ctx, rideID, c.DriverID)
	}

	created := events.New(ctx, events.OfferCreated, rideID, c.DriverID, now)
	log.Info("dispatch.offer_created", "offer_id", o.ID(), "event_id", created.ID,
		"distance_km", c.DistanceKm, "expires_at", o.ExpiresAt())
	s.armTimer(rideID, c.DriverID, created.ID)
	s.events.Publish(ctx, created)
	return true, nil
}

// Accept confirms the offer held by driverID.
func (s *Service) Accept(ctx context.Context, rideID, driverID string) (ride.Ride, error) {
	if done, r, err := s.checkOffer(ctx, rideID, driverID, OfferAccepted); done {
		return r, err
	}
	r, _, err := s.rides.Accept(ctx, rideID, driverID)
	if err != nil {
		return ride.Ride{}, err
	}
	if err := s.finishAccept(ctx, rideID, driverID); err != nil {
		return ride.Ride{}, err
	}
	return r, nil
}

// finishAccept runs the steps after the ride accepted: driver BUSY, offer
// ACCEPTED. It is also the roll-forward path of an expiring offer.
func (s *Service) finishAccept(ctx context.Context, rideID, driverID string) error {
	if _, err := s.drivers.MarkBusy(ctx, driverID, rideID); err != nil {
		return err
	}
	o, changed, err := s.offers.Resolve(ctx, rideID, driverID, OfferAccepted, s.clock.Now())
	if err != nil {
		return err
	}
	s.stopTimer(rideID, driverID)
	if changed {
		accepted := events.New(ctx, events.OfferAccepted, rideID, driverID, s.clock.Now())
		s.logger(ctx, rideID, driverID).Info("offer.accepted", "offer_id", o.ID(), "event_id", accepted.ID)
		s.events.Publish(ctx, accepted)
	}
	return nil
}

// Reject declines the offer held by driverID and triggers rematching.
func (s *Service) Reject(ctx context.Context, rideID, driverID string) (ride.Ride, error) {
	if done, r, err := s.checkOffer(ctx, rideID, driverID, OfferRejected); done {
		return r, err
	}
	if _, _, err := s.rides.ReleaseAssignment(ctx, rideID, driverID); err != nil {
		return ride.Ride{}, err
	}
	// Record the rejection before freeing the driver, so a rematch never
	// sees the driver AVAILABLE without also seeing the exclusion.
	if err := s.closeOffer(ctx, rideID, driverID, OfferRejected, events.OfferRejected, "offer.rejected"); err != nil {
		return ride.Ride{}, err
	}
	return s.rides.Get(ctx, rideID)
}

// Expire is called by the offer timer. It follows the rejection path, unless
// the driver accepted in the meantime, in which case it rolls the accept
// forward. It is also what repairs a saga interrupted half-way.
func (s *Service) Expire(ctx context.Context, rideID, driverID string) error {
	o, err := s.offers.Get(ctx, rideID, driverID)
	if err != nil || o.Status() != OfferPending {
		return nil
	}
	if _, _, err := s.rides.ReleaseAssignment(ctx, rideID, driverID); err != nil {
		if errors.Is(err, apperr.ErrConflict) {
			return s.reconcile(ctx, rideID, driverID)
		}
		return err
	}
	err = s.closeOffer(ctx, rideID, driverID, OfferExpired, events.OfferExpired, "offer.expired")
	if errors.Is(err, apperr.ErrConflict) {
		return nil // a concurrent reject resolved the offer first
	}
	return err
}

// reconcile settles a pending offer whose ride no longer waits for it.
func (s *Service) reconcile(ctx context.Context, rideID, driverID string) error {
	r, err := s.rides.Get(ctx, rideID)
	if err != nil {
		return err
	}
	log := s.logger(ctx, rideID, driverID)
	if r.Status() == ride.DriverAccepted && r.AssignedDriverID() == driverID {
		log.Info("offer.rolled_forward")
		return s.finishAccept(ctx, rideID, driverID)
	}
	if _, _, err := s.offers.Resolve(ctx, rideID, driverID, OfferCancelled, s.clock.Now()); err != nil && !errors.Is(err, apperr.ErrConflict) {
		return err
	}
	s.stopTimer(rideID, driverID)
	if _, err := s.drivers.Release(ctx, driverID, rideID); err != nil {
		return err
	}
	log.Info("offer.closed", "ride_status", r.Status())
	return nil
}

// Cancel ends the ride and frees its driver, if any.
func (s *Service) Cancel(ctx context.Context, rideID string) (ride.Ride, error) {
	r, changed, err := s.rides.Cancel(ctx, rideID)
	if err != nil {
		return ride.Ride{}, err
	}
	if d := r.AssignedDriverID(); d != "" {
		_, _, err := s.offers.Resolve(ctx, rideID, d, OfferCancelled, s.clock.Now())
		if err != nil && !errors.Is(err, apperr.ErrConflict) && !errors.Is(err, apperr.ErrNotFound) {
			return ride.Ride{}, err
		}
		s.stopTimer(rideID, d)
		if _, err := s.drivers.Release(ctx, d, rideID); err != nil {
			return ride.Ride{}, err
		}
	}
	if changed {
		s.publishRideEvent(ctx, r, events.RideCancelled, "ride.cancelled")
	}
	return r, nil
}

// Complete finishes an accepted ride and frees its driver.
func (s *Service) Complete(ctx context.Context, rideID string) (ride.Ride, error) {
	r, changed, err := s.rides.Complete(ctx, rideID)
	if err != nil {
		return ride.Ride{}, err
	}
	if _, err := s.drivers.Release(ctx, r.AssignedDriverID(), rideID); err != nil {
		return ride.Ride{}, err
	}
	if changed {
		s.publishRideEvent(ctx, r, events.RideCompleted, "ride.completed")
	}
	return r, nil
}

// checkOffer looks up the offer a driver is responding to. done=true means
// the command is already answered: no offer (404), the same response was
// already applied (idempotent 200), or the offer is no longer pending (409).
func (s *Service) checkOffer(ctx context.Context, rideID, driverID string, want OfferStatus) (bool, ride.Ride, error) {
	o, err := s.offers.Get(ctx, rideID, driverID)
	if err != nil {
		return true, ride.Ride{}, err
	}
	switch o.Status() {
	case OfferPending:
		return false, ride.Ride{}, nil
	case want:
		r, err := s.rides.Get(ctx, rideID)
		return true, r, err
	default:
		return true, ride.Ride{}, fmt.Errorf("%w: offer of ride %s to driver %s is %s", apperr.ErrConflict, rideID, driverID, o.Status())
	}
}

// closeOffer resolves a pending offer as rejected or expired, frees the driver
// and publishes the event that triggers rematching. Only the caller that
// actually resolves the offer publishes.
func (s *Service) closeOffer(ctx context.Context, rideID, driverID string, to OfferStatus, eventType, logMsg string) error {
	o, changed, err := s.offers.Resolve(ctx, rideID, driverID, to, s.clock.Now())
	if err != nil {
		return err
	}
	s.stopTimer(rideID, driverID)
	if _, err := s.drivers.Release(ctx, driverID, rideID); err != nil {
		return err
	}
	if changed {
		e := events.New(ctx, eventType, rideID, driverID, s.clock.Now())
		s.logger(ctx, rideID, driverID).Info(logMsg, "offer_id", o.ID(), "event_id", e.ID)
		s.events.Publish(ctx, e)
	}
	return nil
}

func (s *Service) publishRideEvent(ctx context.Context, r ride.Ride, eventType, logMsg string) {
	e := events.New(ctx, eventType, r.ID(), r.AssignedDriverID(), s.clock.Now())
	s.logger(ctx, r.ID(), r.AssignedDriverID()).Info(logMsg, "event_id", e.ID)
	s.events.Publish(ctx, e)
}

// armTimer schedules the offer timeout. causationID (the OfferCreated event)
// becomes the cause of the resulting OfferExpired.
func (s *Service) armTimer(rideID, driverID, causationID string) {
	key := offerKey{rideID, driverID}
	t := s.clock.AfterFunc(s.timeout, func() {
		s.mu.Lock()
		delete(s.timers, key)
		s.mu.Unlock()
		ctx := trace.WithCausation(context.Background(), causationID)
		if err := s.Expire(ctx, rideID, driverID); err != nil {
			s.logger(ctx, rideID, driverID).Error("offer.expire_failed", "error", err.Error())
		}
	})
	s.mu.Lock()
	s.timers[key] = t
	s.mu.Unlock()
}

func (s *Service) stopTimer(rideID, driverID string) {
	key := offerKey{rideID, driverID}
	s.mu.Lock()
	t, ok := s.timers[key]
	delete(s.timers, key)
	s.mu.Unlock()
	if ok {
		t.Stop()
	}
}

func (s *Service) logger(ctx context.Context, rideID, driverID string) *slog.Logger {
	l := s.log.With("ride_id", rideID)
	if driverID != "" {
		l = l.With("driver_id", driverID)
	}
	if c := trace.CausationID(ctx); c != "" {
		l = l.With("causation_id", c)
	}
	return l
}
