# Scalability: 1,000 → 100,000 → 1,000,000 drivers

The implementation targets one city on one Postgres primary, one Kafka broker and one Redis node. This document explains what breaks first as the number of drivers grows, and how the design evolves without changing its correctness rules ([ADR-006](adr/ADR-006-concurrency.md), [ADR-007](adr/ADR-007-idempotency.md)).

## Load model

| | 1k drivers | 100k drivers | 1M drivers |
|---|---|---|---|
| Location updates (1 per driver every 4 s) | 250 writes/s | 25k writes/s | 250k writes/s |
| Ride requests at peak (≈ 10 % of drivers per minute) | ~2/s | ~170/s | ~1,700/s |
| Events through the outbox (one per location update, ≈ 7 per ride) | ~265/s | ~26k/s | ~262k/s |
| Drivers inside a 2 km pickup radius (dense city centre) | tens | hundreds | thousands |
| Polling reads (`GET /rides/{id}`, `GET /drivers/{id}`, every 2 s per active client) | ~1k/s | ~100k/s | ~1M/s |

## Bottlenecks

| # | Bottleneck | Where it bites | Why |
|---|---|---|---|
| B1 | **Driver location writes** | ~100k drivers | Every update is an MVCC row version on `drivers`, an outbox row and event, and a second row version on dispatch's `driver_supply` with its `(latitude, longitude)` index; vacuum and WAL volume grow with it. Location is high-churn, low-value-per-write data in a transactional store. |
| B2 | **Matching query cost** | ~100k drivers, dense areas | The bounding-box pre-filter is a B-tree range on two columns, so it scans every driver in the latitude band and then filters longitude. In dense areas the candidate set and the Haversine ranking grow with it. |
| B3 | **Hot-driver contention** | Dense areas at peak | Many concurrent rides rank the same few nearest drivers first; all but one lose the compare-and-set and retry with the next candidate, so work per assignment rises with concurrency. |
| B4 | **Single Postgres primary** | ~100k–1M | All rides, offers, outbox, history and idempotency writes funnel into one primary. |
| B5 | **Outbox relay throughput** | ~100k drivers | One relay publishes at a time per database: an advisory lock keeps per-key order across instances, so adding instances adds standbys, not throughput. Each batch of 100 is sent before one wait for its acknowledgements, which lifts the ceiling from a few hundred to thousands of events per second (not measured). Location updates alone reach 25k events/s at 100k drivers (load model), so that is where the relay gives out. The poll interval also adds latency and polling scans add load. `ridehailing.outbox.pending` and `ridehailing.outbox.oldest.age` show it falling behind. |
| B6 | **Polling read load** | ~100k clients | Without push, clients poll; reads dominate. |
| B7 | **Single broker / partition count** | ~1k events/s | 6 partitions cap dispatch parallelism at 6 consumers (each instance runs 3 listener threads, so two instances already saturate them); one broker is a single point of failure. |
| B8 | **Rematch sweeps during a shortage** | ~100k drivers, peak | `MatchingSweeper` re-runs the full matching query (B2) for every request that has waited 10 s without a driver. Unmatched requests pile up exactly when drivers are scarce, at peak, so sweeper load grows with demand while it finds nothing. Nothing caps the attempts or ends a ride that waits for ever (README § Known limitations). |

## Scaling strategies

| Strategy | Addresses | 1k | 100k | 1M |
|---|---|---|---|---|
| **S1. Live locations out of Postgres.** Keep the current location of `AVAILABLE` drivers in Redis GEO (`GEOADD`/`GEOSEARCH BYRADIUS ASC`), or a cell index such as H3 at resolution 8–9. Postgres keeps driver *status* and a periodically checkpointed location. Already designed as the `candidate-source=redis` flag ([ADR-008](adr/ADR-008-caching.md)). | B1, B2 | not needed | yes | yes, sharded by city/region |
| **S2. Geo-partitioned dispatch.** Key `ride-events` by geo-cell (H3 cell of the pickup, or city) instead of `rideId`, and run dispatch workers per partition group. Rides that compete for the same drivers land on the same worker, which can serialise or batch them locally instead of racing in the database. Two limits: a driver near a cell edge is a candidate for rides in neighbouring cells owned by other workers, so the Postgres claim (ADR-006) still decides there and those races are not removed; and keying by cell or city puts a dense centre on one partition, so cells must be sized by demand, not area. | B3, B7, B8 | no | per city | per H3 cell group |
| **S3. Batch matching windows.** Collect requests in a cell for 1–2 s and solve the assignment for the whole batch (Hungarian / min-cost flow on ETA) instead of greedy nearest-first. Fewer conflicting claims and better global pick-up times. | B3 | no | optional | yes |
| **S4. Shard the transactional store by city/region.** Rides and drivers rarely cross cities, so each city (or region) gets its own Postgres cluster, or a distributed SQL store (Citus, Spanner, CockroachDB) partitioned on `city_id`. Read replicas serve history and reporting. | B4 | no | read replicas | sharded |
| **S5. Batched relay, then change data capture.** Built: the relay sends a whole batch and waits once for its acknowledgements (the idempotent producer keeps per-partition order), logs one line per batch, and holds a Postgres advisory lock so a second instance stays idle instead of publishing out of order ([ADR-005](adr/ADR-005-event-driven-dispatch.md) decision 2). Next: replace `OutboxRelay` with Debezium reading the WAL for the `outbox` table: lower latency, no polling load, same at-least-once semantics, and per-key order without a single publisher. | B5 | batched relay (built) | CDC | CDC |
| **S6. Read caching, then push.** The versioned Redis read cache (already implemented, ADR-008) absorbs polling; beyond that, push ride/offer updates over WebSocket or mobile push driven by the same events. | B6 | cache | cache + push | push |
| **S7. Broker capacity.** Raise partitions (e.g. 6 → 96 → 512 across geo-cells), replicate 3× across availability zones, and scale the dispatch consumer group to match. | B7 | single broker | 3 brokers | cluster per region |
| **S8. Split modules into services.** `driver` (location ingestion and availability, write-heavy) and `dispatch` (CPU-heavy matching) scale independently of `ride` (API and lifecycle). The module boundaries and event catalogue already match this split ([architecture.md](architecture.md#module-rules-verified-by-applicationmodulesverify)). | B1–B4 | no | driver service | all three |
| **S9. Shortage-aware rematching.** Back off a request's sweeps after each empty attempt, cap the attempts, and give a ride that waits too long a terminal outcome. Better: trigger a rematch when supply changes in the request's area (a driver comes online or nearby), not on a timer. | B8 | no | yes | yes |

### Evolution path

1. **1,000 drivers (today):** modular monolith, one Postgres, one Kafka broker, Redis read cache, and the batched relay of S5. Nothing else on this page is needed yet; the priority is correctness and observability, and the delivery metrics (README § Observability) show when the next tier is approaching.
2. **100,000 drivers:** S1 (Redis GEO candidates), S5 (CDC outbox), S6 (cache in front of polling), S7 (3 brokers, more partitions), S9 (shortage-aware rematching), Postgres read replicas, and the driver module split out as a location-ingestion service (S8).
3. **1,000,000 drivers:** S2 geo-partitioned dispatch with S3 batch matching, S4 sharding by city/region, push instead of polling, and per-region deployments.

## Architectural trade-offs

| # | Trade-off | Chosen now | What we give up | When to revisit |
|---|---|---|---|---|
| T1 | **Greedy nearest driver vs. batch-optimal assignment** | Greedy (the brief asks for it) | Globally better pick-up times and fewer conflicting claims | Dense markets where many rides compete for the same drivers (S3) |
| T2 | **Location freshness vs. write cost** | Every update written to Postgres | Write throughput; at scale, drop to "latest location in Redis, checkpoint to Postgres" and accept that matching may use a location a few seconds old | ~100k drivers (S1) |
| T3 | **Modular monolith vs. microservices** | One deployable, verified module boundaries | Independent scaling and deployment. The modules already talk only through events and each decides in its own tables ([ADR-002](adr/ADR-002-decoupled-contexts.md)), so splitting them changes deployment, not the protocol | When one module's load profile diverges (S8) |
| T4 | **Postgres compare-and-set vs. distributed locks** | Single-row atomic updates plus partial unique indexes | Cross-shard assignment; correctness needs ride and driver rows in the same database | Never across cities; within a city shard the same approach holds |
| T5 | **At-least-once + idempotent consumers vs. exactly-once** | At-least-once with `eventId` de-duplication | A little storage (`processed_events`) and the discipline that every handler is idempotent | Not expected; Kafka exactly-once does not cover Postgres side effects |
| T6 | **Cache freshness vs. read load** | Versioned after-commit puts with a 30 s TTL; commands never read the cache | Polling clients may briefly see the previous state if an after-commit put fails | If clients need read-your-writes on every read, serve them from the primary |
| T7 | **Location events through the outbox vs. best-effort** | Every location update is an outbox row, delivered at least once | Relay and WAL capacity: location is ~95 % of all events (load model). A location is overwritten 4 s later, so losing one costs little. Publishing it straight to Kafka after commit, a deliberate dual write for this one event, would take it off the outbox | When the relay or CDC cannot keep up with location volume, or with S1, where location bypasses Postgres anyway |
