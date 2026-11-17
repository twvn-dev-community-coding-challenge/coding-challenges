# ADR-002: Concurrency and Cross-Module Consistency

- **Status:** Accepted
- **Related:** [architecture §4](../architecture.md#4-command-flows-saga-lite), [sequence diagrams 3–5](../sequence-diagram.md), [ADR-001](ADR-001-matching.md)

## Context

Dispatch workers run in parallel (one per bus partition), HTTP requests are
handled concurrently, and offer timers fire on their own goroutines. Creating
or resolving an offer touches **two aggregates owned by two modules**: the ride
and the driver. We need:

1. No driver ever holds two offers or two accepted rides at the same time (AC-09).
2. Every competing action on a ride resolves to exactly one valid outcome.
3. A design that still holds when Ride, Driver and Dispatch become separate services.
   That rules out relying on one shared database transaction.

### Race conditions identified

| # | Race | Without protection |
|---|---|---|
| R1 | Two rides' workers claim the same nearest driver | Driver offered (and later assigned) to two rides |
| R2 | Driver accepts at the moment the offer times out | Ride `DRIVER_ACCEPTED` while driver is back to `AVAILABLE`, or vice versa |
| R3 | Driver accepts while the passenger cancels | Cancelled ride with a `BUSY` driver |
| R4 | Driver switches to `OFFLINE` while being claimed | Offline driver holding an offer |
| R5 | `RideCreated` delivered twice and processed concurrently | Two offers for one ride, two drivers locked |
| R6 | Driver location changes during matching | Driver selected on a stale location |
| R7 | A delayed compensation/timeout releases a driver that has since moved on to another ride | Driver freed while serving someone |
| R8 | Driver rejects at the moment the offer times out | Double rematch, or the offer marked twice |
| R9 | Two `POST /rides` with the same `Idempotency-Key` arrive together | Two rides for one request |
| R10 | Passenger cancels between the saga's `ride.assign` and offer creation | A `PENDING` offer for a cancelled ride. The driver looks `AVAILABLE` but cannot take a new offer |

## Decision

### 1. Saga-lite, orchestrated by Dispatch

Each cross-aggregate command is a short sequence of **single-aggregate atomic steps**,
coordinated by the Dispatch module. Every step is idempotent, and only the offer command
needs an explicit compensation. The full step table is in
[architecture §4](../architecture.md#4-command-flows-saga-lite). The rules behind it:

- **Claim the contended resource first.** Offer = `driver.claim(r)` → `ride.assign(d)`.
  If the ride step fails (cancelled, or already assigned by a duplicate worker), compensate with
  `driver.release(r)`. That compensation is invisible to the passenger.
- **The ride is the single arbiter once an offer exists.** Accept, reject, timeout
  and cancel all begin with a conditional transition on the ride. Exactly one of the
  competing actions wins there. The losers get `ErrConflict` / no-op and never touch
  the driver. This resolves R2, R3 and R8 without locking two objects at once.
- **Release is conditional on the ride id.** `driver.release(r)` only acts when
  `currentRideId == r`. A late compensation or timer can never free a driver who has
  moved on (R7).
- **Timeouts double as crash recovery.** If a command dies between steps, the
  driver is left `OFFERED` (or the offer is left `PENDING`). The offer's timer resolves
  it later, and it looks at the ride (the source of truth) to decide whether to roll
  back (release) or roll forward (finish the accept), see
  [sequence diagram 3](../sequence-diagram.md#3-offer-timeout).

### 2. Atomic conditional updates per aggregate

Every step is an **atomic read-check-write on one aggregate**. The check is the domain
transition method, so the guard lives in the domain, not in the repository.

```go
// Repository port (same shape for Ride and Driver).
// Update loads the aggregate, runs fn, and saves the result atomically.
// fn calls domain transition methods; on error or changed=false nothing is saved.
Update(ctx context.Context, id string, fn func(*Driver) (bool, error)) (Driver, bool, error)
```

**In-memory adapter:** `Update` holds the repository's `sync.Mutex` for the whole
read → `fn` → write, works on a copy, and bumps `version`. Callers never hold
pointers into the store. This short critical section is the in-process equivalent of a
conditional `UPDATE`.

**Postgres equivalent (future adapter).** The same guards become `WHERE` clauses, and
"0 rows affected" maps to `ErrConflict`:

```sql
-- driver.claim(r): R1, R4
UPDATE drivers
   SET status = 'OFFERED', current_ride_id = $ride, version = version + 1
 WHERE id = $driver AND status = 'AVAILABLE';

-- driver.release(r): R7
UPDATE drivers
   SET status = 'AVAILABLE', current_ride_id = NULL, version = version + 1
 WHERE id = $driver AND current_ride_id = $ride AND status IN ('OFFERED', 'BUSY');

-- ride.assign(d): R5
UPDATE rides
   SET status = 'DRIVER_ASSIGNED', assigned_driver_id = $driver, version = version + 1
 WHERE id = $ride AND status = 'MATCHING';

-- ride.accept(d): R2, R3
UPDATE rides
   SET status = 'DRIVER_ACCEPTED', version = version + 1
 WHERE id = $ride AND status = 'DRIVER_ASSIGNED' AND assigned_driver_id = $driver;
```

### 3. Invariants also enforced by the offer store (defence in depth)

```sql
CREATE UNIQUE INDEX one_pending_offer_per_driver ON ride_offers (driver_id) WHERE status = 'PENDING';
CREATE UNIQUE INDEX one_pending_offer_per_ride   ON ride_offers (ride_id)   WHERE status = 'PENDING';
CREATE UNIQUE INDEX one_offer_per_ride_driver    ON ride_offers (ride_id, driver_id);
```

The in-memory offer store checks the same three invariants on insert. A bug in the
saga ordering then surfaces as an error instead of silently producing a double offer.

### 4. How each race is resolved

| # | Resolution |
|---|---|
| R1 | `claim` succeeds only from `AVAILABLE`. The loser moves on to its next candidate or leaves the ride in `MATCHING`. |
| R2 | Both start with a conditional ride transition (`accept` vs `releaseAssignment`). Only one applies. If accept wins after the timer started, the timer rolls forward. |
| R3 | `cancel` and `accept` both transition the ride. Whichever is second fails with `409`. If cancel wins, it releases the driver. |
| R4 | The availability API is also a guarded transition (`AVAILABLE → OFFLINE`). It races with `claim` on the same lock / row, so only one applies. |
| R5 | Per-ride partitioning serialises a ride's events. The inbox skips known `eventId`s. Even if both still run, `ride.assign` succeeds once and the other worker compensates its claim. |
| R6 | Accepted by design: the search reads a snapshot, and the claim checks status only. A driver who moved a few metres is still a fine match. See [architecture §9, trade-off 1](../architecture.md#architectural-trade-offs). |
| R7 | Conditional `release(r)`. |
| R8 | Same as R2: one `releaseAssignment` wins. The offer moves `PENDING → REJECTED` or `PENDING → EXPIRED` exactly once, and only the winner publishes the rematch event. |
| R9 | The idempotency store reserves the key atomically (`IN_PROGRESS`) before the ride is created. The concurrent duplicate gets `409`. |
| R10 | After creating the offer, the saga re-reads the ride. Cancel moves the ride *before* it looks for the offer, so one of the two always sees the other: either cancel closes the offer, or the saga sees `CANCELLED` and closes it itself. Covered by `TestCancelBetweenAssignAndOfferLeavesNoPendingOffer`. |

## Alternatives considered

| Alternative | Why not |
|---|---|
| **One transaction across ride + driver** (lock both, commit together) | Simplest inside a monolith, and always consistent. But it requires both aggregates in one database, which contradicts the module-to-service evolution the brief asks us to demonstrate. Two-phase commit across services is not worth its availability cost here. |
| **Distributed lock** (Redis/Redlock on `driverId`) | An extra moving part with well-known failure modes (lock expiry during GC pauses, clock drift). A conditional update on the owner of the data gives the same guarantee without one. With geo-sharding (architecture §9), claims stay local and never need it. |
| **Pessimistic `SELECT … FOR UPDATE`** | Correct, but holds row locks across application code and round trips. A single conditional `UPDATE` is shorter and cannot deadlock. |
| **Choreography** (modules react to each other's events) | No single place describes the offer lifecycle, and compensation logic is spread over three modules. Harder to test and to explain. |
| **Actor per driver** (one goroutine owns each driver's state) | Idiomatic in Go and race-free by construction, but harder to map onto a database-backed, multi-instance deployment than conditional updates. |
| **Optimistic version check everywhere** (`WHERE version = $v`) | Equivalent guarantee, but it forces read-then-write round trips and retries on any unrelated change (e.g. a location update would fail a claim). Status-based guards only conflict on the field that matters. |

## Consequences

**Positive**
- AC-09 is guaranteed by the data owner, not by timing or by luck.
- Every module keeps a local atomic boundary, so the design survives a split into services.
- Commands are safe to retry end-to-end, which is also the idempotency story for accept.

**Negative / accepted trade-offs**
- **Brief inconsistency windows.** Between saga steps the driver can be `OFFERED`
  while the ride is still `MATCHING` (and so on). These windows are short, invisible to the
  correctness rules above, and repaired by compensation or the timeout.
- **One mutex per in-memory repository** serialises all writes to that repository.
  At 1K drivers each critical section takes microseconds, so this is negligible.
  Per-key lock striping or the Postgres adapter removes it.
- **Timers are in-memory.** A process restart loses them (along with all state).
  With durable storage, a sweeper over `PENDING` offers with `expiresAt < now`
  replaces them **(future)**.

## Verification

Tests live in `internal/dispatch/concurrency_test.go` unless noted. Each one releases
goroutines together with a barrier and repeats the race on a fresh system 1,000×.

| Test | Shape |
|---|---|
| `TestConcurrentRidesCannotShareTheSameDriver` | Two rides, only D001 in range. Exactly one `PENDING` offer for D001. The other ride stays `MATCHING` with no `assignedDriverId`. |
| `TestConcurrentRidesGetDifferentDrivers` | Same race with D002 also in range. The rides end up with **different** drivers. |
| `TestConcurrentDuplicateMatchingCreatesOneOffer` | One ride matched twice at once. One offer, and the other driver is left `AVAILABLE` (claim compensated). |
| `TestConcurrentAcceptAndTimeout` | End state is (`ACCEPTED`, `DRIVER_ACCEPTED`, `BUSY`) or (`EXPIRED`, `MATCHING`, `AVAILABLE`), never a mix. |
| `TestConcurrentAcceptAndCancel` | Ride ends `CANCELLED`, driver `AVAILABLE`, whichever runs first. |
| `TestConcurrentRejectAndTimeout` | The offer resolves exactly once. Ride `MATCHING`, driver `AVAILABLE`. |
| `TestMemoryRepositoryConcurrentClaimHasOneWinner` (driver) | 64 goroutines claim one driver: exactly one wins. |
| `TestConcurrentCreateWithSameKeyCreatesOneRide` (ride) | 20 concurrent creates with one key, 200×: exactly one ride. |

The tests log how often each side of a race wins (e.g. `outcomes: map[ACCEPTED:248 EXPIRED:752]`),
which shows that both interleavings really happen. All tests run under `go test -race ./...`.
