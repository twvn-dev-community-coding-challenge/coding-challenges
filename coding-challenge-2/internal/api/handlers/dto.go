package handlers

import "time"

// ── Request types ─────────────────────────────────────────────────────────────

type SendSMSRequest struct {
	SenderID       string `json:"sender_id"        binding:"required"`
	RecipientPhone string `json:"recipient_phone"  binding:"required"`
	Content        string `json:"content"          binding:"required,min=1,max=1000"`
	CountryCode    string `json:"country_code"`
}

type ProviderCallbackRequest struct {
	ProviderID        string    `json:"provider_id"         binding:"required"`
	ProviderMessageID string    `json:"provider_message_id"`
	MessageID         string    `json:"message_id"`
	Status            string    `json:"status"`
	Event             string    `json:"event"               binding:"required"`
	OccurredAt        time.Time `json:"occurred_at"         binding:"required"`
	ActualCost        string    `json:"actual_cost"`
	Currency          string    `json:"currency"`
	FailureReason     string    `json:"failure_reason"`
}

// ── Response types ─────────────────────────────────────────────────────────────

type Meta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
	Meta  Meta      `json:"meta"`
}

// ── Send SMS ──────────────────────────────────────────────────────────────────

type SendSMSResponse struct {
	MessageID     string    `json:"message_id"`
	Status        string    `json:"status"`
	ProviderID    string    `json:"provider_id"`
	CarrierID     string    `json:"carrier_id"`
	EstimatedCost string    `json:"estimated_cost"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
}

type SendSMSResponseEnvelope struct {
	Data SendSMSResponse `json:"data"`
	Meta Meta            `json:"meta"`
}

// ── Provider callback ─────────────────────────────────────────────────────────

type ProviderCallbackResponse struct {
	MessageID  string    `json:"message_id"`
	Status     string    `json:"status"`
	ActualCost string    `json:"actual_cost,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProviderCallbackResponseEnvelope struct {
	Data ProviderCallbackResponse `json:"data"`
	Meta Meta                     `json:"meta"`
}

// ── Get SMS message ────────────────────────────────────────────────────────────

type SMSMessageDTO struct {
	ID                string    `json:"id"`
	SenderID          string    `json:"sender_id"`
	RecipientID       string    `json:"recipient_id"`
	Content           string    `json:"content"`
	Status            string    `json:"status"`
	EstimatedCost     string    `json:"estimated_cost"`
	ActualCost        string    `json:"actual_cost,omitempty"`
	Currency          string    `json:"currency"`
	ProviderID        string    `json:"provider_id"`
	CarrierID         string    `json:"carrier_id"`
	ProviderMessageID string    `json:"provider_message_id,omitempty"`
	FailureReason     string    `json:"failure_reason,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ProviderSelectionDecisionDTO struct {
	SMSMessageID  string    `json:"sms_message_id"`
	ProviderID    string    `json:"provider_id"`
	EstimatedCost string    `json:"estimated_cost"`
	StrategyName  string    `json:"strategy_name"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}

type GetSMSMessageResponse struct {
	Message  SMSMessageDTO                 `json:"message"`
	Decision *ProviderSelectionDecisionDTO `json:"decision,omitempty"`
}

type GetSMSMessageResponseEnvelope struct {
	Data GetSMSMessageResponse `json:"data"`
	Meta Meta                  `json:"meta"`
}
