package sms

import "testing"

func TestNewBullMQSendPublisher(t *testing.T) {
	p := NewBullMQSendPublisher(nil)
	if p == nil {
		t.Fatal("expected non-nil publisher wrapper")
	}
}
