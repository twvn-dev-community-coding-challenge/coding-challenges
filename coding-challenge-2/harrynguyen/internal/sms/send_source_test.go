package sms

import (
	"context"
	"testing"
)

func TestSendSourceFromContext_unset(t *testing.T) {
	if got := SendSourceFromContext(context.Background()); got != "" {
		t.Fatalf("got %q", got)
	}
}
