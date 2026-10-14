package providers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
	"unicode/utf8"
)

// TwilioAdapter is a simulated implementation of the Twilio SMS gateway.
type TwilioAdapter struct {
	AccountSID string
	AuthToken  string
}

func NewTwilioAdapter(accountSID, authToken string) *TwilioAdapter {
	return &TwilioAdapter{
		AccountSID: accountSID,
		AuthToken:  authToken,
	}
}

func (p *TwilioAdapter) GetProviderName() Provider {
	return ProviderTwilio
}

func (p *TwilioAdapter) Send(ctx context.Context, country string, carrier Carrier, phoneNumber, content string) (string, float64, error) {
	log.Printf("[TWILIO] Sending SMS via Account %s to %s", p.AccountSID, phoneNumber)

	// Simulate network delay
	time.Sleep(150 * time.Millisecond)

	// Twilio message IDs usually start with 'SM'
	providerMsgID := fmt.Sprintf("SM%x", rand.Int63n(1000000000000000))

	// Twilio-specific cost calculation simulation
	// Let's assume Twilio is slightly more expensive or has different rates
	const (
		NormalCharRate  = 0.002
		SpecialCharRate = 0.008
	)

	hasSpecial := false
	for _, r := range content {
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
