package ride

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"ridedispatch/internal/apperr"
)

// IdempotencyStore remembers the result of each POST /rides by its
// Idempotency-Key. A key is reserved (in progress) before the ride is created,
// so two concurrent requests with the same key cannot both create a ride.
type IdempotencyStore struct {
	mu      sync.Mutex
	entries map[string]*idempotencyEntry
}

type idempotencyEntry struct {
	requestHash string
	done        bool
	response    Ride
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{entries: map[string]*idempotencyEntry{}}
}

// Reserve claims key for a request with the given hash. It returns the
// original ride with replay=true if the same request already completed.
func (s *IdempotencyStore) Reserve(key, requestHash string) (Ride, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	switch {
	case !ok:
		s.entries[key] = &idempotencyEntry{requestHash: requestHash}
		return Ride{}, false, nil
	case e.requestHash != requestHash:
		return Ride{}, false, fmt.Errorf("%w: key %q", apperr.ErrIdempotencyKeyReused, key)
	case !e.done:
		return Ride{}, false, fmt.Errorf("%w: key %q", apperr.ErrIdempotencyInProgress, key)
	default:
		return e.response, true, nil
	}
}

// Complete stores the response for a reserved key.
func (s *IdempotencyStore) Complete(key string, response Ride) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.entries[key]; ok {
		e.done = true
		e.response = response
	}
}

// Abandon releases a reserved key after a failed request, so the client can
// retry with the same key.
func (s *IdempotencyStore) Abandon(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.entries[key]; ok && !e.done {
		delete(s.entries, key)
	}
}

// RequestHash fingerprints the parsed request. Hashing the parsed form makes
// the comparison independent of JSON whitespace and key order.
func RequestHash(req CreateRequest) string {
	b, _ := json.Marshal(req) // a struct of strings and floats always marshals
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
