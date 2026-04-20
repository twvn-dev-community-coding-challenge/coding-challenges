package providers

import "testing"

func TestRegisterDefaultRoutes_SkipsNilAdapter(t *testing.T) {
	r := NewSimpleProviderRouter()
	r.RegisterDefaultRoutes(map[Provider]SMSProvider{
		ProviderTwilio: nil,
	})
	_, err := r.Route("VN", CarrierVinaphone)
	if err == nil {
		t.Fatal("expected no route when adapter was nil")
	}
}
