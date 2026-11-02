package ride_test

import (
	"context"
	"errors"
	"testing"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/ride"
)

func TestMemoryRepositoryCreateAndGet(t *testing.T) {
	ctx := context.Background()
	repo := ride.NewMemoryRepository()
	r := ride.New("R001", "P001", pickup, dropoff, t0)

	if err := repo.Create(ctx, r); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, r); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("duplicate create error = %v, want ErrConflict", err)
	}
	got, err := repo.Get(ctx, "R001")
	if err != nil || got.ID() != "R001" || got.Status() != ride.Requested {
		t.Fatalf("Get = %+v, %v", got, err)
	}
	if _, err := repo.Get(ctx, "missing"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}
}

func TestMemoryRepositoryUpdate(t *testing.T) {
	ctx := context.Background()
	repo := ride.NewMemoryRepository()
	_ = repo.Create(ctx, ride.New("R001", "P001", pickup, dropoff, t0))

	updated, changed, err := repo.Update(ctx, "R001", start)
	if err != nil || !changed || updated.Status() != ride.Matching {
		t.Fatalf("Update = %s, %v, %v", updated.Status(), changed, err)
	}

	// A failed transition is not persisted.
	_, _, err = repo.Update(ctx, "R001", complete)
	if !errors.Is(err, apperr.ErrInvalidTransition) {
		t.Fatalf("error = %v, want ErrInvalidTransition", err)
	}
	stored, _ := repo.Get(ctx, "R001")
	if stored.Status() != ride.Matching || stored.Version() != updated.Version() {
		t.Fatalf("failed update leaked: %s v%d", stored.Status(), stored.Version())
	}

	// Mutating a returned copy does not change the store.
	_, _ = stored.Cancel(t0)
	again, _ := repo.Get(ctx, "R001")
	if again.Status() != ride.Matching {
		t.Fatal("store shares state with returned values")
	}

	if _, _, err := repo.Update(ctx, "missing", start); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("Update(missing) error = %v, want ErrNotFound", err)
	}
}
