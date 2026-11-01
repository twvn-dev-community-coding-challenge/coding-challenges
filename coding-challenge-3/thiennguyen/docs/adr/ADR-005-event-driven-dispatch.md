# ADR-005 — Event-driven dispatch: no dual writes, idempotent consumers, self-healing jobs

- **Status:** Accepted
- **Covers:** AC-04, AC-07, AC-10, observability

## Context

The brief asks for asynchronous dispatch (`RideCreated → event → matching`), and for the README to explain why it is asynchronous, what happens when a message is delivered twice, and what happens when processing fails. Writing to the database and to a broker in the same request is a dual write: one can succeed while the other fails.

## Decision

1. **Synchronous create, asynchronous dispatch.** `POST /rides` commits the ride in `REQUESTED` and returns `201`. Matching runs in a Kafka consumer. Reasons:
   - API latency does not depend on driver search.
   - A burst of ride requests is buffered by the broker.
   - Dispatch workers scale independently of the API.
   - A dispatch failure never fails ride creation.

2. **No dual writes.** No request, consumer or job writes to two systems as independent steps, where one can succeed and the other fail. Postgres is the only system written inside a business transaction. Every other write is derived from a committed row, and is either retried until it succeeds or allowed to be briefly stale:

   | Boundary | If written as two steps | How it is closed |
   |---|---|---|
   | State change → its event (commands, consumers, sweepers) | The state commits and the event is lost: a ride is never dispatched. Or the event is sent for a change that rolled back. | **Transactional outbox.** The event is an `outbox` row inserted in the change's own transaction (`OutboxWriter`; `JdbcOutboxWriter` requires `Propagation.MANDATORY`, so it cannot run outside one), and Kafka is never written in the request. The event exists if and only if the change committed. |
   | Outbox row → Kafka (`OutboxRelay`) | The record is sent and the row is not marked, so it is sent again. | `OutboxRelay` polls every 200 ms (`outbox.relay.interval-ms`) and keeps going while batches of 100 (`outbox.relay.batch-size`) come back full; each send waits up to 10 s for its acknowledgement. Each batch is one transaction: take the advisory lock `OutboxRelay.LOCK_KEY` (`pg_try_advisory_xact_lock`; if another instance holds it, this one publishes nothing), claim up to 100 unpublished rows in `id` order, send them all with `acks=all` and an idempotent producer, wait once for every acknowledgement, then set `published_at` on the batch. Any failure before the mark rolls the batch back and re-sends all of it: delivery is **at-least-once and never lost**, and consumers skip the repeats by `eventId` (decision 5). |
   | Kafka record → consumer's state and offset | The work commits and the offset does not, so the record is applied again. | The consumer inserts `(consumer, eventId)` into `processed_events` in the same transaction as its work (decision 5). The offset is committed only after that transaction (`RECORD` ack mode). |
   | Failed record → dead-letter topic and offset | The DLT record is written and the offset is not, so the record is retried and dead-lettered again. | Accepted: the DLT is for inspection and replay, and a replayed record is de-duplicated like any other. |
   | Postgres → Redis read cache | The state commits and the cache keeps an older value. | Not closed, but bounded ([ADR-008](ADR-008-caching.md)): the cache is written only after commit, with a versioned put, a 30 s TTL, and never used to decide a state change. |

   The outbox is custom, not Spring Modulith's Event Publication Registry and externalisation; see *Alternatives considered*. Today the `outbox` table is shared and owned by `common`, and each row names its topic. When a module becomes a service it takes its own copy of the table (see the migration comment).

3. **One topic per owner, keyed for ordering** ([ADR-002](ADR-002-decoupled-contexts.md)), each with 6 partitions locally and a `.DLT` dead-letter topic. Every topic and DLT keeps records for 7 days (`Topics.RETENTION_DAYS`, set as `retention.ms` by `MessagingConfig.ownedTopic`). `spring.kafka.admin.modify-topic-configs` applies it to topics that already exist, so the window never silently depends on a broker default:

   | Topic | Owner | Key | Events |
   |---|---|---|---|
   | `ride-events` | ride | `rideId` | RideRequested, RideCancelled, RideCompleted |
   | `dispatch-events` | dispatch | `rideId` (`driverId` for DriverAssignmentChanged) | MatchingStarted, RideOffered, OfferRejected, OfferExpired, RideAccepted, DriverAssignmentChanged |
   | `driver-events` | driver | `driverId` | DriverAvailabilityChanged, DriverLocationUpdated |

   The full catalogue is in [architecture.md § Event catalogue](../architecture.md#event-catalogue).

4. **Consumers** (listener ack mode `RECORD`, so the offset is committed only after the handler's transaction commits; 3 listener threads per consumer group per instance, `spring.kafka.listener.concurrency`, so 6 partitions keep two instances busy):
   - `DispatchEventListener` (group `dispatch`) consumes all three topics: it opens a request and makes the first matching attempt on RideRequested, rematches on its own OfferRejected/OfferExpired, releases the driver on RideCancelled/RideCompleted, and updates its supply copy from driver-events.
   - `RideDispatchEventsListener` (group `ride`) applies MatchingStarted, RideOffered, OfferRejected, OfferExpired and RideAccepted to the ride.
   - `DriverDispatchEventsListener` (group `driver`) applies DriverAssignmentChanged to the driver's copy.

5. **Duplicate delivery.** Each consumer inserts `(consumer, eventId)` into `processed_events` in the same transaction as its work; a conflict means "already done" and the record is skipped ([ADR-007](ADR-007-idempotency.md)). Even without that, the compare-and-set updates refuse a second offer for a ride that is no longer `WAITING` ([ADR-006](ADR-006-concurrency.md)).

6. **Processing failure.** The transaction rolls back and the offset is not committed. `DefaultErrorHandler` retries 3 times with exponential backoff (0.5 s, 1 s, 2 s), then `DeadLetterPublishingRecoverer` publishes the record to `<topic>.DLT` with the exception in its headers. A malformed record (`MalformedEventException`) goes to the DLT immediately.

7. **Self-healing: three periodic jobs**, so a dead-lettered or lost message never strands a ride:

   | Job | Finds | Does | Covers |
   |---|---|---|---|
   | `UndispatchedRideSweeper` (ride, every 5 s) | Rides still `REQUESTED` more than `ride.undispatched-after` (30 s) after their RideRequested was published; claimed with `FOR UPDATE SKIP LOCKED` | Publishes RideRequested again and resets `rides.dispatch_requested_at` | RideRequested dead-lettered, lost, or published while dispatch was down: dispatch never opened a request, so its own sweeper cannot see the ride |
   | `MatchingSweeper` (dispatch, every 5 s) | `WAITING` requests nobody has tried for `dispatch.sweeper-stale-after` (10 s) | Re-runs matching | A lost OfferRejected/OfferExpired; drivers who became available after a ride had no candidate |
   | `OfferExpiryScheduler` (dispatch, every 1 s) | `PENDING` offers past `expires_at` (30 s) | Expires them like a rejection | Drivers who never answer |

   A repeated RideRequested is harmless: dispatch opens a request at most once per ride (`INSERT … ON CONFLICT DO NOTHING`), so a republish that crosses a slow MatchingStarted is skipped.

8. **A ride that ends before dispatch opens it.** If RideCancelled reaches dispatch before any request exists (the RideRequested was dead-lettered and the passenger cancelled), dispatch inserts the request directly as `CANCELLED` or `COMPLETED`, with no pickup. A late or replayed RideRequested then conflicts on the primary key and is skipped, so a driver is never offered, or held `BUSY` for, a ride that has already ended.

9. **Tracing in logs.** The envelope carries `eventId`, `rideId` and `driverId`. Each listener puts them, and `eventType`, into the logging MDC before handling a record, and the console pattern prints `rideId`, `driverId` and `eventId` on every line, so every log line of a ride carries its identifiers ([README § Observability](../../README.md#observability)). W3C trace-context propagation through Kafka headers is not implemented.

10. **Retention.** `EventRetention` (common, hourly, `events.retention.*`) deletes what delivery no longer needs, in batches of 1000 that skip locked rows, so it never blocks the relay or a consumer and is safe on every instance:
    - outbox rows **published** more than 1 day ago. Unpublished rows are never deleted, however old. Published rows only help debugging, and their payloads hold pickup and driver locations ([ADR-002](ADR-002-decoupled-contexts.md), GDPR).
    - `processed_events` markers older than 14 days. A marker must outlive every possible redelivery of its event, so it must stay longer than `Topics.RETENTION_DAYS` (7 days), which every topic and `.DLT` topic uses (decision 3). `EventRetentionProperties` refuses to start otherwise.

11. **Delivery metrics**, so a stalled relay or a stream of dead letters shows on a dashboard before it shows as stuck rides:

    | Meter | Recorded by | Shows |
    |---|---|---|
    | `ridehailing.outbox.pending`, `ridehailing.outbox.oldest.age` (gauges) | `OutboxBacklog`, every 5 s, independently of the relay | The relay falling behind, or stopped |
    | `ridehailing.events.latency{consumer, event}` (timer) | `InboundEvents`, for every handled event | Time from the outbox insert to the consumer starting on it: outbox wait, relay, broker and consumer queue together |
    | `ridehailing.events.dead.letters{topic}` (counter) | the error handler, after the DLT write succeeds | Records that exhausted their retries, including projection events nothing reconciles |
    | `ridehailing.rides.republished` (counter) | `UndispatchedRideSweeper` | RideRequested events dispatch never saw |

    Dispatch's own meters (attempts, lost claims, time to offer, how offers end) are listed in [README § Observability](../../README.md#observability).

## Consequences

- **Eventual consistency.** A ride is visible in `REQUESTED` before it is offered, and its status follows dispatch's decisions a few milliseconds later; clients poll `GET /rides/{id}`.
- **Outbox polling** adds latency (≈ the poll interval) and database load. Batching lifts the relay from a few hundred to thousands of events per second (not measured), but it is still one publisher for every event in the database; location updates overtake it at about 100,000 drivers, where CDC (Debezium) replaces it ([scalability.md](../scalability.md) B5, S5).
- **One publisher per database, with any number of instances.** Every instance runs a relay, and the advisory lock lets one publish at a time. A batch is claimed only after the previous one committed, which is after Kafka acknowledged it, so events with one key keep their recorded order. This matters most to ride's copy of dispatch's state: it is not versioned, applies dispatch-events as status transitions, and ignores an event that arrives early (ADR-002). The driver copies are versioned and would converge in any order. The cost is that relaying does not scale out: a second instance is a standby. If a relay dies mid-batch, the next one re-sends that batch from its first row, so per key the first delivery of each event is still in order and the repeats are skipped by `eventId`.
- **Dead-lettered projection events are not reconciled.** A dead-lettered DriverAssignmentChanged or ride projection event leaves that copy stale until the next event for the same driver or ride. Dispatch's own state stays correct, and the DLT keeps the record for replay.
- **Republishing** adds one RideRequested per stuck ride every 30 s while dispatch is down, which bounds the extra outbox rows.
- **Marker retention is tied to topic retention.** Raising `Topics.RETENTION_DAYS` to 14 days or more, without raising `events.retention.processed-events`, fails at startup instead of letting a late replay be processed twice. Retention set on the broker or by hand outside the app is not checked. A DLT record older than 7 days is gone, so a failure must be investigated within a week.
- **A long relay outage keeps rows.** Unpublished rows are never deleted, so the outbox grows while Kafka is unreachable and shrinks a day after the relay catches up.
- When modules become separate services, nothing here changes: each service keeps its outbox, topics and consumer group ([ADR-001](ADR-001-module-structure.md)).

## Verification

`AC10EventDrivenDispatchAcceptanceIT`:

- `aNewRideIsDispatchedThroughTheBroker`: RideRequested on `ride-events` and RideOffered on `dispatch-events`, both keyed by `rideId`, carrying the pickup.
- `anUnprocessableEventIsDeadLetteredWithoutBlockingDispatch`: a poison record lands on `ride-events.DLT`, and a following ride is still dispatched.
- `theSweeperRecoversARideWhoseDispatchMessageWasLost`: a `WAITING` request nobody retried is matched by `MatchingSweeper`.
- `aRideWhoseRideRequestedNeverReachedDispatchIsRequestedAgain`: a ride stuck in `REQUESTED` with no dispatch request is republished once and offered.
- `aRideThatEndedBeforeDispatchOpenedItIsNeverOffered`: RideCancelled before RideRequested leaves the request `CANCELLED` and the nearby driver free.

`AC08IdempotencyAcceptanceIT.aRedeliveredEventIsProcessedOnlyOnce` re-publishes a real RideRequested and checks it is processed once.

`KafkaTopicsIT`: every topic and DLT has `retention.ms` of 7 days set on the topic itself, not inherited from the broker default.

`EventRetentionIT`: published outbox rows past retention are deleted while recent and unpublished ones stay; old de-duplication markers are deleted and recent ones stay; a backlog larger than one batch is deleted in one run.

`OutboxRelayIT`: while a second connection holds the relay's advisory lock, as another instance would, nothing is published and the backlog gauges report the waiting rows; once it is released they are published in order. A backlog of more than two batches for one key reaches Kafka in the order it was recorded.

`MetricsIT`: a dead-lettered record, a republished ride and handled events move their meters, along with dispatch's.

## Alternatives considered

| Option | Why not |
|---|---|
| In-memory bus (allowed by the brief) | Loses events on a crash and cannot demonstrate redelivery or partitioning. |
| Publish to Kafka directly, before or after commit | A dual write. Before commit, a rollback leaves an event for a ride that does not exist; after commit, a crash or broker outage loses the dispatch trigger. |
| Kafka transactions, or a chained Kafka and JDBC transaction manager | Kafka transactions make producing and consuming atomic within Kafka only. Chaining them with the JDBC transaction commits one and then the other (best-effort one-phase commit), so a crash between the two is still a dual write. Kafka does not support XA. |
| Synchronous matching inside `POST /rides` | Couples API latency and availability to dispatch; no buffering under load. |
| RabbitMQ / SQS | Fine for work queues; Kafka gives per-key ordering, replay and a natural partition-by-area evolution path. |
| **Spring Modulith events**: publish with `ApplicationEventPublisher`, persist in the Event Publication Registry (`spring-modulith-events-jdbc`), consume with `@ApplicationModuleListener`, externalise to Kafka with `@Externalized` (`spring-modulith-events-kafka`) | It would replace about 250 lines (`OutboxWriter`, `OutboxRelay`, `ProcessedEvents`), resubmit incomplete publications on restart, and make the events visible to Modulith's `Documenter`, `Scenario` tests and observability. Not chosen because: (1) the registry tracks completion per *in-process listener*, while here every module consumes from Kafka in its own consumer group, exactly as a split-out service would, so the in-process half would go unused; (2) per-ride ordering is explicit in the relay (one ordered table, keyed records) and tested; (3) consumers would still need `eventId` de-duplication, since Kafka delivery stays at-least-once. Worth revisiting if modules move to in-process listeners with Kafka only at the edge, which would also reverse ADR-002's "only through the owner's topic" rule. |
| Replay the DLT automatically | Retries a poison record forever; a sweeper that re-derives the event from ride's own state is simpler and covers lost messages too. |
