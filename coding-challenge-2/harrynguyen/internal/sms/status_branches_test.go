package sms

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
)

type addLogFailRepo struct {
	*InMemRepository
	fail bool
}

func (r *addLogFailRepo) AddStatusLog(ctx context.Context, log *providers.StatusLog) error {
	if r.fail {
		return errors.New("status log write failed")
	}
	return r.InMemRepository.AddStatusLog(ctx, log)
}

func TestAddStatusLog_RepoErrorOnlyLogs(t *testing.T) {
	base := NewInMemRepository()
	repo := &addLogFailRepo{InMemRepository: base, fail: true}
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
	ctx := WithSendSource(context.Background(), SendSourceAPI)
	// Create + first AddStatusLog ("Message created") hits failing repo path.
	if _, err := svc.SendSMS(ctx, "PH", "6391712347777", "hi"); err != nil {
		t.Fatal(err)
	}
}

type countingObserver struct{ n *atomic.Int32 }

func (c countingObserver) OnMessageUpdated(context.Context, *providers.SMSMessage) {
	c.n.Add(1)
}

func TestNotifyObservers_SourceScoped(t *testing.T) {
	var srcN, globalN atomic.Int32
	repo := NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
	svc.RegisterGlobalObserver(countingObserver{n: &globalN})
	svc.RegisterSourceObserver(SendSourceAPI, countingObserver{n: &srcN})

	ctx := WithSendSource(context.Background(), SendSourceAPI)
	msg := &providers.SMSMessage{
		ID: "sms-obs", Country: "PH", PhoneNumber: "6391712347778", Content: "x",
		Status: providers.StatusNew, SendSource: SendSourceAPI,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}
	if err := svc.ProcessSendJob(ctx, SendJob{MessageID: msg.ID}); err != nil {
		t.Fatal(err)
	}
	if srcN.Load() < 1 {
		t.Fatalf("source observer calls=%d", srcN.Load())
	}
	if globalN.Load() < 1 {
		t.Fatalf("global observer calls=%d", globalN.Load())
	}
}
