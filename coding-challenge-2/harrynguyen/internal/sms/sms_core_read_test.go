package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
)

type logsErrRepo struct{ *InMemRepository }

func (r *logsErrRepo) GetStatusLogs(ctx context.Context, smsID string) ([]*providers.StatusLog, error) {
	return nil, errors.New("logs unavailable")
}

func TestSMSCoreService_GetMessageAndList(t *testing.T) {
	repo := NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSServiceWithImmediateDispatch(repo, resolver, router, nil)
	ctx := WithSendSource(context.Background(), SendSourceAPI)

	msg, err := svc.SendSMS(ctx, "PH", "6391712345001", "body")
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetMessage(ctx, msg.ID)
	if err != nil || got == nil || got.ID != msg.ID {
		t.Fatalf("GetMessage err=%v got=%v", err, got)
	}

	got2, logs, err := svc.GetMessageWithLogs(ctx, msg.ID)
	if err != nil || got2 == nil || logs == nil {
		t.Fatalf("GetMessageWithLogs err=%v", err)
	}

	list, err := svc.ListMessages(ctx, MessageListParams{Limit: 50, Since: ptrTime(time.Now().Add(-time.Hour))})
	if err != nil || len(list) < 1 {
		t.Fatalf("ListMessages err=%v n=%d", err, len(list))
	}

	svc.RegisterGlobalObserver(readTestObs{})
	svc.RegisterSourceObserver(SendSourceAPI, readTestObs{})
}

func TestSMSCoreService_RegisterSourceObserverEmptyIgnored(t *testing.T) {
	repo := NewInMemRepository()
	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	svc.RegisterSourceObserver("", readTestObs{}) // must not panic or add empty-key bucket
}

func TestSMSCoreService_GetMessageWithLogs_StatusLogsError(t *testing.T) {
	base := NewInMemRepository()
	repo := &logsErrRepo{InMemRepository: base}
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSServiceWithImmediateDispatch(repo, resolver, router, nil)
	ctx := WithSendSource(context.Background(), SendSourceAPI)
	msg, err := svc.SendSMS(ctx, "PH", "6391712345888", "body")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = svc.GetMessageWithLogs(ctx, msg.ID)
	if err == nil || err.Error() != "logs unavailable" {
		t.Fatalf("expected logs error, got %v", err)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

type readTestObs struct{}

func (readTestObs) OnMessageUpdated(context.Context, *providers.SMSMessage) {}
