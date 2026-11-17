package driver_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"ridedispatch/internal/apperr"
	"ridedispatch/internal/driver"
)

func TestMemoryRepositoryAddGetUpdate(t *testing.T) {
	ctx := context.Background()
	repo := driver.NewMemoryRepository()
	if err := repo.Add(ctx, driver.New("D001", here, t0)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Add(ctx, driver.New("D001", here, t0)); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("duplicate add error = %v, want ErrConflict", err)
	}
	if _, err := repo.Get(ctx, "missing"); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrNotFound", err)
	}

	d, changed, err := repo.Update(ctx, "D001", available)
	if err != nil || !changed || d.Status() != driver.Available {
		t.Fatalf("Update = %s, %v, %v", d.Status(), changed, err)
	}
	if _, _, err := repo.Update(ctx, "D001", markBusy("R001")); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("error = %v, want ErrConflict", err)
	}
	stored, _ := repo.Get(ctx, "D001")
	if stored.Status() != driver.Available {
		t.Fatal("failed update was persisted")
	}
}

func TestMemoryRepositoryListsAvailableDrivers(t *testing.T) {
	ctx := context.Background()
	repo := driver.NewMemoryRepository()
	for _, id := range []string{"D001", "D002", "D003"} {
		_ = repo.Add(ctx, driver.New(id, here, t0))
	}
	_, _, _ = repo.Update(ctx, "D001", available)
	_, _, _ = repo.Update(ctx, "D003", available)
	_, _, _ = repo.Update(ctx, "D003", claim("R001"))

	got := repo.Available(ctx)
	if len(got) != 1 || got[0].ID() != "D001" {
		t.Fatalf("Available = %v, want only D001", got)
	}
}

// The repository's atomic read-check-write is what AC-09 relies on: many
// rides racing to claim one driver must produce exactly one winner.
func TestMemoryRepositoryConcurrentClaimHasOneWinner(t *testing.T) {
	ctx := context.Background()
	repo := driver.NewMemoryRepository()
	_ = repo.Add(ctx, driver.New("D001", here, t0))
	_, _, _ = repo.Update(ctx, "D001", available)

	const contenders = 64
	var wins atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range contenders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, changed, err := repo.Update(ctx, "D001", claim(fmt.Sprintf("R%03d", i)))
			if changed {
				wins.Add(1)
			} else if !errors.Is(err, apperr.ErrConflict) {
				t.Errorf("loser error = %v, want ErrConflict", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if wins.Load() != 1 {
		t.Fatalf("%d contenders claimed the driver, want exactly 1", wins.Load())
	}
}
