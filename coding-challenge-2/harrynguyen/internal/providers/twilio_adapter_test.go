package providers

import (
	"context"
	"testing"
)

func TestTwilioAdapter_GetProviderName(t *testing.T) {
	p := NewTwilioAdapter("sid", "token")
	if p.GetProviderName() != ProviderTwilio {
		t.Fatal(p.GetProviderName())
	}
}

func TestTwilioAdapter_Send(t *testing.T) {
	if testing.Short() {
		t.Skip("skips 150ms sleep")
	}
	p := NewTwilioAdapter("ACtest", "tok")
	id, cost, err := p.Send(context.Background(), "US", CarrierUnknown, "+15551234567", "hello!")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" || cost < 0 {
		t.Fatalf("id=%q cost=%v", id, cost)
	}
}
