package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sms-service/internal/api/handlers"
	"sms-service/internal/models"
	"sms-service/internal/repositories"
	"sms-service/internal/services/api_clients"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// seedDeps returns fully wired handler deps with country/carrier/provider/agreement seeded.
func seedDeps() handlers.SMSMessageHandlerDeps {
	countryRepo := repositories.NewInMemoryCountryRepository()
	countryRepo.Save(&models.Country{ID: "country-001", Name: "Viet Nam", CountryCode: "84"})

	carrierRepo := repositories.NewInMemoryCarrierRepository()
	carrierRepo.Save(models.Carrier{
		ID:                "carrier-001",
		Name:              "Viettel",
		CountryId:         "country-001",
		PhoneNumberPrefix: []string{"93", "96", "98"},
	})

	providerRepo := repositories.NewInMemoryProviderRepository()
	providerRepo.Save(models.Provider{ID: string(api_clients.Twilio), Name: "Twilio"})

	providerAgreementRepo := repositories.NewInMemoryProviderAgreementRepository()
	providerAgreementRepo.Save(models.ProviderAgreement{
		ID:         "agree-001",
		CarrierID:  "carrier-001",
		ProviderID: string(api_clients.Twilio),
	})

	return handlers.SMSMessageHandlerDeps{
		SMSMessageRepo:                repositories.NewInMemorySMSMessageRepository(),
		CountryRepo:                   countryRepo,
		SenderRepo:                    repositories.NewInMemorySenderRepository(),
		CarrierRepo:                   carrierRepo,
		ProviderAgreementRepo:         providerAgreementRepo,
		ProviderRepo:                  providerRepo,
		ProviderSelectionDecisionRepo: repositories.NewInMemoryProviderSelectionDecisionRepository(),
	}
}

func buildRouter(deps handlers.SMSMessageHandlerDeps) *gin.Engine {
	h := handlers.NewSMSMessageHandler(deps)
	r := gin.New()
	r.POST("/sms", h.SendSMS)
	r.POST("/callback", h.HandleProviderCallback)
	r.GET("/sms/:id", h.GetByID)
	return r
}

// ── SendSMS ───────────────────────────────────────────────────────────────────

func TestSendSMS_Success(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	body, _ := json.Marshal(handlers.SendSMSRequest{
		SenderID:       "sender-001",
		RecipientPhone: "931234567",
		Content:        "Your OTP is 123456",
		CountryCode:    "84",
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp handlers.SendSMSResponseEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Data.MessageID)
	assert.Equal(t, string(models.StatusSendToProvider), resp.Data.Status)
	assert.Equal(t, string(api_clients.Twilio), resp.Data.ProviderID)
	assert.Equal(t, "carrier-001", resp.Data.CarrierID)
	assert.Equal(t, "500", resp.Data.EstimatedCost)
	assert.Equal(t, "VND", resp.Data.Currency)

	saved, err := deps.SMSMessageRepo.GetById(context.Background(), resp.Data.MessageID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusSendToProvider, saved.Status)
}

func TestSendSMS_MissingRequiredFields(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	body := `{"content":"Hello"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sms", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
}

func TestSendSMS_InvalidJSON(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sms", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendSMS_CountryNotFound(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	body, _ := json.Marshal(handlers.SendSMSRequest{
		SenderID:       "sender-001",
		RecipientPhone: "931234567",
		Content:        "Hello",
		CountryCode:    "999",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "COUNTRY_NOT_FOUND", resp.Error.Code)
}

func TestSendSMS_CarrierNotFound(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	body, _ := json.Marshal(handlers.SendSMSRequest{
		SenderID:       "sender-001",
		RecipientPhone: "701234567", // prefix 70 has no carrier
		Content:        "Hello",
		CountryCode:    "84",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "CARRIER_NOT_FOUND", resp.Error.Code)
}

func TestSendSMS_NoProviderAgreement(t *testing.T) {
	deps := seedDeps()
	// Replace agreement repo with empty one so no agreements exist
	deps.ProviderAgreementRepo = repositories.NewInMemoryProviderAgreementRepository()
	router := buildRouter(deps)

	body, _ := json.Marshal(handlers.SendSMSRequest{
		SenderID:       "sender-001",
		RecipientPhone: "931234567",
		Content:        "Hello",
		CountryCode:    "84",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NO_PROVIDER_AGREEMENT", resp.Error.Code)
}

// ── HandleProviderCallback ────────────────────────────────────────────────────

func callbackBody(messageID, status, event string) string {
	return `{"provider_id":"prov-1","message_id":"` + messageID + `","status":"` + status + `","event":"` + event + `","occurred_at":"2025-01-01T00:00:00Z"}`
}

func TestHandleProviderCallback_InvalidJSON(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
}

func TestHandleProviderCallback_MissingMessageID(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	body := `{"provider_id":"prov-1","message_id":"","status":"QUEUE","event":"DELIVERED","occurred_at":"2025-01-01T00:00:00Z"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "BAD_REQUEST", resp.Error.Code)
}

func TestHandleProviderCallback_MessageNotFound(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(callbackBody("msg-nonexistent", "QUEUE", "DELIVERED")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
}

func TestHandleProviderCallback_InvalidStatusTransition(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusNew,
	})
	router := buildRouter(deps)

	// NEW → SEND_SUCCESS is not a valid transition
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(callbackBody("msg-123", "SEND_SUCCESS", "DELIVERED")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "INVALID_STATUS", resp.Error.Code)
}

func TestHandleProviderCallback_Success_Queue(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusSendToProvider,
	})
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(callbackBody("msg-123", "QUEUE", "QUEUED")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp handlers.ProviderCallbackResponseEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "msg-123", resp.Data.MessageID)
	assert.Equal(t, string(models.StatusQueue), resp.Data.Status)

	saved, _ := deps.SMSMessageRepo.GetById(context.Background(), "msg-123")
	assert.Equal(t, models.StatusQueue, saved.Status)
}

func TestHandleProviderCallback_Success_SendToCarrier(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusQueue,
	})
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(callbackBody("msg-123", "SEND_TO_CARRIER", "DISPATCHED")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	saved, _ := deps.SMSMessageRepo.GetById(context.Background(), "msg-123")
	assert.Equal(t, models.StatusSendToCarrier, saved.Status)
}

func TestHandleProviderCallback_Success_SendSuccessWithCost(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusSendToCarrier,
	})
	router := buildRouter(deps)

	body := `{"provider_id":"prov-1","message_id":"msg-123","status":"SEND_SUCCESS","event":"DELIVERED","occurred_at":"2025-01-01T00:00:00Z","actual_cost":"450.00","currency":"VND"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	saved, _ := deps.SMSMessageRepo.GetById(context.Background(), "msg-123")
	assert.Equal(t, models.StatusSendSuccess, saved.Status)
	require.NotNil(t, saved.ActualCost)
	assert.Equal(t, "450.00", saved.ActualCost.Amount)
}

func TestHandleProviderCallback_Success_SendFailedWithReason(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusSendToCarrier,
	})
	router := buildRouter(deps)

	body := `{"provider_id":"prov-1","message_id":"msg-123","status":"SEND_FAILED","event":"FAILED","occurred_at":"2025-01-01T00:00:00Z","failure_reason":"network timeout"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	saved, _ := deps.SMSMessageRepo.GetById(context.Background(), "msg-123")
	assert.Equal(t, models.StatusSendFailed, saved.Status)
	assert.Equal(t, "network timeout", saved.FailureReason)
}

func TestHandleProviderCallback_Success_SendFailedDefaultReason(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusSendToCarrier,
	})
	router := buildRouter(deps)

	body := `{"provider_id":"prov-1","message_id":"msg-123","status":"SEND_FAILED","event":"FAILED","occurred_at":"2025-01-01T00:00:00Z"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	saved, _ := deps.SMSMessageRepo.GetById(context.Background(), "msg-123")
	assert.Equal(t, models.StatusSendFailed, saved.Status)
	assert.Equal(t, "Unknown failure reason", saved.FailureReason)
}

func TestHandleProviderCallback_Success_CarrierRejected(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusQueue,
	})
	router := buildRouter(deps)

	body := `{"provider_id":"prov-1","message_id":"msg-123","status":"CARRIER_REJECTED","event":"REJECTED","occurred_at":"2025-01-01T00:00:00Z","failure_reason":"blocked"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	saved, _ := deps.SMSMessageRepo.GetById(context.Background(), "msg-123")
	assert.Equal(t, models.StatusCarrierRejected, saved.Status)
	assert.Equal(t, "blocked", saved.FailureReason)
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID_Success(t *testing.T) {
	deps := seedDeps()
	providerID := string(api_clients.Twilio)
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:             "msg-abc",
		SenderID:       "sender-001",
		Content:        "Test content",
		Status:         models.StatusSendToProvider,
		CarrierID:      "carrier-001",
		ProviderID:     &providerID,
		EstimatedCost:  models.Money{Amount: "500", Currency: "VND"},
		RecipientPhone: "931234567",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	})
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sms/msg-abc", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp handlers.GetSMSMessageResponseEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "msg-abc", resp.Data.Message.ID)
	assert.Equal(t, "sender-001", resp.Data.Message.SenderID)
	assert.Equal(t, string(models.StatusSendToProvider), resp.Data.Message.Status)
	assert.Equal(t, "500", resp.Data.Message.EstimatedCost)
	assert.Equal(t, providerID, resp.Data.Message.ProviderID)
}

func TestGetByID_NotFound(t *testing.T) {
	deps := seedDeps()
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sms/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp handlers.ErrorEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
}

func TestGetByID_WithActualCost(t *testing.T) {
	deps := seedDeps()
	actualCost := models.Money{Amount: "450", Currency: "VND"}
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:         "msg-done",
		Status:     models.StatusSendSuccess,
		ActualCost: &actualCost,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sms/msg-done", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp handlers.GetSMSMessageResponseEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "450", resp.Data.Message.ActualCost)
}

func TestGetByID_NilProviderID(t *testing.T) {
	deps := seedDeps()
	deps.SMSMessageRepo.Save(context.Background(), models.SMSMessage{
		ID:         "msg-new",
		Status:     models.StatusNew,
		ProviderID: nil,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})
	router := buildRouter(deps)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sms/msg-new", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp handlers.GetSMSMessageResponseEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "", resp.Data.Message.ProviderID)
}
