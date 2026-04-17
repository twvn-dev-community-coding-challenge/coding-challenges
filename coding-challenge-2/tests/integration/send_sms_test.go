package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sms-service/internal/api/handlers"
	"sms-service/internal/models"
	"sms-service/internal/repositories"
	"sms-service/internal/services/api_clients"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockProviderAPIClient struct {
	estimatedCost models.Money
	called        bool
}

func (m *mockProviderAPIClient) GetCostEstimation(_ context.Context, _ api_clients.EstimationRequest) (models.Money, error) {
	m.called = true
	return m.estimatedCost, nil
}

func (m *mockProviderAPIClient) Send(_ context.Context, _ api_clients.SendRequest) error {
	return nil
}

func TestSendSMS_Success(t *testing.T) {

	// Arrange
	gin.SetMode(gin.TestMode)

	smsMessageRepo := repositories.NewInMemorySMSMessageRepository()

	countryRepo := repositories.NewInMemoryCountryRepository()
	country := models.Country{
		ID:          "country-001",
		Name:        "Viet Nam",
		CountryCode: "84",
	}
	countryRepo.Save(&country)

	carrierRepo := repositories.NewInMemoryCarrierRepository()
	carrier := models.Carrier{
		ID:                "carrier-001",
		Name:              "Viettel",
		CountryId:         country.ID,
		PhoneNumberPrefix: []string{"93", "96", "98", "99"},
	}
	carrierRepo.Save(carrier)

	providerRepo := repositories.NewInMemoryProviderRepository()
	provider := models.Provider{
		ID:   string(api_clients.Twilio),
		Name: "Twilio",
	}
	providerRepo.Save(provider)

	senderRepo := repositories.NewInMemorySenderRepository()
	sender := models.Sender{
		ID:          "sender-001",
		Name:        "Customer Service",
		PhoneNumber: "934567890",
		CountryId:   country.ID,
	}
	senderRepo.Save(sender)

	providerAgreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerAgreement := models.ProviderAgreement{
		ID:         "agreement-001",
		CarrierID:  carrier.ID,
		ProviderID: provider.ID,
	}
	providerAgreementRepo.Save(providerAgreement)

	sut := handlers.NewSMSMessageHandler(handlers.SMSMessageHandlerDeps{
		SMSMessageRepo:        smsMessageRepo,
		CountryRepo:           countryRepo,
		SenderRepo:            senderRepo,
		CarrierRepo:           carrierRepo,
		ProviderAgreementRepo: providerAgreementRepo,
		ProviderRepo:          providerRepo,
	})

	sendSmsRequest := handlers.SendSMSRequest{
		SenderID:       sender.ID,
		RecipientPhone: "931234567",
		Content:        "Your OTP is 123456",
		CountryCode:    "84",
	}

	router := gin.New()
	router.POST("/sms-message", sut.SendSMS)

	// Act
	body, _ := json.Marshal(sendSmsRequest)
	req := httptest.NewRequest(http.MethodPost, "/sms-message", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
		return
	}

	var resp handlers.SendSMSResponseEnvelope
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Assert SMS message is created in repository
	savedMsg, err := smsMessageRepo.GetById(context.Background(), resp.Data.MessageID)
	if err != nil {
		t.Fatalf("Expected SMS message to be saved in repository, but got error: %v", err)
	}

	// Assert SMS message has estimated cost
	if savedMsg.EstimatedCost.Amount == "" {
		t.Error("Expected SMS message to have an estimated cost, but it was empty")
	}
	if savedMsg.EstimatedCost.Amount != resp.Data.EstimatedCost {
		t.Errorf("Expected estimated cost %q, got %q", resp.Data.EstimatedCost, savedMsg.EstimatedCost.Amount)
	}

	// Assert SMS message status is SentToProvider
	if savedMsg.Status != models.StatusSendToProvider {
		t.Errorf("Expected status %q, got %q", models.StatusSendToProvider, savedMsg.Status)
	}
}
