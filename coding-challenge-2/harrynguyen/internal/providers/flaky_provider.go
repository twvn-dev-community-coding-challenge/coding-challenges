package providers

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
)

// FlakyProvider wraps an SMSProvider and fails the first N Send calls, then delegates.
// It is used to simulate error-prone gateways and prove client/service behavior under outages.
type FlakyProvider struct {
	inner     SMSProvider
	failFirst int64
	sendErr   error

	calls atomic.Int64
}

// NewFlakyProvider returns a wrapper that returns sendErr for the first failFirstN invocations
// of Send, then forwards to inner. If sendErr is nil, a default error is used.
func NewFlakyProvider(inner SMSProvider, failFirstN int, sendErr error) *FlakyProvider {
	if failFirstN < 0 {
		failFirstN = 0
	}
	if sendErr == nil {
		sendErr = errors.New("simulated provider outage")
	}
	return &FlakyProvider{
		inner:     inner,
		failFirst: int64(failFirstN),
		sendErr:   sendErr,
	}
}

// Calls returns how many times Send has been invoked (including failures).
func (p *FlakyProvider) Calls() int {
	return int(p.calls.Load())
}

func (p *FlakyProvider) GetProviderName() Provider {
	return p.inner.GetProviderName()
}

func (p *FlakyProvider) Send(ctx context.Context, country string, carrier Carrier, phoneNumber, content string) (string, float64, error) {
	n := p.calls.Add(1)
	if n <= p.failFirst {
		return "", 0, fmt.Errorf("%w (attempt %d/%d)", p.sendErr, n, p.failFirst)
	}
	return p.inner.Send(ctx, country, carrier, phoneNumber, content)
}
