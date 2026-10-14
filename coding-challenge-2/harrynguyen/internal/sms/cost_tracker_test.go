package sms

import (
	"context"
	"testing"

	"github.com/dotdak/sms-otp/internal/providers"
)

func TestCostTracker_Aggregation(t *testing.T) {
	tracker := NewInMemoryCostTracker()
	ctx := context.Background()

	// 1. Simulate sending SMS to Provider Twilio in Vietnam
	msg1 := &providers.SMSMessage{
		ID:            "msg-1",
		Country:       "Vietnam",
		Provider:      providers.ProviderTwilio,
		EstimatedCost: 0.05,
		Status:        providers.StatusQueue,
	}
	tracker.OnMessageUpdated(ctx, msg1) // +1 volume, +0.05 est

	// 2. Simulate sending another SMS to Vonage in Singapore
	msg2 := &providers.SMSMessage{
		ID:            "msg-2",
		Country:       "Singapore",
		Provider:      providers.ProviderVonage,
		EstimatedCost: 0.10,
		Status:        providers.StatusQueue,
	}
	tracker.OnMessageUpdated(ctx, msg2) // +1 volume, +0.10 est

	// 3. Update msg1 to SendSuccess with Actual Cost
	msg1.Status = providers.StatusSendSuccess
	msg1.ActualCost = 0.045
	tracker.OnMessageUpdated(ctx, msg1) // +0.045 actual

	// 4. Send msg3 to Twilio in Vietnam but it falls to CarrierRejected.
	// It was queued first:
	msg3 := &providers.SMSMessage{
		ID:            "msg-3",
		Country:       "Vietnam",
		Provider:      providers.ProviderTwilio,
		EstimatedCost: 0.05,
		Status:        providers.StatusQueue,
	}
	tracker.OnMessageUpdated(ctx, msg3) // +1 volume, +0.05 est

	// Then rejected. It should NOT add to volume/est cost again or actual cost.
	msg3.Status = providers.StatusCarrierRejected
	tracker.OnMessageUpdated(ctx, msg3)

	// Assertions for Provider
	byProvider := tracker.GetProviderMetrics()

	twilioStats := byProvider[providers.ProviderTwilio]
	if twilioStats.TotalVolume != 2 {
		t.Errorf("Twilio volume = %d; want 2", twilioStats.TotalVolume)
	}
	if twilioStats.TotalEstimatedCost != 0.10 {
		t.Errorf("Twilio Est Cost = %f; want 0.10", twilioStats.TotalEstimatedCost)
	}
	if twilioStats.TotalActualCost != 0.045 {
		t.Errorf("Twilio Act Cost = %f; want 0.045", twilioStats.TotalActualCost)
	}

	vonageStats := byProvider[providers.ProviderVonage]
	if vonageStats.TotalVolume != 1 {
		t.Errorf("Vonage volume = %d; want 1", vonageStats.TotalVolume)
	}
	if vonageStats.TotalEstimatedCost != 0.10 {
		t.Errorf("Vonage Est Cost = %f; want 0.10", vonageStats.TotalEstimatedCost)
	}

	// Assertions for Country
	countries := tracker.GetCountryMetrics()

	vnStats := countries["Vietnam"]
	if vnStats.TotalVolume != 2 {
		t.Errorf("Vietnam volume = %d; want 2", vnStats.TotalVolume)
	}
	if vnStats.TotalActualCost != 0.045 {
		t.Errorf("Vietnam Act Cost = %f; want 0.045", vnStats.TotalActualCost)
	}

	sgStats := countries["Singapore"]
	if sgStats.TotalVolume != 1 {
		t.Errorf("Singapore volume = %d; want 1", sgStats.TotalVolume)
	}

	// 5. Revised actual cost on a second terminal callback (mitigation path) adds only the delta
	msg1.ActualCost = 0.05
	tracker.OnMessageUpdated(ctx, msg1)
	byProvider = tracker.GetProviderMetrics()
	twilioStats = byProvider[providers.ProviderTwilio]
	if twilioStats.TotalActualCost < 0.049999 || twilioStats.TotalActualCost > 0.050001 {
		t.Errorf("Twilio Act Cost after revision = %f; want 0.05", twilioStats.TotalActualCost)
	}

	tot := tracker.Totals()
	if tot.TotalEstimatedCost < 0.199999 || tot.TotalEstimatedCost > 0.200001 {
		t.Errorf("Totals().TotalEstimatedCost = %f; want 0.20", tot.TotalEstimatedCost)
	}
	if tot.TotalActualCost < 0.049999 || tot.TotalActualCost > 0.050001 {
		t.Errorf("Totals().TotalActualCost = %f; want 0.05", tot.TotalActualCost)
	}
}

func TestCostTracker_OnMessageUpdated_nilMessage(t *testing.T) {
	tracker := NewInMemoryCostTracker()
	ctx := context.Background()
	tracker.OnMessageUpdated(ctx, nil)
	if tot := tracker.Totals(); tot.TotalVolume != 0 || tot.TotalEstimatedCost != 0 || tot.TotalActualCost != 0 {
		t.Fatalf("expected no aggregates, got %+v", tot)
	}
}
