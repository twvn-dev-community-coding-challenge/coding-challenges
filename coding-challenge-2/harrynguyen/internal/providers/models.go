package providers

import (
	"time"
)

// MessageStatus represents the lifecycle state of an SMS.
type MessageStatus string

const (
	StatusNew             MessageStatus = "New"
	StatusSendToProvider  MessageStatus = "Send-to-provider"
	StatusQueue           MessageStatus = "Queue"
	StatusSendToCarrier   MessageStatus = "Send-to-carrier"
	StatusSendSuccess     MessageStatus = "Send-success"
	StatusSendFailed      MessageStatus = "Send-failed"
	StatusCarrierRejected MessageStatus = "Carrier-rejected"
)

// Carrier represents a Mobile Network Operator.
type Carrier string

const (
	CarrierViettel   Carrier = "Viettel"
	CarrierMobifone  Carrier = "Mobifone"
	CarrierVinaphone Carrier = "Vinaphone"
	CarrierAIS       Carrier = "AIS"
	CarrierDTAC      Carrier = "DTAC"
	CarrierSingtel   Carrier = "Singtel"
	CarrierStarHub   Carrier = "StarHub"
	CarrierGlobe     Carrier = "Globe"
	CarrierSmart     Carrier = "Smart"
	CarrierDITO      Carrier = "DITO"
	CarrierUnknown   Carrier = "Unknown"
)

// Provider represents an SMS gateway provider.
type Provider string

const (
	ProviderTwilio      Provider = "Twilio"
	ProviderVonage      Provider = "Vonage"
	ProviderInfobip     Provider = "Infobip"
	ProviderAWSSNS      Provider = "AWS SNS"
	ProviderTelnyx      Provider = "Telnyx"
	ProviderMessageBird Provider = "MessageBird"
	ProviderSinch       Provider = "Sinch"
)

// SMSMessage represents a message being processed by the system.
type SMSMessage struct {
	ID            string        `json:"id"`
	MessageID     string        `json:"message_id"`
	SendSource    string        `json:"send_source,omitempty"`
	Country       string        `json:"country"`
	PhoneNumber   string        `json:"phone_number"`
	Content       string        `json:"content"`
	Carrier       Carrier       `json:"carrier"`
	Provider      Provider      `json:"provider"`
	Status        MessageStatus `json:"status"`
	EstimatedCost float64       `json:"estimated_cost"`
	ActualCost    float64       `json:"actual_cost"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// StatusLog represents a historical record of status changes.
type StatusLog struct {
	ID        string        `json:"id"`
	SMSID     string        `json:"sms_id"`
	Status    MessageStatus `json:"status"`
	Timestamp time.Time     `json:"timestamp"`
	Metadata  string        `json:"metadata,omitempty"`
}

// IsValidTransition checks if moving from one status to another is valid for the synchronous
// send pipeline. Asynchronous provider callbacks additionally use applyStatusWithRecovery for
// out-of-order events and terminal-state mitigation.
func (s MessageStatus) IsValidTransition(to MessageStatus) bool {
	if s == to {
		return true
	}

	transitions := map[MessageStatus][]MessageStatus{
		StatusNew:             {StatusSendToProvider, StatusSendFailed},
		StatusSendToProvider:  {StatusQueue, StatusSendFailed},
		StatusQueue:           {StatusSendToCarrier, StatusCarrierRejected, StatusSendFailed},
		StatusSendToCarrier:   {StatusSendSuccess, StatusSendFailed},
		StatusCarrierRejected: {StatusSendToProvider},
		StatusSendFailed:      {StatusSendToProvider},
		StatusSendSuccess:     {}, // Terminal
	}

	for _, t := range transitions[s] {
		if t == to {
			return true
		}
	}
	return false
}
