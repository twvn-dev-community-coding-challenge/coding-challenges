package repositories

import (
	"context"
	"testing"

	"sms-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderAgreementRepository_SaveAndFindByCarrierID(t *testing.T) {
	repo := NewInMemoryProviderAgreementRepository()
	ctx := context.Background()

	agreement := models.ProviderAgreement{
		ID:         "agree-001",
		CarrierID:  "carrier-001",
		ProviderID: "provider-twilio",
	}
	require.NoError(t, repo.Save(agreement))

	agreements, err := repo.FindManyByCarrierId(ctx, "carrier-001")
	require.NoError(t, err)
	require.Len(t, agreements, 1)
	assert.Equal(t, "agree-001", agreements[0].ID)
	assert.Equal(t, "provider-twilio", agreements[0].ProviderID)
}

func TestProviderAgreementRepository_FindMany_MultipleProviders(t *testing.T) {
	repo := NewInMemoryProviderAgreementRepository()
	ctx := context.Background()

	repo.Save(models.ProviderAgreement{ID: "agree-001", CarrierID: "carrier-001", ProviderID: "provider-twilio"})
	repo.Save(models.ProviderAgreement{ID: "agree-002", CarrierID: "carrier-001", ProviderID: "provider-vonage"})
	repo.Save(models.ProviderAgreement{ID: "agree-003", CarrierID: "carrier-002", ProviderID: "provider-twilio"})

	agreements, err := repo.FindManyByCarrierId(ctx, "carrier-001")
	require.NoError(t, err)
	assert.Len(t, agreements, 2)
}

func TestProviderAgreementRepository_FindMany_FiltersOtherCarriers(t *testing.T) {
	repo := NewInMemoryProviderAgreementRepository()
	ctx := context.Background()

	repo.Save(models.ProviderAgreement{ID: "agree-001", CarrierID: "carrier-001", ProviderID: "provider-twilio"})
	repo.Save(models.ProviderAgreement{ID: "agree-002", CarrierID: "carrier-002", ProviderID: "provider-vonage"})

	agreements, err := repo.FindManyByCarrierId(ctx, "carrier-002")
	require.NoError(t, err)
	require.Len(t, agreements, 1)
	assert.Equal(t, "agree-002", agreements[0].ID)
}

func TestProviderAgreementRepository_FindMany_EmptyResult(t *testing.T) {
	repo := NewInMemoryProviderAgreementRepository()
	ctx := context.Background()

	agreements, err := repo.FindManyByCarrierId(ctx, "carrier-nonexistent")
	require.NoError(t, err)
	assert.Empty(t, agreements)
}

func TestProviderAgreementRepository_FindMany_EmptyRepo(t *testing.T) {
	repo := NewInMemoryProviderAgreementRepository()
	ctx := context.Background()

	agreements, err := repo.FindManyByCarrierId(ctx, "carrier-001")
	require.NoError(t, err)
	assert.Empty(t, agreements)
}
