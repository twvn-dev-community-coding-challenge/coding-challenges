package api_clients

import (
	"context"
	"sms-service/internal/models"
)

type TwilioAPIClient struct {
}

func (t *TwilioAPIClient) GetCostEstimation(ctx context.Context, req EstimationRequest) (models.Money, error) {
	return models.Money{Amount: "500", Currency: "VND"}, nil
}

func (t *TwilioAPIClient) Send(ctx context.Context, req SendRequest) error {
	return nil
}
