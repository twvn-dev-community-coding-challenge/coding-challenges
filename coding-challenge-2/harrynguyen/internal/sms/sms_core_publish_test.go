package sms

import (
	"context"
	"errors"
	"testing"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
)

type errPublisher struct{}

func (errPublisher) Publish(ctx context.Context, job SendJob) error {
	return errors.New("enqueue failed")
}

func TestSendSMS_NoPublisherReturnsError(t *testing.T) {
	repo := NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSService(repo, resolver, router, nil, nil)
	ctx := WithSendSource(context.Background(), SendSourceAPI)
	_, err := svc.SendSMS(ctx, "PH", "6391712345678", "hello")
	if err == nil {
		t.Fatal("expected error when publisher is nil")
	}
}

func TestSendSMS_PublishFails(t *testing.T) {
	repo := NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSService(repo, resolver, router, nil, errPublisher{})
	ctx := WithSendSource(context.Background(), SendSourceAPI)
	_, err := svc.SendSMS(ctx, "PH", "6391712345678", "hello")
	if err == nil {
		t.Fatal("expected enqueue error")
	}
}

func TestSendSMS_NotifyMessageCreatedHook(t *testing.T) {
	var calls int
	SetTelemetryHooks(func() { calls++ }, nil)
	t.Cleanup(func() { SetTelemetryHooks(nil, nil) })

	repo := NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
	ctx := WithSendSource(context.Background(), SendSourceAPI)
	if _, err := svc.SendSMS(ctx, "PH", "6391712345999", "hello"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("message created hook calls=%d want 1", calls)
	}
}

