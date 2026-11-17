package dispatch_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"ridedispatch/internal/dispatch"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/passenger"
	"ridedispatch/internal/platform/bus"
	"ridedispatch/internal/platform/clock"
	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/ride"
)

// env wires the three modules together with in-memory adapters, a fake clock
// and a real bus, the same way cmd/server does.
type env struct {
	t        *testing.T
	ctx      context.Context
	clock    *clock.Fake
	bus      *bus.Bus
	rides    *ride.Service
	drivers  *driver.Service
	offers   *dispatch.OfferStore
	dispatch *dispatch.Service
	log      *eventLog
	keys     int
}

type envOptions struct {
	cfg dispatch.Config
	// manual leaves the dispatch handler unsubscribed, so a test can call
	// Match itself and control the interleaving.
	manual bool
	// wrapRides lets a test intercept the Ride module calls made by Dispatch,
	// to inject an action between two saga steps.
	wrapRides func(dispatch.Rides) dispatch.Rides
}

// newEnv registers AVAILABLE drivers at the given distances (km due north of
// the pickup).
func newEnv(t *testing.T, distances map[string]float64, opts ...envOptions) *env {
	t.Helper()
	var o envOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	e := &env{t: t, ctx: context.Background(), clock: clock.NewFake(t0), log: &eventLog{}}
	e.bus = bus.New(bus.Options{Partitions: 4, MaxAttempts: 3, Backoff: time.Millisecond})
	e.rides = ride.NewService(ride.Deps{
		Repo:        ride.NewMemoryRepository(),
		Passengers:  passenger.NewDirectory("P001"),
		Idempotency: ride.NewIdempotencyStore(),
		Events:      e.bus,
		Clock:       e.clock,
	})
	e.drivers = driver.NewService(driver.NewMemoryRepository(), e.clock, nil)
	e.offers = dispatch.NewOfferStore()
	var rides dispatch.Rides = e.rides
	if o.wrapRides != nil {
		rides = o.wrapRides(rides)
	}
	e.dispatch = dispatch.NewService(dispatch.Deps{
		Rides:   rides,
		Drivers: e.drivers,
		Offers:  e.offers,
		Events:  e.bus,
		Clock:   e.clock,
		Config:  o.cfg,
	})
	if !o.manual {
		e.bus.Subscribe("dispatch", dispatch.SubscribedEvents, e.dispatch.HandleEvent)
	}
	e.bus.Subscribe("test-log", nil, e.log.handle)
	e.bus.Start()
	t.Cleanup(e.bus.Stop)

	for id, km := range distances {
		e.addDriver(id, km)
	}
	return e
}

func (e *env) addDriver(id string, km float64) {
	e.t.Helper()
	if err := e.drivers.Register(e.ctx, id, geo.OffsetNorthKm(pickup, km)); err != nil {
		e.t.Fatal(err)
	}
	if _, err := e.drivers.SetAvailability(e.ctx, id, driver.Available); err != nil {
		e.t.Fatal(err)
	}
}

// createRide creates a ride at the shared pickup and waits for dispatch.
func (e *env) createRide() string {
	e.t.Helper()
	e.keys++
	r, _, err := e.rides.Create(e.ctx, fmt.Sprintf("key-%d", e.keys), ride.CreateRequest{
		PassengerID: "P001", Pickup: pickup, Destination: dropoff,
	})
	if err != nil {
		e.t.Fatal(err)
	}
	e.settle()
	return r.ID()
}

func (e *env) settle() {
	e.t.Helper()
	ctx, cancel := context.WithTimeout(e.ctx, 5*time.Second)
	defer cancel()
	if err := e.bus.WaitIdle(ctx); err != nil {
		e.t.Fatalf("bus did not settle: %v", err)
	}
	if dead := e.bus.DeadLetters(); len(dead) > 0 {
		e.t.Fatalf("dead letters: %+v", dead)
	}
}

func (e *env) ride(id string) ride.Ride {
	e.t.Helper()
	r, err := e.rides.Get(e.ctx, id)
	if err != nil {
		e.t.Fatal(err)
	}
	return r
}

func (e *env) driver(id string) driver.Driver {
	e.t.Helper()
	d, err := e.drivers.Get(e.ctx, id)
	if err != nil {
		e.t.Fatal(err)
	}
	return d
}

func (e *env) offerStatus(rideID, driverID string) dispatch.OfferStatus {
	e.t.Helper()
	o, err := e.offers.Get(e.ctx, rideID, driverID)
	if err != nil {
		e.t.Fatal(err)
	}
	return o.Status()
}

func (e *env) expectRide(id string, status ride.Status, driverID string) {
	e.t.Helper()
	r := e.ride(id)
	if r.Status() != status || r.AssignedDriverID() != driverID {
		e.t.Fatalf("ride %s = %s/%q, want %s/%q", id, r.Status(), r.AssignedDriverID(), status, driverID)
	}
}

func (e *env) expectDriver(id string, status driver.Status, rideID string) {
	e.t.Helper()
	d := e.driver(id)
	if d.Status() != status || d.CurrentRideID() != rideID {
		e.t.Fatalf("driver %s = %s/%q, want %s/%q", id, d.Status(), d.CurrentRideID(), status, rideID)
	}
}

type eventLog struct {
	mu     sync.Mutex
	events []events.Event
}

func (l *eventLog) handle(_ context.Context, ev events.Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, ev)
	return nil
}

func (l *eventLog) count(rideID, typ string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := 0
	for _, ev := range l.events {
		if ev.RideID == rideID && ev.Type == typ {
			n++
		}
	}
	return n
}

func (l *eventLog) last(rideID, typ string) (events.Event, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.events) - 1; i >= 0; i-- {
		if ev := l.events[i]; ev.RideID == rideID && ev.Type == typ {
			return ev, true
		}
	}
	return events.Event{}, false
}
