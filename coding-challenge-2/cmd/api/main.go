package main

import (
	"context"
	"log"
	"os"
	"sms-service/internal/models"
	"sms-service/internal/repositories"
	"sms-service/internal/services/api_clients"

	"sms-service/internal/api/handlers"
	"sms-service/internal/api/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// stubProviderAPIClient is a simple in-memory stub that returns a fixed cost.
type stubProviderAPIClient struct{}

func (s *stubProviderAPIClient) Send(_ context.Context, _ api_clients.SendRequest) error {
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ── Repositories ────────────────────────────────────────────────────────
	smsRepo := repositories.NewInMemorySMSMessageRepository()
	countryRepo := repositories.NewInMemoryCountryRepository()
	carrierRepo := repositories.NewInMemoryCarrierRepository()
	providerRepo := repositories.NewInMemoryProviderRepository()
	providerAgreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	senderRepo := repositories.NewInMemorySenderRepository()
	decisionRepo := repositories.NewInMemoryProviderSelectionDecisionRepository()

	// ── Seed data ───────────────────────────────────────────────────────────
	seedData(countryRepo, carrierRepo, providerRepo, providerAgreementRepo, senderRepo)

	// ── Handler ─────────────────────────────────────────────────────────────
	smsHandler := handlers.NewSMSMessageHandler(handlers.SMSMessageHandlerDeps{
		SMSMessageRepo:                smsRepo,
		CountryRepo:                   countryRepo,
		SenderRepo:                    senderRepo,
		CarrierRepo:                   carrierRepo,
		ProviderAgreementRepo:         providerAgreementRepo,
		ProviderRepo:                  providerRepo,
		ProviderSelectionDecisionRepo: decisionRepo,
	})

	router := gin.Default()
	routes.Register(router, smsHandler)

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func seedData(
	countryRepo repositories.CountryRepository,
	carrierRepo repositories.CarrierRepository,
	providerRepo repositories.ProviderRepository,
	providerAgreementRepo repositories.ProviderAgreementRepository,
	senderRepo repositories.SenderRepository,
) {
	country := &models.Country{ID: "country-vn", Name: "Viet Nam", CountryCode: "84"}
	countryRepo.Save(country)

	carrier := models.Carrier{
		ID:                "carrier-viettel",
		Name:              "Viettel",
		CountryId:         country.ID,
		PhoneNumberPrefix: []string{"93", "96", "97", "98", "86", "32", "33", "34", "35", "36", "37", "38", "39"},
	}
	carrierRepo.Save(carrier)

	provider := models.Provider{ID: "provider-twilio", Name: "Twilio"}
	providerRepo.Save(provider)

	providerAgreementRepo.Save(models.ProviderAgreement{
		ID:         "agreement-001",
		CarrierID:  carrier.ID,
		ProviderID: provider.ID,
	})

	senderRepo.Save(models.Sender{
		ID:          "sender-001",
		Name:        "Customer Service",
		PhoneNumber: "934567890",
		CountryId:   country.ID,
	})

	log.Println("Seeded: country=VN, carrier=Viettel, provider=Twilio, agreement, sender")
}
