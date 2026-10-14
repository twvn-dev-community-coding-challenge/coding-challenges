package sms

import (
	"context"
	"fmt"
	"time"

	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/providers"
)

func (s *SMSCoreServiceInstance) UpdateStatus(ctx context.Context, msg *providers.SMSMessage, newStatus providers.MessageStatus, metadata string) error {
	if !msg.Status.IsValidTransition(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", msg.Status, newStatus)
	}

	msg.Status = newStatus
	if err := s.repo.Update(ctx, msg); err != nil {
		return err
	}

	s.AddStatusLog(ctx, msg.ID, newStatus, metadata)

	s.notifyObservers(ctx, msg)

	// Automatic fallback triggered on delivery failure (User Story 4)
	if newStatus == providers.StatusCarrierRejected || newStatus == providers.StatusSendFailed {
		return s.UpdateStatus(ctx, msg, providers.StatusSendToProvider, "Automatic fallback triggered due to delivery failure")
	}

	return nil
}

func (s *SMSCoreServiceInstance) AddStatusLog(ctx context.Context, smsID string, status providers.MessageStatus, metadata string) {
	log := &providers.StatusLog{
		SMSID:     smsID,
		Status:    status,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	if err := s.repo.AddStatusLog(ctx, log); err != nil {
		obslog.Init()
		obslog.L.ErrorContext(ctx, obslog.MsgStatusLogWriteFailed, "sms_id", smsID, "err", err)
	}
}

func (s *SMSCoreServiceInstance) notifyObservers(ctx context.Context, msg *providers.SMSMessage) {
	for _, o := range s.globalObservers {
		o.OnMessageUpdated(ctx, msg)
	}
	if msg == nil || msg.SendSource == "" {
		return
	}
	for _, o := range s.sourceObservers[msg.SendSource] {
		o.OnMessageUpdated(ctx, msg)
	}
}

// callbackLogAttrs builds Loki-friendly log fields for SMS provider webhook handling (masked phone).
func callbackLogAttrs(ctx context.Context, msg *providers.SMSMessage) []any {
	obslog.Init()
	out := []any{
		"event", obslog.EventSMSCallback,
		"sms_id", msg.ID,
		"provider_message_id", msg.MessageID,
		"send_source", msg.SendSource,
		"provider", string(msg.Provider),
		"country", msg.Country,
		"phone_suffix", maskCallbackPhone(msg.PhoneNumber),
	}
	if rid := obslog.RequestID(ctx); rid != "" {
		out = append(out, "request_id", rid)
	}
	if cid := obslog.CorrelationID(ctx); cid != "" {
		out = append(out, "correlation_id", cid)
	}
	return out
}

func maskCallbackPhone(digits string) string {
	if len(digits) < 4 {
		return "****"
	}
	return "***" + digits[len(digits)-4:]
}
