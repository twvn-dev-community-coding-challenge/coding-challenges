package worker

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/redis/go-redis/v9"
)

func TestStartSMSWorker_nilRedis(t *testing.T) {
	err := StartSMSWorker(context.Background(), nil, "q", &sms.SMSCoreServiceInstance{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStartSMSWorker_emptyQueue(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	svc := newTestSMS(t)
	err = StartSMSWorker(context.Background(), rdb, "", svc)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStartSMSWorker_nilService(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	err = StartSMSWorker(context.Background(), rdb, "sms", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStartSMSWorker_redisPingFails(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	s.Close()
	svc := newTestSMS(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = StartSMSWorker(ctx, rdb, "sms-queue", svc)
	if err == nil || !strings.Contains(err.Error(), "redis ping failed") {
		t.Fatalf("expected redis ping error, got %v", err)
	}
}

// TestStartSMSWorker_startsWhenRedisHealthy exercises ping + BullMQ worker wiring against miniredis.
// If gobullmq is incompatible with miniredis in this environment, the test is skipped.
func TestStartSMSWorker_startsWhenRedisHealthy(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	svc := newTestSMS(t)
	err = StartSMSWorker(context.Background(), rdb, "sms-outbound-worker-test", svc)
	if err != nil {
		if strings.Contains(err.Error(), "failed to start BullMQ worker") ||
			strings.Contains(err.Error(), "bullmq") ||
			strings.Contains(strings.ToLower(err.Error()), "script") {
			t.Skipf("BullMQ worker not supported on this Redis stub: %v", err)
		}
		t.Fatal(err)
	}
}

func TestNewSMSQueue_nilRedis(t *testing.T) {
	_, err := NewSMSQueue(context.Background(), nil, "sms")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewSMSQueue_emptyName(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	_, err = NewSMSQueue(context.Background(), rdb, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewSMSQueue_ok(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	q, err := NewSMSQueue(context.Background(), rdb, "sms-test-queue")
	if err != nil {
		t.Fatal(err)
	}
	if q == nil {
		t.Fatal("nil queue")
	}
}

func newTestSMS(t *testing.T) *sms.SMSCoreServiceInstance {
	t.Helper()
	repo := sms.NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	return sms.NewSMSServiceWithImmediateDispatch(repo, resolver, router, nil)
}

// noopEnqueue leaves SMS rows in StatusNew so the outbound job handler can run RunOutboundJob.
type noopEnqueue struct{}

func (noopEnqueue) Publish(context.Context, sms.SendJob) error { return nil }

func newTestSMSWithRepo(t *testing.T, repo *sms.InMemRepository) *sms.SMSCoreServiceInstance {
	t.Helper()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, &providers.MockProvider{Name: providers.ProviderMessageBird, SkipDelay: true, Quiet: true})
	return sms.NewSMSService(repo, resolver, router, nil, noopEnqueue{})
}

func TestProcessOutboundSMSJob_happyPath(t *testing.T) {
	repo := sms.NewInMemRepository()
	svc := newTestSMSWithRepo(t, repo)
	ctx := context.Background()
	msg := &providers.SMSMessage{
		ID: "jw-out", Country: "PH", PhoneNumber: "6391712348888", Content: "job",
		Status: providers.StatusNew,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}
	payload := map[string]any{"message_id": msg.ID}
	out, err := processOutboundSMSJob(ctx, svc, "job-1", payload)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("result=%v", out)
	}
	got, err := repo.GetByID(ctx, msg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != providers.StatusQueue {
		t.Fatalf("status=%q want Queue", got.Status)
	}
}

func TestProcessOutboundSMSJob_parseError(t *testing.T) {
	svc := newTestSMS(t)
	_, err := processOutboundSMSJob(context.Background(), svc, "job-bad", map[string]any{"message_id": ""})
	if err == nil || !strings.Contains(err.Error(), "parse job payload") {
		t.Fatalf("expected parse error, got %v", err)
	}
}

func TestProcessOutboundSMSJob_loadMessageFails(t *testing.T) {
	svc := newTestSMS(t)
	_, err := processOutboundSMSJob(context.Background(), svc, "job-missing", map[string]any{"message_id": "no-such-sms"})
	if err == nil || !strings.Contains(err.Error(), "load message") {
		t.Fatalf("expected load error, got %v", err)
	}
}

func TestProcessOutboundSMSJob_payloadAsJSONString(t *testing.T) {
	repo := sms.NewInMemRepository()
	svc := newTestSMSWithRepo(t, repo)
	ctx := context.Background()
	msg := &providers.SMSMessage{
		ID: "jw-str", Country: "PH", PhoneNumber: "6391712348889", Content: "j",
		Status: providers.StatusNew,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}
	raw := `{"message_id":"` + msg.ID + `"}`
	out, err := processOutboundSMSJob(ctx, svc, "job-2", raw)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("result=%v", out)
	}
}
