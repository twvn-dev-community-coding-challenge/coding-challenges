package models

import "time"

// ProviderSelectionDecision is an immutable audit record of which provider was chosen for a message.
type ProviderSelectionDecision struct {
	SMSMessageID  string
	ProviderID    string
	EstimatedCost Money
	CreatedAt     time.Time
}
