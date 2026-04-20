package sms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dotdak/sms-otp/internal/carrier"
	"github.com/dotdak/sms-otp/internal/obslog"
	"github.com/dotdak/sms-otp/internal/providers"
	"github.com/dotdak/sms-otp/internal/ratelimit"
	"github.com/ktbsomen/gobullmq"
	"github.com/ktbsomen/gobullmq/types"
)

const SMSOutboundJobName = "sms-outbound"

// SendJob is the BullMQ job payload for asynchronous outbound SMS processing.
type SendJob struct {
	MessageID     string `json:"message_id"`
	TestSMSMode   string `json:"test_sms_mode,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

// SendJobPublisher enqueues outbound SMS work for async processing.
type SendJobPublisher interface {
	Publish(ctx context.Context, job SendJob) error
}

// BullMQSendPublisher is the production publisher — enqueues jobs into BullMQ / Redis.
type BullMQSendPublisher struct {
	queue *gobullmq.Queue
}

func NewBullMQSendPublisher(q *gobullmq.Queue) *BullMQSendPublisher {
	return &BullMQSendPublisher{queue: q}
}

// Publish implements SendJobPublisher.
func (p *BullMQSendPublisher) Publish(ctx context.Context, job SendJob) error {
	// gobullmq's worker dereferences job.Opts.RemoveOnComplete on completion;
	// setting explicit retention prevents a nil-pointer panic in the vendor library.
	_, err := p.queue.Add(ctx, SMSOutboundJobName, job,
		gobullmq.AddWithJobID(job.MessageID),
		gobullmq.AddWithAttempts(3),
		gobullmq.AddWithRemoveOnComplete(types.KeepJobs{Count: 1000}),
		gobullmq.AddWithRemoveOnFail(types.KeepJobs{Count: 500}),
	)
	if err != nil {
		return fmt.Errorf("bullmq enqueue: %w", err)
	}
	return nil
}

// immediateSendPublisher runs the outbound pipeline synchronously in the same process.
// Used by cmd/simulate and tests that don't want a Redis dependency.
type immediateSendPublisher struct {
	svc *SMSCoreServiceInstance
}

func (p *immediateSendPublisher) Publish(ctx context.Context, job SendJob) error {
	return RunOutboundJob(ctx, p.svc, job)
}

// NewSMSServiceWithImmediateDispatch wires the immediate (non-BullMQ) publisher.
// Use this for cmd/simulate and unit tests; use NewSMSService + BullMQSendPublisher in production.
func NewSMSServiceWithImmediateDispatch(repo Repository, resolver carrier.CarrierResolver, router providers.ProviderRouter, limiter *ratelimit.Limiter) *SMSCoreServiceInstance {
	s := NewSMSService(repo, resolver, router, limiter, nil)
	s.publisher = &immediateSendPublisher{svc: s}
	return s
}

// ParseSendJobPayload decodes a BullMQ job.Data value into a SendJob.
// gobullmq can deliver the payload as a string, []byte, or map[string]interface{}.
func ParseSendJobPayload(data any) (SendJob, error) {
	if data == nil {
		return SendJob{}, fmt.Errorf("empty job data")
	}
	var raw []byte
	switch v := data.(type) {
	case string:
		raw = []byte(v)
	case []byte:
		raw = v
	case json.RawMessage:
		raw = []byte(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return SendJob{}, fmt.Errorf("marshal job data: %w", err)
		}
		raw = b
	}
	var job SendJob
	if err := json.Unmarshal(raw, &job); err != nil {
		return SendJob{}, fmt.Errorf("unmarshal send job: %w", err)
	}
	if job.MessageID == "" {
		return SendJob{}, fmt.Errorf("missing message_id")
	}
	return job, nil
}

// workerContext rebuilds request/trace context values that were stored in the job payload.
func workerContext(ctx context.Context, job SendJob) context.Context {
	out := ctx
	if job.RequestID != "" {
		out = obslog.WithRequestID(out, job.RequestID)
	}
	if job.CorrelationID != "" {
		out = obslog.WithCorrelationID(out, job.CorrelationID)
	}
	return WithTestSMSMode(out, job.TestSMSMode)
}
