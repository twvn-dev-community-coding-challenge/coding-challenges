# ADR-007 — Idempotency: durable keys in PostgreSQL, event IDs for consumers, no-op transitions

- **Status:** Accepted
- **Covers:** AC-01, AC-05, AC-06, AC-08, AC-10

## Context

Duplicates arrive at three layers: client retries of `POST /rides`, repeated driver commands (`accept`, `reject`), and Kafka redelivery (at-least-once). Each layer needs its own identity — a client request key is **not** an event identity, because one ride produces many events.

## Decision

### 1. Ride creation — `Idempotency-Key` header

- Header required: `400 IDEMPOTENCY_KEY_REQUIRED` if missing, `400 IDEMPOTENCY_KEY_TOO_LONG` above 128 characters.
- Table `idempotency_keys(key PK, request_hash, ride_id, created_at)`.
- `request_hash = SHA-256(canonical JSON of the validated body)`.
- In **one transaction**: `INSERT … ON CONFLICT DO NOTHING`; if inserted → create ride + outbox row and set `ride_id`. No status column is needed: the key row and the ride commit together, so a committed key always has its ride. If the key already exists:

| Existing record | Response |
|---|---|
| same hash | replay original: `201` + same ride (header `Idempotent-Replayed: true`) |
| concurrent duplicate | its insert blocks on the unique index until the first transaction commits, then takes the replay path |
| different hash | `422 Unprocessable Entity` (`IDEMPOTENCY_KEY_REUSED`) |

- Keys are meant to be kept 24 h. The Redis replay entries expire after 24 h; the Postgres rows are **not purged yet** (README § Known limitations).
- Validation failures (`400`) do not consume the key.
- A Redis replay cache (`rh:idem:{key}`) may answer a repeat before this table is touched; a cache miss always falls through to this path ([ADR-008](ADR-008-caching.md)).

### 2. Commands — natural idempotency through state

- `accept` (dispatch) by the driver whose acceptance dispatch already recorded (request `ACCEPTED` with that driver) → `200` with the offer, no event.
- `reject` (dispatch) repeated by a driver already excluded from the ride → `200` with that driver's latest offer, no event.
- `cancel` / `complete` (ride) repeated when the ride is already `CANCELLED` / `COMPLETED` → `200` with the ride, no event.
- Anything else → `409` with the current status in the problem body: `OFFER_NOT_HELD` or `OFFER_EXPIRED` from dispatch, `INVALID_TRANSITION` from ride, `INVALID_DRIVER_TRANSITION` from driver.
- The ride and driver rules are `RideTransition.evaluate(status)` and `DriverTransition.evaluate(status)` (the kernel's `StateTransition`), returning `APPLY | ALREADY_APPLIED | REJECTED` (`TransitionOutcome`); unit-tested in `RideTransitionTest` and `DriverTransitionTest`. Dispatch's rules are `DispatchRequest.evaluate(transition, driverId)` over `DispatchRequestTransition` (only the holder may answer an offer), unit-tested in `DispatchRequestTest`; each decision is then applied as a compare-and-set ([ADR-006](ADR-006-concurrency.md), [ADR-003](ADR-003-clean-architecture.md)).

### 3. Event consumers — `eventId`

- Every event carries a UUID `eventId`, assigned when the outbox row is written.
- Consumer inserts `processed_events(consumer, event_id)` **in the same transaction** as its state change; a conflict means duplicate → skip and commit the offset.
- Conditional updates (ADR-006) make even an un-deduplicated replay harmless; `processed_events` additionally prevents duplicate side effects (e.g. a second offer).
- The driver copies kept by driver and dispatch also carry their owner's version and apply an event only if it is newer, so a correct but out-of-order delivery is harmless there too. Ride's copy applies dispatch-events as status transitions: a duplicate is a no-op, but an out-of-order event is ignored, so it relies on per-ride order ([ADR-002](ADR-002-decoupled-contexts.md), [ADR-005](ADR-005-event-driven-dispatch.md) § Consequences).
- Dispatch opens at most one request per ride (`INSERT … ON CONFLICT DO NOTHING` on `ride_id`), so a RideRequested republished by `UndispatchedRideSweeper`, which carries a new `eventId`, is also a no-op ([ADR-005](ADR-005-event-driven-dispatch.md)).

## Consequences

- Idempotency is durable and transactional with the data it protects — survives restarts and has no cache-consistency window.
- One extra indexed insert per create and per consumed event.
- `processed_events` is pruned by `EventRetention` after 14 days, longer than the topics' 7-day retention ([ADR-005](ADR-005-event-driven-dispatch.md) decision 10).
- `idempotency_keys` grows; it should be pruned after 24 hours (the replay window the cache already uses, [ADR-008](ADR-008-caching.md)), which is **not implemented yet**.

## Verification

- `AC01CreateRideAcceptanceIT`: same key and body replays the original ride; same key with a different body → `422`; replay stays correct when the cache entry is lost.
- `AC08IdempotencyAcceptanceIT`: concurrent duplicate creates produce one ride; duplicate acceptance and rejection are harmless; a redelivered event is processed once.
- `AC06StateMachineAcceptanceIT`: duplicate transitions are answered without change; cancelling twice is harmless.
- `RequestFingerprintTest`: identical requests share a fingerprint and any difference changes it.

## Alternatives considered

| Option | Why not |
|---|---|
| Redis `SETNX` on the key at the API edge | Not atomic with the ride insert or the Kafka publish (dual write → lost rides); volatile; stores no request hash. Redis is used instead as a read-through **replay cache** in front of this table, populated after commit ([ADR-008](ADR-008-caching.md)); it never decides. |
| Use `idempotency_key` as the event ID | Breaks rematching — legitimate later events for the same ride would be discarded. |
| Kafka exactly-once (transactions) | Only covers Kafka→Kafka; our side effects are in Postgres, so idempotent consumers are still required. |
