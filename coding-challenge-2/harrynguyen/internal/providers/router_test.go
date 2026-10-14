package providers

import "testing"

func TestSimpleProviderRouter_Route(t *testing.T) {
	router := NewSimpleProviderRouter()
	adapters := map[Provider]SMSProvider{
		ProviderVonage:      NewMockProvider(ProviderVonage),
		ProviderInfobip:     NewMockProvider(ProviderInfobip),
		ProviderTwilio:      NewMockProvider(ProviderTwilio),
		ProviderAWSSNS:      NewMockProvider(ProviderAWSSNS),
		ProviderTelnyx:      NewMockProvider(ProviderTelnyx),
		ProviderMessageBird: NewMockProvider(ProviderMessageBird),
		ProviderSinch:       NewMockProvider(ProviderSinch),
	}
	router.RegisterDefaultRoutes(adapters)

	tests := []struct {
		country      string
		carrier      Carrier
		wantProvider Provider
		wantErr      bool
	}{
		{"VN", CarrierViettel, ProviderVonage, false},
		{"VN", CarrierMobifone, ProviderInfobip, false},
		{"VN", CarrierVinaphone, ProviderTwilio, false},
		{"TH", CarrierAIS, ProviderInfobip, false},
		{"TH", CarrierDTAC, ProviderAWSSNS, false},
		{"SG", CarrierSingtel, ProviderTwilio, false},
		{"SG", CarrierStarHub, ProviderTelnyx, false},
		{"PH", CarrierGlobe, ProviderMessageBird, false},
		{"PH", CarrierSmart, ProviderSinch, false},
		{"PH", CarrierDITO, ProviderMessageBird, false},
		{"Unknown", CarrierViettel, Provider(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.country+"-"+string(tt.carrier), func(t *testing.T) {
			got, err := router.Route(tt.country, tt.carrier)
			if (err != nil) != tt.wantErr {
				t.Errorf("Route() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantProvider {
				t.Errorf("Route() = %v, want %v", got, tt.wantProvider)
			}
		})
	}
}

func TestSimpleProviderRouter_RegisterRouter(t *testing.T) {
	r := NewSimpleProviderRouter()
	mock := NewMockProvider(ProviderTwilio)
	r.RegisterRouter("XX", CarrierUnknown, ProviderTwilio, mock)
	p, err := r.Route("XX", CarrierUnknown)
	if err != nil || p != ProviderTwilio {
		t.Fatalf("Route = %v, %v", p, err)
	}
	got, ok := r.Adapter(ProviderTwilio)
	if !ok || got != mock {
		t.Fatalf("Adapter = %v, ok=%v", got, ok)
	}
}
