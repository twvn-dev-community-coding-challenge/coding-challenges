package models

import "time"

// DomainEvent is used by the in-memory event bus for async simulation.
type DomainEvent struct {
	Name       string
	MessageID  string
	OccurredAt time.Time
	Payload    map[string]string
}
