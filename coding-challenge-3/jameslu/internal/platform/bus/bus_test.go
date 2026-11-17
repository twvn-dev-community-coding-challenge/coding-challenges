package bus_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"ridedispatch/internal/platform/bus"
	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/platform/trace"
)

func newBus(t *testing.T, partitions int) *bus.Bus {
	t.Helper()
	b := bus.New(bus.Options{Partitions: partitions, MaxAttempts: 3, Backoff: time.Millisecond})
	b.Start()
	t.Cleanup(b.Stop)
	return b
}

func waitIdle(t *testing.T, b *bus.Bus) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.WaitIdle(ctx); err != nil {
		t.Fatalf("bus did not become idle: %v", err)
	}
}

type recorder struct {
	mu   sync.Mutex
	seen []events.Event
}

func (r *recorder) handle(_ context.Context, e events.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seen = append(r.seen, e)
	return nil
}

func (r *recorder) events() []events.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]events.Event(nil), r.seen...)
}

func ev(id, typ, rideID string) events.Event {
	return events.Event{ID: id, Type: typ, RideID: rideID}
}

func TestDeliversToMatchingSubscribers(t *testing.T) {
	b := newBus(t, 4)
	created, all := &recorder{}, &recorder{}
	b.Subscribe("dispatch", []string{events.RideCreated}, created.handle)
	b.Subscribe("timeline", nil, all.handle)

	b.Publish(context.Background(), ev("E1", events.RideCreated, "R001"), ev("E2", events.OfferCreated, "R001"))
	waitIdle(t, b)

	if got := created.events(); len(got) != 1 || got[0].ID != "E1" {
		t.Fatalf("typed subscriber got %v", got)
	}
	if got := all.events(); len(got) != 2 {
		t.Fatalf("wildcard subscriber got %v", got)
	}
}

func TestPreservesOrderPerRide(t *testing.T) {
	b := newBus(t, 4)
	rec := &recorder{}
	b.Subscribe("c", nil, rec.handle)

	for i := range 200 {
		b.Publish(context.Background(), ev(fmt.Sprintf("E%03d", i), events.OfferCreated, "R001"))
	}
	waitIdle(t, b)

	got := rec.events()
	if len(got) != 200 {
		t.Fatalf("got %d events, want 200", len(got))
	}
	for i, e := range got {
		if e.ID != fmt.Sprintf("E%03d", i) {
			t.Fatalf("event %d is %s: order not preserved", i, e.ID)
		}
	}
}

func TestDifferentRidesAreProcessedInParallel(t *testing.T) {
	b := newBus(t, 2)
	r1, r2 := "R001", ""
	for i := 2; r2 == ""; i++ {
		if id := fmt.Sprintf("R%03d", i); b.PartitionFor(id) != b.PartitionFor(r1) {
			r2 = id
		}
	}

	r2Done := make(chan struct{})
	b.Subscribe("c", nil, func(_ context.Context, e events.Event) error {
		if e.RideID == r1 {
			// Blocks until R2 is handled: only possible if partitions run in parallel.
			select {
			case <-r2Done:
				return nil
			case <-time.After(2 * time.Second):
				return errors.New("R2 was not processed while R1 was in progress")
			}
		}
		close(r2Done)
		return nil
	})

	b.Publish(context.Background(), ev("E1", events.RideCreated, r1))
	b.Publish(context.Background(), ev("E2", events.RideCreated, r2))
	waitIdle(t, b)
	if len(b.DeadLetters()) != 0 {
		t.Fatalf("dead letters: %+v", b.DeadLetters())
	}
}

func TestRetriesFailedHandlerUntilSuccess(t *testing.T) {
	b := newBus(t, 1)
	attempts := 0
	b.Subscribe("c", nil, func(context.Context, events.Event) error {
		attempts++
		if attempts < 3 {
			return errors.New("transient")
		}
		return nil
	})

	b.Publish(context.Background(), ev("E1", events.RideCreated, "R001"))
	waitIdle(t, b)

	if attempts != 3 || len(b.DeadLetters()) != 0 {
		t.Fatalf("attempts = %d, dead letters = %d", attempts, len(b.DeadLetters()))
	}
}

func TestDeadLettersAfterMaxAttempts(t *testing.T) {
	b := newBus(t, 1)
	attempts := 0
	b.Subscribe("dispatch", nil, func(context.Context, events.Event) error {
		attempts++
		return errors.New("permanent")
	})

	b.Publish(context.Background(), ev("E1", events.RideCreated, "R001"))
	waitIdle(t, b)

	dead := b.DeadLetters()
	if attempts != 3 || len(dead) != 1 {
		t.Fatalf("attempts = %d, dead letters = %d", attempts, len(dead))
	}
	if dead[0].Event.ID != "E1" || dead[0].Consumer != "dispatch" || dead[0].Err == "" {
		t.Fatalf("dead letter = %+v", dead[0])
	}
}

func TestRecoversFromPanickingHandler(t *testing.T) {
	b := newBus(t, 1)
	attempts := 0
	b.Subscribe("c", nil, func(context.Context, events.Event) error {
		attempts++
		if attempts == 1 {
			panic("boom")
		}
		return nil
	})

	b.Publish(context.Background(), ev("E1", events.RideCreated, "R001"))
	waitIdle(t, b)

	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2 (panic then success)", attempts)
	}
}

func TestSkipsDuplicateEventIDPerConsumer(t *testing.T) {
	b := newBus(t, 2)
	first, second := &recorder{}, &recorder{}
	b.Subscribe("first", nil, first.handle)
	b.Subscribe("second", nil, second.handle)

	e := ev("E1", events.RideCreated, "R001")
	b.Publish(context.Background(), e)
	b.Publish(context.Background(), e)
	b.Publish(context.Background(), e)
	waitIdle(t, b)

	if len(first.events()) != 1 || len(second.events()) != 1 {
		t.Fatalf("deliveries: first=%d second=%d, want 1 each", len(first.events()), len(second.events()))
	}
}

func TestHandlerContextCarriesCausation(t *testing.T) {
	b := newBus(t, 1)
	var causation string
	b.Subscribe("c", nil, func(ctx context.Context, e events.Event) error {
		causation = trace.CausationID(ctx)
		return nil
	})
	b.Publish(context.Background(), ev("E1", events.RideCreated, "R001"))
	waitIdle(t, b)
	if causation != "E1" {
		t.Fatalf("causation = %q, want E1", causation)
	}
}
