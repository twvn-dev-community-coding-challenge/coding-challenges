package handlers

import (
	"context"

	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/sms"
)

// stubSMS is a minimal SMSCoreService for handler tests that do not exercise the real pipeline.
type stubSMS struct {
	sendErr        error
	msg            *providers.SMSMessage
	listErr        error
	cbErr          error
	gwErr          error
	notFoundWithGW bool
}

func (s *stubSMS) SendSMS(ctx context.Context, country, phoneNumber, content string) (*providers.SMSMessage, error) {
	if s.sendErr != nil {
		return nil, s.sendErr
	}
	if s.msg != nil {
		return s.msg, nil
	}
	return &providers.SMSMessage{ID: "stub-sms-1", Status: providers.StatusNew}, nil
}

func (s *stubSMS) HandleCallback(ctx context.Context, messageID string, status providers.MessageStatus, actualCost float64) error {
	return s.cbErr
}

func (s *stubSMS) GetMessage(ctx context.Context, id string) (*providers.SMSMessage, error) {
	if s.gwErr != nil {
		return nil, s.gwErr
	}
	return nil, sms.ErrNotFound
}

func (s *stubSMS) GetMessageWithLogs(ctx context.Context, id string) (*providers.SMSMessage, []*providers.StatusLog, error) {
	if s.gwErr != nil {
		return nil, nil, s.gwErr
	}
	if s.notFoundWithGW {
		return nil, nil, sms.ErrNotFound
	}
	return &providers.SMSMessage{ID: id, Status: providers.StatusQueue}, nil, nil
}

func (s *stubSMS) ListMessages(ctx context.Context, params sms.MessageListParams) ([]*providers.SMSMessage, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return []*providers.SMSMessage{}, nil
}

func (s *stubSMS) RegisterGlobalObserver(o sms.Observer) {}

func (s *stubSMS) RegisterSourceObserver(source string, o sms.Observer) {}
