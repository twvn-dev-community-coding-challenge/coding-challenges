package sms

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/dotdak/sms-otp/internal/providers"
)

func TestApplyStatusWithRecovery_QueueSkipsSendToCarrier(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:          "sms-q",
		MessageID:   "prov-q",
		Country:     "PH",
		PhoneNumber: "639171234567",
		Content:     "hi",
		Carrier:     providers.CarrierGlobe,
		Provider:    providers.ProviderTwilio,
		Status:      providers.StatusQueue,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}

	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	if err := svc.HandleCallback(ctx, "prov-q", providers.StatusSendSuccess, 0.02); err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}

	updated, err := repo.GetByID(ctx, "sms-q")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != providers.StatusSendSuccess {
		t.Fatalf("status = %q, want Send-success", updated.Status)
	}

	logs, err := repo.GetStatusLogs(ctx, "sms-q")
	if err != nil {
		t.Fatal(err)
	}
	var sawRecovery bool
	for _, lg := range logs {
		if strings.Contains(lg.Metadata, "recovery: inferred send-to-carrier") {
			sawRecovery = true
			break
		}
	}
	if !sawRecovery {
		t.Fatal("expected a status log for inferred send-to-carrier recovery step")
	}
}

func TestApplyStatusWithRecovery_SendToProviderSkipsToDelivered(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:            "sms-p",
		MessageID:     "prov-p",
		Country:       "PH",
		PhoneNumber:   "639171234567",
		Content:       "hi",
		Carrier:       providers.CarrierGlobe,
		Provider:      providers.ProviderTwilio,
		Status:        providers.StatusSendToProvider,
		EstimatedCost: 0.01,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}

	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	if err := svc.HandleCallback(ctx, "prov-p", providers.StatusSendSuccess, 0.03); err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}

	updated, err := repo.GetByID(ctx, "sms-p")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != providers.StatusSendSuccess {
		t.Fatalf("status = %q, want Send-success", updated.Status)
	}
}

func TestApplyStatusWithRecovery_MitigateLateFailureAfterSuccess(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:          "sms-ok",
		MessageID:   "prov-ok",
		Country:     "PH",
		PhoneNumber: "639171234567",
		Content:     "hi",
		Carrier:     providers.CarrierGlobe,
		Provider:    providers.ProviderTwilio,
		Status:      providers.StatusSendSuccess,
		ActualCost:  0.04,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}

	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	msg.ActualCost = 0.05
	if err := svc.HandleCallback(ctx, "prov-ok", providers.StatusSendFailed, 0.05); err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}

	updated, err := repo.GetByID(ctx, "sms-ok")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != providers.StatusSendSuccess {
		t.Fatalf("status should remain Send-success, got %q", updated.Status)
	}
	if updated.ActualCost != 0.05 {
		t.Fatalf("actual cost = %v, want revised 0.05", updated.ActualCost)
	}

	logs, err := repo.GetStatusLogs(ctx, "sms-ok")
	if err != nil {
		t.Fatal(err)
	}
	var sawMitigation bool
	for _, lg := range logs {
		if strings.HasPrefix(lg.Metadata, "mitigation:") {
			sawMitigation = true
			break
		}
	}
	if !sawMitigation {
		t.Fatal("expected mitigation status log")
	}
}

func TestApplyStatusWithRecovery_CarrierRejectedAtSendToProvider(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:          "sms-cr",
		MessageID:   "prov-cr",
		Country:     "PH",
		PhoneNumber: "639171234567",
		Content:     "hi",
		Carrier:     providers.CarrierGlobe,
		Provider:    providers.ProviderTwilio,
		Status:      providers.StatusSendToProvider,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}

	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	if err := svc.HandleCallback(ctx, "prov-cr", providers.StatusCarrierRejected, 0); err != nil {
		t.Fatalf("HandleCallback: %v", err)
	}

	updated, err := repo.GetByID(ctx, "sms-cr")
	if err != nil {
		t.Fatal(err)
	}
	// Send-failed triggers automatic fallback to Send-to-provider in updateStatus.
	if updated.Status != providers.StatusSendToProvider {
		t.Fatalf("after recovery + fallback, status = %q, want Send-to-provider", updated.Status)
	}
}

func TestApplyStatusWithRecovery_SendToProviderMissingMessageID(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:          "sms-nomid",
		MessageID:   "prov-nomid",
		Country:     "PH",
		PhoneNumber: "639171234567",
		Content:     "hi",
		Carrier:     providers.CarrierGlobe,
		Provider:    providers.ProviderTwilio,
		Status:      providers.StatusSendToProvider,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}
	// Simulate inconsistent state: handoff without internal provider id on the row.
	msg.MessageID = ""
	if err := repo.Update(ctx, msg); err != nil {
		t.Fatal(err)
	}

	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	err := svc.HandleCallback(ctx, "prov-nomid", providers.StatusSendSuccess, 0)
	if err == nil {
		t.Fatal("expected error when recovering to Send-success without provider message_id on message")
	}
}

func TestRecoveryTelemetryHooks_QueueRecoveryThenMitigation(t *testing.T) {
	var mu sync.Mutex
	var kinds []string
	SetTelemetryHooks(nil, func(kind string) {
		mu.Lock()
		kinds = append(kinds, kind)
		mu.Unlock()
	})
	t.Cleanup(func() { SetTelemetryHooks(nil, nil) })

	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:          "sms-tel",
		MessageID:   "prov-tel",
		Country:     "PH",
		PhoneNumber: "639171234567",
		Content:     "hi",
		Carrier:     providers.CarrierGlobe,
		Provider:    providers.ProviderTwilio,
		Status:      providers.StatusQueue,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}

	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	if err := svc.HandleCallback(ctx, "prov-tel", providers.StatusSendSuccess, 0.01); err != nil {
		t.Fatalf("HandleCallback success: %v", err)
	}
	if err := svc.HandleCallback(ctx, "prov-tel", providers.StatusSendFailed, 0.02); err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if !slices.Contains(kinds, "recovered_queue_to_delivered") {
		t.Fatalf("expected recovered_queue_to_delivered in hooks, got %v", kinds)
	}
	if !slices.Contains(kinds, "mitigation") {
		t.Fatalf("expected mitigation in hooks, got %v", kinds)
	}
}

func TestApplyStatusWithRecovery_UnrecoverableReturnsError(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemRepository()
	msg := &providers.SMSMessage{
		ID:          "sms-new",
		MessageID:   "prov-new",
		Country:     "PH",
		PhoneNumber: "639171234567",
		Content:     "hi",
		Carrier:     providers.CarrierGlobe,
		Provider:    providers.ProviderTwilio,
		Status:      providers.StatusNew,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatal(err)
	}

	svc := NewSMSService(repo, nil, providers.NewSimpleProviderRouter(), nil, nil)
	err := svc.HandleCallback(ctx, "prov-new", providers.StatusSendSuccess, 0)
	if err == nil {
		t.Fatal("expected error for unrecoverable transition from New to Send-success")
	}
}
