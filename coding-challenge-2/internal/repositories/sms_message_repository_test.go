package repositories

import (
	"context"
	"testing"
	"time"

	"sms-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMSMessageRepository_Save_NewMessage(t *testing.T) {
	repo := NewInMemorySMSMessageRepository()
	ctx := context.Background()

	msg := models.SMSMessage{
		ID:       "msg-001",
		SenderID: "sender-001",
		Content:  "Hello",
		Status:   models.StatusNew,
	}

	require.NoError(t, repo.Save(ctx, msg))

	got, err := repo.GetById(ctx, "msg-001")
	require.NoError(t, err)
	assert.Equal(t, msg.ID, got.ID)
	assert.Equal(t, msg.SenderID, got.SenderID)
	assert.Equal(t, msg.Content, got.Content)
	assert.Equal(t, models.StatusNew, got.Status)
}

func TestSMSMessageRepository_Save_UpdatesSpecificFields(t *testing.T) {
	repo := NewInMemorySMSMessageRepository()
	ctx := context.Background()
	providerID := "provider-twilio"

	original := models.SMSMessage{
		ID:        "msg-001",
		Content:   "Hello",
		SenderID:  "sender-001",
		CarrierID: "carrier-001",
		Status:    models.StatusNew,
	}
	require.NoError(t, repo.Save(ctx, original))

	update := models.SMSMessage{
		ID:            "msg-001",
		Status:        models.StatusSendToProvider,
		ProviderID:    &providerID,
		EstimatedCost: models.Money{Amount: "500", Currency: "VND"},
		UpdatedAt:     time.Now().UTC(),
	}
	require.NoError(t, repo.Save(ctx, update))

	got, err := repo.GetById(ctx, "msg-001")
	require.NoError(t, err)
	assert.Equal(t, models.StatusSendToProvider, got.Status)
	assert.Equal(t, &providerID, got.ProviderID)
	assert.Equal(t, "500", got.EstimatedCost.Amount)
	assert.Equal(t, "VND", got.EstimatedCost.Currency)
	// Original immutable fields must not change
	assert.Equal(t, "Hello", got.Content)
	assert.Equal(t, "sender-001", got.SenderID)
	assert.Equal(t, "carrier-001", got.CarrierID)
}

func TestSMSMessageRepository_Save_UpdateActualCost(t *testing.T) {
	repo := NewInMemorySMSMessageRepository()
	ctx := context.Background()

	msg := models.SMSMessage{ID: "msg-002", Status: models.StatusSendToCarrier}
	require.NoError(t, repo.Save(ctx, msg))

	actualCost := &models.Money{Amount: "450", Currency: "VND"}
	msg.Status = models.StatusSendSuccess
	msg.ActualCost = actualCost
	require.NoError(t, repo.Save(ctx, msg))

	got, err := repo.GetById(ctx, "msg-002")
	require.NoError(t, err)
	assert.Equal(t, models.StatusSendSuccess, got.Status)
	require.NotNil(t, got.ActualCost)
	assert.Equal(t, "450", got.ActualCost.Amount)
}

func TestSMSMessageRepository_Save_UpdateFailureReason(t *testing.T) {
	repo := NewInMemorySMSMessageRepository()
	ctx := context.Background()

	msg := models.SMSMessage{ID: "msg-003", Status: models.StatusSendToCarrier}
	require.NoError(t, repo.Save(ctx, msg))

	msg.Status = models.StatusSendFailed
	msg.FailureReason = "network timeout"
	require.NoError(t, repo.Save(ctx, msg))

	got, err := repo.GetById(ctx, "msg-003")
	require.NoError(t, err)
	assert.Equal(t, models.StatusSendFailed, got.Status)
	assert.Equal(t, "network timeout", got.FailureReason)
}

func TestSMSMessageRepository_GetById_NotFound(t *testing.T) {
	repo := NewInMemorySMSMessageRepository()
	ctx := context.Background()

	_, err := repo.GetById(ctx, "nonexistent")
	assert.ErrorIs(t, err, models.ErrNotFound)
}
