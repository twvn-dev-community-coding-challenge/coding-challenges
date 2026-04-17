package repositories

import (
	"sync"

	"sms-service/internal/models"
)

type CountryRepository interface {
	GetCountryByCode(countryCode string) (*models.Country, error)
	Save(country *models.Country) error
}

type InMemoryCountryRepository struct {
	mu        sync.RWMutex
	countries map[string]*models.Country // keyed by CountryCode
}

func NewInMemoryCountryRepository() CountryRepository {
	return &InMemoryCountryRepository{
		countries: make(map[string]*models.Country),
	}
}

func (r *InMemoryCountryRepository) Save(country *models.Country) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.countries[country.CountryCode] = country
	return nil
}

func (r *InMemoryCountryRepository) GetCountryByCode(countryCode string) (*models.Country, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, exists := r.countries[countryCode]
	if !exists {
		return nil, models.ErrNotFound
	}
	return c, nil
}
