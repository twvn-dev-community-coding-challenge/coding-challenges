package repositories

import (
	"context"
	"testing"

	"sms-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSenderRepository_SaveAndGetByID(t *testing.T) {
	repo := NewInMemorySenderRepository()
	ctx := context.Background()

	sender := models.Sender{
		ID:          "sender-001",
		Name:        "Support",
		PhoneNumber: "934567890",
		CountryId:   "c-001",
	}
	require.NoError(t, repo.Save(sender))

	got, err := repo.GetByID(ctx, "sender-001")
	require.NoError(t, err)
	assert.Equal(t, "sender-001", got.ID)
	assert.Equal(t, "Support", got.Name)
	assert.Equal(t, "934567890", got.PhoneNumber)
	assert.Equal(t, "c-001", got.CountryId)
}

func TestSenderRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemorySenderRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestSenderRepository_GetByID_EmptyRepo(t *testing.T) {
	repo := NewInMemorySenderRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "sender-001")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestSenderRepository_Save_OverwritesExisting(t *testing.T) {
	repo := NewInMemorySenderRepository()
	ctx := context.Background()

	repo.Save(models.Sender{ID: "sender-001", Name: "Old Name"})
	repo.Save(models.Sender{ID: "sender-001", Name: "New Name"})

	got, err := repo.GetByID(ctx, "sender-001")
	require.NoError(t, err)
	assert.Equal(t, "New Name", got.Name)
}
