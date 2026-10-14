package sms

import (
	"context"
	"errors"
	"testing"

	"github.com/dotdak/sms-otp/internal/providers"
)

// getByProvErrRepo returns a non-ErrNotFound error from GetByMessageID (repository failure path).
type getByProvErrRepo struct {
	*InMemRepository
}

func (r *getByProvErrRepo) GetByMessageID(ctx context.Context, messageID string) (*providers.SMSMessage, error) {
	return nil, errors.New("database unavailable")
}

func TestHandleCallback_LookupNotFound(t *testing.T) {
	repo := NewInMemRepository()
	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	ctx := context.Background()
	err := svc.HandleCallback(ctx, "unknown-provider-id", providers.StatusSendSuccess, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleCallback_LookupRepositoryError(t *testing.T) {
	base := NewInMemRepository()
	repo := &getByProvErrRepo{InMemRepository: base}
	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	ctx := context.Background()
	err := svc.HandleCallback(ctx, "any-id", providers.StatusSendSuccess, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHandleCallback_ShortPhoneMaskInLogs(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:          "sms-shortph",
		MessageID:   "prov-short",
		Country:     "PH",
		PhoneNumber: "12",
		Content:     "x",
		Carrier:     providers.CarrierGlobe,
		Provider:    providers.ProviderTwilio,
		Status:      providers.StatusQueue,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}
	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	if err := svc.HandleCallback(ctx, "prov-short", providers.StatusSendSuccess, 0.01); err != nil {
		t.Fatal(err)
	}
}
