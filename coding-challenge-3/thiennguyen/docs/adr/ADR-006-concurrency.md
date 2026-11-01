# ADR-006 — Concurrency: compare-and-set on dispatch's own tables, guarded by partial unique indexes

- **Status:** Accepted
- **Covers:** AC-02, AC-05, AC-06, AC-09
- **Relates to:** [ADR-003](ADR-003-clean-architecture.md) (the lifecycle rules each statement is built from)

## Context

The races the system must survive:

| Race | Example | Bad outcome |
|---|---|---|
| Two rides claim one driver (AC-09) | Dispatch for R1 and for R2 both rank D001 first | D001 holds two offers |
| Accept vs. timeout | Driver accepts at 29.9 s, the expiry job runs at 30 s | Ride accepted *and* re-matched |
| Accept vs. reject | The same driver's two taps arrive together | Two outcomes for one offer |
| Accept vs. cancel | Passenger cancels while the driver accepts | Driver `BUSY` on a cancelled ride |
| Availability vs. offer | Driver goes `OFFLINE` while being offered | Offer held by an offline driver |
| Duplicate commands and events | Accept sent twice; an event redelivered | Two `RideAccepted` events, two offers |

Each module owns its tables and no transaction spans modules ([ADR-002](ADR-002-decoupled-contexts.md)). Every assignment decision is therefore made inside **dispatch**, on `dispatch_requests`, `driver_supply` and `ride_offers`; ride's `rides` and driver's `drivers` hold eventually consistent copies.

## Decision

1. **Every state change is a compare-and-set in SQL.** The `WHERE` clause states the status the change expects, and the number of affected rows is the outcome: `1` means this transaction won, `0` means another one changed the row first. The row lock taken by the `UPDATE` serialises competitors without explicit locks or retries.

   The guard of each statement is not written by hand: the persistence adapters build it from the domain's transition enums (`RideTransition`, `DispatchRequestTransition`, `OfferTransition`, `AssignmentTransition`) — `status IN (:from)` from the transition's `from()` statuses (`fromNames()`), plus the holder condition where the transition requires it ([ADR-003](ADR-003-clean-architecture.md)). The application asks the aggregate first (`DispatchRequest.evaluate(transition, driverId)`, `RideTransition.evaluate(status)`) to tell a duplicate from a conflict; the statement then makes the decision atomic.

   | Change | Statement (simplified) |
   |---|---|
   | Claim a driver | `UPDATE driver_supply SET assignment='OFFERED', current_ride_id=:ride WHERE driver_id=:d AND online AND assignment='NONE' AND latitude IS NOT NULL` |
   | Start a matching attempt | `UPDATE dispatch_requests SET version=version+1 WHERE ride_id=:r AND status='WAITING' RETURNING …` (locks the request) |
   | Offer, accept, return to waiting | `UPDATE dispatch_requests SET status=… WHERE ride_id=:r AND status=:expected [AND driver_id=:d]` |
   | Accept an offer | `UPDATE ride_offers SET status='ACCEPTED' WHERE … AND status='PENDING' AND expires_at > :now` |
   | Expire an offer | `UPDATE ride_offers SET status='EXPIRED' WHERE id=:id AND status='PENDING' AND expires_at <= :now` |
   | Start or end a trip | `UPDATE driver_supply SET assignment=… WHERE driver_id=:d AND assignment IN (…) AND current_ride_id=:ride` |
   | Ride status (ride module) | `UPDATE rides … WHERE id=:id AND status IN (…) [AND assigned_driver_id=:d]` |
   | Availability (driver module) | `UPDATE drivers SET availability=:to WHERE id=:id AND availability=:from AND assignment='NONE'` |

2. **The AC-09 claim is one local transaction in dispatch, spanning two aggregates.** Lock the `WAITING` request, rank free drivers from `driver_supply` ([ADR-004](ADR-004-matching.md)), claim the first one still free, move the request to `OFFERED`, insert the `PENDING` offer and write `RideOffered` and `DriverAssignmentChanged` to the outbox. A dispatch that loses the claim falls through to the next candidate; with none left the request stays `WAITING` and the ride stays `MATCHING`.
   - This transaction changes two aggregates, the dispatch request and the supply driver, because the AC-09 invariant (a driver holds at most one ride) spans both. It is a deliberate exception to "one aggregate per transaction": both live in dispatch's database, the request is always locked first (4), and partial unique indexes back the invariant (3). Splitting it would need a reservation and compensation for no gain inside one context ([ADR-003](ADR-003-clean-architecture.md) § 4).

3. **Constraints are the last line of defence**, so a regression cannot produce a double assignment:
   - `dispatch_requests_active_driver_uq`: `UNIQUE (driver_id) WHERE status IN ('OFFERED','ACCEPTED')`. No driver holds two active rides.
   - `ride_offers_pending_driver_uq` and `ride_offers_pending_ride_uq`: at most one `PENDING` offer per driver and per ride.
   - `driver_supply_current_ride_uq`: a ride holds at most one driver.
   - `CHECK` constraints tie the columns together: an `OFFERED`/`ACCEPTED` request has a driver and a `WAITING` one has none; `driver_supply.assignment <> 'NONE'` exactly when `current_ride_id` is set.
   - The `rides` and `drivers` copies deliberately carry **no** uniqueness on the driver: they are updated per ride and per driver on different partitions, and a correct but out-of-order delivery must not be rejected there (ADR-002).

4. **Lock order: request row first, then supply row**, in every dispatch transaction: a matching attempt, accept, reject, expiry and the release when a ride ends. An availability change reads the supply row without locking it and, if the driver holds an offer, expires it through the same request-then-supply path. Deadlocks between these paths are therefore impossible.

5. **`READ COMMITTED`, one transaction per command, inside one module.** Every write is conditional, so stronger isolation adds retries without adding safety. Each transaction also writes its outbox rows, never Kafka ([ADR-005](ADR-005-event-driven-dispatch.md)).

6. **Kafka keys give per-entity ordering.** `ride-events` and the ride-level `dispatch-events` are keyed by `rideId`; `driver-events` and `DriverAssignmentChanged` by `driverId`. One consumer thread handles a given ride at a time, which removes same-ride races between messages; races across rides on one driver are settled by (1).

7. **Copies apply only newer versions.** `driver_supply` keeps `availability_version` and `location_version` (the driver row's version); `drivers` keeps `assignment_version` (the supply row's version). A stale or duplicate copy is a no-op.

8. **Scheduled jobs are safe to run concurrently.** `OfferExpiryScheduler` and `MatchingSweeper` read candidate IDs without locks; each attempt then locks the request row, so a second run, or a run on another instance, waits and then finds nothing to do. `UndispatchedRideSweeper` and `OutboxRelay` claim rows with `FOR UPDATE SKIP LOCKED`.

### How each race resolves

| Race | Outcome |
|---|---|
| Two rides, one driver | Exactly one `driver_supply` claim succeeds; the other dispatch tries its next candidate. |
| Accept vs. timeout | Both lock the request first. Accept requires `expires_at > now` and expiry requires `expires_at <= now`, so exactly one applies; the loser sees the request already moved on. |
| Accept vs. reject | Both lock the request; the second finds it no longer `OFFERED` for that driver and returns the offer unchanged or `409 OFFER_NOT_HELD`. |
| Accept vs. cancel | The ride's terminal state wins. Ride ignores a `RideAccepted` that arrives after `CANCELLED`; dispatch, on `RideCancelled`, cancels the pending offer or releases the `BUSY` driver. |
| Availability vs. offer | The driver module refuses `OFFLINE` while its copy shows `OFFERED`/`BUSY`. If the copy lags, going offline succeeds and dispatch expires the offer (`OfferExpiredV1.reason = DRIVER_OFFLINE`). **Residual race:** an offline event read while dispatch is claiming that same driver sees the driver free and only clears the online flag; the new offer then times out after 30 s. Accepted, rather than locking supply before the request and inverting the lock order. |
| Duplicates | Repeated commands return the current offer or ride with no new event ([ADR-007](ADR-007-idempotency.md)); redelivered events are skipped by `processed_events`, and the compare-and-set would refuse them anyway. |

## Verification

- `AC09ConcurrentAssignmentAcceptanceIT` (Testcontainers): `twoRidesCompetingForOneDriverProduceExactlyOneOffer` (repeated 20×, barrier-synchronised), `theLosingRideFallsThroughToTheNextDriver` (10×), `acceptAndRejectRacingForOneOfferHaveOneWinner` (10×).
- `PersistenceConstraintsIT` bypasses the services and writes a second active assignment directly; the indexes above reject it.
- `SystemInvariants` runs after the acceptance tests: no driver holds two active rides, at most one pending offer per ride and per driver, supply and requests agree, no excluded driver is re-offered, history contains only legal edges.
- Not yet covered by a dedicated race test: accept vs. timeout and accept vs. cancel. Their outcomes are covered one after the other by `AC05RideOfferAcceptanceIT`, and the concurrent case relies on the request-row lock in (4).

## Consequences

- No distributed lock service; correctness depends on dispatch's tables living in one database. That holds per city or region ([scalability.md](../scalability.md), T4).
- Losing a claim costs one extra `UPDATE` per contested candidate.
- Hot drivers in dense areas cause row-lock contention at very high load; the mitigation is geo-partitioned dispatch (scalability S2).
- Ride and driver statuses lag dispatch by the outbox and consumer latency (milliseconds); tests wait for them.
- Accept, reject, expiry and the ride ending likewise touch the request, an offer and the supply driver in one transaction, under the same lock order.

## Alternatives considered

| Option | Why not |
|---|---|
| `SELECT … FOR UPDATE` then update | Same safety, more round trips and a wider lock window. Used only where dispatch must read before deciding (the request row). |
| Optimistic `@Version` everywhere plus retry | Turns contention into exceptions and retry loops; a compare-and-set on the status is more direct. |
| Redis or ZooKeeper distributed lock | Extra infrastructure and lease-expiry hazards, and still needs database constraints for safety. |
| `SERIALIZABLE` isolation | Correct, but serialization failures need retry logic everywhere and cost throughput. |
| Driver module owns `OFFERED`/`BUSY` and dispatch asks it to claim | Puts the AC-09 race across two modules; needs a distributed lock or a two-phase reservation (ADR-002). |
