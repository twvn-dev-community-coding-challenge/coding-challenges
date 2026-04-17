package repositories

import (
	"testing"

	"sms-service/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountryRepository_SaveAndGetByCode(t *testing.T) {
	repo := NewInMemoryCountryRepository()

	country := &models.Country{ID: "c-001", Name: "Vietnam", CountryCode: "84"}
	require.NoError(t, repo.Save(country))

	got, err := repo.GetCountryByCode("84")
	require.NoError(t, err)
	assert.Equal(t, "c-001", got.ID)
	assert.Equal(t, "Vietnam", got.Name)
	assert.Equal(t, "84", got.CountryCode)
}

func TestCountryRepository_GetCountryByCode_NotFound(t *testing.T) {
	repo := NewInMemoryCountryRepository()

	_, err := repo.GetCountryByCode("999")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestCountryRepository_GetCountryByCode_EmptyRepo(t *testing.T) {
	repo := NewInMemoryCountryRepository()

	_, err := repo.GetCountryByCode("84")
	assert.ErrorIs(t, err, models.ErrNotFound)
}

func TestCountryRepository_Save_OverwritesSameCode(t *testing.T) {
	repo := NewInMemoryCountryRepository()

	first := &models.Country{ID: "c-001", Name: "Vietnam", CountryCode: "84"}
	second := &models.Country{ID: "c-002", Name: "Vietnam Updated", CountryCode: "84"}
	require.NoError(t, repo.Save(first))
	require.NoError(t, repo.Save(second))

	got, err := repo.GetCountryByCode("84")
	require.NoError(t, err)
	assert.Equal(t, "c-002", got.ID)
	assert.Equal(t, "Vietnam Updated", got.Name)
}

func TestCountryRepository_MultipleCountries(t *testing.T) {
	repo := NewInMemoryCountryRepository()

	repo.Save(&models.Country{ID: "c-001", Name: "Vietnam", CountryCode: "84"})
	repo.Save(&models.Country{ID: "c-002", Name: "USA", CountryCode: "1"})

	vn, err := repo.GetCountryByCode("84")
	require.NoError(t, err)
	assert.Equal(t, "c-001", vn.ID)

	us, err := repo.GetCountryByCode("1")
	require.NoError(t, err)
	assert.Equal(t, "c-002", us.ID)
}
