package models

// ProviderAgreement captures the routing contract between a carrier and a provider.
type ProviderAgreement struct {
	ID         string
	CarrierID  string
	ProviderID string
}
