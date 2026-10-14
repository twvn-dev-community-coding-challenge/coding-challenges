package sms

import (
	"context"
	"fmt"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/providers"
)

// applyStatusWithRecovery applies a provider callback status. It uses the same transition
// rules as the synchronous send path, then recovers common out-of-order webhook sequences,
// and mitigates conflicting events after terminal Send-success.
func (s *SMSCoreServiceInstance) applyStatusWithRecovery(ctx context.Context, msg *providers.SMSMessage, newStatus providers.MessageStatus, metadata string) error {
	if msg.Status.IsValidTransition(newStatus) {
		return s.UpdateStatus(ctx, msg, newStatus, metadata)
	}

	// Mitigation: delivery already confirmed — ignore late failures, duplicate noise, etc.
	if msg.Status == providers.StatusSendSuccess {
		if err := s.repo.Update(ctx, msg); err != nil {
			return err
		}
		s.AddStatusLog(ctx, msg.ID, msg.Status, fmt.Sprintf("mitigation: ignored invalid transition to %s — %s", newStatus, metadata))
		notifyRecoveryEvent(obslog.RecoveryKindMitigation)
		obslog.RecoveryEvent(ctx, obslog.RecoveryKindMitigation, msg.ID, msg.MessageID, string(msg.Status), string(newStatus))
		s.notifyObservers(ctx, msg)
		return nil
	}

	// Carrier rejected before we ever reached Queue — map to send-failed (valid from Send-to-provider).
	if newStatus == providers.StatusCarrierRejected && msg.Status == providers.StatusSendToProvider {
		notifyRecoveryEvent(obslog.RecoveryKindMappedCarrierReject)
		obslog.RecoveryEvent(ctx, obslog.RecoveryKindMappedCarrierReject, msg.ID, msg.MessageID, string(msg.Status), string(providers.StatusSendFailed))
		return s.UpdateStatus(ctx, msg, providers.StatusSendFailed, fmt.Sprintf("recovery: mapped carrier-rejected at send-to-provider — %s", metadata))
	}

	// Recovery: webhook reports delivered while intermediate states were skipped.
	if newStatus == providers.StatusSendSuccess {
		switch msg.Status {
		case providers.StatusQueue:
			notifyRecoveryEvent(obslog.RecoveryKindRecoveredQueueToDelivered)
			obslog.RecoveryEvent(ctx, obslog.RecoveryKindRecoveredQueueToDelivered, msg.ID, msg.MessageID, string(msg.Status), string(newStatus))
			if err := s.UpdateStatus(ctx, msg, providers.StatusSendToCarrier, "recovery: inferred send-to-carrier before delivery confirmation"); err != nil {
				return err
			}
			return s.UpdateStatus(ctx, msg, providers.StatusSendSuccess, metadata)
		case providers.StatusSendToProvider:
			if msg.MessageID == "" {
				s.AddStatusLog(ctx, msg.ID, msg.Status, fmt.Sprintf("recovery failed: missing provider message_id for transition to %s — %s", newStatus, metadata))
				notifyRecoveryEvent(obslog.RecoveryKindRecoveryFailedMissingProviderID)
				obslog.RecoveryEvent(ctx, obslog.RecoveryKindRecoveryFailedMissingProviderID, msg.ID, msg.MessageID, string(msg.Status), string(newStatus))
				return fmt.Errorf("cannot recover transition from %s to %s: empty provider message_id", msg.Status, newStatus)
			}
			notifyRecoveryEvent(obslog.RecoveryKindRecoveredProviderToDelivered)
			obslog.RecoveryEvent(ctx, obslog.RecoveryKindRecoveredProviderToDelivered, msg.ID, msg.MessageID, string(msg.Status), string(newStatus))
			if err := s.UpdateStatus(ctx, msg, providers.StatusQueue, "recovery: reconciled queued state before delivery confirmation"); err != nil {
				return err
			}
			if err := s.UpdateStatus(ctx, msg, providers.StatusSendToCarrier, "recovery: inferred send-to-carrier before delivery confirmation"); err != nil {
				return err
			}
			return s.UpdateStatus(ctx, msg, providers.StatusSendSuccess, metadata)
		}
	}

	s.AddStatusLog(ctx, msg.ID, msg.Status, fmt.Sprintf("unrecoverable invalid transition to %s — %s", newStatus, metadata))
	notifyRecoveryEvent(obslog.RecoveryKindUnrecoverable)
	obslog.RecoveryEvent(ctx, obslog.RecoveryKindUnrecoverable, msg.ID, msg.MessageID, string(msg.Status), string(newStatus))
	return fmt.Errorf("invalid status transition from %s to %s", msg.Status, newStatus)
}
