package carrier

import (
	"testing"

	"github.com/dotdak/sms-otp/internal/providers"
)

func TestPrefixCarrierResolver_Resolve(t *testing.T) {
	resolver := NewPrefixCarrierResolver()

	tests := []struct {
		phoneNumber string
		wantCarrier providers.Carrier
	}{
		{"+84961234567", providers.CarrierViettel},
		{"84901234567", providers.CarrierMobifone},
		{"+84912345678", providers.CarrierVinaphone},
		{"+66812345678", providers.CarrierAIS},
		{"+66822345678", providers.CarrierDTAC},
		{"+6581234567", providers.CarrierSingtel},
		{"+6591234567", providers.CarrierStarHub},
		{"+639171234567", providers.CarrierGlobe},
		{"+639181234567", providers.CarrierSmart},
		{"+639911234567", providers.CarrierDITO},
		{"+1234567890", providers.CarrierUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.phoneNumber, func(t *testing.T) {
			got, _ := resolver.Resolve(tt.phoneNumber)
			if got != tt.wantCarrier {
				t.Errorf("Resolve() = %v, want %v", got, tt.wantCarrier)
			}
		})
	}
}
