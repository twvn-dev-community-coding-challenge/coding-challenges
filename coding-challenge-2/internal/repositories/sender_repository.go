package repositories

import (
	"context"
	"sync"

	"sms-service/internal/models"
)

type SenderRepository interface {
	Save(sender models.Sender) error
	GetByID(ctx context.Context, id string) (models.Sender, error)
}

type InMemorySenderRepository struct {
	mu      sync.RWMutex
	senders map[string]models.Sender
}

func NewInMemorySenderRepository() SenderRepository {
	return &InMemorySenderRepository{
		senders: make(map[string]models.Sender),
	}
}

func (r *InMemorySenderRepository) Save(sender models.Sender) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.senders[sender.ID] = sender
	return nil
}

func (r *InMemorySenderRepository) GetByID(ctx context.Context, id string) (models.Sender, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, exists := r.senders[id]
	if !exists {
		return models.Sender{}, models.ErrNotFound
	}
	return s, nil
}
