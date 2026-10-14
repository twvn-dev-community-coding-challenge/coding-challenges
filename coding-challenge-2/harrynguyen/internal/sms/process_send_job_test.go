package sms

import (
	"context"
	"testing"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
)

func TestProcessSendJob_IdempotentWhenNotNew(t *testing.T) {
	repo := NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSServiceWithImmediateDispatch(repo, resolver, router, nil)
	ctx := WithSendSource(context.Background(), SendSourceAPI)
	msg, err := svc.SendSMS(ctx, "PH", "6391712345678", "x")
	if err != nil {
		t.Fatal(err)
	}
	// After immediate dispatch, status is no longer New
	if err := svc.ProcessSendJob(ctx, SendJob{MessageID: msg.ID}); err != nil {
		t.Fatal(err)
	}
}
