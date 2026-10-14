package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrLimited is returned when a fixed-window counter exceeds the limit.
var ErrLimited = errors.New("rate limit exceeded")

// LimitedError carries Retry-After hint for HTTP 429.
type LimitedError struct {
	RetryAfter time.Duration
}

func (e *LimitedError) Error() string {
	return fmt.Sprintf("%v (retry after %s)", ErrLimited, e.RetryAfter.Round(time.Second))
}

func (e *LimitedError) Unwrap() error {
	return ErrLimited
}

// Limiter implements fixed-window rate limiting with Redis INCR + EXPIRE.
type Limiter struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Limiter {
	return &Limiter{rdb: rdb}
}

// Allow increments key and returns nil if within limit. If over limit, decrements
// the key to avoid counting rejected over-quota attempts, then returns *LimitedError.
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) error {
	if l == nil || l.rdb == nil {
		return nil
	}
	if limit <= 0 || window <= 0 {
		return nil
	}
	n, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("ratelimit incr: %w", err)
	}
	if n == 1 {
		if err := l.rdb.Expire(ctx, key, window).Err(); err != nil {
			return fmt.Errorf("ratelimit expire: %w", err)
		}
	}
	if n > int64(limit) {
		if _, err := l.rdb.Decr(ctx, key).Result(); err != nil {
			return fmt.Errorf("ratelimit decr: %w", err)
		}
		ttl, err := l.rdb.TTL(ctx, key).Result()
		if err != nil || ttl < 0 {
			ttl = window
		}
		return &LimitedError{RetryAfter: ttl}
	}
	return nil
}
