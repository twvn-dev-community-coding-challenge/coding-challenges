package handlers

import (
	"context"
	"testing"

	"github.com/dotdak/sms-otp/internal/providers"
)

func TestAuthSMSObserver_OnMessageUpdated(t *testing.T) {
	var o authSMSObserver
	o.OnMessageUpdated(context.Background(), &providers.SMSMessage{ID: "m1", Status: providers.StatusQueue})
	o.OnMessageUpdated(context.Background(), nil)
}
