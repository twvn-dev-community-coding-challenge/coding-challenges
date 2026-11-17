# James Lu - Ride-Hailing Dispatch Service

A ride-hailing dispatch service in Go. Passengers request rides, and the service
offers each ride to the nearest available driver. Drivers accept or reject, and
the ride is rematched until it is accepted or no eligible driver remains.

It is built as a **modular monolith whose modules talk to each other as if they
were already separate services**. Everything runs in one process with in-memory
storage and an in-memory message bus, so no external infrastructure is needed.

> **Design docs start here: [docs/architecture.md](docs/architecture.md).**
> Also: [sequence diagrams](docs/sequence-diagram.md),
> [ADR-001 Matching](docs/adr/ADR-001-matching.md),
> [ADR-002 Concurrency](docs/adr/ADR-002-concurrency.md).

## Team Members

- James Lu

## How to Run

Requires Go 1.24 or later (`brew install go`). No other dependencies.

```bash
cd coding-challenge-3/jameslu
go run ./cmd/server
```

The server listens on `:8080` and logs JSON to stdout. Configuration is through
environment variables:

| Variable | Default | Meaning |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DISPATCH_RADIUS_KM` | `2` | Matching radius |
| `OFFER_TIMEOUT` | `30s` | Time a driver has to respond to an offer |
| `DISPATCH_WORKERS` | `4` | Bus partitions (parallel dispatch workers) |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn` or `error` |

### Seed data

Registration is out of scope, so passengers `P001`–`P003` and four drivers are
created at startup. The drivers sit due north of the brief's pickup point
(10.7769, 106.7009), at the distances used in the brief's matching example:

| Driver | Distance from pickup | Status |
|---|---|---|
| D001 | 0.8 km | AVAILABLE |
| D002 | 1.5 km | AVAILABLE |
| D003 | 3.2 km (outside the 2 km radius) | AVAILABLE |
| D004 | 0.3 km | OFFLINE |

### Walkthrough

```bash
# 1. Request a ride. The response is 201 with status REQUESTED; matching runs asynchronously.
curl -s -X POST localhost:8080/rides \
  -H 'Idempotency-Key: demo-1' \
  -d '{"passengerId":"P001","pickup":{"latitude":10.7769,"longitude":106.7009},"destination":{"latitude":10.8231,"longitude":106.6297}}'

# 2. The ride is now offered to the nearest driver (DRIVER_ASSIGNED, D001).
curl -s localhost:8080/rides/<rideId>

# 3. D001 rejects; the ride is rematched to D002.
curl -s -X POST localhost:8080/rides/<rideId>/offers/D001/reject

# 4. D002 accepts, then the trip completes.
curl -s -X POST localhost:8080/rides/<rideId>/offers/D002/accept
curl -s -X POST localhost:8080/rides/<rideId>/complete

# 5. What happened to this ride?
curl -s localhost:8080/rides/<rideId>/timeline
```

Other endpoints: `POST /rides/{id}/cancel`, `GET /drivers/{id}`,
`PUT /drivers/{id}/availability` (`{"status":"AVAILABLE"|"OFFLINE"}`) and
`PUT /drivers/{id}/location` (`{"latitude":…,"longitude":…}`). The full list with
status codes is in [architecture §7](docs/architecture.md#7-http-api).

## How to Test

```bash
cd coding-challenge-3/jameslu
go test -race ./...
```

The suite takes roughly 15 seconds, because each concurrency test repeats its
race 1,000 times on a fresh system. To see how often each side of a race wins:

```bash
go test -race -v -run Concurrent ./internal/dispatch/
```

Overall statement coverage is about 93%
(`go test -coverpkg=./internal/... ./...`). Timeouts are tested with a fake
clock, so no test sleeps.

| Area | Where |
|---|---|
| State machines (valid, invalid, duplicate transitions) | `internal/ride/ride_test.go`, `internal/driver/driver_test.go` |
| Matching (nearest, radius, offline/busy excluded, none available, tie-break, latest location) | `internal/dispatch/matcher_test.go` |
| Offer saga, rematching, timeout, cancel, complete | `internal/dispatch/service_test.go` |
| Concurrency (AC-09 and the other races) | `internal/dispatch/concurrency_test.go` |
| Idempotency (ride creation, duplicate accept, duplicate events) | `internal/ride/service_test.go`, `internal/dispatch/service_test.go` |
| Event bus (ordering, parallelism, retry, dead letter, inbox) | `internal/platform/bus/bus_test.go` |
| HTTP API contract and timeline | `internal/app/api_test.go` |

The commit history follows red → green: each `test(...)` commit adds failing
tests, and the following `feat(...)` commit makes them pass.

## Approach

The modules are Ride, Driver and Dispatch, plus a small platform layer (bus,
clock, ids). Each module owns its data and exposes a service API. Dispatch
orchestrates the offer lifecycle by calling Ride and Driver. It never writes
their state directly.

- **Architecture.** Module boundaries, ownership and dependency direction, see
  [architecture §2](docs/architecture.md#2-system-context-and-module-boundaries).
- **Matching strategy.** Nearest `AVAILABLE` driver within the radius
  (Haversine), ties broken by `driverId`. The matcher returns an ordered candidate
  list, so a lost claim falls through to the next driver. Drivers who rejected or
  timed out are excluded for that ride only, derived from offer history. See
  [ADR-001](docs/adr/ADR-001-matching.md).
- **Concurrency strategy.** Every step is an atomic read-check-write on one
  aggregate (the in-memory equivalent of a conditional `UPDATE`). The offer saga
  **claims the driver first** (`AVAILABLE → OFFERED`), then assigns the ride, and
  compensates if the ride has moved on. After an offer exists, **the ride is the
  arbiter**: accept, reject, timeout and cancel race on a ride transition and only
  the winner touches the driver. Driver release is conditional on the ride id, so
  a late compensation can never free a driver serving another ride. See
  [ADR-002](docs/adr/ADR-002-concurrency.md).
- **Idempotency strategy.** Three layers.
  - HTTP: the `Idempotency-Key` is reserved before the ride is created. A replay
    returns the original response with `Idempotent-Replayed: true`. A different
    body with the same key gets `422`. A concurrent duplicate gets `409`.
  - Commands: every transition is a no-op when it has already happened, so a
    repeated accept returns `200` with no new effects.
  - Events: a per-consumer inbox skips known event ids, and the state machine
    guards make a duplicate harmless even if the inbox is bypassed.

  See [architecture §6](docs/architecture.md#6-idempotency-strategy).
- **Event-driven approach.** `RideCreated`, `OfferRejected` and `OfferExpired`
  trigger matching on a dispatch worker. The bus is partitioned by `rideId`, like
  a Kafka partition key, so a ride's events are handled in order while different
  rides run in parallel. Details in
  [architecture §5](docs/architecture.md#5-event-driven-dispatch).
  - *Why asynchronous:* creating a ride never waits for, or fails because of, the
    geo search. Matching scales separately from the API, and a failed attempt is
    retried without the passenger resubmitting.
  - *If a message is delivered twice:* the consumer's inbox skips known event ids.
    Even without it, the duplicate finds the ride already past matching (or loses
    the conditional `assign`) and does nothing.
  - *If processing fails:* the handler is retried with exponential backoff (5
    attempts, panics included), then the event is dead-lettered with an `ERROR` log.
    Steps are idempotent, so a retry after partial progress is safe. The ride stays
    in a visible, safe state (`REQUESTED`/`MATCHING`).
- **Observability.** JSON logs carry `ride_id`, `driver_id`, `event_id` and
  `causation_id`. Every event records its cause, so the chain
  `HTTP request → RideCreated → OfferCreated → OfferExpired → …` can be followed.
  `GET /rides/{id}/timeline` returns the same story as data. See
  [architecture §8](docs/architecture.md#8-observability).
- **Scalability considerations.** How the design evolves from 1K to 1M drivers
  (geo-sharded location store, regional dispatch workers, partitioned broker),
  with bottlenecks and trade-offs, is in
  [architecture §9](docs/architecture.md#9-scalability-and-performance-design).

## Design Decisions

- **Go.** Real threads make the concurrency tests meaningful on in-memory storage,
  and `go test -race` checks the code for data races. The standard library covers
  HTTP routing, JSON and structured logging, so there are no external dependencies.
- **In-memory only.** The brief allows it, and real infrastructure would add setup
  cost without strengthening the concurrency proof. The repositories have the
  semantics of conditional SQL updates, and ADR-002 shows the equivalent SQL.
- **Saga-lite, orchestrated by Dispatch**, rather than one transaction across ride
  and driver. A shared transaction is simpler inside a monolith but would not
  survive splitting the modules into services. Orchestration keeps the offer
  lifecycle readable in one place. Choreography would spread it over three modules.
- **Spec interpretations.** Where the brief is ambiguous (cancel after acceptance,
  timeout exclusion, duplicate commands, concurrent idempotency keys, …), each
  decision and its reasoning is in
  [architecture §10](docs/architecture.md#10-spec-interpretations).

## Known Limitations

- State and offer timers live in memory and are lost on restart.
- A ride left in `MATCHING` is not retried when a new driver becomes available.
- One offer at a time: time-to-accept is bounded by `OFFER_TIMEOUT` × attempts.
- Drivers learn about offers by polling `GET /drivers/{id}` (`currentRideId`).
  There is no push channel.
- No authentication: the `driverId` in the path is trusted.
- Idempotency keys are never expired.

## Challenges Faced

- **Ordering the saga steps.** Claiming the driver before assigning the ride makes
  the compensation invisible to the passenger. Recording a rejection before
  releasing the driver stops a racing rematch from offering the same ride back to
  them.
- **Duplicate matching for the same ride.** Two workers can claim the same driver
  for the same ride. That claim is a no-op for the second worker, so the saga uses
  the `changed` flag from the ride assignment to decide which worker creates the
  offer. Without it, the loser's compensation would undo the winner's work.
- **Proving races, not just passing tests.** The concurrency tests count how often
  each side wins, to confirm that both interleavings actually occur.

## With More Time, We Would...

- Add a Postgres adapter (conditional `UPDATE`s, partial unique indexes on
  offers) and a transactional outbox.
- Add a sweeper that re-drives rides stuck in `MATCHING` and offers past
  `expiresAt`. This would also make timers survive a restart.
- Export metrics: time-to-offer, acceptance rate, lost claims, dead letters.
- Replace the linear scan with a geo index behind the existing `DriverLocator` port.

## AI Tools Used

Claude Code (Claude Opus 5.5) was used throughout: to analyse the brief and past
reviews, brainstorm and write the design documents, and pair on the test-first
implementation. Every design decision was discussed and chosen before any code
was written. The design docs were committed first, and the code followed them
test by test.
