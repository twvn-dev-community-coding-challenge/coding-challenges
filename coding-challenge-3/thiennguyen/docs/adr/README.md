# Architecture decision records

Read in order: each ADR builds on the ones before it, from the shape of the system down to technical detail.

| # | Decision | Level | Covers |
|---|---|---|---|
| [ADR-001](ADR-001-module-structure.md) | **Modular monolith**: one Gradle project per module, assembled by `:app`; each module owns its schema | System shape | boundaries the brief asks to demonstrate |
| [ADR-002](ADR-002-decoupled-contexts.md) | **Decoupled bounded contexts**: ride, driver and dispatch each own their state and react to each other's events; dispatch owns the driver assignment | Context ownership | AC-02, AC-05, AC-09 |
| [ADR-003](ADR-003-clean-architecture.md) | **DDD and Clean Architecture inside each context**: domain / application (+ port) / adapters (web, messaging, persistence, cache) / frameworks, DDD aggregates, lifecycle transitions as the single source of rules | Code structure | AC-02, AC-05 – AC-07 |
| [ADR-004](ADR-004-matching.md) | **Matching**: nearest available driver within 2 km, Haversine, tie by smallest `driverId`, permanent exclusions | Domain policy | AC-03, AC-04, AC-07 |
| [ADR-005](ADR-005-event-driven-dispatch.md) | **Event-driven dispatch**: no dual writes (state and its event commit together, relayed to Kafka from the outbox), idempotent consumers, sweepers for lost messages | Integration | AC-04, AC-07, AC-10 |
| [ADR-006](ADR-006-concurrency.md) | **Concurrency**: compare-and-set on dispatch's tables, partial unique indexes, lock order | Consistency | AC-02, AC-05, AC-06, AC-09 |
| [ADR-007](ADR-007-idempotency.md) | **Idempotency**: durable `Idempotency-Key` records, `eventId` de-duplication, no-op repeated transitions | Consistency | AC-01, AC-08 |
| [ADR-008](ADR-008-caching.md) | **Caching**: Redis as a versioned, fail-open read cache, never the source of truth | Performance | AC-01, AC-08, scalability |
