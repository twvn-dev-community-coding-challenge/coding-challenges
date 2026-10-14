package ratelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestLimitedError_ErrorAndUnwrap(t *testing.T) {
	e := &LimitedError{RetryAfter: 1500 * time.Millisecond}
	if e.Error() == "" {
		t.Fatal("empty error string")
	}
	if !errors.Is(e, ErrLimited) {
		t.Fatal("Unwrap should expose ErrLimited")
	}
}

func TestLimiter_NilReceiver(t *testing.T) {
	var l *Limiter
	if err := l.Allow(context.Background(), "k", 5, time.Hour); err != nil {
		t.Fatalf("nil limiter: %v", err)
	}
}

func TestLimiter_NewNilRedis(t *testing.T) {
	l := New(nil)
	if err := l.Allow(context.Background(), "k", 5, time.Hour); err != nil {
		t.Fatalf("nil redis: %v", err)
	}
}

func TestLimiter_InvalidWindow(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer rdb.Close()
	l := New(rdb)
	ctx := context.Background()
	if err := l.Allow(ctx, "k", 5, 0); err != nil {
		t.Fatalf("zero window: %v", err)
	}
	if err := l.Allow(ctx, "k2", 0, time.Hour); err != nil {
		t.Fatalf("zero limit: %v", err)
	}
}
