package sms

import (
	"context"
	"sync"

	"github.com/dotdak/sms-otp/internal/providers"
)

// MetricSnapshot represents a snapshot of aggregated cost and volume metrics.
type MetricSnapshot struct {
	TotalVolume        int
	TotalEstimatedCost float64
	TotalActualCost    float64
}

// InMemoryCostTracker implements the Observer interface to track SMS usage metrics.
// It maintains thread-safe aggregates of volume and costs.
type InMemoryCostTracker struct {
	mu              sync.RWMutex
	byProvider      map[providers.Provider]*MetricSnapshot
	byCountry       map[string]*MetricSnapshot
	processedMsgID  map[string]providers.MessageStatus // tracks the latest status applied to avoid double counting
	lastActualByMsg map[string]float64      // terminal callbacks may revise actual cost without a status change
}

// NewInMemoryCostTracker creates a new initialized CostTracker.
func NewInMemoryCostTracker() *InMemoryCostTracker {
	return &InMemoryCostTracker{
		byProvider:      make(map[providers.Provider]*MetricSnapshot),
		byCountry:       make(map[string]*MetricSnapshot),
		processedMsgID:  make(map[string]providers.MessageStatus),
		lastActualByMsg: make(map[string]float64),
	}
}

// OnMessageUpdated is called by the SMSService when a message state changes.
func (c *InMemoryCostTracker) OnMessageUpdated(ctx context.Context, msg *providers.SMSMessage) {
	if msg == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	lastStatus, exists := c.processedMsgID[msg.ID]

	// Determine what to aggregate based on the transition
	var incVol int
	var incEstCost float64
	var incActCost float64

	// If this is the first time we see this message entering the provider queue
	if (!exists || lastStatus == providers.StatusSendToProvider) && msg.Status == providers.StatusQueue {
		incVol = 1
		incEstCost = msg.EstimatedCost
	}

	// Actual cost: first delivery to Send-success, or supplemental cost from a later callback / mitigation path
	if msg.Status == providers.StatusSendSuccess && msg.ActualCost > 0 {
		prevAct := c.lastActualByMsg[msg.ID]
		if msg.ActualCost > prevAct {
			incActCost = msg.ActualCost - prevAct
			c.lastActualByMsg[msg.ID] = msg.ActualCost
		}
	}

	c.processedMsgID[msg.ID] = msg.Status

	if incVol == 0 && incEstCost == 0 && incActCost == 0 {
		return // Nothing new to aggregate for costs/volume
	}

	// Update Provider metrics
	if msg.Provider != "" {
		if _, ok := c.byProvider[msg.Provider]; !ok {
			c.byProvider[msg.Provider] = &MetricSnapshot{}
		}
		c.byProvider[msg.Provider].TotalVolume += incVol
		c.byProvider[msg.Provider].TotalEstimatedCost += incEstCost
		c.byProvider[msg.Provider].TotalActualCost += incActCost
	}

	// Update Country metrics
	if msg.Country != "" {
		if _, ok := c.byCountry[msg.Country]; !ok {
			c.byCountry[msg.Country] = &MetricSnapshot{}
		}
		c.byCountry[msg.Country].TotalVolume += incVol
		c.byCountry[msg.Country].TotalEstimatedCost += incEstCost
		c.byCountry[msg.Country].TotalActualCost += incActCost
	}
}

// GetProviderMetrics returns a copy of metrics aggregated by Provider
func (c *InMemoryCostTracker) GetProviderMetrics() map[providers.Provider]MetricSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make(map[providers.Provider]MetricSnapshot, len(c.byProvider))
	for k, v := range c.byProvider {
		res[k] = *v
	}
	return res
}

// GetCountryMetrics returns a copy of metrics aggregated by Country
func (c *InMemoryCostTracker) GetCountryMetrics() map[string]MetricSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make(map[string]MetricSnapshot, len(c.byCountry))
	for k, v := range c.byCountry {
		res[k] = *v
	}
	return res
}

// Totals returns global aggregates summed from byProvider only (provider and country maps mirror the same increments; summing both would double-count).
func (c *InMemoryCostTracker) Totals() MetricSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var s MetricSnapshot
	for _, v := range c.byProvider {
		s.TotalVolume += v.TotalVolume
		s.TotalEstimatedCost += v.TotalEstimatedCost
		s.TotalActualCost += v.TotalActualCost
	}
	return s
}
