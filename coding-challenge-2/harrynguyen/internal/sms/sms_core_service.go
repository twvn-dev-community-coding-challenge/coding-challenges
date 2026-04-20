package sms

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/phone"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/ratelimit"
)

// SMSCoreService defines the interface for the SMS core coordination layer.
type SMSCoreService interface {
	SendSMS(ctx context.Context, country, phoneNumber, content string) (*providers.SMSMessage, error)
	HandleCallback(ctx context.Context, messageID string, status providers.MessageStatus, actualCost float64) error
	GetMessage(ctx context.Context, id string) (*providers.SMSMessage, error)
	GetMessageWithLogs(ctx context.Context, id string) (*providers.SMSMessage, []*providers.StatusLog, error)
	ListMessages(ctx context.Context, params MessageListParams) ([]*providers.SMSMessage, error)
	RegisterGlobalObserver(o Observer)
	RegisterSourceObserver(source string, o Observer)
}

type SMSCoreServiceInstance struct {
	repo            Repository
	resolver        carrier.CarrierResolver
	router          providers.ProviderRouter
	limiter         *ratelimit.Limiter
	publisher       SendJobPublisher
	globalObservers []Observer
	sourceObservers map[string][]Observer
}

const smsPerPhonePerHour = 10

func NewSMSService(repo Repository, resolver carrier.CarrierResolver, router providers.ProviderRouter, limiter *ratelimit.Limiter, publisher SendJobPublisher) *SMSCoreServiceInstance {
	return &SMSCoreServiceInstance{
		repo:            repo,
		resolver:        resolver,
		router:          router,
		limiter:         limiter,
		publisher:       publisher,
		globalObservers: make([]Observer, 0),
		sourceObservers: make(map[string][]Observer),
	}
}

func (s *SMSCoreServiceInstance) RegisterGlobalObserver(o Observer) {
	s.globalObservers = append(s.globalObservers, o)
}

func (s *SMSCoreServiceInstance) RegisterSourceObserver(source string, o Observer) {
	if source == "" {
		return
	}
	s.sourceObservers[source] = append(s.sourceObservers[source], o)
}

func (s *SMSCoreServiceInstance) SendSMS(ctx context.Context, country, phoneNumber, content string) (*providers.SMSMessage, error) {
	iso, err := phone.NormalizeCountry(country)
	if err != nil {
		return nil, fmt.Errorf("country: %w", err)
	}
	canonical, err := phone.Validate(iso, phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("phone: %w", err)
	}

	if s.limiter != nil {
		hour := time.Now().Unix() / 3600
		smsKey := fmt.Sprintf("sms:1h:phone:%s:%d", canonical, hour)
		if err := s.limiter.Allow(ctx, smsKey, smsPerPhonePerHour, time.Hour); err != nil {
			return nil, err
		}
	}

	// Create "New" message record (store canonical E.164 digits)
	msg := &providers.SMSMessage{
		SendSource:  SendSourceFromContext(ctx),
		Country:     iso,
		PhoneNumber: canonical,
		Content:     content,
		Status:      providers.StatusNew,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}
	notifyMessageCreated()
	s.AddStatusLog(ctx, msg.ID, providers.StatusNew, "Message created")

	if s.publisher == nil {
		if err := s.UpdateStatus(ctx, msg, providers.StatusSendFailed, "SMS send job publisher is not configured"); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("sms send job publisher is not configured")
	}

	job := SendJob{
		MessageID:     msg.ID,
		TestSMSMode:   GetTestSMSMode(ctx),
		RequestID:     obslog.RequestID(ctx),
		CorrelationID: obslog.CorrelationID(ctx),
	}
	if err := s.publisher.Publish(ctx, job); err != nil {
		if uerr := s.UpdateStatus(ctx, msg, providers.StatusSendFailed, fmt.Sprintf("Queue enqueue failed: %v", err)); uerr != nil {
			return nil, uerr
		}
		return nil, fmt.Errorf("queue enqueue failed: %w", err)
	}
	s.AddStatusLog(ctx, msg.ID, providers.StatusNew, "Queued for outbound processing")

	return msg, nil
}

// ProcessSendJob loads the message, runs carrier resolution, routing, and provider send.
// It is idempotent: if the message is no longer in StatusNew, it returns nil without error.
func (s *SMSCoreServiceInstance) ProcessSendJob(ctx context.Context, job SendJob) error {
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

func (s *SMSCoreServiceInstance) HandleCallback(ctx context.Context, messageID string, status providers.MessageStatus, actualCost float64) error {
	obslog.Init()
	msg, err := s.repo.GetByMessageID(ctx, messageID)
	if err != nil {
		args := []any{
			"event", obslog.EventSMSCallback,
			"step", obslog.StepCallbackLookup,
			"provider_message_id", messageID,
			"callback_status", string(status),
			"actual_cost", actualCost,
			"err", err,
		}
		if rid := obslog.RequestID(ctx); rid != "" {
			args = append(args, "request_id", rid)
		}
		if cid := obslog.CorrelationID(ctx); cid != "" {
			args = append(args, "correlation_id", cid)
		}
		if errors.Is(err, ErrNotFound) {
			args = append(args, "reason", "not_found")
			obslog.L.WarnContext(ctx, obslog.MsgSMSCallbackLookupFailed, args...)
		} else {
			args = append(args, "reason", "repository_error")
			obslog.L.ErrorContext(ctx, obslog.MsgSMSCallbackLookupFailed, args...)
		}
		return fmt.Errorf("failed to find message by ID %s: %w", messageID, err)
	}

	base := callbackLogAttrs(ctx, msg)
	obslog.L.InfoContext(ctx, obslog.MsgSMSCallbackDispatch,
		append(base,
			"step", obslog.StepCallbackDispatch,
			"previous_status", string(msg.Status),
			"callback_status", string(status),
			"actual_cost", actualCost,
		)...)

	if actualCost > 0 {
		msg.ActualCost = actualCost
	}

	err = s.applyStatusWithRecovery(ctx, msg, status, fmt.Sprintf("Asynchronous update received. Actual cost: %.4f", actualCost))
	if err != nil {
		obslog.L.ErrorContext(ctx, obslog.MsgSMSCallbackApplyFailed, append(base,
			"step", obslog.StepCallbackApply,
			"callback_status", string(status),
			"err", err,
		)...)
		return err
	}

	obslog.L.InfoContext(ctx, obslog.MsgSMSCallbackApplied, append(base,
		"step", obslog.StepCallbackApplied,
		"final_status", string(msg.Status),
		"actual_cost", msg.ActualCost,
	)...)
	return nil
}

func (s *SMSCoreServiceInstance) GetMessage(ctx context.Context, id string) (*providers.SMSMessage, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SMSCoreServiceInstance) GetMessageWithLogs(ctx context.Context, id string) (*providers.SMSMessage, []*providers.StatusLog, error) {
	msg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	logs, err := s.repo.GetStatusLogs(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return msg, logs, nil
}

func (s *SMSCoreServiceInstance) ListMessages(ctx context.Context, params MessageListParams) ([]*providers.SMSMessage, error) {
	return s.repo.ListMessages(ctx, params)
}
