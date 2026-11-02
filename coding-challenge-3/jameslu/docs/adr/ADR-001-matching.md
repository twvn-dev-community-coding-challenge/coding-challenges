# ADR-001: Driver Matching Strategy

- **Status:** Accepted
- **Related:** [architecture §3–§4](../architecture.md#3-domain-model-and-state-machines), [ADR-002](ADR-002-concurrency.md)

## Context

When a ride enters `MATCHING`, dispatch must pick one driver to offer it to. The
brief fixes the policy: *nearest available driver within a configurable radius
(default 2 km)*, with ties broken by the lexicographically smallest `driverId`.
It also requires:

- Drivers who reject (or time out on) a ride are never offered that ride again.
- Matching uses the latest known driver location.
- If no driver is eligible, the ride stays in `MATCHING` with no `assignedDriverId`.

Four questions remain open: how to compute and search distances, how many drivers to
offer at once, where the exclusion rule lives, and how the strategy can evolve.

## Decision

### 1. Selection algorithm

```
candidates = drivers where
               status == AVAILABLE
               and driverId ∉ excluded(ride)
               and haversine(pickup, driver.location) ≤ radius
order by (distance asc, driverId asc)
```

- **Haversine** with Earth radius 6,371 km. It is accurate to ~0.5 % at city
  scale, which is far below the precision of GPS input.
- The comparison is `≤ radius`. A driver at exactly the radius is a candidate.
- Equal distances are compared exactly (`float64`, same inputs → same result), and
  then ordered by `driverId`. This makes selection fully deterministic, so tests can
  assert on it.

### 2. Behind a port

The matcher depends on one narrow interface, not on the driver store
(`internal/dispatch/matcher.go`):

```go
type DriverLocator interface {
    // AVAILABLE drivers that may be near the point. Implementations may
    // over-approximate (e.g. a whole geo cell); the matcher filters exactly.
    AvailableNear(ctx context.Context, p geo.Location, radiusKm float64) []driver.Candidate
}
```

The excluded set is passed in by the saga, which reads it from the offer store
(`OfferStore.Excluded`).

The in-memory `DriverLocator` is a linear scan. A geo-indexed implementation (Redis
`GEOSEARCH`, H3 cells) replaces it without touching the matcher or its tests (see
[architecture §9](../architecture.md#9-scalability-and-performance-design)).

### 3. The result is an ordered list, not one driver

The matcher returns the **whole ordered candidate list**. The saga tries the
candidates in order and stops at the first successful claim. The snapshot can be stale
(the nearest driver may have just been claimed by another ride or gone offline), so
"the nearest suitable driver" means **the nearest one we can actually claim**. A lost
claim moves on to the next candidate immediately, without searching again.

### 4. One offer at a time

Only one driver holds an offer for a ride at any time, as the brief's state model
requires (`DRIVER_ASSIGNED` has a single `assignedDriverId`).

### 5. Exclusion is derived from offer history

`excluded(ride)` = drivers with a `REJECTED` or `EXPIRED` offer for that ride. It
is computed from the offer store, so there is no second list to keep consistent. It
is scoped per ride: the same drivers stay eligible for other rides. Offers expire
into the exclusion set as well, because the brief says a timeout *"is treated as
rejected"*.

### 6. Configuration

`DISPATCH_RADIUS_KM` (default `2`) and `OFFER_TIMEOUT` (default `30s`) are read at
startup and injected into the matcher and the saga.

## Alternatives considered

| Alternative | Why not (now) |
|---|---|
| **Broadcast to the top N drivers, first accept wins** | Shorter time-to-accept, but contradicts the single-`assignedDriverId` model. N drivers would be locked in `OFFERED`, reducing supply for other rides. It also needs a "you were too late" path. A candidate for **(future)** high-demand zones. |
| **Spatial index now** (grid, geohash, k-d tree) | Not needed at 1K drivers, where a linear scan takes microseconds. The `DriverLocator` port keeps the option open without paying the complexity today. |
| **Road distance / ETA** instead of straight-line | Explicitly out of scope ("Traffic/ETA calculation"). It would slot in as a re-ranking step after the radius filter. |
| **Global batch optimisation** (assign many rides at once, e.g. Hungarian algorithm) | Better total pickup distance under heavy load, but adds latency (batch window) and complexity. The brief prescribes nearest-driver. Listed as an evolution path. |
| **Equirectangular approximation** instead of Haversine | Slightly faster, but Haversine is what the brief suggests, and the cost is negligible. |
| **Separate "rejected drivers" list on the ride** | Duplicates information the offer store already has, and can drift out of sync. |

## Consequences

**Positive**
- Deterministic and easy to test: every matching rule maps to one unit test.
- Stale snapshots are harmless: correctness comes from the claim ([ADR-002](ADR-002-concurrency.md)), not from the search.
- The search implementation can be swapped for scale without changing behaviour.

**Negative / accepted trade-offs**
- **Greedy.** The nearest driver for ride A may be the only driver near ride B.
  Acceptable for the brief. Batch matching is the evolution path.
- **Sequential offers.** A non-responsive driver delays the passenger by up to
  `OFFER_TIMEOUT` per attempt.
- **Timeout = permanent exclusion.** In a sparse area, a driver who simply missed the
  notification is never asked again for that ride. We follow the brief literally. A
  softer rule (re-include after all other candidates are exhausted) is a possible
  extension.
- **No retry when supply appears.** A ride left in `MATCHING` waits until something
  triggers matching again (out of scope, see
  [limitations](../architecture.md#12-known-limitations)).
