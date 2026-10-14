package sms

import (
	"testing"

	"github.com/dotdak/sms-otp/internal/providers"
)

func TestMessageStatus_IsValidTransition(t *testing.T) {
	tests := []struct {
		name string
		from providers.MessageStatus
		to   providers.MessageStatus
		want bool
	}{
		{"New to SendToProvider", providers.StatusNew, providers.StatusSendToProvider, true},
		{"New to SendSuccess (Invalid)", providers.StatusNew, providers.StatusSendSuccess, false},
		{"Queue to SendToCarrier", providers.StatusQueue, providers.StatusSendToCarrier, true},
		{"Queue to CarrierRejected", providers.StatusQueue, providers.StatusCarrierRejected, true},
		{"CarrierRejected to SendToProvider", providers.StatusCarrierRejected, providers.StatusSendToProvider, true},
		{"SendSuccess to New (Invalid)", providers.StatusSendSuccess, providers.StatusNew, false},
		{"SendFailed to SendToProvider", providers.StatusSendFailed, providers.StatusSendToProvider, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.IsValidTransition(tt.to); got != tt.want {
				t.Errorf("IsValidTransition() from %v to %v = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
