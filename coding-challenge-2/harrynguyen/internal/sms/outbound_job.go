package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/providers"
)

// RunOutboundJob loads the persisted message and drives the full outbound send pipeline:
// carrier resolution → provider routing → adapter.Send → status update to Queue.
// Called from the BullMQ worker (worker/bullmq.go) and from the immediate in-process
// publisher (send_job.go) used by cmd/simulate and unit tests.
//
// Idempotent: if the message is no longer StatusNew (e.g. already processed by a duplicate
// job), the function returns nil without error.
func RunOutboundJob(ctx context.Context, s *SMSCoreServiceInstance, job SendJob) error {
	obslog.Init()
	ctx = workerContext(ctx, job)
	testMode := GetTestSMSMode(ctx)

	msg, err := s.repo.GetByID(ctx, job.MessageID)
	if err != nil {
		return fmt.Errorf("load message %s: %w", job.MessageID, err)
	}
	if msg.Status != providers.StatusNew {
		return nil
	}

	canonical := msg.PhoneNumber
	iso := msg.Country
	content := msg.Content

	resolvedCarrier, err := s.resolver.Resolve(canonical)
	if err != nil || resolvedCarrier == providers.CarrierUnknown {
		s.UpdateStatus(ctx, msg, providers.StatusSendFailed, "Carrier resolution failed")
		return fmt.Errorf("carrier resolution failed for: %s", canonical)
	}
	msg.Carrier = resolvedCarrier

	providerName, err := s.router.Route(iso, resolvedCarrier)
	if err != nil {
		s.UpdateStatus(ctx, msg, providers.StatusSendFailed, fmt.Sprintf("Routing failed: %v", err))
		return fmt.Errorf("routing failed: %w", err)
	}
	msg.Provider = providerName

	if err := s.UpdateStatus(ctx, msg, providers.StatusSendToProvider, fmt.Sprintf("Selected provider: %s", providerName)); err != nil {
		return err
	}

	provider, ok := s.router.Adapter(providerName)
	if !ok {
		s.UpdateStatus(ctx, msg, providers.StatusSendFailed, fmt.Sprintf("Missing adapter for: %s", providerName))
		return fmt.Errorf("missing adapter for provider: %s", providerName)
	}

	if testMode == TestSMSModeAlwaysFail {
		s.UpdateStatus(ctx, msg, providers.StatusSendFailed, "Forced failure from x-test-sms-mode=always-fail")
		return fmt.Errorf("forced failure from test mode: %s", testMode)
	}

	if testMode == TestSMSModeTransientFailureOnce {
		s.AddStatusLog(ctx, msg.ID, providers.StatusSendFailed, "Forced transient failure on first send attempt from x-test-sms-mode=transient-failure-once")
	}

	providerMsgID, estimatedCost, err := provider.Send(ctx, iso, resolvedCarrier, canonical, content)
	if err != nil {
		if testMode == TestSMSModeAlwaysSuccess {
			providerMsgID = fmt.Sprintf("forced-success-%d", time.Now().UnixNano())
			estimatedCost = 0
			s.AddStatusLog(ctx, msg.ID, providers.StatusSendToProvider, "Provider error bypassed by x-test-sms-mode=always-success")
		} else {
			s.UpdateStatus(ctx, msg, providers.StatusSendFailed, fmt.Sprintf("Provider send failed: %v", err))
			return fmt.Errorf("provider send failed: %w", err)
		}
	}

	msg.MessageID = providerMsgID
	msg.EstimatedCost = estimatedCost
	if err := s.UpdateStatus(ctx, msg, providers.StatusQueue, fmt.Sprintf("Accepted by provider. Provider ID: %s", providerMsgID)); err != nil {
		return err
	}

	return nil
}
