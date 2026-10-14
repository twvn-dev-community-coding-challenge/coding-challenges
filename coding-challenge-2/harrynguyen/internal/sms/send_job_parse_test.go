package sms

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dotdak/sms-otp/internal/obslog"
)

func TestParseSendJobPayload_nil(t *testing.T) {
	_, err := ParseSendJobPayload(nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseSendJobPayload_string(t *testing.T) {
	j := SendJob{MessageID: "mid-1", RequestID: "r", CorrelationID: "c", TestSMSMode: "x"}
	raw, _ := json.Marshal(j)
	got, err := ParseSendJobPayload(string(raw))
	if err != nil || got.MessageID != "mid-1" {
		t.Fatalf("err=%v got=%+v", err, got)
	}
}

func TestParseSendJobPayload_bytes(t *testing.T) {
	raw := []byte(`{"message_id":"m2"}`)
	got, err := ParseSendJobPayload(raw)
	if err != nil || got.MessageID != "m2" {
		t.Fatalf("err=%v got=%+v", err, got)
	}
}

func TestParseSendJobPayload_rawMessage(t *testing.T) {
	got, err := ParseSendJobPayload(json.RawMessage(`{"message_id":"m3"}`))
	if err != nil || got.MessageID != "m3" {
		t.Fatalf("err=%v got=%+v", err, got)
	}
}

func TestParseSendJobPayload_map(t *testing.T) {
	got, err := ParseSendJobPayload(map[string]any{"message_id": "m4"})
	if err != nil || got.MessageID != "m4" {
		t.Fatalf("err=%v got=%+v", err, got)
	}
}

func TestParseSendJobPayload_missingMessageID(t *testing.T) {
	_, err := ParseSendJobPayload(`{}`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWorkerContext(t *testing.T) {
	ctx := context.Background()
	job := SendJob{RequestID: "a", CorrelationID: "b", TestSMSMode: string(TestSMSModeAlwaysFail)}
	out := workerContext(ctx, job)
	if obslog.RequestID(out) != "a" || obslog.CorrelationID(out) != "b" {
		t.Fatal("ids not propagated")
	}
	if GetTestSMSMode(out) != TestSMSModeAlwaysFail {
		t.Fatalf("test mode: %q", GetTestSMSMode(out))
	}
}
