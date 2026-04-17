package services_test

import (
	"testing"

	"sms-service/internal/models"
	"sms-service/internal/repositories"
	"sms-service/internal/services"
	"sms-service/internal/services/api_clients"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildSelectorRepos() (repositories.ProviderAgreementRepository, repositories.ProviderRepository) {
	agreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerRepo := repositories.NewInMemoryProviderRepository()

	providerRepo.Save(models.Provider{ID: string(api_clients.Twilio), Name: "Twilio"})
	agreementRepo.Save(models.ProviderAgreement{
		ID:         "agree-001",
		CarrierID:  "carrier-001",
		ProviderID: string(api_clients.Twilio),
	})

	return agreementRepo, providerRepo
}

func TestGetFirstProviderSelector_Select_Success_Twilio(t *testing.T) {
	agreementRepo, providerRepo := buildSelectorRepos()
	selector := services.NewGetFirstProviderSelector(agreementRepo, providerRepo)

	msg := &models.SMSMessage{
		ID:             "msg-001",
		CarrierID:      "carrier-001",
		RecipientPhone: "931234567",
		Content:        "Hello",
		Status:         models.StatusNew,
	}

	decision, err := selector.Select(msg)
	require.NoError(t, err)
	assert.Equal(t, "msg-001", decision.SMSMessageID)
	assert.Equal(t, string(api_clients.Twilio), decision.ProviderID)
	assert.Equal(t, "500", decision.EstimatedCost.Amount)
	assert.Equal(t, "VND", decision.EstimatedCost.Currency)
	assert.False(t, decision.CreatedAt.IsZero())
}

func TestGetFirstProviderSelector_Select_Success_Vonage(t *testing.T) {
	agreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerRepo := repositories.NewInMemoryProviderRepository()
	providerRepo.Save(models.Provider{ID: string(api_clients.Vonage), Name: "Vonage"})
	agreementRepo.Save(models.ProviderAgreement{
		ID:         "agree-002",
		CarrierID:  "carrier-001",
		ProviderID: string(api_clients.Vonage),
	})

	selector := services.NewGetFirstProviderSelector(agreementRepo, providerRepo)
	msg := &models.SMSMessage{
		ID:        "msg-002",
		CarrierID: "carrier-001",
		Status:    models.StatusNew,
	}

	decision, err := selector.Select(msg)
	require.NoError(t, err)
	assert.Equal(t, string(api_clients.Vonage), decision.ProviderID)
	assert.Equal(t, "1000", decision.EstimatedCost.Amount)
}

func TestGetFirstProviderSelector_Select_MessageNotNew(t *testing.T) {
	agreementRepo, providerRepo := buildSelectorRepos()
	selector := services.NewGetFirstProviderSelector(agreementRepo, providerRepo)

	for _, status := range []models.SMSStatus{
		models.StatusSendToProvider,
		models.StatusQueue,
		models.StatusSendToCarrier,
		models.StatusSendSuccess,
		models.StatusSendFailed,
		models.StatusCarrierRejected,
	} {
		msg := &models.SMSMessage{ID: "msg-001", CarrierID: "carrier-001", Status: status}
		_, err := selector.Select(msg)
		assert.Error(t, err, "expected error for status %s", status)
	}
}

func TestGetFirstProviderSelector_Select_NoAgreements(t *testing.T) {
	agreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerRepo := repositories.NewInMemoryProviderRepository()
	providerRepo.Save(models.Provider{ID: string(api_clients.Twilio), Name: "Twilio"})

	selector := services.NewGetFirstProviderSelector(agreementRepo, providerRepo)
	msg := &models.SMSMessage{
		ID:        "msg-001",
		CarrierID: "carrier-no-agreement",
		Status:    models.StatusNew,
	}

	_, err := selector.Select(msg)
	assert.Error(t, err)
}

func TestGetFirstProviderSelector_Select_ProviderNotFound(t *testing.T) {
	agreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerRepo := repositories.NewInMemoryProviderRepository()
	// Agreement points to a provider not in providerRepo
	agreementRepo.Save(models.ProviderAgreement{
		ID:         "agree-001",
		CarrierID:  "carrier-001",
		ProviderID: string(api_clients.Twilio),
	})

	selector := services.NewGetFirstProviderSelector(agreementRepo, providerRepo)
	msg := &models.SMSMessage{
		ID:        "msg-001",
		CarrierID: "carrier-001",
		Status:    models.StatusNew,
	}

	_, err := selector.Select(msg)
	assert.Error(t, err)
}

func TestGetFirstProviderSelector_Select_UnsupportedProvider(t *testing.T) {
	agreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerRepo := repositories.NewInMemoryProviderRepository()
	providerRepo.Save(models.Provider{ID: "provider-unsupported", Name: "Unknown"})
	agreementRepo.Save(models.ProviderAgreement{
		ID:         "agree-001",
		CarrierID:  "carrier-001",
		ProviderID: "provider-unsupported",
	})

	selector := services.NewGetFirstProviderSelector(agreementRepo, providerRepo)
	msg := &models.SMSMessage{
		ID:        "msg-001",
		CarrierID: "carrier-001",
		Status:    models.StatusNew,
	}

	_, err := selector.Select(msg)
	assert.Error(t, err)
}

func TestGetFirstProviderSelector_Select_UsesFirstAgreement(t *testing.T) {
	agreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerRepo := repositories.NewInMemoryProviderRepository()
	providerRepo.Save(models.Provider{ID: string(api_clients.Twilio), Name: "Twilio"})
	providerRepo.Save(models.Provider{ID: string(api_clients.Vonage), Name: "Vonage"})
	// First agreement is Twilio
	agreementRepo.Save(models.ProviderAgreement{ID: "agree-001", CarrierID: "carrier-001", ProviderID: string(api_clients.Twilio)})
	agreementRepo.Save(models.ProviderAgreement{ID: "agree-002", CarrierID: "carrier-001", ProviderID: string(api_clients.Vonage)})

	selector := services.NewGetFirstProviderSelector(agreementRepo, providerRepo)
	msg := &models.SMSMessage{ID: "msg-001", CarrierID: "carrier-001", Status: models.StatusNew}

	decision, err := selector.Select(msg)
	require.NoError(t, err)
	assert.Equal(t, string(api_clients.Twilio), decision.ProviderID)
}
