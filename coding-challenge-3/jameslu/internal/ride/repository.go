package ride

import (
	"context"
	"fmt"
	"sync"

	"ridedispatch/internal/apperr"
)

// Repository stores rides. Update is the only way to change a stored ride:
// it loads the ride, runs fn (which calls transition methods) and saves the
// result atomically. If fn returns an error or changed=false, nothing is saved.
//
// A Postgres adapter implements Update as a conditional UPDATE whose WHERE
// clause carries the same guard (see docs/adr/ADR-002-concurrency.md).
type Repository interface {
	Create(ctx context.Context, r Ride) error
	Get(ctx context.Context, id string) (Ride, error)
	Update(ctx context.Context, id string, fn func(*Ride) (bool, error)) (Ride, bool, error)
}

// MemoryRepository holds rides by value, so callers never share state with
// the store. One mutex makes each Update an atomic read-check-write.
type MemoryRepository struct {
	mu    sync.Mutex
	rides map[string]Ride
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{rides: map[string]Ride{}}
}

func (m *MemoryRepository) Create(_ context.Context, r Ride) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.rides[r.id]; exists {
		return fmt.Errorf("%w: ride %s already exists", apperr.ErrConflict, r.id)
	}
	m.rides[r.id] = r
	return nil
}

func (m *MemoryRepository) Get(_ context.Context, id string) (Ride, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rides[id]
	if !ok {
		return Ride{}, fmt.Errorf("%w: ride %s", apperr.ErrNotFound, id)
	}
	return r, nil
}

func (m *MemoryRepository) Update(_ context.Context, id string, fn func(*Ride) (bool, error)) (Ride, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.rides[id]
	if !ok {
		return Ride{}, false, fmt.Errorf("%w: ride %s", apperr.ErrNotFound, id)
	}
	next := current
	changed, err := fn(&next)
	if err != nil {
		return current, false, err
	}
	if changed {
		m.rides[id] = next
	}
	return next, changed, nil
}
