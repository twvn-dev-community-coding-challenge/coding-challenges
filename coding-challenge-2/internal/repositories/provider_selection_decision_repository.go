package repositories

import (
	"context"
	"sync"

	"sms-service/internal/models"
)

type ProviderSelectionDecisionRepository interface {
	Save(ctx context.Context, decision models.ProviderSelectionDecision) error
	GetBySMSMessageID(ctx context.Context, smsMessageID string) (models.ProviderSelectionDecision, error)
}

type InMemoryProviderSelectionDecisionRepository struct {
	mu        sync.RWMutex
	decisions map[string]models.ProviderSelectionDecision
}

func NewInMemoryProviderSelectionDecisionRepository() ProviderSelectionDecisionRepository {
	return &InMemoryProviderSelectionDecisionRepository{
		decisions: make(map[string]models.ProviderSelectionDecision),
	}
}

func (r *InMemoryProviderSelectionDecisionRepository) Save(ctx context.Context, decision models.ProviderSelectionDecision) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.decisions[decision.SMSMessageID] = decision
	return nil
}

func (r *InMemoryProviderSelectionDecisionRepository) GetBySMSMessageID(ctx context.Context, smsMessageID string) (models.ProviderSelectionDecision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, exists := r.decisions[smsMessageID]
	if !exists {
		return models.ProviderSelectionDecision{}, models.ErrNotFound
	}
	return d, nil
}
