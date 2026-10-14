package obslog

import (
	"context"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	ctx := WithRequestID(context.Background(), "r1")
	if RequestID(ctx) != "r1" {
		t.Fatalf("RequestID=%q", RequestID(ctx))
	}
	if RequestID(context.Background()) != "" {
		t.Fatal("empty ctx should have no request id")
	}
}

func TestWithCorrelationID(t *testing.T) {
	ctx := WithCorrelationID(context.Background(), "c1")
	if CorrelationID(ctx) != "c1" {
		t.Fatalf("CorrelationID=%q", CorrelationID(ctx))
	}
	if CorrelationID(context.Background()) != "" {
		t.Fatal("empty ctx should have no correlation id")
	}
}

func TestInit_Idempotent(t *testing.T) {
	Init()
	Init()
	if L == nil {
		t.Fatal("L should be initialized")
	}
}

func TestRecoveryEvent(t *testing.T) {
	ctx := WithRequestID(WithCorrelationID(context.Background(), "cc"), "rr")
	RecoveryEvent(ctx, "mitigation", "sms1", "p1", "Queue", "Send-success")
}
