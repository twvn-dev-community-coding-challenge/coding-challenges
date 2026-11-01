# ADR-002 — Decoupled bounded contexts: each module owns its state and the others react to events

- **Status:** Accepted
- **Relates to:** [ADR-001](ADR-001-module-structure.md) (one Gradle project per module)

## Context

After ADR-001, `dispatch` was its own Gradle project but still depended on `ride` and `driver`. Every dispatch command locked the ride row, compare-and-set the driver row and inserted the offer **in one database transaction**, and dispatch read the pickup and the nearby drivers through other modules' APIs. Cancel and complete also lived in dispatch. None of that survives being split into separate services.

## Decision

### Ownership

| Module | Owns (source of truth) | Keeps a copy of | Publishes on |
|---|---|---|---|
| **ride** | ride lifecycle, history, idempotency keys; **cancel** and **complete** | dispatch's offer state (DRIVER_ASSIGNED / DRIVER_ACCEPTED / back to MATCHING) | `ride-events`: RideRequested (carries the pickup), RideCancelled, RideCompleted |
| **driver** | availability (OFFLINE/AVAILABLE), location | dispatch's assignment of the driver (OFFERED/BUSY) | `driver-events`: DriverAvailabilityChanged, DriverLocationUpdated |
| **dispatch** | `dispatch_requests`, offers, exclusions, **driver assignment** (`driver_supply.assignment`), matching | drivers' online flag and location (`driver_supply`) | `dispatch-events`: MatchingStarted, RideOffered, OfferRejected, OfferExpired, RideAccepted (key rideId), DriverAssignmentChanged (key driverId) |

- `dispatch` depends only on `:common` and `:contracts`. Its `package-info.java` declares `@ApplicationModule(allowedDependencies = {"common", "contracts"})`, which Spring Modulith verifies, and `DispatchModuleIT` boots it on its own ([ADR-001](ADR-001-module-structure.md) decision 7).
- `:contracts` is plain Java: topic names, wire names, `EventEnvelope`, and versioned payload records (`RideRequestedV1`, …). Any future service, in any language, can mirror it.

### The assignment stays atomic

The claim (`driver_supply` NONE → OFFERED), the request update (WAITING → OFFERED) and the offer insert happen in **one local transaction** inside dispatch. Accept, reject, expiry and the release on ride end do the same. So AC-09 and every rule in ADR-006 still hold, on dispatch's tables:

- partial unique indexes: one active request per driver, one pending offer per driver and per ride, one driver per ride;
- lock order: request row first, then supply row.

The `rides_active_driver_uq` and `drivers_current_ride_uq` indexes were removed. Those tables are now copies, updated per ride or per driver on different partitions; a correct but out-of-order delivery must not be rejected there.

### Consistency rules

| Situation | Rule |
|---|---|
| Ride status after an offer, rejection, expiry or acceptance | Updated by ride's projection of dispatch-events. Eventually consistent, typically within milliseconds. Each event is applied at most once (`processed_events`) and only as a compare-and-set. |
| **Cancel vs accept** | The ride's terminal state wins. Ride ignores a RideAccepted that arrives after cancellation. Dispatch, on RideCancelled, cancels a pending offer or releases a BUSY driver, whichever applies. |
| Complete | Allowed once the ride shows DRIVER_ACCEPTED. Dispatch releases the driver on RideCompleted. |
| **AC-02: OFFERED/BUSY driver goes offline** | The driver module refuses using its own copy of the assignment. If dispatch has just made an offer the copy has not seen yet, going offline succeeds, and dispatch treats DriverAvailabilityChanged(offline) for an OFFERED driver as the offer expiring (`OfferExpiredV1.reason = DRIVER_OFFLINE`): the driver is excluded and the ride re-matched. A driver who goes offline mid-trip keeps the trip and just receives no new offers. |
| **Ride ends before dispatch opens it** | RideRequested was dead-lettered and the passenger cancelled. Dispatch records the request directly as `CANCELLED`/`COMPLETED`, so a late RideRequested is skipped. A ride still `REQUESTED` 30 s after its RideRequested is republished by ride's `UndispatchedRideSweeper` ([ADR-005](ADR-005-event-driven-dispatch.md) § 7–8). |
| Out-of-order or duplicate copies | The driver copies carry the owner's version (driver row version for availability and location in `driver_supply`, supply version for the assignment in `drivers`) and are applied only if newer. Ride's copy is not versioned: each dispatch-event is a status compare-and-set, so a duplicate is a no-op, and order comes from the `rideId` key. An event that arrives out of order is ignored, which is why the outbox relay publishes from one instance at a time, under a Postgres advisory lock ([ADR-005](ADR-005-event-driven-dispatch.md) decision 2). |
| Residual race | An offline event that arrives while dispatch is claiming that same driver sees the driver as free and only flips the online flag. The new offer then expires after `dispatch.offer-timeout` (30 s). This is accepted rather than locking supply before the request, which would invert the lock order. |

### API changes

- `POST /rides/{id}/cancel` and `/complete` are served by `ride`.
- `POST /rides/{id}/offers/{driverId}/accept|reject` and `GET /rides/{id}/offers` are served by `dispatch` and return the **offer** (`status: ACCEPTED|REJECTED|…`), not the ride. Clients read the ride's new status from `GET /rides/{id}` a moment later.

## Consequences

- `dispatch` can be extracted as described in ADR-001: give it its own application, database and consumer groups. Nothing in it calls ride or driver.
- **Eventual consistency** of ride status and of the driver API's OFFERED/BUSY. Tests wait for it (`awaitRideStatus`, `awaitDriverStatus`); `SystemInvariants.assertProjectionsConverged` proves the copies catch up.
- **More events:** about seven per ride, plus one per driver location update. Location volume is the first scaling concern ([scalability.md](../scalability.md) S1).
- **GDPR:** `RideRequestedV1.pickup` and `DriverLocationUpdatedV1.location` put personal location data on Kafka and in the outbox.
  - Mitigations: topics carry identifiers plus only the location fields; every topic and DLT keeps records for 7 days (`Topics.RETENTION_DAYS`, [ADR-005](ADR-005-event-driven-dispatch.md) decision 3); published `outbox` rows are deleted after 1 day, and `processed_events` markers, which hold no personal data, after 14 days (`EventRetention`, [ADR-005](ADR-005-event-driven-dispatch.md) decision 10).
  - Coordinates are never logged.
  - Encrypting the location fields, or replacing them with a geocell such as H3, are the options if retention cannot be bounded.
- Dispatch consumes its own `dispatch-events` to rematch after a rejection or expiry, so a rematch is asynchronous just like the first match.

## Alternatives considered

| Option | Why not |
|---|---|
| Keep the cross-module transaction and split later | Every correctness guarantee depended on it; splitting would mean redesigning under pressure. |
| Dispatch calls ride and driver over HTTP | Turns one transaction into a distributed one and couples dispatch's availability to two other services. |
| Driver module owns OFFERED/BUSY and dispatch asks it to claim | Puts the AC-09 race across two services; it needs a distributed lock or a two-phase reservation. |
| Ride owns the offer state | Moves matching's core decision into the ride service; dispatch would only rank drivers. |
| Saga orchestrator | Unnecessary while each step is owned by exactly one module; choreography over owned topics is enough. |
