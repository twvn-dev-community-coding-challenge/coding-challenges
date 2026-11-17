// Package timeline records every event per ride, so "what happened to ride
// R001?" can be answered with one API call as well as from the logs.
package timeline

import (
	"context"
	"slices"
	"sync"

	"ridedispatch/internal/platform/events"
)

type Store struct {
	mu     sync.Mutex
	byRide map[string][]events.Event
}

func NewStore() *Store {
	return &Store{byRide: map[string][]events.Event{}}
}

// Append is a bus handler subscribed to every event type.
func (s *Store) Append(_ context.Context, e events.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byRide[e.RideID] = append(s.byRide[e.RideID], e)
	return nil
}

// ForRide returns the ride's events in the order they were delivered.
func (s *Store) ForRide(rideID string) []events.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.byRide[rideID])
}
