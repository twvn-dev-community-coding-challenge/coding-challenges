# ADR-008 — Redis as a cache, never as the source of truth

- **Status:** Accepted
- **Covers:** AC-01 (read-back), AC-08 (replay), scalability
- **Relates to:** [ADR-007](ADR-007-idempotency.md) (Postgres stays the idempotency authority), [ADR-004](ADR-004-matching.md) (matching)

## Context

Without push notifications, passengers and drivers poll `GET /rides/{id}` and `GET /drivers/{id}` — the dominant read load. Client retries of `POST /rides` re-read the idempotency record. At larger scale, the matching pre-filter (available drivers near a point) becomes the hottest query ([scalability.md](../scalability.md)).

Using Redis as the idempotency **authority** (`SETNX` on the key) would lose or duplicate rides ([ADR-007](ADR-007-idempotency.md) § Alternatives considered). Redis is still useful — as a cache with strict rules.

## Decision

### Rules that apply to every cache

1. **Postgres decides; Redis only answers reads.** No state transition, claim or idempotency decision depends on a cache hit being present. `CacheIT.everyFlowStillWorksWhileRedisIsUnavailable` pauses the Redis container and runs create, dispatch, accept and read; the other tests run with Redis up.
2. **Write to Redis only after the Postgres transaction commits** (`TransactionSynchronization.afterCommit`). A rolled-back transaction never reaches the cache.
3. **Versioned puts.** Every cached value carries the row's `version`; a Lua script stores it only if it is newer than what is cached. This stops a slow reader from overwriting a fresh value with an old one.
4. **Fail open.** Lettuce command and connect timeouts are 100 ms; any Redis error is logged (`Cache … failed, falling back to the database`), counted as `ridehailing.cache.requests{result=error}`, and the request falls through to Postgres. The Redis health indicator is off (`management.health.redis.enabled: false`), so `/actuator/health` stays `UP` without Redis. Redis being down degrades latency, never correctness.
5. **Bounded staleness.** Every key has a TTL, so a missed invalidation heals itself.

### What is cached (MVP)

| Cache | Key | Value | Written | Invalidated | TTL |
|---|---|---|---|---|---|
| Ride read model | `rh:ride:{rideId}` | ride JSON + `version` | after create and after every ride transition (versioned put) | overwritten by the newer version; TTL | 30 s |
| Driver read model | `rh:driver:{driverId}` | driver JSON (availability, location, assignment copy) + `version` | after availability / location / status change | versioned put | 30 s |
| Idempotency replay | `rh:idem:{key}` | `request_hash`, `rideId` | after the create transaction commits | never (immutable once committed) | 24 h (the intended Postgres retention, not yet enforced there) |

- `GET /rides/{id}` → cache hit returns immediately; miss → Postgres → versioned put.
- `POST /rides` → cache lookup on `Idempotency-Key` first:
  - hit with the same hash → replay from the ride cache / Postgres without opening a write transaction;
  - hit with a different hash → `422` (safe: entries exist only for committed keys and never change);
  - miss → the full Postgres path of ADR-007, which stays authoritative (a cold cache or flushed Redis cannot cause a duplicate).
- Command endpoints (`accept`, `reject`, `cancel`, `complete`, availability) **always** read Postgres; they never consult the cache.

### What could be cached later (designed, not built)

A flag such as `dispatch.candidate-source = postgres | redis` (default `postgres`) would switch the matching pre-filter:

- `rh:drivers:available` — a Redis GEO set of `AVAILABLE` drivers, updated after commit on availability, location and status changes.
- Matching uses `GEOSEARCH … BYRADIUS 2 km ASC` to get candidate IDs, then the **same** Postgres CAS claim and exclusion check as today (ADR-006). A stale entry therefore costs one failed claim, never a wrong assignment.
- Redis error → fall back to the SQL bounding-box query.
- This is the first scaling step in [scalability.md](../scalability.md) (S1). It was left out of this implementation so the core is proven without it.

## Verification

| Test | Asserts |
|---|---|
| `CacheIT.readingARideCachesItWithItsVersion` | a read stores the ride with its row version |
| `CacheIT.everyCommittedStateChangeRefreshesTheCachedRide` | each committed transition refreshes the cached ride |
| `CacheIT.anOlderVersionNeverOverwritesANewerOne` | the Lua versioned put refuses an older version |
| `CacheIT.everyFlowStillWorksWhileRedisIsUnavailable` | with the Redis container paused, create, dispatch, accept and `GET` still succeed |
| `AC01CreateRideAcceptanceIT.replayStaysCorrectWhenTheCacheEntryIsLost` | a replay after the idempotency entry is gone is still answered from Postgres, with one ride |

Not covered by a dedicated test yet: a rolled-back transaction leaving the cache untouched, and a `422` answered from the cache.

## Consequences

- One more container in `compose.yaml` and Testcontainers setup.
- Read latency and DB read load drop for polling clients; writes gain one after-commit Redis call (async-tolerant, failure-ignored).
- Cache metrics (`ridehailing.cache.requests{cache, result}`) exposed via Actuator/Micrometer.
- Two places hold ride state; the version rule and TTL bound the inconsistency, and only reads see it.

## Alternatives considered

| Option | Why not |
|---|---|
| Redis as idempotency authority | Not atomic with the ride insert; volatile ([ADR-007](ADR-007-idempotency.md)). |
| Spring `@Cacheable` with `transactionAware()` cache manager | Defers puts to after commit, but cannot do versioned puts, so the stale-write race remains. May still be used for the static passenger directory. |
| Write-through (update Redis inside the transaction) | Cache updated even when the transaction later rolls back. |
| No cache | Valid for the brief's scale; rejected because polling is the dominant load and the cache is the cheapest scaling lever. |
