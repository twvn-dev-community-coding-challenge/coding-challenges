package dispatch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ridedispatch/internal/apperr"
)

type OfferStatus string

const (
	OfferPending   OfferStatus = "PENDING"
	OfferAccepted  OfferStatus = "ACCEPTED"
	OfferRejected  OfferStatus = "REJECTED"
	OfferExpired   OfferStatus = "EXPIRED"
	OfferCancelled OfferStatus = "CANCELLED"
)

// Offer is a ride offered to one driver. (rideID, driverID) identifies it:
// a driver who rejects or misses an offer is never offered that ride again.
type Offer struct {
	id          string
	rideID      string
	driverID    string
	status      OfferStatus
	createdAt   time.Time
	expiresAt   time.Time
	respondedAt time.Time
}

func NewOffer(id, rideID, driverID string, now, expiresAt time.Time) Offer {
	return Offer{id: id, rideID: rideID, driverID: driverID, status: OfferPending, createdAt: now, expiresAt: expiresAt}
}

func (o Offer) ID() string             { return o.id }
func (o Offer) RideID() string         { return o.rideID }
func (o Offer) DriverID() string       { return o.driverID }
func (o Offer) Status() OfferStatus    { return o.status }
func (o Offer) CreatedAt() time.Time   { return o.createdAt }
func (o Offer) ExpiresAt() time.Time   { return o.expiresAt }
func (o Offer) RespondedAt() time.Time { return o.respondedAt }

// resolve moves a PENDING offer to a final status. Resolving to the status it
// already has is a no-op; any other change of a final offer is a conflict.
func (o *Offer) resolve(to OfferStatus, now time.Time) (bool, error) {
	switch {
	case o.status == to:
		return false, nil
	case o.status != OfferPending:
		return false, fmt.Errorf("%w: offer of ride %s to driver %s is %s", apperr.ErrConflict, o.rideID, o.driverID, o.status)
	}
	o.status = to
	o.respondedAt = now
	return true, nil
}

// OfferStore keeps offers and enforces, on insert, the invariants that also
// back the saga (docs/adr/ADR-002-concurrency.md §3): one pending offer per
// driver, one pending offer per ride, one offer per (ride, driver).
type OfferStore struct {
	mu     sync.Mutex
	offers map[offerKey]Offer
}

type offerKey struct{ rideID, driverID string }

func NewOfferStore() *OfferStore {
	return &OfferStore{offers: map[offerKey]Offer{}}
}

func (s *OfferStore) Create(_ context.Context, o Offer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.offers[offerKey{o.rideID, o.driverID}]; exists {
		return fmt.Errorf("%w: ride %s was already offered to driver %s", apperr.ErrConflict, o.rideID, o.driverID)
	}
	for _, other := range s.offers {
		if other.status != OfferPending {
			continue
		}
		if other.driverID == o.driverID {
			return fmt.Errorf("%w: driver %s already holds a pending offer for ride %s", apperr.ErrConflict, o.driverID, other.rideID)
		}
		if other.rideID == o.rideID {
			return fmt.Errorf("%w: ride %s already has a pending offer to driver %s", apperr.ErrConflict, o.rideID, other.driverID)
		}
	}
	s.offers[offerKey{o.rideID, o.driverID}] = o
	return nil
}

func (s *OfferStore) Get(_ context.Context, rideID, driverID string) (Offer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.offers[offerKey{rideID, driverID}]
	if !ok {
		return Offer{}, fmt.Errorf("%w: no offer of ride %s to driver %s", apperr.ErrNotFound, rideID, driverID)
	}
	return o, nil
}

// Resolve atomically moves the offer to a final status.
func (s *OfferStore) Resolve(_ context.Context, rideID, driverID string, to OfferStatus, now time.Time) (Offer, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := offerKey{rideID, driverID}
	o, ok := s.offers[key]
	if !ok {
		return Offer{}, false, fmt.Errorf("%w: no offer of ride %s to driver %s", apperr.ErrNotFound, rideID, driverID)
	}
	changed, err := o.resolve(to, now)
	if err != nil {
		return s.offers[key], false, err
	}
	s.offers[key] = o
	return o, changed, nil
}

// Excluded returns the drivers who rejected or let expire an offer for the
// ride. They are never offered that ride again.
func (s *OfferStore) Excluded(_ context.Context, rideID string) map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]bool{}
	for k, o := range s.offers {
		if k.rideID == rideID && (o.status == OfferRejected || o.status == OfferExpired) {
			out[k.driverID] = true
		}
	}
	return out
}
