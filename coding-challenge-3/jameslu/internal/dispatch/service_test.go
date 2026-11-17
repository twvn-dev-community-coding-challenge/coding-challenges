package dispatch_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/dispatch"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/ride"
)

// Distances used throughout: D001 is nearest, D002 next, D003 is outside the
// default 2 km radius.
var standardFleet = map[string]float64{"D001": 0.8, "D002": 1.5, "D003": 3.2}

func TestRideCreatedEventTriggersOfferToNearestDriver(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()

	e.expectRide(r, ride.DriverAssigned, "D001")
	e.expectDriver("D001", driver.Offered, r)
	e.expectDriver("D002", driver.Available, "")

	o, err := e.offers.Get(e.ctx, r, "D001")
	if err != nil || o.Status() != dispatch.OfferPending || !o.ExpiresAt().Equal(t0.Add(30*time.Second)) {
		t.Fatalf("offer = %+v, %v", o, err)
	}

	created, _ := e.log.last(r, events.RideCreated)
	offered, ok := e.log.last(r, events.OfferCreated)
	if !ok || offered.DriverID != "D001" || offered.CausationID != created.ID {
		t.Fatalf("OfferCreated = %+v, want driver D001 caused by %s", offered, created.ID)
	}
}

func TestAcceptOffer(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()

	got, err := e.dispatch.Accept(e.ctx, r, "D001")
	if err != nil || got.Status() != ride.DriverAccepted {
		t.Fatalf("Accept = %s, %v", got.Status(), err)
	}
	e.settle()
	e.expectRide(r, ride.DriverAccepted, "D001")
	e.expectDriver("D001", driver.Busy, r)
	if e.offerStatus(r, "D001") != dispatch.OfferAccepted {
		t.Fatal("offer not accepted")
	}
}

func TestDuplicateAcceptIsIdempotent(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()

	for range 3 {
		got, err := e.dispatch.Accept(e.ctx, r, "D001")
		if err != nil || got.Status() != ride.DriverAccepted {
			t.Fatalf("Accept = %s, %v", got.Status(), err)
		}
	}
	e.settle()
	e.expectDriver("D001", driver.Busy, r)
	if n := e.log.count(r, events.OfferAccepted); n != 1 {
		t.Fatalf("OfferAccepted published %d times, want 1", n)
	}
}

func TestOnlyTheOfferHolderCanRespond(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()

	if _, err := e.dispatch.Accept(e.ctx, r, "D002"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("accept by non-holder error = %v, want ErrNotFound", err)
	}
	if _, err := e.dispatch.Reject(e.ctx, r, "D002"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("reject by non-holder error = %v, want ErrNotFound", err)
	}
	e.expectRide(r, ride.DriverAssigned, "D001")
	e.expectDriver("D002", driver.Available, "")
}

func TestRejectReturnsRideToMatchingAndOffersNextDriver(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()

	got, err := e.dispatch.Reject(e.ctx, r, "D001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status() != ride.Matching || got.AssignedDriverID() != "" {
		t.Fatalf("after reject ride = %s/%q, want MATCHING with no driver", got.Status(), got.AssignedDriverID())
	}
	e.settle()

	e.expectDriver("D001", driver.Available, "")
	e.expectRide(r, ride.DriverAssigned, "D002")
	e.expectDriver("D002", driver.Offered, r)

	if _, err := e.dispatch.Accept(e.ctx, r, "D002"); err != nil {
		t.Fatal(err)
	}
	e.expectRide(r, ride.DriverAccepted, "D002")
}

func TestDuplicateRejectIsIdempotent(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()
	_, _ = e.dispatch.Reject(e.ctx, r, "D001")
	e.settle()

	if _, err := e.dispatch.Reject(e.ctx, r, "D001"); err != nil {
		t.Fatalf("duplicate reject error = %v", err)
	}
	e.settle()
	if n := e.log.count(r, events.OfferRejected); n != 1 {
		t.Fatalf("OfferRejected published %d times, want 1", n)
	}
	e.expectRide(r, ride.DriverAssigned, "D002")
}

func TestRejectedDriversAreExcludedForThatRideOnly(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()

	_, _ = e.dispatch.Reject(e.ctx, r, "D001")
	e.settle()
	_, _ = e.dispatch.Reject(e.ctx, r, "D002")
	e.settle()

	// No eligible driver remains: D003 is outside the radius.
	e.expectRide(r, ride.Matching, "")
	e.expectDriver("D001", driver.Available, "")
	e.expectDriver("D002", driver.Available, "")
	if e.log.count(r, events.MatchingExhausted) != 1 {
		t.Fatal("MatchingExhausted not published")
	}

	// Matching again (e.g. a redelivered event) still never re-offers them.
	if err := e.dispatch.Match(e.ctx, r); err != nil {
		t.Fatal(err)
	}
	e.settle()
	e.expectRide(r, ride.Matching, "")

	// They remain eligible for other rides.
	other := e.createRide()
	e.expectRide(other, ride.DriverAssigned, "D001")
}

func TestNoDriverAvailableLeavesRideMatching(t *testing.T) {
	e := newEnv(t, map[string]float64{"FAR": 5})
	r := e.createRide()
	e.expectRide(r, ride.Matching, "")
	e.expectDriver("FAR", driver.Available, "")
	if e.log.count(r, events.MatchingExhausted) != 1 {
		t.Fatal("MatchingExhausted not published")
	}
}

func TestOfferTimeoutIsTreatedAsRejection(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()
	offered, _ := e.log.last(r, events.OfferCreated)

	e.clock.Advance(29 * time.Second)
	e.settle()
	e.expectRide(r, ride.DriverAssigned, "D001")

	e.clock.Advance(time.Second)
	e.settle()

	if e.offerStatus(r, "D001") != dispatch.OfferExpired {
		t.Fatal("offer not expired")
	}
	e.expectDriver("D001", driver.Available, "")
	e.expectRide(r, ride.DriverAssigned, "D002")

	expired, ok := e.log.last(r, events.OfferExpired)
	if !ok || expired.DriverID != "D001" || expired.CausationID != offered.ID {
		t.Fatalf("OfferExpired = %+v, want driver D001 caused by %s", expired, offered.ID)
	}

	// The expired driver can no longer accept.
	if _, err := e.dispatch.Accept(e.ctx, r, "D001"); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("accept after expiry error = %v, want ErrConflict", err)
	}
}

func TestOfferTimeoutIsConfigurable(t *testing.T) {
	e := newEnv(t, standardFleet, envOptions{cfg: dispatch.Config{OfferTimeout: 10 * time.Second}})
	r := e.createRide()
	e.clock.Advance(10 * time.Second)
	e.settle()
	e.expectRide(r, ride.DriverAssigned, "D002")
}

func TestRespondedOfferDoesNotExpire(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()
	_, _ = e.dispatch.Accept(e.ctx, r, "D001")
	e.clock.Advance(time.Minute)
	e.settle()
	e.expectRide(r, ride.DriverAccepted, "D001")
	e.expectDriver("D001", driver.Busy, r)
}

func TestCancelPendingOfferReleasesDriver(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()

	got, err := e.dispatch.Cancel(e.ctx, r)
	if err != nil || got.Status() != ride.Cancelled {
		t.Fatalf("Cancel = %s, %v", got.Status(), err)
	}
	e.settle()
	e.expectDriver("D001", driver.Available, "")
	if e.offerStatus(r, "D001") != dispatch.OfferCancelled {
		t.Fatal("pending offer not cancelled")
	}
	if _, err := e.dispatch.Accept(e.ctx, r, "D001"); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("accept after cancel error = %v, want ErrConflict", err)
	}

	// The timer finds nothing to do.
	e.clock.Advance(time.Minute)
	e.settle()
	e.expectDriver("D001", driver.Available, "")
}

func TestCancelAcceptedRideReleasesDriver(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()
	_, _ = e.dispatch.Accept(e.ctx, r, "D001")

	if _, err := e.dispatch.Cancel(e.ctx, r); err != nil {
		t.Fatal(err)
	}
	e.settle()
	e.expectRide(r, ride.Cancelled, "D001")
	e.expectDriver("D001", driver.Available, "")
}

func TestCancelMatchingRide(t *testing.T) {
	e := newEnv(t, nil)
	r := e.createRide()
	if _, err := e.dispatch.Cancel(e.ctx, r); err != nil {
		t.Fatal(err)
	}
	if _, err := e.dispatch.Cancel(e.ctx, r); err != nil {
		t.Fatalf("duplicate cancel error = %v", err)
	}
	e.settle()
	e.expectRide(r, ride.Cancelled, "")
	if n := e.log.count(r, events.RideCancelled); n != 1 {
		t.Fatalf("RideCancelled published %d times, want 1", n)
	}
}

func TestCompleteReleasesDriver(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()
	_, _ = e.dispatch.Accept(e.ctx, r, "D001")

	if _, err := e.dispatch.Complete(e.ctx, r); err != nil {
		t.Fatal(err)
	}
	if _, err := e.dispatch.Complete(e.ctx, r); err != nil {
		t.Fatalf("duplicate complete error = %v", err)
	}
	e.settle()
	e.expectRide(r, ride.Completed, "D001")
	e.expectDriver("D001", driver.Available, "")
	if _, err := e.dispatch.Cancel(e.ctx, r); !errors.Is(err, apperr.ErrInvalidTransition) {
		t.Fatalf("cancel completed ride error = %v", err)
	}
}

func TestCompleteBeforeAcceptIsRejected(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()
	if _, err := e.dispatch.Complete(e.ctx, r); !errors.Is(err, apperr.ErrInvalidTransition) {
		t.Fatalf("error = %v, want ErrInvalidTransition", err)
	}
	e.expectDriver("D001", driver.Offered, r)
}

func TestDuplicateRideCreatedEventCreatesOneOffer(t *testing.T) {
	e := newEnv(t, standardFleet)
	r := e.createRide()
	created, _ := e.log.last(r, events.RideCreated)

	// Redelivery through the bus is skipped by the inbox...
	e.bus.Publish(e.ctx, created)
	e.settle()
	// ...and a duplicate that bypasses the inbox is a no-op for the state machine.
	if err := e.dispatch.HandleEvent(e.ctx, created); err != nil {
		t.Fatal(err)
	}
	e.settle()

	e.expectRide(r, ride.DriverAssigned, "D001")
	e.expectDriver("D002", driver.Available, "")
	if n := e.log.count(r, events.OfferCreated); n != 1 {
		t.Fatalf("OfferCreated published %d times, want 1", n)
	}
}

func TestMatchSkipsRideCancelledBeforeDispatch(t *testing.T) {
	e := newEnv(t, standardFleet, envOptions{manual: true})
	r := e.createRide()
	if _, err := e.dispatch.Cancel(e.ctx, r); err != nil {
		t.Fatal(err)
	}
	if err := e.dispatch.Match(e.ctx, r); err != nil {
		t.Fatal(err)
	}
	e.expectRide(r, ride.Cancelled, "")
	e.expectDriver("D001", driver.Available, "")
}

// cancelAfterAssign cancels the ride right after the saga assigns it, i.e.
// between step 2 (assign) and step 3 (create offer) of the offer saga.
type cancelAfterAssign struct {
	dispatch.Rides
	cancel func(rideID string)
	once   sync.Once
}

func (c *cancelAfterAssign) Assign(ctx context.Context, id, driverID string) (ride.Ride, bool, error) {
	r, changed, err := c.Rides.Assign(ctx, id, driverID)
	if err == nil && changed {
		c.once.Do(func() { c.cancel(id) })
	}
	return r, changed, err
}

func TestCancelBetweenAssignAndOfferLeavesNoPendingOffer(t *testing.T) {
	var e *env
	e = newEnv(t, map[string]float64{"D001": 0.5}, envOptions{
		wrapRides: func(r dispatch.Rides) dispatch.Rides {
			return &cancelAfterAssign{Rides: r, cancel: func(id string) { _, _ = e.dispatch.Cancel(e.ctx, id) }}
		},
	})
	r := e.createRide()

	e.expectRide(r, ride.Cancelled, "D001")
	e.expectDriver("D001", driver.Available, "")
	if status := e.offerStatus(r, "D001"); status == dispatch.OfferPending {
		t.Fatal("a PENDING offer was left behind for a cancelled ride")
	}

	// D001 must be able to take the next ride straight away.
	next := e.createRide()
	e.expectRide(next, ride.DriverAssigned, "D001")
}
