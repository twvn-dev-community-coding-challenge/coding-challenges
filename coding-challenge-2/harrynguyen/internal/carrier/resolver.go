package carrier

import (
	"strings"

	"github.com/dotdak/sms-otp/internal/providers"
)

// CarrierResolver maps E.164-style digits to a mobile network operator (carrier).
type CarrierResolver interface {
	Resolve(phoneNumber string) (providers.Carrier, error)
}

// PrefixCarrierResolver uses leading-digit prefixes to infer the carrier (simulation; not exhaustive).
type PrefixCarrierResolver struct{}

// NewPrefixCarrierResolver returns a prefix-based CarrierResolver.
func NewPrefixCarrierResolver() *PrefixCarrierResolver {
	return &PrefixCarrierResolver{}
}

// Resolve implements CarrierResolver.
func (r *PrefixCarrierResolver) Resolve(phoneNumber string) (providers.Carrier, error) {
	// Simulation based on common prefixes (not exhaustive)
	cleanPhone := strings.TrimLeft(phoneNumber, "+")

	switch {
	// Vietnam (+84)
	case strings.HasPrefix(cleanPhone, "84"):
		suffix := cleanPhone[2:]
		switch {
		case strings.HasPrefix(suffix, "96"), strings.HasPrefix(suffix, "97"), strings.HasPrefix(suffix, "98"), strings.HasPrefix(suffix, "86"):
			return providers.CarrierViettel, nil
		case strings.HasPrefix(suffix, "90"), strings.HasPrefix(suffix, "93"), strings.HasPrefix(suffix, "89"):
			return providers.CarrierMobifone, nil
		case strings.HasPrefix(suffix, "91"), strings.HasPrefix(suffix, "94"), strings.HasPrefix(suffix, "88"):
			return providers.CarrierVinaphone, nil
		}
	// Thailand (+66)
	case strings.HasPrefix(cleanPhone, "66"):
		suffix := cleanPhone[2:]
		switch {
		case strings.HasPrefix(suffix, "81"), strings.HasPrefix(suffix, "91"):
			return providers.CarrierAIS, nil
		case strings.HasPrefix(suffix, "82"), strings.HasPrefix(suffix, "92"):
			return providers.CarrierDTAC, nil
		}
	// Singapore (+65)
	case strings.HasPrefix(cleanPhone, "65"):
		suffix := cleanPhone[2:]
		switch {
		case strings.HasPrefix(suffix, "8"), strings.HasPrefix(suffix, "9"):
			// In reality, this is more complex. Simple simulation: 8 is Singtel, 9 is StarHub
			if strings.HasPrefix(suffix, "8") {
				return providers.CarrierSingtel, nil
			}
			return providers.CarrierStarHub, nil
		}
	// Philippines (+63)
	case strings.HasPrefix(cleanPhone, "63"):
		suffix := cleanPhone[2:]
		switch {
		case strings.HasPrefix(suffix, "917"), strings.HasPrefix(suffix, "905"):
			return providers.CarrierGlobe, nil
		case strings.HasPrefix(suffix, "918"), strings.HasPrefix(suffix, "999"):
			return providers.CarrierSmart, nil
		case strings.HasPrefix(suffix, "991"), strings.HasPrefix(suffix, "992"):
			return providers.CarrierDITO, nil
		}
	}

	return providers.CarrierUnknown, nil
}
