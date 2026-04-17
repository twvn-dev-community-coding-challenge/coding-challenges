package repositories

import (
	"context"
	"sync"

	"sms-service/internal/models"
)

type ProviderAgreementRepository interface {
	Save(agreement models.ProviderAgreement) error
	FindManyByCarrierId(ctx context.Context, carrierID string) ([]models.ProviderAgreement, error)
}

type InMemoryProviderAgreementRepository struct {
	mu         sync.RWMutex
	agreements []models.ProviderAgreement
}

func NewInMemoryProviderAgreementRepository() ProviderAgreementRepository {
	return &InMemoryProviderAgreementRepository{}
}

func (r *InMemoryProviderAgreementRepository) Save(agreement models.ProviderAgreement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agreements = append(r.agreements, agreement)
	return nil
}

func (r *InMemoryProviderAgreementRepository) FindManyByCarrierId(ctx context.Context, carrierID string) ([]models.ProviderAgreement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var agreements []models.ProviderAgreement
	for _, a := range r.agreements {
		if a.CarrierID == carrierID {
			agreements = append(agreements, a)
		}
	}
	return agreements, nil
}
