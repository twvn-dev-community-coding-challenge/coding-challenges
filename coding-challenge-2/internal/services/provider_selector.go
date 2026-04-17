package services

import (
	"context"
	"fmt"
	"log"
	"sms-service/internal/models"
	"sms-service/internal/repositories"
	"sms-service/internal/services/api_clients"
	"time"
)

type ProviderSelector interface {
	Select(message *models.SMSMessage) (models.ProviderSelectionDecision, error)
}

type GetFirstProviderSelector struct {
	providerAgreementRepo repositories.ProviderAgreementRepository
	providerRepo          repositories.ProviderRepository
}

func NewGetFirstProviderSelector(
	providerAgreementRepo repositories.ProviderAgreementRepository,
	providerRepo repositories.ProviderRepository,
) *GetFirstProviderSelector {
	return &GetFirstProviderSelector{
		providerAgreementRepo: providerAgreementRepo,
		providerRepo:          providerRepo,
	}
}

func (g *GetFirstProviderSelector) Select(message *models.SMSMessage) (models.ProviderSelectionDecision, error) {
	ctx := context.Background()

	if message.Status != models.StatusNew {
		return models.ProviderSelectionDecision{}, fmt.Errorf("message is not in correct status")
	}

	agreements, err := g.providerAgreementRepo.FindManyByCarrierId(ctx, message.CarrierID)
	if err != nil || len(agreements) == 0 {
		return models.ProviderSelectionDecision{}, fmt.Errorf("no provider agreements for carrier")
	}

	provider, err := g.providerRepo.GetByID(ctx, agreements[0].ProviderID)
	if err != nil {
		return models.ProviderSelectionDecision{}, fmt.Errorf("no provider found for carrier")
	}

	enum := api_clients.ProviderEnum(provider.ID)
	apiClient, err := api_clients.GetProviderAndClient(enum)
	if err != nil {
		return models.ProviderSelectionDecision{}, fmt.Errorf("failed to get provider client for provider ID %s", provider.ID)
	}

	estimatedCost, err := apiClient.GetCostEstimation(ctx, api_clients.EstimationRequest{
		ProviderID:     provider.ID,
		CarrierID:      message.CarrierID,
		RecipientPhone: message.RecipientPhone,
		Content:        message.Content,
	})
	if err != nil {
		return models.ProviderSelectionDecision{}, fmt.Errorf("failed to get cost estimation from provider")
	}

	log.Println("Provider selected:", provider.ID)

	return models.ProviderSelectionDecision{
		SMSMessageID:  message.ID,
		ProviderID:    provider.ID,
		EstimatedCost: estimatedCost,
		CreatedAt:     time.Now().UTC(),
	}, nil
}
