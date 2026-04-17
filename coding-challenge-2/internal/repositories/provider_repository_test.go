package repositories

import (
	"context"
	"testing"

	"sms-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRepository_SaveAndGetByID(t *testing.T) {
	repo := NewInMemoryProviderRepository()
	ctx := context.Background()

	provider := models.Provider{ID: "provider-twilio", Name: "Twilio"}
	require.NoError(t, repo.Save(provider))

	got, err := repo.GetByID(ctx, "provider-twilio")
	require.NoError(t, err)
	assert.Equal(t, "provider-twilio", got.ID)
	assert.Equal(t, "Twilio", got.Name)
}

func TestProviderRepository_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryProviderRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestProviderRepository_GetByID_EmptyRepo(t *testing.T) {
	repo := NewInMemoryProviderRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "provider-twilio")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestProviderRepository_Save_OverwritesExisting(t *testing.T) {
	repo := NewInMemoryProviderRepository()
	ctx := context.Background()

	repo.Save(models.Provider{ID: "p-001", Name: "OldName"})
	repo.Save(models.Provider{ID: "p-001", Name: "NewName"})

	got, err := repo.GetByID(ctx, "p-001")
	require.NoError(t, err)
	assert.Equal(t, "NewName", got.Name)
}

func TestProviderRepository_MultipleProviders(t *testing.T) {
	repo := NewInMemoryProviderRepository()
	ctx := context.Background()

	repo.Save(models.Provider{ID: "provider-twilio", Name: "Twilio"})
	repo.Save(models.Provider{ID: "provider-vonage", Name: "Vonage"})

	twilio, err := repo.GetByID(ctx, "provider-twilio")
	require.NoError(t, err)
	assert.Equal(t, "Twilio", twilio.Name)

	vonage, err := repo.GetByID(ctx, "provider-vonage")
	require.NoError(t, err)
	assert.Equal(t, "Vonage", vonage.Name)
}
