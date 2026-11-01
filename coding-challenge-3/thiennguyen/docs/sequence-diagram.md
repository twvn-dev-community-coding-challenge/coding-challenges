# Sequence diagrams

The main flows, at the level of modules and their transactions. A `TX` box is one Postgres transaction inside **one** module; nothing inside it talks to Kafka or to another module. Modules talk only through events on the owner's topic ([ADR-002](adr/ADR-002-decoupled-contexts.md)). Components are described in [`architecture.md`](architecture.md).

## 1. Create ride (idempotent) → asynchronous dispatch

```mermaid
sequenceDiagram
    autonumber
    actor P as Passenger
    participant R as ride
    participant DB as PostgreSQL
    participant K as Kafka
    participant D as dispatch

    P->>R: POST /rides with an Idempotency-Key
    rect rgb(225, 245, 238)
    Note over R,DB: TX
    R->>DB: record the key, create the ride REQUESTED, add RideRequested to the outbox
    end
    R-->>P: 201 Created
    R->>K: the outbox relay publishes RideRequested
    K->>D: RideRequested (diagram 2)
```

- **Retries.** The key, the ride and its event commit together, so a crash leaves either nothing or all three. A retry with the same key and body returns the same ride; a different body returns `422`; an unknown passenger returns `400` without consuming the key. A concurrent duplicate waits on the key's unique index, then replays.
- **Cache.** Redis can answer a replay before Postgres is touched, and is written only after commit, so a flushed cache can never create a second ride ([ADR-008](adr/ADR-008-caching.md)).
- **Relay.** The event reaches Kafka at least once and is never lost ([ADR-005](adr/ADR-005-event-driven-dispatch.md) decision 2).

## 1b. Read a ride (polling clients)

```mermaid
sequenceDiagram
    autonumber
    actor P as Passenger
    participant R as ride
    participant C as Redis
    participant DB as PostgreSQL

    P->>R: GET /rides/{id}
    R->>C: cached ride?
    alt miss or Redis down
        R->>DB: read the ride
        R->>C: store it, if newer than what is cached
    end
    R-->>P: 200 ride
```

Commands (`accept`, `reject`, `cancel`, `complete`) never read the cache; they always run against Postgres.

## 2. Dispatch: open the request, match, offer

```mermaid
sequenceDiagram
    autonumber
    participant K as Kafka
    participant D as dispatch
    participant R as ride
    participant DR as driver

    K->>D: RideRequested
    rect rgb(225, 245, 238)
    Note over D: TX in dispatch
    D->>D: open the request (once per ride), claim the nearest free driver, create a 30 s offer
    end
    D->>K: MatchingStarted, RideOffered, DriverAssignmentChanged
    K->>R: ride REQUESTED → MATCHING → DRIVER_ASSIGNED
    K->>DR: driver shows OFFERED
```

- **Duplicates.** A redelivered event, or a second RideRequested for a ride dispatch already opened or already saw end, is skipped in the same transaction.
- **Matching.** Free drivers (online, located, not held) within 2 km, minus those who already rejected or let this ride's offer expire, sorted by distance and then id ([ADR-004](adr/ADR-004-matching.md)). The claim succeeds only if the driver is still free ([ADR-006](adr/ADR-006-concurrency.md)).
- **No driver.** The request stays waiting (the ride shows `MATCHING`), and the sweeper retries later.

## 3. Reject → rematch → accept → complete

With the seed data, as in `make demo`: D001 (0.8 km) is offered first; after it rejects, the next nearest eligible driver is D005 (1.2 km).

```mermaid
sequenceDiagram
    autonumber
    actor D1 as Driver D001
    actor D2 as Driver D005
    actor P as Passenger
    participant DISP as dispatch
    participant RIDE as ride

    D1->>DISP: reject
    DISP->>DISP: TX: offer rejected, D001 excluded and released
    DISP-->>RIDE: OfferRejected → ride back to MATCHING
    DISP->>DISP: rematch without D001 → offer to D005
    DISP-->>RIDE: RideOffered → DRIVER_ASSIGNED (D005)

    D2->>DISP: accept
    DISP->>DISP: TX: offer accepted, D005 BUSY
    DISP-->>RIDE: RideAccepted → DRIVER_ACCEPTED
    Note over D2,DISP: accepting again returns the same offer, with no new event

    P->>RIDE: complete
    RIDE->>RIDE: TX: ride COMPLETED
    RIDE-->>DISP: RideCompleted → request closed, D005 free
```

Dispatch rematches by consuming its own OfferRejected from dispatch-events.

## 4. Offer timeout, and a driver going offline while offered

```mermaid
sequenceDiagram
    autonumber
    participant S as OfferExpiryScheduler
    participant DISP as dispatch
    participant DRV as driver
    participant RIDE as ride

    loop every second
        S->>DISP: offers past their 30 s
        DISP->>DISP: TX per offer: expire it, exactly like a rejection
    end

    DRV-->>DISP: DriverAvailabilityChanged (offline) for a driver dispatch holds OFFERED
    DISP->>DISP: TX: expire that offer, mark the driver offline
    DISP-->>RIDE: OfferExpired → ride back to MATCHING, rematch without this driver
```

A driver can go offline while offered because the driver module checked an older copy of the assignment. An `accept` racing either expiry is resolved by locking the request first: whichever commits first wins, and the other gets `409` or is skipped.

## 5. Two rides compete for the same driver (AC-09)

```mermaid
sequenceDiagram
    autonumber
    participant WA as Worker A (ride R1)
    participant DB as dispatch tables
    participant WB as Worker B (ride R2)

    par
        WA->>DB: claim D001 if free
    and
        WB->>DB: claim D001 if free
    end
    Note over DB: the row lock serialises the two claims
    DB-->>WA: claimed → R1 offered to D001
    DB-->>WB: not free → try the next driver, or leave R2 waiting
```

The claim is one compare-and-set update in dispatch's own tables. Partial unique indexes reject a double assignment even if the code regresses ([ADR-006](adr/ADR-006-concurrency.md)).

## 6. Cancel races accept

```mermaid
sequenceDiagram
    autonumber
    actor P as Passenger
    actor D as Driver
    participant RIDE as ride
    participant DISP as dispatch

    par
        P->>RIDE: cancel → ride CANCELLED
    and
        D->>DISP: accept → request ACCEPTED, driver BUSY
    end
    DISP-->>RIDE: RideAccepted → ignored, the ride is already CANCELLED
    RIDE-->>DISP: RideCancelled → request closed, driver free
```

If RideCancelled reaches dispatch first, the pending offer is cancelled and the accept returns `409 OFFER_NOT_HELD`. Either order ends with the ride `CANCELLED` and the driver free.

## 7. Processing failure

```mermaid
sequenceDiagram
    autonumber
    participant K as Kafka
    participant H as consumer
    participant DLT as dead-letter topic

    K->>H: event
    H--xH: fails, transaction rolled back
    K->>H: redelivered 3 times with backoff
    H->>DLT: still failing → dead-lettered, offset committed
    Note over K,DLT: the sweepers recover what was lost (table below)
```

The retries wait 0.5 s, 1 s and 2 s. A record that cannot be decoded goes to the DLT at once. DLT records are kept for 7 days. What recovers which loss:

| Lost or dead-lettered | Recovered by |
|---|---|
| RideRequested | `UndispatchedRideSweeper` (ride, every 5 s) republishes it for a ride still `REQUESTED` 30 s later; dispatch opens the request at most once, so a repeat is harmless. If the ride was cancelled first, dispatch has already recorded it as `CANCELLED` and skips the late RideRequested. |
| OfferRejected / OfferExpired (dispatch's own rematch trigger) | `MatchingSweeper` (dispatch, every 5 s): the request is back to waiting. |
| An unanswered offer | `OfferExpiryScheduler` after 30 s. |
| A projection event (ride's or driver's copy) | Not reconciled: the copy stays stale until the next event for that ride or driver; dispatch's state is correct and the record waits on the DLT ([ADR-005](adr/ADR-005-event-driven-dispatch.md) § Consequences). |
