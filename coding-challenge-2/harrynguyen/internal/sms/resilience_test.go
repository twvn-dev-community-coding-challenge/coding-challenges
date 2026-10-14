package sms

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/ratelimit"
	"github.com/redis/go-redis/v9"
)

// Resilience and load tests prove behavior under flaky providers and concurrency.
// Run: go test ./internal/sms/... -run '^(TestResilience|TestLoad)_' -v
// Race: go test ./internal/sms/... -race -run '^TestResilience_ConcurrentSendSMS$'

func fastMock(name providers.Provider) *providers.MockProvider {
	return &providers.MockProvider{Name: name, SkipDelay: true, Quiet: true}
}

func phGlobePhone(seq int) string {
	return fmt.Sprintf("63917%07d", seq%10_000_000)
}

func newResilienceService(tb testing.TB, provider providers.SMSProvider, limiter *ratelimit.Limiter) (*SMSCoreServiceInstance, *InMemRepository) {
	tb.Helper()
	repo := NewInMemRepository()
	resolver := carrier.NewPrefixCarrierResolver()
	router := providers.NewSimpleProviderRouter()
	router.RegisterRouter("PH", providers.CarrierGlobe, providers.ProviderMessageBird, provider)
	svc := NewSMSServiceWithImmediateDispatch(repo, resolver, router, limiter)
	return svc, repo
}

// TestResilience_FlakyProviderEventualSuccess shows that after N simulated provider failures,
// the next SendSMS completes and reaches Queue (each attempt is a distinct outbound message).
func TestResilience_FlakyProviderEventualSuccess(t *testing.T) {
	ctx := context.Background()
	failN := 5
	inner := fastMock(providers.ProviderMessageBird)
	flaky := providers.NewFlakyProvider(inner, failN, errors.New("upstream timeout"))
	svc, _ := newResilienceService(t, flaky, nil)

	var failures, successes int
	for i := 0; i < failN+3; i++ {
		_, err := svc.SendSMS(ctx, "PH", phGlobePhone(i+1), "resilience probe")
		if err != nil {
			failures++
			continue
		}
		successes++
	}

	if failures != failN {
		t.Fatalf("want %d failed sends from flaky provider, got %d", failN, failures)
	}
	if successes < 3 {
		t.Fatalf("want at least 3 successes after failures, got %d", successes)
	}
	if flaky.Calls() != failN+3 {
		t.Fatalf("provider calls = %d, want %d", flaky.Calls(), failN+3)
	}
}

// TestResilience_FailedSendLeavesRecoverableState checks that a failed provider call
// moves the message through Send-failed and automatic fallback to Send-to-provider.
func TestResilience_FailedSendLeavesRecoverableState(t *testing.T) {
	ctx := context.Background()
	flaky := providers.NewFlakyProvider(fastMock(providers.ProviderMessageBird), 1, errors.New("boom"))
	svc, repo := newResilienceService(t, flaky, nil)

	_, err := svc.SendSMS(ctx, "PH", phGlobePhone(42), "x")
	if err == nil {
		t.Fatal("expected first send to fail")
	}

	msgs, err := repo.ListMessages(ctx, MessageListParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("messages = %d, want 1", len(msgs))
	}
	if msgs[0].Status != providers.StatusSendToProvider {
		t.Fatalf("after failure, want status Send-to-provider for retry routing, got %q", msgs[0].Status)
	}
}

// TestResilience_ConcurrentSendSMS exercises the in-memory repository and service under parallel load.
func TestResilience_ConcurrentSendSMS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent load test in -short mode")
	}

	ctx := context.Background()
	const workers = 256
	svc, repo := newResilienceService(t, fastMock(providers.ProviderMessageBird), nil)

	var wg sync.WaitGroup
	var ok atomic.Int32
	var fail atomic.Int32
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			_, err := svc.SendSMS(ctx, "PH", phGlobePhone(10_000+seq), fmt.Sprintf("c%d", seq))
			if err != nil {
				fail.Add(1)
				return
			}
			ok.Add(1)
		}(i)
	}
	wg.Wait()

	if int(ok.Load()) != workers {
		t.Fatalf("successful sends = %d, want %d (failures=%d)", ok.Load(), workers, fail.Load())
	}
	msgs, err := repo.ListMessages(ctx, MessageListParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != workers {
		t.Fatalf("persisted messages = %d, want %d", len(msgs), workers)
	}
	queued := 0
	for _, m := range msgs {
		if m.Status == providers.StatusQueue {
			queued++
		}
	}
	if queued != workers {
		t.Fatalf("messages in Queue = %d, want %d", queued, workers)
	}
}

// TestResilience_ConcurrentDeliveryCallbacks fires duplicate delivery webhooks in parallel;
// the message must end in Send-success without panic (some callbacks may error on transitions).
func TestResilience_ConcurrentDeliveryCallbacks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent callback test in -short mode")
	}

	ctx := context.Background()
	svc, repo := newResilienceService(t, fastMock(providers.ProviderMessageBird), nil)
	msg, err := svc.SendSMS(ctx, "PH", phGlobePhone(5000), "cb storm")
	if err != nil {
		t.Fatal(err)
	}
	if msg.Status != providers.StatusQueue {
		t.Fatalf("precondition: want Queue, got %s", msg.Status)
	}
	provID := msg.MessageID

	const workers = 64
	var wg sync.WaitGroup
	var errs atomic.Int32
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := svc.HandleCallback(ctx, provID, providers.StatusSendSuccess, 0.01); err != nil {
				errs.Add(1)
			}
		}()
	}
	wg.Wait()

	final, err := repo.GetByID(ctx, msg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if final.Status != providers.StatusSendSuccess {
		t.Fatalf("after concurrent callbacks, want Send-success, got %q (parallel errs=%d)", final.Status, errs.Load())
	}
}

// TestResilience_RedisRateLimitUnderConcurrency shows per-phone limits stay consistent under parallel SendSMS.
func TestResilience_RedisRateLimitUnderConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping redis load test in -short mode")
	}

	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	lim := ratelimit.New(rdb)
	ctx := context.Background()

	sharedPhone := phGlobePhone(777)
	svc, _ := newResilienceService(t, fastMock(providers.ProviderMessageBird), lim)

	const perHour = 10
	const extra = 4
	var wg sync.WaitGroup
	var allowed atomic.Int32
	var limited atomic.Int32

	for i := 0; i < perHour+extra; i++ {
		wg.Add(1)
		go func(seq int) {
			defer wg.Done()
			_, err := svc.SendSMS(ctx, "PH", sharedPhone, fmt.Sprintf("rate %d", seq))
			if err == nil {
				allowed.Add(1)
				return
			}
			var le *ratelimit.LimitedError
			if errors.As(err, &le) {
				limited.Add(1)
			}
		}(i)
	}
	wg.Wait()

	if int(allowed.Load()) != perHour {
		t.Fatalf("allowed sends for same phone = %d, want %d", allowed.Load(), perHour)
	}
	if int(limited.Load()) != extra {
		t.Fatalf("rate-limited sends = %d, want %d", limited.Load(), extra)
	}
}

// TestLoad_SustainedThroughput runs many sequential sends to approximate steady-state throughput (not a benchmark).
func TestLoad_SustainedThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sustained load test in -short mode")
	}

	ctx := context.Background()
	const n = 400
	svc, _ := newResilienceService(t, fastMock(providers.ProviderMessageBird), nil)

	start := time.Now()
	for i := 0; i < n; i++ {
		if _, err := svc.SendSMS(ctx, "PH", phGlobePhone(20_000+i), "load"); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	t.Logf("%d sequential SendSMS in %v (%.0f msg/s)", n, elapsed, float64(n)/elapsed.Seconds())
}
