// Package events defines the event envelope and the event types exchanged
// between modules. It is the contract a real broker would carry.
package events

import (
	"context"
	"time"

	"ridedispatch/internal/platform/ids"
	"ridedispatch/internal/platform/trace"
)

const (
	RideCreated       = "RideCreated"
	OfferCreated      = "OfferCreated"
	OfferAccepted     = "OfferAccepted"
	OfferRejected     = "OfferRejected"
	OfferExpired      = "OfferExpired"
	MatchingExhausted = "MatchingExhausted"
	RideCancelled     = "RideCancelled"
	RideCompleted     = "RideCompleted"
)

// Event is the envelope. RideID is also the partition key.
type Event struct {
	ID          string    `json:"eventId"`
	Type        string    `json:"type"`
	RideID      string    `json:"rideId"`
	DriverID    string    `json:"driverId,omitempty"`
	CausationID string    `json:"causationId,omitempty"`
	OccurredAt  time.Time `json:"occurredAt"`
}

// Publisher delivers events at least once, in order per RideID.
type Publisher interface {
	Publish(ctx context.Context, events ...Event)
}

// New builds an event caused by whatever ctx carries (a request or an event).
func New(ctx context.Context, typ, rideID, driverID string, now time.Time) Event {
	return Event{
		ID:          ids.New("E"),
		Type:        typ,
		RideID:      rideID,
		DriverID:    driverID,
		CausationID: trace.CausationID(ctx),
		OccurredAt:  now,
	}
}
