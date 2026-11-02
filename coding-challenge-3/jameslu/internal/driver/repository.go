package driver

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"ridedispatch/internal/apperr"
)

// Repository stores drivers. Update is an atomic read-check-write: it loads
// the driver, runs fn (which calls transition methods) and saves the result.
// If fn returns an error or changed=false, nothing is saved.
//
// A Postgres adapter implements Update as a conditional UPDATE, e.g. a claim is
// "UPDATE drivers SET status='OFFERED' ... WHERE id=$1 AND status='AVAILABLE'"
// (see docs/adr/ADR-002-concurrency.md).
type Repository interface {
	Add(ctx context.Context, d Driver) error
	Get(ctx context.Context, id string) (Driver, error)
	Update(ctx context.Context, id string, fn func(*Driver) (bool, error)) (Driver, bool, error)
	// Available returns a snapshot of AVAILABLE drivers ordered by id. The
	// snapshot may be stale by the time it is used; claims re-check status.
	Available(ctx context.Context) []Driver
}

// MemoryRepository holds drivers by value behind one mutex.
type MemoryRepository struct {
	mu      sync.Mutex
	drivers map[string]Driver
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{drivers: map[string]Driver{}}
}

func (m *MemoryRepository) Add(_ context.Context, d Driver) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.drivers[d.id]; exists {
		return fmt.Errorf("%w: driver %s already exists", apperr.ErrConflict, d.id)
	}
	m.drivers[d.id] = d
	return nil
}

func (m *MemoryRepository) Get(_ context.Context, id string) (Driver, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.drivers[id]
	if !ok {
		return Driver{}, fmt.Errorf("%w: driver %s", apperr.ErrNotFound, id)
	}
	return d, nil
}

func (m *MemoryRepository) Update(_ context.Context, id string, fn func(*Driver) (bool, error)) (Driver, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.drivers[id]
	if !ok {
		return Driver{}, false, fmt.Errorf("%w: driver %s", apperr.ErrNotFound, id)
	}
	next := current
	changed, err := fn(&next)
	if err != nil {
		return current, false, err
	}
	if changed {
		m.drivers[id] = next
	}
	return next, changed, nil
}

func (m *MemoryRepository) Available(_ context.Context) []Driver {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Driver
	for _, d := range m.drivers {
		if d.status == Available {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].id < out[j].id })
	return out
}
