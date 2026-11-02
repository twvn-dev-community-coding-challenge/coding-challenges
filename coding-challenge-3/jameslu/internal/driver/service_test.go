package driver_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/driver"
	"ridedispatch/internal/geo"
	"ridedispatch/internal/platform/clock"
)

func newService(t *testing.T, ids ...string) (*driver.Service, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(t0)
	svc := driver.NewService(driver.NewMemoryRepository(), clk, nil)
	for _, id := range ids {
		if err := svc.Register(context.Background(), id, here); err != nil {
			t.Fatal(err)
		}
	}
	return svc, clk
}

func TestServiceAvailability(t *testing.T) {
	svc, _ := newService(t, "D001")
	ctx := context.Background()

	d, err := svc.SetAvailability(ctx, "D001", driver.Available)
	if err != nil || d.Status() != driver.Available {
		t.Fatalf("go available = %s, %v", d.Status(), err)
	}
	if _, err := svc.Claim(ctx, "D001", "R001"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SetAvailability(ctx, "D001", driver.Offline); !errors.Is(err, apperr.ErrInvalidTransition) {
		t.Fatalf("offered driver going offline error = %v", err)
	}
	if _, err := svc.SetAvailability(ctx, "missing", driver.Available); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("unknown driver error = %v", err)
	}
}

func TestServiceUpdateLocation(t *testing.T) {
	svc, clk := newService(t, "D001")
	ctx := context.Background()
	clk.Advance(time.Minute)
	moved := geo.OffsetNorthKm(here, 1)

	d, err := svc.UpdateLocation(ctx, "D001", moved)
	if err != nil || d.Location() != moved || !d.LocationUpdatedAt().Equal(t0.Add(time.Minute)) {
		t.Fatalf("UpdateLocation = %+v, %v", d.Location(), err)
	}
	if _, err := svc.UpdateLocation(ctx, "D001", geo.Location{Latitude: 100}); !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("invalid location error = %v", err)
	}
	stored, _ := svc.Get(ctx, "D001")
	if stored.Location() != moved {
		t.Fatal("invalid update overwrote the latest location")
	}
}

func TestServiceOfferLifecycle(t *testing.T) {
	svc, _ := newService(t, "D001")
	ctx := context.Background()
	_, _ = svc.SetAvailability(ctx, "D001", driver.Available)

	mustChange := func(name string, changed bool, err error) {
		t.Helper()
		if err != nil || !changed {
			t.Fatalf("%s: changed=%v err=%v", name, changed, err)
		}
	}
	changed, err := svc.Claim(ctx, "D001", "R001")
	mustChange("claim", changed, err)
	changed, err = svc.MarkBusy(ctx, "D001", "R001")
	mustChange("mark busy", changed, err)
	changed, err = svc.Release(ctx, "D001", "R001")
	mustChange("release", changed, err)

	d, _ := svc.Get(ctx, "D001")
	if d.Status() != driver.Available || d.CurrentRideID() != "" {
		t.Fatalf("after release: %s %q", d.Status(), d.CurrentRideID())
	}
}

func TestServiceAvailableNearListsOnlyAvailableDrivers(t *testing.T) {
	svc, _ := newService(t, "D001", "D002", "D003")
	ctx := context.Background()
	_, _ = svc.SetAvailability(ctx, "D001", driver.Available)
	_, _ = svc.SetAvailability(ctx, "D002", driver.Available)
	_, _ = svc.Claim(ctx, "D002", "R001")

	got := svc.AvailableNear(ctx, here, 2)
	if len(got) != 1 || got[0].ID != "D001" || got[0].Location != here {
		t.Fatalf("AvailableNear = %+v, want only D001", got)
	}
}
