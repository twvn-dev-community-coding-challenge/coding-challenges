package sms

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
)

type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, SendJob) error { return nil }

type errSendProvider struct{ name providers.Provider }

func (e *errSendProvider) GetProviderName() providers.Provider { return e.name }
func (e *errSendProvider) Send(context.Context, string, providers.Carrier, string, string) (string, float64, error) {
	return "", 0, errors.New("simulated provider send failure")
}

// eachPipeline runs subtests for ProcessSendJob and RunOutboundJob (same outbound logic, different entrypoints).
func eachPipeline(t *testing.T, fn func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error)) {
	t.Helper()
	t.Run("ProcessSendJob", func(t *testing.T) {
		fn(t, func(s *SMSCoreServiceInstance, ctx context.Context, job SendJob) error {
			return s.ProcessSendJob(ctx, job)
		})
	})
	t.Run("RunOutboundJob", func(t *testing.T) {
		fn(t, func(s *SMSCoreServiceInstance, ctx context.Context, job SendJob) error {
			return RunOutboundJob(ctx, s, job)
		})
	})
}

func TestOutboundPipeline_LoadNotFound(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		err := run(svc, ctx, SendJob{MessageID: "does-not-exist"})
		if err == nil || !strings.Contains(err.Error(), "load message") {
			t.Fatalf("expected load error, got %v", err)
		}
	})
}

func TestOutboundPipeline_CarrierUnknown(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		msg := &providers.SMSMessage{
			ID: "sms-carrier-x", Country: "PH", PhoneNumber: "6390000000000", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		err := run(svc, ctx, SendJob{MessageID: msg.ID})
		if err == nil || !strings.Contains(err.Error(), "carrier resolution failed") {
			t.Fatalf("expected carrier resolution error, got %v", err)
		}
		updated, _ := repo.GetByID(ctx, msg.ID)
		// SendFailed triggers automatic fallback to Send-to-provider (UpdateStatus).
		if updated.Status != providers.StatusSendToProvider {
			t.Fatalf("status=%q want Send-to-provider", updated.Status)
		}
	})
}

func TestOutboundPipeline_RoutingFailed(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		msg := &providers.SMSMessage{
			ID: "sms-route-x", Country: "PH", PhoneNumber: "6391712345001", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		err := run(svc, ctx, SendJob{MessageID: msg.ID})
		if err == nil || !strings.Contains(err.Error(), "routing failed") {
			t.Fatalf("expected routing error, got %v", err)
		}
	})
}

func TestOutboundPipeline_MissingAdapter(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, nil)
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		msg := &providers.SMSMessage{
			ID: "sms-adapter-x", Country: "PH", PhoneNumber: "6391712345002", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		err := run(svc, ctx, SendJob{MessageID: msg.ID})
		if err == nil || !strings.Contains(err.Error(), "missing adapter") {
			t.Fatalf("expected missing adapter error, got %v", err)
		}
	})
}

func TestOutboundPipeline_TestModeAlwaysFail(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		msg := &providers.SMSMessage{
			ID: "sms-failmode", Country: "PH", PhoneNumber: "6391712345003", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		job := SendJob{MessageID: msg.ID, TestSMSMode: TestSMSModeAlwaysFail}
		err := run(svc, ctx, job)
		if err == nil || !strings.Contains(err.Error(), "forced failure") {
			t.Fatalf("expected forced failure, got %v", err)
		}
	})
}

func TestOutboundPipeline_ProviderSendError(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &errSendProvider{name: providers.ProviderMessageBird})
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		msg := &providers.SMSMessage{
			ID: "sms-senderr", Country: "PH", PhoneNumber: "6391712345004", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		err := run(svc, ctx, SendJob{MessageID: msg.ID})
		if err == nil || !strings.Contains(err.Error(), "provider send failed") {
			t.Fatalf("expected provider send error, got %v", err)
		}
	})
}

func TestOutboundPipeline_TestModeAlwaysSuccessBypassesProviderError(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &errSendProvider{name: providers.ProviderMessageBird})
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		msg := &providers.SMSMessage{
			ID: "sms-bypass", Country: "PH", PhoneNumber: "6391712345005", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		job := SendJob{MessageID: msg.ID, TestSMSMode: TestSMSModeAlwaysSuccess}
		if err := run(svc, ctx, job); err != nil {
			t.Fatal(err)
		}
		updated, _ := repo.GetByID(ctx, msg.ID)
		if updated.Status != providers.StatusQueue {
			t.Fatalf("status=%q want Queue", updated.Status)
		}
		if !strings.HasPrefix(updated.MessageID, "forced-success-") {
			t.Fatalf("unexpected provider message id %q", updated.MessageID)
		}
	})
}

func TestOutboundPipeline_TransientModeAddsLogThenSucceeds(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := context.Background()
		msg := &providers.SMSMessage{
			ID: "sms-transient", Country: "PH", PhoneNumber: "6391712345006", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		job := SendJob{MessageID: msg.ID, TestSMSMode: TestSMSModeTransientFailureOnce}
		if err := run(svc, ctx, job); err != nil {
			t.Fatal(err)
		}
		logs, err := repo.GetStatusLogs(ctx, msg.ID)
		if err != nil {
			t.Fatal(err)
		}
		var sawForced bool
		for _, lg := range logs {
			if strings.Contains(lg.Metadata, "transient-failure-once") {
				sawForced = true
				break
			}
		}
		if !sawForced {
			t.Fatal("expected transient-failure-once status log")
		}
	})
}

func TestOutboundPipeline_HappyPathThenIdempotent(t *testing.T) {
	eachPipeline(t, func(t *testing.T, run func(*SMSCoreServiceInstance, context.Context, SendJob) error) {
		repo := NewInMemRepository()
		resolver := carrier.NewPrefixCarrierResolver()
		router := providers.NewSimpleProviderRouter()
		router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
		svc := NewSMSService(repo, resolver, router, nil, noopPublisher{})
		ctx := WithSendSource(context.Background(), SendSourceAPI)
		msg := &providers.SMSMessage{
			ID: "sms-happy", Country: "PH", PhoneNumber: "6391712345007", Content: "hi",
			Status: providers.StatusNew,
		}
		if err := repo.Create(ctx, msg); err != nil {
			t.Fatal(err)
		}
		job := SendJob{MessageID: msg.ID}
		if err := run(svc, ctx, job); err != nil {
			t.Fatal(err)
		}
		updated, _ := repo.GetByID(ctx, msg.ID)
		if updated.Status != providers.StatusQueue {
			t.Fatalf("first run: status=%q", updated.Status)
		}
		if err := run(svc, ctx, job); err != nil {
			t.Fatal(err)
		}
	})
}
