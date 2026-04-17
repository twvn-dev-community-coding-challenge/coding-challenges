package repositories

import (
	"context"
	"strings"
	"sync"

	"sms-service/internal/models"
)

type CarrierRepository interface {
	Save(carrier models.Carrier) error
	FindByCountryAndPhone(ctx context.Context, countryID string, phone string) (models.Carrier, error)
}

type InMemoryCarrierRepository struct {
	mu       sync.RWMutex
	carriers []models.Carrier
}

func NewInMemoryCarrierRepository() CarrierRepository {
	return &InMemoryCarrierRepository{}
}

func (r *InMemoryCarrierRepository) Save(carrier models.Carrier) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.carriers = append(r.carriers, carrier)
	return nil
}

func (r *InMemoryCarrierRepository) FindByCountryAndPhone(ctx context.Context, countryID string, phone string) (models.Carrier, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.carriers {
		if c.CountryId != countryID {
			continue
		}
		for _, prefix := range c.PhoneNumberPrefix {
			if strings.HasPrefix(phone, prefix) {
				return c, nil
			}
		}
	}
	return models.Carrier{}, models.ErrNotFound
}
