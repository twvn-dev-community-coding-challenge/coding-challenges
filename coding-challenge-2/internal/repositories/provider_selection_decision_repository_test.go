package repositories

import (
	"context"
	"testing"
	"time"

	"sms-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderSelectionDecisionRepository_SaveAndGet(t *testing.T) {
	repo := NewInMemoryProviderSelectionDecisionRepository()
	ctx := context.Background()

	decision := models.ProviderSelectionDecision{
		SMSMessageID:  "msg-001",
		ProviderID:    "provider-twilio",
		EstimatedCost: models.Money{Amount: "500", Currency: "VND"},
		CreatedAt:     time.Now().UTC(),
	}
	require.NoError(t, repo.Save(ctx, decision))

	got, err := repo.GetBySMSMessageID(ctx, "msg-001")
	require.NoError(t, err)
	assert.Equal(t, "msg-001", got.SMSMessageID)
	assert.Equal(t, "provider-twilio", got.ProviderID)
	assert.Equal(t, "500", got.EstimatedCost.Amount)
	assert.Equal(t, "VND", got.EstimatedCost.Currency)
}

func TestProviderSelectionDecisionRepository_GetBySMSMessageID_NotFound(t *testing.T) {
	repo := NewInMemoryProviderSelectionDecisionRepository()
	ctx := context.Background()

	_, err := repo.GetBySMSMessageID(ctx, "nonexistent")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestProviderSelectionDecisionRepository_GetBySMSMessageID_EmptyRepo(t *testing.T) {
	repo := NewInMemoryProviderSelectionDecisionRepository()
	ctx := context.Background()

	_, err := repo.GetBySMSMessageID(ctx, "msg-001")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestProviderSelectionDecisionRepository_Save_OverwritesExisting(t *testing.T) {
	repo := NewInMemoryProviderSelectionDecisionRepository()
	ctx := context.Background()

	repo.Save(ctx, models.ProviderSelectionDecision{
		SMSMessageID:  "msg-001",
		ProviderID:    "provider-twilio",
		EstimatedCost: models.Money{Amount: "500", Currency: "VND"},
	})
	repo.Save(ctx, models.ProviderSelectionDecision{
		SMSMessageID:  "msg-001",
		ProviderID:    "provider-vonage",
		EstimatedCost: models.Money{Amount: "1000", Currency: "VND"},
	})

	got, err := repo.GetBySMSMessageID(ctx, "msg-001")
	require.NoError(t, err)
	assert.Equal(t, "provider-vonage", got.ProviderID)
	assert.Equal(t, "1000", got.EstimatedCost.Amount)
}

func TestProviderSelectionDecisionRepository_MultipleMessages(t *testing.T) {
	repo := NewInMemoryProviderSelectionDecisionRepository()
	ctx := context.Background()

	repo.Save(ctx, models.ProviderSelectionDecision{SMSMessageID: "msg-001", ProviderID: "provider-twilio"})
	repo.Save(ctx, models.ProviderSelectionDecision{SMSMessageID: "msg-002", ProviderID: "provider-vonage"})

	got1, err := repo.GetBySMSMessageID(ctx, "msg-001")
	require.NoError(t, err)
	assert.Equal(t, "provider-twilio", got1.ProviderID)

	got2, err := repo.GetBySMSMessageID(ctx, "msg-002")
	require.NoError(t, err)
	assert.Equal(t, "provider-vonage", got2.ProviderID)
}
