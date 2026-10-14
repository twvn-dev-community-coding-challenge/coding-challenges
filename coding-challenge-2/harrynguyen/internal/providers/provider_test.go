package providers

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMockProvider_Send(t *testing.T) {
	provider := NewMockProvider(ProviderTwilio)
	ctx := context.Background()

	tests := []struct {
		name       string
		content    string
		wantRate   float64
		hasSpecial bool
	}{
		{
			name:       "Normal characters only",
			content:    "Hello world 123, . !",
			wantRate:   0.001,
			hasSpecial: false,
		},
		{
			name:       "With special characters (emoji)",
			content:    "Hello 💎",
			wantRate:   0.005,
			hasSpecial: true,
		},
		{
			name:       "With other special characters",
			content:    "Cost is $100",
			wantRate:   0.005,
			hasSpecial: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgID, cost, err := provider.Send(ctx, "Vietnam", CarrierViettel, "+8496111222", tt.content)
			if err != nil {
				t.Fatalf("Send() error = %v", err)
			}

			if !strings.HasPrefix(msgID, string(ProviderTwilio)+"-msg-") {
				t.Errorf("Send() msgID = %v, want prefix %v", msgID, string(ProviderTwilio)+"-msg-")
			}

			runeCount := float64(utf8.RuneCountInString(tt.content))
			expectedCost := runeCount * tt.wantRate
			if cost != expectedCost {
				t.Errorf("Send() cost = %v, want %v (rate: %v, runes: %v)", cost, expectedCost, tt.wantRate, runeCount)
			}
		})
	}
}

func TestMockProvider_GetProviderName(t *testing.T) {
	name := ProviderInfobip
	provider := NewMockProvider(name)
	if provider.GetProviderName() != name {
		t.Errorf("GetProviderName() = %v, want %v", provider.GetProviderName(), name)
	}
}
