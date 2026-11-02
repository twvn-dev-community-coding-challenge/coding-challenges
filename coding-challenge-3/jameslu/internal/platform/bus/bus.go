// Package bus is an in-memory message bus with the delivery semantics we
// would expect from a real broker:
//
//   - partitioned by RideID: one worker per partition, so events of a ride are
//     handled in order while different rides are handled in parallel;
//   - at-least-once: a failing or panicking handler is retried with
//     exponential backoff, then dead-lettered;
//   - an inbox per consumer that skips event ids it has already processed.
package bus

import (
	"context"
	"fmt"
	"hash/fnv"
	"log/slog"
	"slices"
	"sync"
	"time"

	"ridedispatch/internal/platform/events"
	"ridedispatch/internal/platform/trace"
)

type Handler func(ctx context.Context, e events.Event) error

type Options struct {
	Partitions  int           // default 4
	MaxAttempts int           // default 5
	Backoff     time.Duration // first retry delay, doubled per attempt; default 50ms
	Logger      *slog.Logger
}

type DeadLetter struct {
	Event    events.Event
	Consumer string
	Attempts int
	Err      string
}

type Bus struct {
	opts       Options
	log        *slog.Logger
	partitions []*partition
	stop       chan struct{}
	workers    sync.WaitGroup

	mu        sync.Mutex
	subs      []*subscription
	inflight  int
	dead      []DeadLetter
	isStopped bool
}

type subscription struct {
	consumer string
	types    []string // nil means every type
	handler  Handler

	mu        sync.Mutex
	processed map[string]bool
}

type delivery struct {
	event   events.Event
	sub     *subscription
	attempt int
}

type partition struct {
	mu    sync.Mutex
	queue []delivery
	wake  chan struct{}
}

func New(opts Options) *Bus {
	if opts.Partitions <= 0 {
		opts.Partitions = 4
	}
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 5
	}
	if opts.Backoff <= 0 {
		opts.Backoff = 50 * time.Millisecond
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}
	b := &Bus{opts: opts, log: opts.Logger, stop: make(chan struct{})}
	for range opts.Partitions {
		b.partitions = append(b.partitions, &partition{wake: make(chan struct{}, 1)})
	}
	return b
}

// Subscribe registers handler for the given event types (nil = all types).
// consumer names the subscriber's inbox.
func (b *Bus) Subscribe(consumer string, types []string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, &subscription{consumer: consumer, types: types, handler: handler, processed: map[string]bool{}})
}

func (b *Bus) Start() {
	for _, p := range b.partitions {
		b.workers.Add(1)
		go b.run(p)
	}
}

// Stop stops the workers. Undelivered events are dropped, like an in-memory
// broker losing its state.
func (b *Bus) Stop() {
	b.mu.Lock()
	if b.isStopped {
		b.mu.Unlock()
		return
	}
	b.isStopped = true
	b.mu.Unlock()
	close(b.stop)
	b.workers.Wait()
}

// Publish enqueues one delivery per matching subscriber. It never blocks, so
// handlers may publish follow-up events to their own partition.
func (b *Bus) Publish(_ context.Context, evs ...events.Event) {
	for _, e := range evs {
		b.mu.Lock()
		var targets []*subscription
		for _, s := range b.subs {
			if s.types == nil || slices.Contains(s.types, e.Type) {
				targets = append(targets, s)
			}
		}
		b.inflight += len(targets)
		b.mu.Unlock()

		for _, s := range targets {
			b.enqueue(delivery{event: e, sub: s, attempt: 1})
		}
	}
}

// PartitionFor returns the partition index that handles rideID.
func (b *Bus) PartitionFor(rideID string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(rideID))
	return int(h.Sum32() % uint32(len(b.partitions)))
}

// WaitIdle blocks until every published event has been handled or
// dead-lettered, including scheduled retries.
func (b *Bus) WaitIdle(ctx context.Context) error {
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		b.mu.Lock()
		idle := b.inflight == 0
		b.mu.Unlock()
		if idle {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
		}
	}
}

func (b *Bus) DeadLetters() []DeadLetter {
	b.mu.Lock()
	defer b.mu.Unlock()
	return slices.Clone(b.dead)
}

func (b *Bus) enqueue(d delivery) {
	p := b.partitions[b.PartitionFor(d.event.RideID)]
	p.mu.Lock()
	p.queue = append(p.queue, d)
	p.mu.Unlock()
	select {
	case p.wake <- struct{}{}:
	default:
	}
}

func (b *Bus) run(p *partition) {
	defer b.workers.Done()
	for {
		p.mu.Lock()
		if len(p.queue) == 0 {
			p.mu.Unlock()
			select {
			case <-p.wake:
				continue
			case <-b.stop:
				return
			}
		}
		d := p.queue[0]
		p.queue = p.queue[1:]
		p.mu.Unlock()

		b.deliver(d)
	}
}

func (b *Bus) deliver(d delivery) {
	s, e := d.sub, d.event
	log := b.log.With("event_id", e.ID, "event_type", e.Type, "ride_id", e.RideID, "consumer", s.consumer)

	s.mu.Lock()
	duplicate := s.processed[e.ID]
	s.mu.Unlock()
	if duplicate {
		log.Info("bus.duplicate_skipped")
		b.done()
		return
	}

	err := safeCall(s.handler, trace.WithCausation(context.Background(), e.ID), e)
	if err == nil {
		s.mu.Lock()
		s.processed[e.ID] = true
		s.mu.Unlock()
		b.done()
		return
	}

	if d.attempt >= b.opts.MaxAttempts {
		log.Error("bus.dead_lettered", "attempt", d.attempt, "error", err.Error())
		b.mu.Lock()
		b.dead = append(b.dead, DeadLetter{Event: e, Consumer: s.consumer, Attempts: d.attempt, Err: err.Error()})
		b.mu.Unlock()
		b.done()
		return
	}

	delay := b.opts.Backoff << (d.attempt - 1)
	log.Warn("bus.redelivery", "attempt", d.attempt, "retry_in", delay.String(), "error", err.Error())
	d.attempt++
	time.AfterFunc(delay, func() { b.enqueue(d) })
}

func (b *Bus) done() {
	b.mu.Lock()
	b.inflight--
	b.mu.Unlock()
}

func safeCall(h Handler, ctx context.Context, e events.Event) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v", r)
		}
	}()
	return h(ctx, e)
}
