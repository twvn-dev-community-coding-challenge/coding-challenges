package providers

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrNoRuleFound is returned when no routing rule matches country and carrier.
	ErrNoRuleFound = errors.New("no routing rule found for the given country and carrier")
)

// ProviderRouter selects an SMS gateway implementation from country and carrier,
// and holds the SMSProvider adapters for each Provider enum value.
type ProviderRouter interface {
	Route(country string, carrier Carrier) (Provider, error)
	RegisterRouter(country string, carrier Carrier, provider Provider, adapter SMSProvider)
	Adapter(provider Provider) (SMSProvider, bool)
}

// SimpleProviderRouter stores (country, carrier) → Provider routing rules and
// Provider → SMSProvider adapters. It is safe for concurrent Route/Adapter calls
// after registration (typical: wire in main before serving traffic).
type SimpleProviderRouter struct {
	mu       sync.RWMutex
	routes   map[routeKey]Provider
	adapters map[Provider]SMSProvider
}

type routeKey struct {
	country string
	carrier Carrier
}

// NewSimpleProviderRouter returns an empty router; use RegisterRouter and/or RegisterDefaultRoutes.
func NewSimpleProviderRouter() *SimpleProviderRouter {
	return &SimpleProviderRouter{
		routes:   make(map[routeKey]Provider),
		adapters: make(map[Provider]SMSProvider),
	}
}

// RegisterRouter adds or replaces a route for (country, carrier) → provider and
// registers the SMS adapter for that provider value.
func (r *SimpleProviderRouter) RegisterRouter(country string, carrier Carrier, provider Provider, adapter SMSProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[routeKey{country: country, carrier: carrier}] = provider
	if adapter != nil {
		r.adapters[provider] = adapter
	}
}

// RegisterDefaultRoutes applies the built-in VN / TH / SG / PH table using adapters from the map.
// Every Provider used in that table must have a non-nil entry in adapters.
func (r *SimpleProviderRouter) RegisterDefaultRoutes(adapters map[Provider]SMSProvider) {
	rules := []struct {
		country string
		carrier Carrier
		p       Provider
	}{
		{"VN", CarrierViettel, ProviderVonage},
		{"VN", CarrierMobifone, ProviderInfobip},
		{"VN", CarrierVinaphone, ProviderTwilio},
		{"TH", CarrierAIS, ProviderInfobip},
		{"TH", CarrierDTAC, ProviderAWSSNS},
		{"SG", CarrierSingtel, ProviderTwilio},
		{"SG", CarrierStarHub, ProviderTelnyx},
		{"PH", CarrierGlobe, ProviderMessageBird},
		{"PH", CarrierSmart, ProviderSinch},
		{"PH", CarrierDITO, ProviderMessageBird},
	}
	for _, rule := range rules {
		a := adapters[rule.p]
		if a == nil {
			continue
		}
		r.RegisterRouter(rule.country, rule.carrier, rule.p, a)
	}
}

// Route implements ProviderRouter.
func (r *SimpleProviderRouter) Route(country string, carrier Carrier) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.routes[routeKey{country: country, carrier: carrier}]
	if !ok {
		return Provider(""), fmt.Errorf("%w: country=%s, carrier=%s", ErrNoRuleFound, country, carrier)
	}
	return p, nil
}

// Adapter returns the SMSProvider registered for provider, if any.
func (r *SimpleProviderRouter) Adapter(provider Provider) (SMSProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[provider]
	return a, ok
}
