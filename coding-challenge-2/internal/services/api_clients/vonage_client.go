package api_clients

import (
	"context"
	"sms-service/internal/models"
)

type VonageAPIClient struct {
}

func (s *VonageAPIClient) GetCostEstimation(ctx context.Context, req EstimationRequest) (models.Money, error) {
	return models.Money{Amount: "1000", Currency: "VND"}, nil
}

func (s *VonageAPIClient) Send(ctx context.Context, req SendRequest) error {
	return nil
}
