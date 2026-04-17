package repositories

import (
	"context"
	"log"
	"sync"

	"sms-service/internal/models"
)

// SMSMessageRepository abstracts persistence of SMS messages.
type SMSMessageRepository interface {
	GetById(ctx context.Context, messageID string) (models.SMSMessage, error)
	Save(ctx context.Context, message models.SMSMessage) error
}

type InMemorySMSRepository struct {
	mu       sync.RWMutex
	messages map[string]models.SMSMessage // Fake indexing hashmap, retrieve by ID
}

// NewInMemorySMSMessageRepository Constructor (singleton-like usage if you reuse it)
func NewInMemorySMSMessageRepository() SMSMessageRepository {
	return &InMemorySMSRepository{
		messages: make(map[string]models.SMSMessage),
	}
}

func (repository *InMemorySMSRepository) GetById(ctx context.Context, id string) (models.SMSMessage, error) {

	// Lock the mutex
	repository.mu.RLock()

	// Release the mutex when the function returns
	defer repository.mu.RUnlock()

	msg, exists := repository.messages[id]
	if !exists {
		log.Printf("[SMSRepo] GetById(%s): not found", id)
		return models.SMSMessage{}, models.ErrNotFound
	}

	log.Printf("[SMSRepo] GetById(%s): found, status=%s, provider=%v", id, msg.Status, msg.ProviderID)
	return msg, nil
}

func (repository *InMemorySMSRepository) Save(ctx context.Context, message models.SMSMessage) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	_, exists := repository.messages[message.ID]
	if !exists {
		log.Printf("[SMSRepo] Save(%s): new message, status=%s", message.ID, message.Status)
		repository.messages[message.ID] = message
		return nil
	}

	log.Printf("[SMSRepo] Save(%s): updating, status=%s -> %s, provider=%v", message.ID, repository.messages[message.ID].Status, message.Status, message.ProviderID)
	existingMessage := repository.messages[message.ID]
	existingMessage.EstimatedCost = message.EstimatedCost
	existingMessage.Status = message.Status
	existingMessage.ActualCost = message.ActualCost
	existingMessage.FailureReason = message.FailureReason
	existingMessage.ProviderID = message.ProviderID
	existingMessage.UpdatedAt = message.UpdatedAt
	repository.messages[message.ID] = existingMessage
	return nil
}
