package handlers

import (
	"context"

	"github.com/dotdak/sms-otp/internal/obslog"
)

// authFlowAttrs returns common JSON log fields for auth flows (Loki: filter by correlation_id / request_id).
func authFlowAttrs(ctx context.Context, flow, step string) []any {
	obslog.Init()
	out := []any{
		"event", obslog.EventAuthFlow,
		"flow", flow,
		"step", step,
		"request_id", obslog.RequestID(ctx),
	}
	if cid := obslog.CorrelationID(ctx); cid != "" {
		out = append(out, "correlation_id", cid)
	}
	return out
}

// maskPhoneSuffix keeps only the last 4 digits for logs (E.164 digits).
func maskPhoneSuffix(digits string) string {
	if len(digits) < 4 {
		return "****"
	}
	return "***" + digits[len(digits)-4:]
}

// smsCallbackAttrs returns common JSON log fields for POST /api/sms/callback (Loki: correlation_id / request_id).
func smsCallbackAttrs(ctx context.Context, step string) []any {
	obslog.Init()
	out := []any{
		"event", obslog.EventSMSCallback,
		"flow", obslog.FlowSMSCallback,
		"step", step,
		"request_id", obslog.RequestID(ctx),
	}
	if cid := obslog.CorrelationID(ctx); cid != "" {
		out = append(out, "correlation_id", cid)
	}
	return out
}
