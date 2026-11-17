package timeline_test

import (
	"context"
	"testing"

	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/timeline"
)

func TestTimelineKeepsEventsPerRideInOrder(t *testing.T) {
	s := timeline.NewStore()
	ctx := context.Background()
	for _, e := range []events.Event{
		{ID: "E1", Type: events.RideCreated, RideID: "R1"},
		{ID: "E2", Type: events.RideCreated, RideID: "R2"},
		{ID: "E3", Type: events.OfferCreated, RideID: "R1"},
	} {
		if err := s.Append(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	got := s.ForRide("R1")
	if len(got) != 2 || got[0].ID != "E1" || got[1].ID != "E3" {
		t.Fatalf("ForRide(R1) = %+v", got)
	}
	got[0].ID = "mutated"
	if s.ForRide("R1")[0].ID != "E1" {
		t.Fatal("ForRide must return a copy")
	}
	if len(s.ForRide("unknown")) != 0 {
		t.Fatal("unknown ride should have an empty timeline")
	}
}
