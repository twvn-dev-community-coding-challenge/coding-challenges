package sms

import (
	"context"
	"strings"
)

type ctxKey string

const testSMSModeContextKey ctxKey = "test_sms_mode"

const (
	TestSMSModeTransientFailureOnce = "transient-failure-once"
	TestSMSModeAlwaysFail           = "always-fail"
	TestSMSModeAlwaysSuccess        = "always-success"
)

func WithTestSMSMode(ctx context.Context, mode string) context.Context {
	normalized := strings.TrimSpace(strings.ToLower(mode))
	if normalized == "" {
		return ctx
	}
	return context.WithValue(ctx, testSMSModeContextKey, normalized)
}

func GetTestSMSMode(ctx context.Context) string {
	v := ctx.Value(testSMSModeContextKey)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
