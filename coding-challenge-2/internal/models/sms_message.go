package models

import "time"

// SMSStatus is the canonical lifecycle status of an SMS message.
type SMSStatus string

const (
	StatusNew             SMSStatus = "NEW"
	StatusSendToProvider  SMSStatus = "SEND_TO_PROVIDER"
	StatusQueue           SMSStatus = "QUEUE"
	StatusSendToCarrier   SMSStatus = "SEND_TO_CARRIER"
	StatusSendSuccess     SMSStatus = "SEND_SUCCESS"
	StatusSendFailed      SMSStatus = "SEND_FAILED"
	StatusCarrierRejected SMSStatus = "CARRIER_REJECTED"
)

// SMSMessage is the central aggregate.
type SMSMessage struct {
	ID                string
	SenderID          string
	RecipientID       string
	RecipientPhone    string
	Content           string
	Status            SMSStatus
	EstimatedCost     Money
	ActualCost        *Money
	ProviderID        *string
	CarrierID         string
	ProviderMessageID string
	FailureReason     string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ListSMSMessagesFilter is used by both the services and repositories layers.
type ListSMSMessagesFilter struct {
	Status     *SMSStatus
	ProviderID string
	CarrierID  string
	From       *time.Time
	To         *time.Time
	Page       int
	PageSize   int
}

func (f *SMSMessage) CanTransitionTo(status SMSStatus) bool {
	currentStatus := f.Status

	switch currentStatus {
	case StatusNew:
		return status == StatusSendToProvider
	case StatusSendToProvider:
		return status == StatusQueue
	case StatusQueue:
		return status == StatusSendToCarrier || status == StatusCarrierRejected
	case StatusSendToCarrier:
		return status == StatusSendSuccess || status == StatusSendFailed
	case StatusSendSuccess:
		return false
	case StatusSendFailed:
		return status == StatusSendToProvider
	case StatusCarrierRejected:
		return status == StatusSendToProvider
	default:
		return false
	}
}

//StatusNew             SMSStatus = "NEW"
//StatusSendToProvider  SMSStatus = "SEND_TO_PROVIDER"
//StatusQueue           SMSStatus = "QUEUE"
//StatusSendToCarrier   SMSStatus = "SEND_TO_CARRIER"
//StatusSendSuccess     SMSStatus = "SEND_SUCCESS"
//StatusSendFailed      SMSStatus = "SEND_FAILED"
//StatusCarrierRejected SMSStatus = "CARRIER_REJECTED"
