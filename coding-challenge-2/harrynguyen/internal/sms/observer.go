package sms

import (
	"context"

	"github.com/dotdak/sms-otp/internal/providers"
)

// Observer defines the interface for components notified when an SMSMessage's state updates.
// Register with RegisterGlobalObserver (e.g. cost across all sends) or RegisterSourceObserver
// (only messages whose SendSource matches; set via WithSendSource on the SendSMS context).
type Observer interface {
	OnMessageUpdated(ctx context.Context, msg *providers.SMSMessage)
}
