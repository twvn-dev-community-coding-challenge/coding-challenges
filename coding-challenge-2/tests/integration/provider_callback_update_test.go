package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sms-service/internal/api/handlers"
	"sms-service/internal/models"
	"sms-service/internal/repositories"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSMSMessageCallbackStatus_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)

	repo := repositories.NewInMemorySMSMessageRepository()
	handler := handlers.NewSMSMessageHandler(handlers.SMSMessageHandlerDeps{
		SMSMessageRepo: repo,
	})

	// Seed data
	repo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusSendToProvider,
	})

	router := gin.New()
	router.POST("/sms-message/webhooks/provider-callback", handler.HandleProviderCallback)

	// Act
	body := `{"provider_id":"prov-1","message_id":"msg-123","status":"QUEUE","event":"DELIVERED","occurred_at":"2025-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/sms-message/webhooks/provider-callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	msg, _ := repo.GetById(context.Background(), "msg-123")

	assert.Equal(t, "msg-123", msg.ID)
	assert.Equal(t, models.StatusQueue, msg.Status)
	assert.Equal(t, (*models.Money)(nil), msg.ActualCost)

	// Assert response body contains expected fields
	var resp handlers.ProviderCallbackResponseEnvelope
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "msg-123", resp.Data.MessageID)
	assert.Equal(t, string(models.StatusQueue), resp.Data.Status)
}

func TestSMSMessageCallbackStatus_Failure(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)

	repo := repositories.NewInMemorySMSMessageRepository()
	handler := handlers.NewSMSMessageHandler(handlers.SMSMessageHandlerDeps{
		SMSMessageRepo: repo,
	})

	// Seed data
	repo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusSendToCarrier,
	})

	router := gin.New()
	router.POST("/sms-message/webhooks/provider-callback", handler.HandleProviderCallback)

	// Act
	body := `{"provider_id":"prov-1","message_id":"msg-123","status":"SEND_FAILED","event":"FAILED","occurred_at":"2025-01-01T00:00:00Z","failure_reason":"timeout"}`
	req := httptest.NewRequest(http.MethodPost, "/sms-message/webhooks/provider-callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	msg, _ := repo.GetById(context.Background(), "msg-123")

	assert.Equal(t, "msg-123", msg.ID)
	assert.Equal(t, models.StatusSendFailed, msg.Status)
	assert.Equal(t, "timeout", msg.FailureReason)
	assert.Equal(t, (*models.Money)(nil), msg.ActualCost)
}

func TestSMSMessageCallbackStatus_UpdateActualCost(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)

	repo := repositories.NewInMemorySMSMessageRepository()
	handler := handlers.NewSMSMessageHandler(handlers.SMSMessageHandlerDeps{
		SMSMessageRepo: repo,
	})

	// Seed data
	repo.Save(context.Background(), models.SMSMessage{
		ID:     "msg-123",
		Status: models.StatusSendToCarrier,
	})

	expectedActualCost := models.Money{
		Amount:   "123.45",
		Currency: "VND",
	}

	router := gin.New()
	router.POST("/sms-message/webhooks/provider-callback", handler.HandleProviderCallback)

	// Act
	body := `{"provider_id":"prov-1","message_id":"msg-123","status":"SEND_SUCCESS","event":"DELIVERED","occurred_at":"2025-01-01T00:00:00Z","actual_cost":"123.45","currency":"VND"}`
	req := httptest.NewRequest(http.MethodPost, "/sms-message/webhooks/provider-callback", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	msg, _ := repo.GetById(context.Background(), "msg-123")

	assert.Equal(t, "msg-123", msg.ID)
	assert.NotNil(t, msg.ActualCost)
	assert.Equal(t, expectedActualCost, *msg.ActualCost)
}
