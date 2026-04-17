package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sms-service/internal/api/handlers"
	"sms-service/internal/models"
	"sms-service/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSMSMessageByID_Success(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)

	repo := repositories.NewInMemorySMSMessageRepository()
	testMessageID := "msg-123"
	providerId := "provider-001"
	testMessage := models.SMSMessage{
		ID:             testMessageID,
		SenderID:       "sender-001",
		RecipientID:    "recipient-001",
		RecipientPhone: "+84901234567",
		Content:        "Your OTP is 123456",
		Status:         models.StatusSendSuccess,
		EstimatedCost: models.Money{
			Amount:   "350.00",
			Currency: "VND",
		},
		ActualCost: &models.Money{
			Amount:   "345.00",
			Currency: "VND",
		},
		ProviderID:        &providerId,
		CarrierID:         "carrier-001",
		ProviderMessageID: "provider-msg-789",
		FailureReason:     "",
		CreatedAt:         time.Now().Add(-1 * time.Hour),
		UpdatedAt:         time.Now(),
	}

	err := repo.Save(context.Background(), testMessage)
	require.NoError(t, err)

	handler := handlers.NewSMSMessageHandler(handlers.SMSMessageHandlerDeps{
		SMSMessageRepo: repo,
	})

	router := gin.New()
	router.GET("/sms-message/:id", handler.GetByID)

	// Act
	req := httptest.NewRequest(http.MethodGet, "/sms-message/"+testMessageID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.GetSMSMessageResponseEnvelope
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, testMessageID, response.Data.Message.ID)
	assert.Equal(t, "sender-001", response.Data.Message.SenderID)
	assert.Equal(t, "recipient-001", response.Data.Message.RecipientID)
	assert.Equal(t, "Your OTP is 123456", response.Data.Message.Content)
	assert.Equal(t, "SEND_SUCCESS", response.Data.Message.Status)
	assert.Equal(t, "350.00", response.Data.Message.EstimatedCost)
	assert.Equal(t, "345.00", response.Data.Message.ActualCost)
	assert.Equal(t, "VND", response.Data.Message.Currency)
	assert.Equal(t, "provider-001", response.Data.Message.ProviderID)
	assert.Equal(t, "carrier-001", response.Data.Message.CarrierID)
	assert.Equal(t, "provider-msg-789", response.Data.Message.ProviderMessageID)
	assert.Empty(t, response.Data.Message.FailureReason)
	assert.NotEmpty(t, response.Meta.RequestID)
	assert.NotZero(t, response.Meta.Timestamp)
}

func TestGetSMSMessageByID_NotFound(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)

	repo := repositories.NewInMemorySMSMessageRepository()
	handler := handlers.NewSMSMessageHandler(handlers.SMSMessageHandlerDeps{
		SMSMessageRepo: repo,
	})

	router := gin.New()
	router.GET("/sms-message/:id", handler.GetByID)

	nonExistentID := "msg-999"

	// Act
	req := httptest.NewRequest(http.MethodGet, "/sms-message/"+nonExistentID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response handlers.ErrorEnvelope
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Contains(t, response.Error.Message, "msg-999")
	assert.Contains(t, response.Error.Message, "not found")
	assert.NotEmpty(t, response.Meta.RequestID)
	assert.NotZero(t, response.Meta.Timestamp)
}
