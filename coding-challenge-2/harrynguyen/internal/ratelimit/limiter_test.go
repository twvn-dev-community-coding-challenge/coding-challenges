package ratelimit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestLimiter_Allow(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	l := New(rdb)
	ctx := context.Background()
	key := "t:key"
	window := time.Hour

	for i := 0; i < 5; i++ {
		if err := l.Allow(ctx, key, 5, window); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
	}
	err = l.Allow(ctx, key, 5, window)
	var lim *LimitedError
	if !errors.As(err, &lim) {
		t.Fatalf("expected LimitedError, got %v", err)
	}
}

func TestLimiter_Allow_incrRedisError(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	l := New(rdb)
	ctx := context.Background()

	s.SetError("ERR simulated redis failure")
	defer s.SetError("")

	err = l.Allow(ctx, "incr-fail-key", 5, time.Hour)
	if err == nil || !strings.Contains(err.Error(), "ratelimit incr") {
		t.Fatalf("expected incr wrap error, got %v", err)
	}
}

func TestLimiter_Allow_incrFailsWhenServerClosed(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	s.Close()

	l := New(rdb)
	err = l.Allow(context.Background(), "k", 5, time.Hour)
	if err == nil || !strings.Contains(err.Error(), "ratelimit incr") {
		t.Fatalf("expected incr error, got %v", err)
	}
}

// Pre-seed the counter without TTL so INCR crosses the limit on first Allow; TTL is -1 and code uses window as Retry-After.
func TestLimiter_Allow_overLimitUsesWindowWhenKeyHasNoTTL(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	l := New(rdb)
	ctx := context.Background()
	key := "preseed-no-ttl"
	window := 37 * time.Second

	if err := rdb.Set(ctx, key, "5", 0).Err(); err != nil {
		t.Fatal(err)
	}
	err = l.Allow(ctx, key, 5, window)
	var lim *LimitedError
	if !errors.As(err, &lim) {
		t.Fatalf("expected LimitedError, got %v", err)
	}
	if lim.RetryAfter != window {
		t.Fatalf("RetryAfter=%v want %v", lim.RetryAfter, window)
	}
}

func TestLimiter_Allow_retryAfterFromTTLWhenLimited(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	l := New(rdb)
	ctx := context.Background()
	key := "ttl-retry-key"
	window := time.Hour

	for i := 0; i < 5; i++ {
		if err := l.Allow(ctx, key, 5, window); err != nil {
			t.Fatalf("iter %d: %v", i, err)
		}
	}
	err = l.Allow(ctx, key, 5, window)
	var lim *LimitedError
	if !errors.As(err, &lim) {
		t.Fatalf("expected LimitedError, got %v", err)
	}
	if lim.RetryAfter <= 0 || lim.RetryAfter > window {
		t.Fatalf("RetryAfter=%v not in (0,%v]", lim.RetryAfter, window)
	}
}

// failFirstExpireHook fails the first EXPIRE issued on the client (covers Allow expire error path).
type failFirstExpireHook struct{ n int }

func (f *failFirstExpireHook) DialHook(next redis.DialHook) redis.DialHook { return next }

func (f *failFirstExpireHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (f *failFirstExpireHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if strings.EqualFold(cmd.Name(), "expire") {
			f.n++
			if f.n == 1 {
				return errors.New("injected expire failure")
			}
		}
		return next(ctx, cmd)
	}
}

func TestLimiter_Allow_expireErrorFromRedis(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	rdb.AddHook(&failFirstExpireHook{})
	defer rdb.Close()
	l := New(rdb)
	ctx := context.Background()

	err = l.Allow(ctx, "expire-fail-key", 5, time.Hour)
	if err == nil || !strings.Contains(err.Error(), "ratelimit expire") {
		t.Fatalf("expected expire wrap error, got %v", err)
	}
}

// failFirstDecrHook fails the first DECR (covers Allow decr error path when over quota).
type failFirstDecrHook struct{ n int }

func (f *failFirstDecrHook) DialHook(next redis.DialHook) redis.DialHook { return next }

func (f *failFirstDecrHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (f *failFirstDecrHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if strings.EqualFold(cmd.Name(), "decr") {
			f.n++
			if f.n == 1 {
				return errors.New("injected decr failure")
			}
		}
		return next(ctx, cmd)
	}
}

func TestLimiter_Allow_decrErrorWhenOverLimit(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	rdb.AddHook(&failFirstDecrHook{})
	defer rdb.Close()
	l := New(rdb)
	ctx := context.Background()
	key := "decr-fail-key"
	window := time.Hour

	if err := l.Allow(ctx, key, 2, window); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := l.Allow(ctx, key, 2, window); err != nil {
		t.Fatalf("second: %v", err)
	}
	err = l.Allow(ctx, key, 2, window)
	if err == nil || !strings.Contains(err.Error(), "ratelimit decr") {
		t.Fatalf("expected decr wrap error, got %v", err)
	}
}

func TestLimitedError_RetryAfterRoundedInMessage(t *testing.T) {
	e := &LimitedError{RetryAfter: 2*time.Minute + 499*time.Millisecond}
	msg := e.Error()
	if !strings.Contains(msg, "rate limit exceeded") || !strings.Contains(msg, "retry after") {
		t.Fatalf("unexpected message: %q", msg)
	}
}
