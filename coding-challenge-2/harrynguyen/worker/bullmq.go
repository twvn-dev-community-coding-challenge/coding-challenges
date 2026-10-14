package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/dotdak/sms-otp/internal/sms"
	"github.com/ktbsomen/gobullmq"
	"github.com/ktbsomen/gobullmq/types"
	"github.com/redis/go-redis/v9"
)

// StartSMSWorker starts a BullMQ consumer that dequeues sms-outbound jobs and runs the
// outbound send pipeline (sms.RunOutboundJob) for each one.
func StartSMSWorker(ctx context.Context, rdb *redis.Client, queueName string, svc *sms.SMSCoreServiceInstance) error {
	if rdb == nil {
		return fmt.Errorf("redis client is nil")
	}
	if queueName == "" {
		return fmt.Errorf("queue name must be non-empty")
	}
	if svc == nil {
		return fmt.Errorf("sms service is nil")
	}

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	log.Printf("BullMQ SMS worker connecting (queue: %s)", queueName)

	workerOpts := gobullmq.WorkerOptions{
		Autorun:         true,
		Concurrency:     10,
		StalledInterval: 30000,
	}

	_, err := gobullmq.NewWorker(ctx, queueName, workerOpts, rdb,
		func(ctx context.Context, job *types.Job, _ gobullmq.WorkerProcessAPI) (interface{}, error) {
			return processOutboundSMSJob(ctx, svc, job.Id, job.Data)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to start BullMQ worker: %w", err)
	}

	log.Printf("BullMQ SMS worker started (queue: %s)", queueName)
	return nil
}

// processOutboundSMSJob is shared by the BullMQ worker callback and unit tests (miniredis cannot run gobullmq enqueue Lua).
func processOutboundSMSJob(ctx context.Context, svc *sms.SMSCoreServiceInstance, jobID string, jobData any) (interface{}, error) {
	log.Printf("SMS worker processing job [%s]", jobID)
	sendJob, err := sms.ParseSendJobPayload(jobData)
	if err != nil {
		return nil, fmt.Errorf("parse job payload: %w", err)
	}
	if err := sms.RunOutboundJob(ctx, svc, sendJob); err != nil {
		return nil, err
	}
	return "ok", nil
}

// NewSMSQueue creates a BullMQ queue client used by BullMQSendPublisher.
func NewSMSQueue(ctx context.Context, rdb *redis.Client, queueName string) (*gobullmq.Queue, error) {
	if rdb == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	if queueName == "" {
		return nil, fmt.Errorf("queue name must be non-empty")
	}
	return gobullmq.NewQueue(ctx, queueName, rdb)
}
