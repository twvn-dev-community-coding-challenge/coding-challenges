package sms

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dotdak/sms-otp/internal/providers"
)

var (
	ErrNotFound = errors.New("not found")
)

// Repository defines the operations for SMS persistence.
type Repository interface {
	Create(ctx context.Context, msg *providers.SMSMessage) error
	Update(ctx context.Context, msg *providers.SMSMessage) error
	GetByID(ctx context.Context, id string) (*providers.SMSMessage, error)
	GetByMessageID(ctx context.Context, messageID string) (*providers.SMSMessage, error)
	ListMessages(ctx context.Context, params MessageListParams) ([]*providers.SMSMessage, error)

	AddStatusLog(ctx context.Context, log *providers.StatusLog) error
	GetStatusLogs(ctx context.Context, smsID string) ([]*providers.StatusLog, error)
}

// InMemRepository is an in-memory implementation of the Repository.
type InMemRepository struct {
	mu         sync.RWMutex
	createSeq  atomic.Uint64
	messages   map[string]*providers.SMSMessage
	messageIDs map[string]string // message_id -> internal id mapping
	statusLogs map[string][]*providers.StatusLog
}

// NewInMemRepository creates a new instance of InMemRepository.
func NewInMemRepository() *InMemRepository {
	return &InMemRepository{
		messages:   make(map[string]*providers.SMSMessage),
		messageIDs: make(map[string]string),
		statusLogs: make(map[string][]*providers.StatusLog),
	}
}

func (r *InMemRepository) Create(ctx context.Context, msg *providers.SMSMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if msg.ID == "" {
		// Monotonic suffix: UnixNano alone collides under heavy concurrent Create on the same host.
		msg.ID = fmt.Sprintf("sms_%d_%d", time.Now().UnixNano(), r.createSeq.Add(1))
	}
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()

	r.messages[msg.ID] = msg
	if msg.MessageID != "" {
		r.messageIDs[msg.MessageID] = msg.ID
	}
	return nil
}

func (r *InMemRepository) Update(ctx context.Context, msg *providers.SMSMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.messages[msg.ID]; !ok {
		return ErrNotFound
	}
	msg.UpdatedAt = time.Now()
	r.messages[msg.ID] = msg
	if msg.MessageID != "" {
		r.messageIDs[msg.MessageID] = msg.ID
	}
	return nil
}

func (r *InMemRepository) GetByID(ctx context.Context, id string) (*providers.SMSMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, ok := r.messages[id]
	if !ok {
		return nil, ErrNotFound
	}
	return msg, nil
}

func (r *InMemRepository) GetByMessageID(ctx context.Context, messageID string) (*providers.SMSMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.messageIDs[messageID]
	if !ok {
		return nil, ErrNotFound
	}
	return r.messages[id], nil
}

func (r *InMemRepository) ListMessages(ctx context.Context, params MessageListParams) ([]*providers.SMSMessage, error) {
	_ = ctx
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs := make([]*providers.SMSMessage, 0, len(r.messages))
	for _, m := range r.messages {
		if params.Status != "" && m.Status != params.Status {
			continue
		}
		if params.Phone != "" && !phoneFilterMatch(m.PhoneNumber, params.Phone) {
			continue
		}
		if params.Since != nil && m.CreatedAt.Before(*params.Since) {
			continue
		}
		msgs = append(msgs, m)
	}
	// Newest first (match Postgres ordering)
	slicesSortSMSByCreatedDesc(msgs)
	start := params.Offset
	if start < 0 {
		start = 0
	}
	if start > len(msgs) {
		return []*providers.SMSMessage{}, nil
	}
	msgs = msgs[start:]
	if params.Limit > 0 && len(msgs) > params.Limit {
		msgs = msgs[:params.Limit]
	}
	return msgs, nil
}

func phoneFilterMatch(phoneNumber, filter string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(phoneNumber), strings.ToLower(filter))
}

func slicesSortSMSByCreatedDesc(msgs []*providers.SMSMessage) {
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].CreatedAt.After(msgs[j].CreatedAt)
	})
}

func (r *InMemRepository) AddStatusLog(ctx context.Context, log *providers.StatusLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.ID = fmt.Sprintf("log_%d", time.Now().UnixNano())
	log.Timestamp = time.Now()
	r.statusLogs[log.SMSID] = append(r.statusLogs[log.SMSID], log)
	return nil
}

func (r *InMemRepository) GetStatusLogs(ctx context.Context, smsID string) ([]*providers.StatusLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logs, ok := r.statusLogs[smsID]
	if !ok {
		return []*providers.StatusLog{}, nil
	}
	return logs, nil
}
