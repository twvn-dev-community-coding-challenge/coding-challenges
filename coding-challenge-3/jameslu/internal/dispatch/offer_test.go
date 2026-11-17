package dispatch_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/dispatch"
)

func offer(rideID, driverID string) dispatch.Offer {
	return dispatch.NewOffer("O-"+rideID+"-"+driverID, rideID, driverID, t0, t0.Add(30*time.Second))
}

func TestOfferStoreEnforcesInvariants(t *testing.T) {
	ctx := context.Background()
	s := dispatch.NewOfferStore()
	if err := s.Create(ctx, offer("R1", "D1")); err != nil {
		t.Fatal(err)
	}

	cases := map[string]dispatch.Offer{
		"second pending offer for the driver": offer("R2", "D1"),
		"second pending offer for the ride":   offer("R1", "D2"),
	}
	for name, o := range cases {
		if err := s.Create(ctx, o); !errors.Is(err, apperr.ErrConflict) {
			t.Fatalf("%s: error = %v, want ErrConflict", name, err)
		}
	}

	// Once resolved, the driver and the ride can take new offers, but the same
	// (ride, driver) pair never gets a second offer.
	if _, _, err := s.Resolve(ctx, "R1", "D1", dispatch.OfferRejected, t0); err != nil {
		t.Fatal(err)
	}
	if err := s.Create(ctx, offer("R2", "D1")); err != nil {
		t.Fatalf("driver blocked after its offer was resolved: %v", err)
	}
	if err := s.Create(ctx, offer("R1", "D2")); err != nil {
		t.Fatalf("ride blocked after its offer was resolved: %v", err)
	}
	if err := s.Create(ctx, offer("R1", "D1")); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("repeated (ride, driver) offer error = %v, want ErrConflict", err)
	}
}

func TestOfferResolution(t *testing.T) {
	ctx := context.Background()
	s := dispatch.NewOfferStore()
	_ = s.Create(ctx, offer("R1", "D1"))

	o, changed, err := s.Resolve(ctx, "R1", "D1", dispatch.OfferAccepted, t0)
	if err != nil || !changed || o.Status() != dispatch.OfferAccepted {
		t.Fatalf("accept = %s changed=%v err=%v", o.Status(), changed, err)
	}
	if _, changed, err := s.Resolve(ctx, "R1", "D1", dispatch.OfferAccepted, t0); err != nil || changed {
		t.Fatalf("duplicate accept: changed=%v err=%v", changed, err)
	}
	if _, _, err := s.Resolve(ctx, "R1", "D1", dispatch.OfferExpired, t0); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("expire accepted offer error = %v, want ErrConflict", err)
	}
	if _, _, err := s.Resolve(ctx, "R1", "D9", dispatch.OfferAccepted, t0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("unknown offer error = %v, want ErrNotFound", err)
	}
}

func TestExclusionIsDerivedFromRejectedAndExpiredOffers(t *testing.T) {
	ctx := context.Background()
	s := dispatch.NewOfferStore()
	for _, step := range []struct {
		driver string
		final  dispatch.OfferStatus
	}{{"D1", dispatch.OfferRejected}, {"D2", dispatch.OfferExpired}, {"D3", dispatch.OfferCancelled}} {
		_ = s.Create(ctx, offer("R1", step.driver))
		_, _, _ = s.Resolve(ctx, "R1", step.driver, step.final, t0)
	}
	_ = s.Create(ctx, offer("R2", "D4"))
	_, _, _ = s.Resolve(ctx, "R2", "D4", dispatch.OfferRejected, t0)

	got := s.Excluded(ctx, "R1")
	if len(got) != 2 || !got["D1"] || !got["D2"] {
		t.Fatalf("Excluded(R1) = %v, want D1 and D2", got)
	}
}
