package sms

import "context"

type sendSourceCtxKey struct{}

// Send-source values identify which entry path created an SMS so callbacks notify the right scoped observers.
const (
	SendSourceAuth = "auth"
	SendSourceAPI  = "api"
)

// WithSendSource returns a child context that carries the send source for SendSMS.
func WithSendSource(ctx context.Context, source string) context.Context {
	return context.WithValue(ctx, sendSourceCtxKey{}, source)
}

// SendSourceFromContext returns the send source from ctx, or "" if unset.
func SendSourceFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(sendSourceCtxKey{}).(string)
	return v
}
