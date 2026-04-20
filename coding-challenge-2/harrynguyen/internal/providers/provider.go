package providers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
	"unicode/utf8"
)

// SMSProvider defines the interface for interacting with an SMS gateway.
type SMSProvider interface {
	Send(ctx context.Context, country string, carrier Carrier, phoneNumber, content string) (providerMsgID string, estimatedCost float64, err error)
	GetProviderName() Provider
}

// MockProvider is a simulated implementation of an SMS gateway.
type MockProvider struct {
	Name Provider
	// SkipDelay, when true, avoids the default 100ms simulated latency (for load / resilience tests).
	SkipDelay bool
	// Quiet, when true, skips provider Send logging (keeps high-concurrency tests readable).
	Quiet bool
}

func NewMockProvider(name Provider) *MockProvider {
	return &MockProvider{Name: name}
}

func (p *MockProvider) GetProviderName() Provider {
	return p.Name
}

func (p *MockProvider) Send(ctx context.Context, country string, carrier Carrier, phoneNumber, content string) (string, float64, error) {
	if !p.Quiet {
		log.Printf("[PROVIDER: %s] Sending SMS to %s (%s, %s): %s", p.Name, phoneNumber, country, carrier, content)
	}

	if !p.SkipDelay {
		time.Sleep(100 * time.Millisecond)
	}

	// Simulate provider-specific message ID
	providerMsgID := fmt.Sprintf("%s-msg-%d", p.Name, rand.Int63n(1000000))

	// Cost is calculated based on content length and character complexity
	const (
		NormalCharRate  = 0.001
		SpecialCharRate = 0.005
	)

	hasSpecial := false
	for _, r := range content {
		// A simple definition of "normal": alphanumeric and basic punctuation
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '.' || r == ',' || r == '!') {
			hasSpecial = true
			break
		}
	}

	rate := NormalCharRate
	if hasSpecial {
		rate = SpecialCharRate
	}

	estimatedCost := float64(utf8.RuneCountInString(content)) * rate

	return providerMsgID, estimatedCost, nil
}
