package ride_test

import (
	"errors"
	"testing"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/ride"
)

func TestIdempotencyStoreLifecycle(t *testing.T) {
	s := ride.NewIdempotencyStore()
	r := ride.New("R001", "P001", pickup, dropoff, t0)

	if _, replay, err := s.Reserve("K1", "hash-a"); err != nil || replay {
		t.Fatalf("first reserve = replay %v, err %v", replay, err)
	}
	if _, _, err := s.Reserve("K1", "hash-a"); !errors.Is(err, apperr.ErrIdempotencyInProgress) {
		t.Fatalf("reserve while in progress error = %v", err)
	}
	if _, _, err := s.Reserve("K1", "hash-b"); !errors.Is(err, apperr.ErrIdempotencyKeyReused) {
		t.Fatalf("reserve with other hash error = %v", err)
	}

	s.Complete("K1", r)

	got, replay, err := s.Reserve("K1", "hash-a")
	if err != nil || !replay || got.ID() != "R001" {
		t.Fatalf("replay = %s, %v, %v", got.ID(), replay, err)
	}
	if _, _, err := s.Reserve("K1", "hash-b"); !errors.Is(err, apperr.ErrIdempotencyKeyReused) {
		t.Fatalf("reuse after completion error = %v", err)
	}
}

func TestIdempotencyStoreAbandonFreesKey(t *testing.T) {
	s := ride.NewIdempotencyStore()
	_, _, _ = s.Reserve("K1", "hash-a")
	s.Abandon("K1")
	if _, replay, err := s.Reserve("K1", "hash-b"); err != nil || replay {
		t.Fatalf("reserve after abandon = replay %v, err %v", replay, err)
	}
}

func TestRequestHashIsStableAndSensitive(t *testing.T) {
	a := ride.CreateRequest{PassengerID: "P001", Pickup: pickup, Destination: dropoff}
	b := a
	if ride.RequestHash(a) != ride.RequestHash(b) {
		t.Fatal("equal requests must hash equally")
	}
	b.Destination.Latitude += 0.0001
	if ride.RequestHash(a) == ride.RequestHash(b) {
		t.Fatal("different requests must hash differently")
	}
}
