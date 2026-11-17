package ride_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/passenger"
	"ridedispatch/internal/platform/clock"
	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/ride"
)

type publisher struct {
	mu     sync.Mutex
	events []events.Event
}

func (p *publisher) Publish(_ context.Context, evs ...events.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, evs...)
}

func (p *publisher) count(typ string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for _, e := range p.events {
		if e.Type == typ {
			n++
		}
	}
	return n
}

func newService() (*ride.Service, *publisher) {
	pub := &publisher{}
	svc := ride.NewService(ride.Deps{
		Repo:        ride.NewMemoryRepository(),
		Passengers:  passenger.NewDirectory("P001", "P002"),
		Idempotency: ride.NewIdempotencyStore(),
		Events:      pub,
		Clock:       clock.NewFake(t0),
	})
	return svc, pub
}

func validRequest() ride.CreateRequest {
	return ride.CreateRequest{PassengerID: "P001", Pickup: pickup, Destination: dropoff}
}

func TestCreateRide(t *testing.T) {
	svc, pub := newService()
	ctx := context.Background()

	r, replayed, err := svc.Create(ctx, "K1", validRequest())
	if err != nil || replayed {
		t.Fatalf("Create = replayed %v, err %v", replayed, err)
	}
	if r.ID() == "" || r.Status() != ride.Requested || r.PassengerID() != "P001" {
		t.Fatalf("unexpected ride %s %s %s", r.ID(), r.Status(), r.PassengerID())
	}
	got, err := svc.Get(ctx, r.ID())
	if err != nil || got.ID() != r.ID() {
		t.Fatalf("created ride cannot be retrieved: %v", err)
	}
	if pub.count(events.RideCreated) != 1 {
		t.Fatalf("RideCreated published %d times", pub.count(events.RideCreated))
	}
	pub.mu.Lock()
	e := pub.events[0]
	pub.mu.Unlock()
	if e.RideID != r.ID() || e.ID == "" {
		t.Fatalf("RideCreated envelope = %+v", e)
	}
}

func TestCreateRideGeneratesUniqueIDs(t *testing.T) {
	svc, _ := newService()
	a, _, _ := svc.Create(context.Background(), "K1", validRequest())
	b, _, _ := svc.Create(context.Background(), "K2", validRequest())
	if a.ID() == b.ID() {
		t.Fatalf("both rides got id %s", a.ID())
	}
}

func TestCreateRideValidation(t *testing.T) {
	cases := []struct {
		name    string
		key     string
		mutate  func(*ride.CreateRequest)
		wantErr error
	}{
		{"missing idempotency key", "", func(*ride.CreateRequest) {}, apperr.ErrIdempotencyKeyMissing},
		{"missing passenger", "K1", func(r *ride.CreateRequest) { r.PassengerID = "" }, apperr.ErrValidation},
		{"unknown passenger", "K1", func(r *ride.CreateRequest) { r.PassengerID = "P999" }, apperr.ErrNotFound},
		{"invalid pickup latitude", "K1", func(r *ride.CreateRequest) { r.Pickup = geo.Location{Latitude: 91, Longitude: 0} }, apperr.ErrValidation},
		{"invalid destination longitude", "K1", func(r *ride.CreateRequest) { r.Destination = geo.Location{Latitude: 0, Longitude: -181} }, apperr.ErrValidation},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc, pub := newService()
			req := validRequest()
			c.mutate(&req)
			if _, _, err := svc.Create(context.Background(), c.key, req); !errors.Is(err, c.wantErr) {
				t.Fatalf("error = %v, want %v", err, c.wantErr)
			}
			if pub.count(events.RideCreated) != 0 {
				t.Fatal("rejected request must not create a ride")
			}
		})
	}
}

func TestCreateRideIsIdempotent(t *testing.T) {
	svc, pub := newService()
	ctx := context.Background()

	first, _, _ := svc.Create(ctx, "K1", validRequest())
	again, replayed, err := svc.Create(ctx, "K1", validRequest())
	if err != nil || !replayed || again.ID() != first.ID() || again.Status() != ride.Requested {
		t.Fatalf("replay = %s replayed=%v err=%v", again.ID(), replayed, err)
	}
	if pub.count(events.RideCreated) != 1 {
		t.Fatalf("replay created another ride (%d RideCreated)", pub.count(events.RideCreated))
	}

	other := validRequest()
	other.PassengerID = "P002"
	if _, _, err := svc.Create(ctx, "K1", other); !errors.Is(err, apperr.ErrIdempotencyKeyReused) {
		t.Fatalf("reused key error = %v", err)
	}
}

func TestFailedCreateDoesNotBurnTheKey(t *testing.T) {
	svc, _ := newService()
	bad := validRequest()
	bad.PassengerID = "P999"
	_, _, _ = svc.Create(context.Background(), "K1", bad)

	if _, replayed, err := svc.Create(context.Background(), "K1", validRequest()); err != nil || replayed {
		t.Fatalf("retry after failure = replayed %v, err %v", replayed, err)
	}
}

func TestConcurrentCreateWithSameKeyCreatesOneRide(t *testing.T) {
	for range 200 {
		svc, pub := newService()
		const clients = 20
		var created atomic.Int32
		var wg sync.WaitGroup
		start := make(chan struct{})
		for range clients {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, replayed, err := svc.Create(context.Background(), "K1", validRequest())
				switch {
				case err == nil && !replayed:
					created.Add(1)
				case err == nil && replayed, errors.Is(err, apperr.ErrIdempotencyInProgress):
				default:
					t.Errorf("unexpected result: replayed=%v err=%v", replayed, err)
				}
			}()
		}
		close(start)
		wg.Wait()
		if created.Load() != 1 || pub.count(events.RideCreated) != 1 {
			t.Fatalf("created=%d RideCreated=%d, want 1", created.Load(), pub.count(events.RideCreated))
		}
	}
}

func TestRideServiceTransitionsPersist(t *testing.T) {
	svc, _ := newService()
	ctx := context.Background()
	r, _, _ := svc.Create(ctx, "K1", validRequest())

	steps := []struct {
		name string
		do   func() (ride.Ride, bool, error)
		want ride.Status
	}{
		{"start", func() (ride.Ride, bool, error) { return svc.StartMatching(ctx, r.ID()) }, ride.Matching},
		{"assign", func() (ride.Ride, bool, error) { return svc.Assign(ctx, r.ID(), "D001") }, ride.DriverAssigned},
		{"release", func() (ride.Ride, bool, error) { return svc.ReleaseAssignment(ctx, r.ID(), "D001") }, ride.Matching},
		{"assign again", func() (ride.Ride, bool, error) { return svc.Assign(ctx, r.ID(), "D002") }, ride.DriverAssigned},
		{"accept", func() (ride.Ride, bool, error) { return svc.Accept(ctx, r.ID(), "D002") }, ride.DriverAccepted},
		{"complete", func() (ride.Ride, bool, error) { return svc.Complete(ctx, r.ID()) }, ride.Completed},
	}
	for _, s := range steps {
		got, changed, err := s.do()
		if err != nil || !changed || got.Status() != s.want {
			t.Fatalf("%s = %s changed=%v err=%v", s.name, got.Status(), changed, err)
		}
		stored, _ := svc.Get(ctx, r.ID())
		if stored.Status() != s.want {
			t.Fatalf("%s not persisted: stored %s", s.name, stored.Status())
		}
	}
	if _, _, err := svc.Cancel(ctx, r.ID()); !errors.Is(err, apperr.ErrInvalidTransition) {
		t.Fatalf("cancel completed ride error = %v", err)
	}
}
