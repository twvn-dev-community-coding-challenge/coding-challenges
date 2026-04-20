// Package obslog configures structured logging (slog) for the API and recovery paths.
package obslog

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

type ctxKey int

const (
	requestIDKey     ctxKey = 1
	correlationIDKey ctxKey = 2
)

// L is the application logger. Initialized by Init.
var L *slog.Logger

var initOnce sync.Once

// Init configures the default handler from LOG_FORMAT (json or text). Safe to call multiple times.
func Init() {
	initOnce.Do(func() {
		opts := &slog.HandlerOptions{Level: slog.LevelInfo}
		var h slog.Handler
		switch os.Getenv("LOG_FORMAT") {
		case "json":
			h = slog.NewJSONHandler(os.Stdout, opts)
		default:
			h = slog.NewTextHandler(os.Stdout, opts)
		}
		L = slog.New(h).With("service", "sms-otp")
	})
}

// WithRequestID returns ctx carrying X-Request-ID (per-request id).
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns X-Request-ID from ctx, if any.
func RequestID(ctx context.Context) string {
	v, _ := ctx.Value(requestIDKey).(string)
	return v
}

// WithCorrelationID returns ctx carrying X-Correlation-ID (cross-service trace id).
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationID returns X-Correlation-ID from ctx, if any.
func CorrelationID(ctx context.Context) string {
	v, _ := ctx.Value(correlationIDKey).(string)
	return v
}

// RecoveryEvent logs a callback recovery or mitigation outcome (Loki-friendly key=value in JSON).
func RecoveryEvent(ctx context.Context, kind, smsID, providerMsgID, fromStatus, toStatus string) {
	Init()
	args := []any{
		slog.String("event", EventSMSRecovery),
		slog.String("recovery_kind", kind),
		slog.String("sms_id", smsID),
		slog.String("provider_message_id", providerMsgID),
		slog.String("from_status", fromStatus),
		slog.String("to_status", toStatus),
	}
	if rid := RequestID(ctx); rid != "" {
		args = append(args, slog.String("request_id", rid))
	}
	if cid := CorrelationID(ctx); cid != "" {
		args = append(args, slog.String("correlation_id", cid))
	}
	L.InfoContext(ctx, MsgSMSRecoveryEvent, args...)
}
