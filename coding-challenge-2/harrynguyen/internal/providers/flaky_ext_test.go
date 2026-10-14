package providers

import (
	"context"
	"errors"
	"testing"
)

func TestNewFlakyProvider_NegativeFailsBecomesZero(t *testing.T) {
	inner := NewMockProvider(ProviderVonage)
	p := NewFlakyProvider(inner, -3, errors.New("x"))
	if _, _, err := p.Send(context.Background(), "PH", CarrierGlobe, "6391712345678", "hi"); err != nil {
		t.Fatal(err)
	}
}

func TestNewFlakyProvider_DefaultErr(t *testing.T) {
	inner := NewMockProvider(ProviderVonage)
	p := NewFlakyProvider(inner, 1, nil)
	_, _, err := p.Send(context.Background(), "PH", CarrierGlobe, "6391712345678", "hi")
	if err == nil {
		t.Fatal("expected error")
	}
	_, _, err = p.Send(context.Background(), "PH", CarrierGlobe, "6391712345678", "hi")
	if err != nil {
		t.Fatal(err)
	}
}

func TestFlakyProvider_GetProviderName(t *testing.T) {
	inner := NewMockProvider(ProviderInfobip)
	p := NewFlakyProvider(inner, 0, errors.New("x"))
	if p.GetProviderName() != ProviderInfobip {
		t.Fatal(p.GetProviderName())
	}
}
