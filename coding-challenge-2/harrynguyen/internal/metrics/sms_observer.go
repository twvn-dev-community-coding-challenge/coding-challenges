package metrics

import (
	"context"

	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
)

type smsStatusObserver struct{}

// NewSMSStatusObserver records each persisted SMS status for Grafana (lifecycle / resilience views).
func NewSMSStatusObserver() sms.Observer {
	return smsStatusObserver{}
}

func (smsStatusObserver) OnMessageUpdated(ctx context.Context, msg *providers.SMSMessage) {
	if msg == nil {
		return
	}
	SMSMessageStatusTotal.WithLabelValues(string(msg.Status)).Inc()
}
