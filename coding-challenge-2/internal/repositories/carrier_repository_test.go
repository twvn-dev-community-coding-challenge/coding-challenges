package repositories

import (
	"context"
	"testing"

	"sms-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCarrierRepository_FindByCountryAndPhone_Found(t *testing.T) {
	repo := NewInMemoryCarrierRepository()
	ctx := context.Background()

	carrier := models.Carrier{
		ID:                "carrier-001",
		Name:              "Viettel",
		CountryId:         "country-001",
		PhoneNumberPrefix: []string{"93", "96", "98"},
	}
	require.NoError(t, repo.Save(carrier))

	got, err := repo.FindByCountryAndPhone(ctx, "country-001", "931234567")
	require.NoError(t, err)
	assert.Equal(t, "carrier-001", got.ID)
	assert.Equal(t, "Viettel", got.Name)
}

func TestCarrierRepository_FindByCountryAndPhone_MatchesSecondPrefix(t *testing.T) {
	repo := NewInMemoryCarrierRepository()
	ctx := context.Background()

	repo.Save(models.Carrier{ID: "carrier-001", CountryId: "c-001", PhoneNumberPrefix: []string{"93", "96", "98"}})

	got, err := repo.FindByCountryAndPhone(ctx, "c-001", "961234567")
	require.NoError(t, err)
	assert.Equal(t, "carrier-001", got.ID)
}

func TestCarrierRepository_FindByCountryAndPhone_SelectsCorrectAmongMultiple(t *testing.T) {
	repo := NewInMemoryCarrierRepository()
	ctx := context.Background()

	repo.Save(models.Carrier{ID: "carrier-001", CountryId: "c-001", PhoneNumberPrefix: []string{"93"}})
	repo.Save(models.Carrier{ID: "carrier-002", CountryId: "c-001", PhoneNumberPrefix: []string{"96"}})

	got, err := repo.FindByCountryAndPhone(ctx, "c-001", "961234567")
	require.NoError(t, err)
	assert.Equal(t, "carrier-002", got.ID)
}

func TestCarrierRepository_FindByCountryAndPhone_WrongCountry(t *testing.T) {
	repo := NewInMemoryCarrierRepository()
	ctx := context.Background()

	repo.Save(models.Carrier{ID: "carrier-001", CountryId: "country-001", PhoneNumberPrefix: []string{"93"}})

	_, err := repo.FindByCountryAndPhone(ctx, "country-002", "931234567")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestCarrierRepository_FindByCountryAndPhone_WrongPrefix(t *testing.T) {
	repo := NewInMemoryCarrierRepository()
	ctx := context.Background()

	repo.Save(models.Carrier{ID: "carrier-001", CountryId: "country-001", PhoneNumberPrefix: []string{"93"}})

	_, err := repo.FindByCountryAndPhone(ctx, "country-001", "701234567")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestCarrierRepository_FindByCountryAndPhone_EmptyRepo(t *testing.T) {
	repo := NewInMemoryCarrierRepository()
	ctx := context.Background()

	_, err := repo.FindByCountryAndPhone(ctx, "country-001", "931234567")
	assert.ErrorIs(t, err, models.ErrNotFound)
}
