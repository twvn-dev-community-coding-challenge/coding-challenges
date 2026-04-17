package repositories

import (
	"context"
	"sync"

	"sms-service/internal/models"
)

type ProviderRepository interface {
	Save(provider models.Provider) error
	GetByID(ctx context.Context, id string) (models.Provider, error)
}

type InMemoryProviderRepository struct {
	mu        sync.RWMutex
	providers map[string]models.Provider
}

func NewInMemoryProviderRepository() ProviderRepository {
	return &InMemoryProviderRepository{
		providers: make(map[string]models.Provider),
	}
}

func (r *InMemoryProviderRepository) Save(provider models.Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.ID] = provider
	return nil
}

func (r *InMemoryProviderRepository) GetByID(ctx context.Context, id string) (models.Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.providers[id]
	if !exists {
		return models.Provider{}, models.ErrNotFound
	}
	return p, nil
}
