# Sequence Diagrams

Back to [architecture](architecture.md). Step names match the command table in
[architecture §4](architecture.md#4-command-flows-saga-lite).

1. [Create ride → match → offer → accept](#1-create-ride--match--offer--accept)
2. [Reject → rematch](#2-reject--rematch)
3. [Offer timeout](#3-offer-timeout)
4. [Two rides race for the same driver (AC-09)](#4-two-rides-race-for-the-same-driver-ac-09)
5. [Cancel during a pending offer](#5-cancel-during-a-pending-offer)
6. [Duplicate delivery and idempotent replay](#6-duplicate-delivery-and-idempotent-replay)

---

## 1. Create ride → match → offer → accept

```mermaid
sequenceDiagram
    autonumber
    actor P as Passenger
    participant API as HTTP API
    participant R as Ride module
    participant Bus as Event bus
    participant D as Dispatch
    participant Dr as Driver module
    actor DD as Driver D001

    P->>API: POST /rides (Idempotency-Key: K1)
    API->>R: reserve K1 (IN_PROGRESS)
    R->>R: create ride R001 (REQUESTED)
    R->>R: store K1 → response
    API-->>P: 201 {R001, REQUESTED}
    R--)Bus: RideCreated(R001)

    Bus--)D: RideCreated(R001) [worker = hash(R001) % N]
    D->>R: startMatching(R001) → MATCHING
    D->>Dr: snapshot AVAILABLE drivers
    D->>D: filter radius ≤ 2 km, exclude rejected, sort (distance, id)
    D->>Dr: claim(D001, R001): AVAILABLE → OFFERED
    D->>R: assign(R001, D001): MATCHING → DRIVER_ASSIGNED
    D->>D: create offer (PENDING, expiresAt = now + 30s), arm timer
    D--)Bus: OfferCreated(R001, D001)

    DD->>API: POST /rides/R001/offers/D001/accept
    API->>D: accept(R001, D001)
    D->>R: accept(D001): DRIVER_ASSIGNED → DRIVER_ACCEPTED
    D->>Dr: markBusy(D001, R001): OFFERED → BUSY
    D->>D: offer → ACCEPTED, cancel timer
    D--)Bus: OfferAccepted(R001, D001)
    API-->>DD: 200 {R001, DRIVER_ACCEPTED}
```

## 2. Reject → rematch

The rejected offer is recorded **before** D001 becomes `AVAILABLE` again. A rematch that
runs immediately afterwards already sees D001 in the exclusion set.

```mermaid
sequenceDiagram
    autonumber
    actor D1 as Driver D001
    participant API as HTTP API
    participant D as Dispatch
    participant R as Ride module
    participant Dr as Driver module
    participant Bus as Event bus

    D1->>API: POST /rides/R001/offers/D001/reject
    API->>D: reject(R001, D001)
    D->>R: releaseAssignment(D001): DRIVER_ASSIGNED → MATCHING, assignedDriverId = nil
    D->>D: offer(R001, D001) → REJECTED
    D->>Dr: release(D001, R001): OFFERED → AVAILABLE
    D--)Bus: OfferRejected(R001, D001)
    API-->>D1: 200 {R001, MATCHING}

    Bus--)D: OfferRejected(R001, D001)
    D->>D: exclusion(R001) = {D001}
    D->>Dr: claim(D002, R001): AVAILABLE → OFFERED
    D->>R: assign(R001, D002): MATCHING → DRIVER_ASSIGNED
    D->>D: create offer(R001, D002) PENDING
    D--)Bus: OfferCreated(R001, D002)

    Note over D: If no eligible driver remains:<br/>ride stays MATCHING, assignedDriverId = nil,<br/>publish MatchingExhausted(R001)
```

## 3. Offer timeout

The timer goes through the same conditional ride transition as a manual reject. If
the driver accepted in the same instant, only one of the two wins on the ride.

```mermaid
sequenceDiagram
    autonumber
    participant T as Offer timer
    participant D as Dispatch
    participant R as Ride module
    participant Dr as Driver module
    participant Bus as Event bus

    T->>D: expire(R001, D001) at expiresAt
    D->>D: offer(R001, D001) still PENDING?
    alt ride is DRIVER_ASSIGNED to D001
        D->>R: releaseAssignment(D001) → MATCHING
        D->>D: offer → EXPIRED
        D->>Dr: release(D001, R001) → AVAILABLE
        D--)Bus: OfferExpired(R001, D001) → rematch as in diagram 2
    else ride is DRIVER_ACCEPTED by D001 (accept won, saga unfinished)
        D->>Dr: markBusy(D001, R001) (idempotent)
        D->>D: offer → ACCEPTED (roll forward)
    else ride is CANCELLED / COMPLETED
        D->>Dr: release(D001, R001) (no-op if already released)
        D->>D: offer → CANCELLED
    end
```

## 4. Two rides race for the same driver (AC-09)

Workers for R001 and R002 run in parallel because the rides hash to different
partitions. Both see D001 as the nearest available driver in their snapshots. The
claim is a conditional update, so exactly one of them wins.

```mermaid
sequenceDiagram
    autonumber
    participant WA as Worker A (R001)
    participant WB as Worker B (R002)
    participant Dr as Driver module
    participant R as Ride module

    par snapshot
        WA->>Dr: candidates near R001 → [D001, D002]
    and
        WB->>Dr: candidates near R002 → [D001]
    end

    par claim
        WA->>Dr: claim(D001, R001) — AVAILABLE?
    and
        WB->>Dr: claim(D001, R002) — AVAILABLE?
    end
    Note over Dr: Atomic read-check-write on D001.<br/>Only the first succeeds.
    Dr-->>WA: ok (D001 OFFERED, currentRideId = R001)
    Dr-->>WB: ErrConflict (status = OFFERED)

    WA->>R: assign(R001, D001) → DRIVER_ASSIGNED
    WB->>WB: log dispatch.claim_lost, try next candidate
    WB->>WB: no candidates left
    Note over WB: R002 stays MATCHING, assignedDriverId = nil<br/>(with D002 in range it would be offered D002 instead)
```

## 5. Cancel during a pending offer

```mermaid
sequenceDiagram
    autonumber
    actor P as Passenger
    participant API as HTTP API
    participant D as Dispatch
    participant R as Ride module
    participant Dr as Driver module
    participant T as Offer timer

    P->>API: POST /rides/R001/cancel
    API->>D: cancel(R001)
    D->>R: cancel(): DRIVER_ASSIGNED → CANCELLED (returns assignedDriverId = D001)
    D->>D: pending offer(R001, D001) → CANCELLED, cancel timer
    D->>Dr: release(D001, R001): OFFERED → AVAILABLE
    API-->>P: 200 {R001, CANCELLED}

    Note over T,D: If the timer fires anyway it finds the offer<br/>not PENDING and does nothing.
    Note over D,R: A concurrent accept from D001 fails on<br/>ride.accept (ride is CANCELLED) → 409.
```

## 6. Duplicate delivery and idempotent replay

```mermaid
sequenceDiagram
    autonumber
    actor P as Passenger
    participant API as HTTP API
    participant R as Ride module
    participant Bus as Event bus
    participant D as Dispatch

    P->>API: POST /rides (K1, body B)
    API-->>P: 201 {R001}
    P->>API: POST /rides (K1, body B) — client retry
    API->>R: lookup K1, hash(B) matches
    API-->>P: 201 {R001} + Idempotent-Replayed: true (no new ride)
    P->>API: POST /rides (K1, body B')
    API-->>P: 422 IDEMPOTENCY_KEY_REUSED

    Bus--)D: RideCreated(R001, eventId E1)
    D->>D: inbox: mark E1, start matching
    Bus--)D: RideCreated(R001, eventId E1) — redelivery
    D->>D: inbox: E1 already processed → skip
    Note over D,R: Even if the inbox were bypassed, startMatching is a no-op<br/>and assign() on an already-assigned ride fails → the extra<br/>driver claim is compensated with release().
```
