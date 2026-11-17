package dispatch_test

import (
	"sync"
	"testing"

	"ridedispatch/internal/dispatch"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/ride"
)

// Each test repeats a race many times on a fresh system. Goroutines are
// released together by a barrier so the operations genuinely overlap; run
// with `go test -race` to also check for data races.
const iterations = 1000

func race(fns ...func()) {
	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, fn := range fns {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			fn()
		}()
	}
	close(start)
	wg.Wait()
}

// AC-09: two rides compete for the only nearby driver.
func TestConcurrentRidesCannotShareTheSameDriver(t *testing.T) {
	wins := map[string]int{}
	defer func() { t.Logf("first ride won %d times, second ride won %d times", wins["first"], wins["second"]) }()
	for range iterations {
		e := newEnv(t, map[string]float64{"D001": 0.5}, envOptions{manual: true})
		r1, r2 := e.createRide(), e.createRide()

		race(
			func() { _ = e.dispatch.Match(e.ctx, r1) },
			func() { _ = e.dispatch.Match(e.ctx, r2) },
		)

		winner, loser := r1, r2
		wins["first"]++
		if e.ride(r2).AssignedDriverID() == "D001" {
			winner, loser = r2, r1
			wins["first"]--
			wins["second"]++
		}
		e.expectRide(winner, ride.DriverAssigned, "D001")
		e.expectRide(loser, ride.Matching, "")
		e.expectDriver("D001", driver.Offered, winner)
		if e.offerStatus(winner, "D001") != dispatch.OfferPending {
			t.Fatal("winner has no pending offer")
		}
		if _, err := e.offers.Get(e.ctx, loser, "D001"); err == nil {
			t.Fatal("loser also created an offer for D001")
		}
	}
}

// With a second driver in range the loser falls through to it.
func TestConcurrentRidesGetDifferentDrivers(t *testing.T) {
	for range iterations {
		e := newEnv(t, map[string]float64{"D001": 0.5, "D002": 1.0}, envOptions{manual: true})
		r1, r2 := e.createRide(), e.createRide()

		race(
			func() { _ = e.dispatch.Match(e.ctx, r1) },
			func() { _ = e.dispatch.Match(e.ctx, r2) },
		)

		a, b := e.ride(r1).AssignedDriverID(), e.ride(r2).AssignedDriverID()
		if a == "" || b == "" || a == b {
			t.Fatalf("rides assigned to %q and %q, want two different drivers", a, b)
		}
		e.expectDriver(a, driver.Offered, r1)
		e.expectDriver(b, driver.Offered, r2)
	}
}

// The same ride matched twice at once (duplicate RideCreated on two workers):
// one offer, and the extra claim is compensated.
func TestConcurrentDuplicateMatchingCreatesOneOffer(t *testing.T) {
	for range iterations {
		e := newEnv(t, map[string]float64{"D001": 0.5, "D002": 1.0}, envOptions{manual: true})
		r := e.createRide()

		race(
			func() { _ = e.dispatch.Match(e.ctx, r) },
			func() { _ = e.dispatch.Match(e.ctx, r) },
		)

		got := e.ride(r)
		if got.Status() != ride.DriverAssigned {
			t.Fatalf("ride = %s, want DRIVER_ASSIGNED", got.Status())
		}
		holder := got.AssignedDriverID()
		for _, id := range []string{"D001", "D002"} {
			if id == holder {
				e.expectDriver(id, driver.Offered, r)
			} else {
				e.expectDriver(id, driver.Available, "")
			}
		}
	}
}

// Accept and timeout at the same instant: exactly one outcome, never a mix.
func TestConcurrentAcceptAndTimeout(t *testing.T) {
	outcomes := map[dispatch.OfferStatus]int{}
	defer func() { t.Logf("outcomes: %v", outcomes) }()
	for range iterations {
		e := newEnv(t, map[string]float64{"D001": 0.5}, envOptions{manual: true})
		r := e.createRide()
		_ = e.dispatch.Match(e.ctx, r)

		race(
			func() { _, _ = e.dispatch.Accept(e.ctx, r, "D001") },
			func() { _ = e.dispatch.Expire(e.ctx, r, "D001") },
		)

		got, d := e.ride(r), e.driver("D001")
		status := e.offerStatus(r, "D001")
		outcomes[status]++
		switch status {
		case dispatch.OfferAccepted:
			if got.Status() != ride.DriverAccepted || d.Status() != driver.Busy {
				t.Fatalf("accepted offer but ride=%s driver=%s", got.Status(), d.Status())
			}
		case dispatch.OfferExpired:
			if got.Status() != ride.Matching || got.AssignedDriverID() != "" || d.Status() != driver.Available {
				t.Fatalf("expired offer but ride=%s/%q driver=%s", got.Status(), got.AssignedDriverID(), d.Status())
			}
		default:
			t.Fatalf("offer ended %s", status)
		}
	}
}

// Accept and cancel at the same instant.
func TestConcurrentAcceptAndCancel(t *testing.T) {
	for range iterations {
		e := newEnv(t, map[string]float64{"D001": 0.5}, envOptions{manual: true})
		r := e.createRide()
		_ = e.dispatch.Match(e.ctx, r)

		race(
			func() { _, _ = e.dispatch.Accept(e.ctx, r, "D001") },
			func() { _, _ = e.dispatch.Cancel(e.ctx, r) },
		)

		if got := e.ride(r); got.Status() != ride.Cancelled {
			t.Fatalf("ride = %s, want CANCELLED (cancel is allowed from both states)", got.Status())
		}
		e.expectDriver("D001", driver.Available, "")
	}
}

// Reject and timeout at the same instant: the offer resolves once.
func TestConcurrentRejectAndTimeout(t *testing.T) {
	outcomes := map[dispatch.OfferStatus]int{}
	defer func() { t.Logf("outcomes: %v", outcomes) }()
	for range iterations {
		e := newEnv(t, map[string]float64{"D001": 0.5}, envOptions{manual: true})
		r := e.createRide()
		_ = e.dispatch.Match(e.ctx, r)

		race(
			func() { _, _ = e.dispatch.Reject(e.ctx, r, "D001") },
			func() { _ = e.dispatch.Expire(e.ctx, r, "D001") },
		)

		status := e.offerStatus(r, "D001")
		outcomes[status]++
		if status != dispatch.OfferRejected && status != dispatch.OfferExpired {
			t.Fatalf("offer ended %s", status)
		}
		e.expectRide(r, ride.Matching, "")
		e.expectDriver("D001", driver.Available, "")
	}
}
