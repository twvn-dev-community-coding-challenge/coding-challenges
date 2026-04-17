package api_clients

import (
	"context"
	"fmt"
	"sms-service/internal/models"
)

type EstimationRequest struct {
	ProviderID     string
	CarrierID      string
	RecipientPhone string
	Content        string
}

type SendRequest struct {
	MessageID      string
	Content        string
	RecipientPhone string
}

type ProviderAPIClient interface {
	GetCostEstimation(ctx context.Context, req EstimationRequest) (models.Money, error)
	Send(ctx context.Context, req SendRequest) error
}

type ProviderEnum string

const (
	Twilio ProviderEnum = "provider-twilio"
	Vonage ProviderEnum = "provider-vonage"
)

func GetProviderAndClient(enum ProviderEnum) (ProviderAPIClient, error) {
	switch enum {
	case Twilio:
		return &TwilioAPIClient{}, nil
	case Vonage:
		return &VonageAPIClient{}, nil
	default:
		return nil, fmt.Errorf("provider not supported: %s", enum)
	}
}
