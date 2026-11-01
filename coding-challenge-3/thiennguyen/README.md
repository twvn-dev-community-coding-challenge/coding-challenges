
# [thiennguyen] - [Coding Challenge 3] Ride-Hailing Dispatch Service

## Team Members

- Thien Nguyen (Developer)

## How to run

### Setup (once)

**Prerequisites:**
- **Docker**, for running the stack and the integration tests.
- **JDK 25**, only to run Gradle on your machine (`brew install openjdk@25`; `make` finds it, or pass `make JAVA_HOME=/path/to/jdk-25 …`). Running the app with `make up` needs Docker alone: the image is built inside Docker.

```bash
make setup       # one call: checks JDK 25, downloads Gradle and every dependency, pulls the Postgres/Kafka/Redis images
make help        # lists every target
```

The Gradle wrapper (`gradlew`, `gradle/wrapper/`) is committed, so a fresh clone needs no Gradle install. If it is ever missing, any Gradle target recreates it through Docker (`make wrapper` does it explicitly).

### Run the project

```bash
make up          # builds the image inside Docker, starts app + Postgres + Kafka + Redis, waits until healthy
make demo        # plays one ride end to end (see below)
make trace RIDE=<rideId>   # the story of one ride, from the logs
make logs        # follow the app's logs
make down        # stop and remove containers and volumes (this also deletes the logs trace reads)
make clean       # delete every Gradle build output (all build/ and .kotlin/ dirs); keeps .gradle/ and the wrapper
```

The API listens on `http://localhost:8080`. Actuator exposes `health` (with liveness and readiness probes), `info`, `metrics` and `modulith`; Redis is left out of health on purpose ([ADR-008](docs/adr/ADR-008-caching.md)). Every tunable (`ride.*`, `dispatch.*`, `outbox.*`, `events.retention.*`, `cache.*`) is in [application.yml](app/src/main/resources/application.yml) and [cache.yml](app/src/main/resources/cache.yml). Seed data (`app/src/main/resources/db/seed/`): passengers `P001–P005`; drivers `D001–D010` around the brief's pickup point in Ho Chi Minh City (`D001` 0.8 km, `D002` 1.5 km, `D003` 3.2 km, `D004` offline, …).

For development, `./gradlew :app:bootRun` starts Postgres, Kafka and Redis from `compose.yaml` automatically (Spring Boot Docker Compose support) and runs the app on the host. `make demo` works against it too; `make trace` does not, because it reads the app container's logs.

### Demo and trace

**`make demo`** ([scripts/demo.sh](scripts/demo.sh)) plays one ride against the running stack with seed data, and checks first that the API is healthy:

1. Passenger `P001` requests a ride at the brief's pickup point; the ride starts `REQUESTED`.
2. Dispatch offers it to the nearest available driver, **D001** (0.8 km); the ride shows `DRIVER_ASSIGNED`.
3. D001 rejects. The ride returns to `MATCHING` and D001 is excluded from this ride.
4. Dispatch rematches to the next nearest driver, **D005** (1.2 km). D003 (3.2 km) and D007 (2.5 km) are outside the 2 km radius, and D004 is offline.
5. D005 accepts. The accept call returns the offer; the ride shows `DRIVER_ACCEPTED` a moment later, once ride has applied dispatch's event (ADR-002).
6. The trip completes, and the script prints the ride's history, offers and trace commands.

It can be re-run: a rejection excludes D001 from that one ride only, and completing frees D005.

**`make trace RIDE=<rideId>`** ([scripts/trace-ride.sh](scripts/trace-ride.sh)) answers the brief's "What happened to ride R001?" from the logs alone. It prints the app container's log lines that carry `rideId=<rideId>`, across every module, the Kafka listeners and the outbox: the request, each outbox event and its publication, the dispatch attempt, each offer, the rejection, the acceptance and every ride status change.

Trace needs the stack started with `make up`: it reads `docker compose logs app`, so it cannot see `./gradlew :app:bootRun` output, and `make down` deletes those logs. For the persisted state changes instead of the logs, use `GET /rides/{rideId}/history`.

### API

| Method & path | Purpose |
|---|---|
| `POST /rides` + `Idempotency-Key` header | Create a ride (`201`, `Location`, `Idempotent-Replayed`) |
| `GET /rides/{rideId}` · `/history` · `/offers` | Current state (cached), state history, offers (`/offers` is served by dispatch and returns `404` until dispatch has opened the ride) |
| `POST /rides/{rideId}/offers/{driverId}/accept` · `/reject` | Driver responds to the pending offer (served by dispatch; returns the **offer** — the ride shows `DRIVER_ACCEPTED` or `MATCHING` a moment later) |
| `POST /rides/{rideId}/cancel` · `/complete` | Passenger cancels / trip completes (served by ride) |
| `GET /drivers/{driverId}` | Driver state |
| `PUT /drivers/{driverId}/availability` `{"status":"AVAILABLE"\|"OFFLINE"}` | Go online / offline |
| `PUT /drivers/{driverId}/location` `{"latitude":…, "longitude":…}` | Report location |

Errors are RFC 9457 problem bodies with a stable `code`:

| Status | Codes |
|---|---|
| 400 | `VALIDATION_FAILED`, `INVALID_LATITUDE`, `INVALID_LONGITUDE`, `UNKNOWN_PASSENGER`, `IDEMPOTENCY_KEY_REQUIRED`, `IDEMPOTENCY_KEY_TOO_LONG`, `UNSUPPORTED_AVAILABILITY` |
| 404 | `RIDE_NOT_FOUND`, `DRIVER_NOT_FOUND` |
| 409 | `INVALID_TRANSITION`, `INVALID_DRIVER_TRANSITION`, `OFFER_NOT_HELD`, `OFFER_EXPIRED`, `CONCURRENT_UPDATE` |
| 422 | `IDEMPOTENCY_KEY_REUSED` |

```bash
curl -s -X POST localhost:8080/rides -H 'Content-Type: application/json' -H 'Idempotency-Key: 7f4d9c2e-ride-request-001' \
  -d '{"passengerId":"P001","pickup":{"latitude":10.7769,"longitude":106.7009},"destination":{"latitude":10.8231,"longitude":106.6297}}'
```

### How to run the tests

```bash
make build            # compile, unit + architecture tests, boot jar — no Docker, seconds
make test             # unit + architecture tests only
make it               # all integration tests (Testcontainers: acceptance, cache, observability, constraints, module boot)
make acceptance       # acceptance tests only (AC-01 … AC-10)
make ac AC=09         # one acceptance criterion
make verify           # what CI runs: make build, then make it
./gradlew :ride:test  # one module's unit tests
```

Every acceptance criterion has its own test class (`AC01CreateRideAcceptanceIT` … `AC10EventDrivenDispatchAcceptanceIT`, tagged `AC-01` … `AC-10`). Integration tests share one Spring context and one set of containers; each test works in its own patch of the map (`TestCity`) and checks global safety invariants (`SystemInvariants`) afterwards. See [docs/testing-strategy.md](docs/testing-strategy.md).

## Approach

A small dispatch service that receives ride requests, offers each ride to the nearest available driver, handles accept / reject / timeout / cancel / complete, and stays correct under duplicate requests, duplicate events and concurrent dispatch. It is built as a **modular monolith** whose module boundaries are verified in tests, so `ride`, `driver` and `dispatch` could each become a service.

**Stack:** Java 25 · Spring Boot 4 · Spring Modulith 2 · PostgreSQL 18 · Kafka 4 (KRaft) · Redis 8 · Lombok (bean constructors and loggers only) · Gradle 9 (Kotlin DSL, version catalog) · Testcontainers 2 · Docker Compose.

| Document | Contents |
|---|---|
| [docs/architecture.md](docs/architecture.md) | C4 context / container / component views, module rules, state machines, data model, event catalogue |
| [docs/sequence-diagram.md](docs/sequence-diagram.md) | Create, dispatch, reject → rematch → accept → complete, timeout and driver offline, concurrent assignment, cancel vs accept, failure |
| [docs/scalability.md](docs/scalability.md) | 1k → 100k → 1M drivers: bottlenecks, strategies, trade-offs |
| [docs/testing-strategy.md](docs/testing-strategy.md) | Test levels, invariants, quality gates, CI |
| [docs/adr/](docs/adr/README.md) | In reading order, high-level first: [001 modular monolith](docs/adr/ADR-001-module-structure.md) · [002 decoupled contexts](docs/adr/ADR-002-decoupled-contexts.md) · [003 clean architecture](docs/adr/ADR-003-clean-architecture.md) · [004 matching](docs/adr/ADR-004-matching.md) · [005 event-driven dispatch](docs/adr/ADR-005-event-driven-dispatch.md) · [006 concurrency](docs/adr/ADR-006-concurrency.md) · [007 idempotency](docs/adr/ADR-007-idempotency.md) · [008 caching](docs/adr/ADR-008-caching.md) |

## Design Decisions

### Project layout

```
build.gradle.kts      settings shared by every project (Java 25 toolchain, Spring BOMs, test setup)
modules/contracts/    topic names, event wire names, envelope and versioned payload records (plain Java)
modules/common/       shared kernel (common.kernel: geo, StateTransition) + platform: API errors, transactional outbox,
                      de-duplication, Redis cache, logging (+ test fixtures)
modules/driver/       availability and location; copy of dispatch's assignment
modules/ride/         passengers (seeded), ride creation (idempotent), lifecycle, history, read API, undispatched-ride sweeper
modules/dispatch/     dispatch requests, matching, offers, the driver claim, timeouts, sweeper, Kafka consumer
app/                  Spring Boot application: assembles the modules, config, seed data, cross-module tests
gradle/libs.versions.toml   single version catalog for every project
```

Each module is its own Gradle project with its own dependencies, unit tests and database migrations (`src/main/resources/db/migration/<module>/`, timestamp versions, no foreign keys across modules), so the build fails as soon as one module reaches a module it does not depend on, and any module can be extracted into a service by giving it its own Spring Boot application and database. See [ADR-001](docs/adr/ADR-001-module-structure.md).

Inside `ride`, `driver` and `dispatch` the code has four layers, DDD's names for Clean Architecture's rings, with dependencies pointing inwards ([ADR-003](docs/adr/ADR-003-clean-architecture.md)):

```
domain/             aggregates, value objects, lifecycle transitions, matching (plain Java)
application/        use cases and their transactions, event handlers, schedulers
  port/             interfaces the use cases need: repositories, caches, passengers
adapters/
  web/              REST controllers
    model/          their request and response models
  messaging/        the module's Kafka listener
  persistence/      JDBC implementations of the ports
  cache/            Redis implementations of the ports
frameworks/         Spring wiring: owned topics, properties, beans
```

Each lifecycle (ride, dispatch request, offer, driver assignment) is an enum of transitions; the SQL compare-and-set guards are built from it, so a rule lives in one place.

### Architecture

```
HTTP ──► ride ────── ride-events ──────► dispatch ◄───── driver-events ────── driver ◄── HTTP
          ▲                               │ ▲  ▲                                ▲
          │                               │ │  └── HTTP: accept / reject        │
          │                               ├─┘ own dispatch-events (rematch)     │
          └──────── dispatch-events ──────┴────────── dispatch-events ──────────┘
                 (offer outcomes)                      (driver assignment)

Each module: its own tables in PostgreSQL (plus one shared outbox) · Kafka: one topic per owner · Redis: read cache only
```

Modules:
- `contracts`: the published language (topics, events, versioned payloads).
- `driver`: availability and location, plus a copy of dispatch's assignment.
- `ride`: passengers (seeded), creation, cancel, complete, history, plus a projection of dispatch's offer state.
- `dispatch`: dispatch requests, offers, the driver claim, matching, timeouts, sweeper.
- `common`: outbox, de-duplication, cache, geo, errors, logging.

**No business module calls another**: each decides in its own transaction and announces it on its own topic; the others react ([ADR-002](docs/adr/ADR-002-decoupled-contexts.md)). `dispatch` depends only on `common` and `contracts`, so it can be extracted as a service. Each module declares its allowed dependencies to Spring Modulith; `ModularityTest` fails the build on a violation, together with the ArchUnit rules (run in the same `verify()`) that layers depend inwards, domains are plain Java, only infrastructure uses SQL, Kafka or Redis, only the outbox relay publishes to Kafka and only the cache talks to Redis. `DispatchModuleIT` boots dispatch alone (`@ApplicationModuleTest`), `./gradlew test` writes Modulith's generated module diagrams to `app/build/spring-modulith-docs`, and `GET /actuator/modulith` shows the module structure at runtime. Full C4 views: [docs/architecture.md](docs/architecture.md).

### Matching strategy

Nearest free driver (online in dispatch's supply copy, holding no offer or trip) with a known location within `dispatch.max-radius-km` (default **2 km**, inclusive), by Haversine distance, ties broken by the lexicographically smallest `driverId`, skipping drivers excluded for this ride. A bounding box around the pickup is the indexed SQL pre-filter; `NearestDriverMatcher` (pure Java, unit- and property-tested) ranks the candidates. Dispatch walks the ranked list and claims the first driver that is still available. Details and alternatives (PostGIS, Redis GEO, H3, batch assignment): [ADR-004](docs/adr/ADR-004-matching.md).

### Concurrency strategy

Every state change is a **compare-and-set** in SQL. The driver claim runs entirely in dispatch's own `driver_supply` table: `UPDATE driver_supply SET assignment='OFFERED' … WHERE driver_id=? AND online AND assignment='NONE'` succeeds for exactly one of two racing dispatches, and the loser moves on to the next candidate or leaves its ride waiting. Partial unique indexes in dispatch's tables make a double assignment impossible even if code regresses: one active request per driver, one pending offer per driver and per ride, one driver per ride. Every dispatch transaction locks the request row before the supply row, which rules out deadlocks. Kafka is keyed by `rideId` (`driverId` for driver events), so events about one ride or one driver are processed in order. The copies held by ride and driver follow from events and never decide anything ([ADR-002](docs/adr/ADR-002-decoupled-contexts.md)). Proven by `AC09ConcurrentAssignmentAcceptanceIT` (repeated barrier-synchronised races). [ADR-006](docs/adr/ADR-006-concurrency.md).

### Idempotency strategy

| Duplicate | Mechanism |
|---|---|
| `POST /rides` retried | `idempotency_keys` row (key + SHA-256 of the request) inserted **in the same transaction** as the ride and its event; same body → original ride replayed (`Idempotent-Replayed: true`), different body → `422`. Concurrent duplicates block on the primary key and then replay. A Redis replay cache answers repeats first but never decides. |
| Accept / reject / cancel / complete repeated | Already applied → the current offer (dispatch) or ride (ride) is returned with no side effects and no new event. |
| Copies applied out of order | The driver copies carry their owner's version and are applied only if newer (driver row version for availability/location, supply version for assignment). Ride's copy applies each dispatch-event as a status transition and relies on per-ride order from the `rideId` key, which the single outbox publisher preserves (see Event-driven approach). |
| Kafka event redelivered | Consumers insert `(consumer, eventId)` into `processed_events` in the same transaction as their work; a conflict means "already done". |

[ADR-007](docs/adr/ADR-007-idempotency.md).

### Event-driven approach

**Why asynchronous:** creating a ride is fast and never depends on driver search; bursts of requests are buffered by the broker; dispatch workers scale independently; a dispatch failure never fails ride creation.

**How:** every state change and the event announcing it commit together in the `outbox` table (no dual write). `OutboxRelay` publishes the rows in batches to their owner's topic (`ride-events`, `dispatch-events`, `driver-events`), keyed by `rideId` or `driverId`. It holds a Postgres advisory lock while it publishes, so with several app instances only one publishes at a time and each key's events keep their order (`OutboxRelayIT`). Dispatch opens a request and matches on RideRequested, rematches on its own OfferRejected/OfferExpired, and releases the driver on RideCancelled/RideCompleted. Ride and driver keep their copies current from dispatch-events. Every event, its key, producer and consumers: [docs/architecture.md § Event catalogue](docs/architecture.md#event-catalogue).

**If a message is delivered twice:** the second delivery finds its `eventId` in `processed_events` and is skipped; even without that, the conditional updates would refuse to offer a ride dispatch is no longer waiting to match (`AC08IdempotencyAcceptanceIT` re-publishes a real event to prove it).

**If processing fails:** the transaction rolls back and the offset is not committed; the container retries with exponential backoff (0.5 s, 1 s, 2 s), then publishes the record to `<topic>.DLT` with the exception in its headers. Malformed records go to the DLT immediately. Periodic jobs make sure a dead-lettered or lost message never strands a ride (`AC10EventDrivenDispatchAcceptanceIT`):
- `UndispatchedRideSweeper` (ride) publishes RideRequested again for a ride still `REQUESTED` 30 s after its last RideRequested, i.e. one dispatch never opened.
- `MatchingSweeper` (dispatch) re-runs matching for rides still waiting that nobody has tried recently.
- `OfferExpiryScheduler` (dispatch) expires offers nobody answers after `dispatch.offer-timeout` (default 30 s); they are treated as rejections.

`EventRetention` (common, hourly) deletes outbox rows a day after they are published, never unpublished ones, and `processed_events` markers after 14 days, twice the topics' retention (`EventRetentionIT`).

If a ride is cancelled before dispatch ever saw it, dispatch records it as ended, so a late RideRequested never produces an offer. [ADR-005](docs/adr/ADR-005-event-driven-dispatch.md).

### Caching

Redis is a **cache only**: ride and driver read models for polling clients and an idempotency replay cache. Values are written after the database commit, carry the row version and are only stored when newer (Lua compare), every key has a TTL, and every Redis error falls back to Postgres within 100 ms. Commands never read the cache. Settings (timeouts and TTLs) are in [`app/src/main/resources/cache.yml`](app/src/main/resources/cache.yml). `CacheIT` pauses the Redis container and runs the whole flow. [ADR-008](docs/adr/ADR-008-caching.md).

### Observability

Logs are plain text in Spring Boot's console layout (`CONSOLE_LOG_PATTERN` in [logback-spring.xml](app/src/main/resources/logback-spring.xml)), with the class that logged in place of the application name. Every line carries `[rideId=… eventId=… driverId=…]`, with `-` where a value does not apply, filled from the logging context by the HTTP filter, the Kafka listeners, the outbox and every service:

```
2026-10-30T15:09:12.591+07:00  INFO 78389 --- [RideCreationService    ] [omcat-handler-0] [rideId=1915df74-4b7e-4976-987e-41b5350a8ecc eventId=-                                    driverId=-     ] : Ride created in REQUESTED
2026-10-30T15:09:12.789+07:00  INFO 78389 --- [DispatchService        ] [        kafka-3] [rideId=1915df74-4b7e-4976-987e-41b5350a8ecc eventId=e409be13-a6eb-4583-a73a-41641e2be0b1 driverId=-     ] : Dispatch request opened
```

Columns have fixed widths, so the lines read as a table. Class names are padded to 23 characters and longer third-party names are cut at the end. Ride and event ids are padded to a UUID's 36 characters. `driverId` comes last because driver ids have no fixed length. Ids are never cut, so `make trace` still finds every line.

In a terminal the level is coloured by severity (ERROR red, WARN yellow, INFO and DEBUG green). The class is coloured by its layer through `%layerColor` ([LayerColorConverter](modules/common/src/main/java/com/ridehailing/common/logging/LayerColorConverter.java), registered in the same file):

| Layer | Colour |
|---|---|
| `adapters.web` | blue |
| `adapters.messaging` | cyan |
| `application`, `domain` | bright magenta |
| `adapters.persistence`, `adapters.cache` | magenta |
| `frameworks`, `common` | bright blue |
| other application classes | default |
| third-party (Spring, Kafka, Hikari…) | grey |

Colours follow `spring.output.ansi.enabled`. `./gradlew :app:bootRun` and the app in `compose.yaml` force them on, so `make logs` is coloured. `make trace` keeps the colours in a terminal and strips them when its output is piped or redirected. Turn colours off (`detect` or `never`) wherever logs are shipped to an aggregator.

`make trace RIDE=<id>` prints the story of one ride; `GET /rides/{id}/history` returns its persisted state changes. Coordinates and other personal data are never logged. Event payloads carry identifiers, except the pickup in RideRequested and the driver location in DriverLocationUpdated, which dispatch needs; topic retention and outbox pruning must bound how long they are kept (GDPR, ADR-002). `ObservabilityIT` asserts the trace is complete.

**Metrics** answer what the logs cannot show at a glance: whether delivery is keeping up and how dispatch is doing. They are exposed at `GET /actuator/metrics/<name>`, next to Spring Boot's JVM, HTTP, connection-pool and Kafka-client metrics. Dispatch's meters are recorded only after the transaction that decided them commits. `MetricsIT` and `OutboxRelayIT` check that each one moves.

| Meter | Type | Answers |
|---|---|---|
| `ridehailing.outbox.pending`, `ridehailing.outbox.oldest.age` | gauges | Is the outbox relay keeping up, or stopped? |
| `ridehailing.events.latency{consumer, event}` | timer | How long from an event being recorded to its consumer starting on it? |
| `ridehailing.events.dead.letters{topic}` | counter | Are records being dead-lettered? |
| `ridehailing.rides.republished` | counter | Are RideRequested events being lost before dispatch sees them? |
| `ridehailing.dispatch.attempts{outcome=offered\|no_driver}` | counter | How often does matching find no driver? |
| `ridehailing.dispatch.claims.lost` | counter | How contested are the nearest drivers? |
| `ridehailing.dispatch.time.to.offer{offer=first\|rematch}` | timer | How long from dispatch opening a request to an offer? |
| `ridehailing.offers.ended{result}` | counter | How do offers end: `accepted`, `rejected`, `timed_out`, `driver_offline`, `cancelled`? |
| `ridehailing.cache.requests{cache, result}` | counter | What is each cache's hit rate, and is Redis failing? |

## Challenges Faced

### Scalability considerations

From 1k to 1M drivers the first bottlenecks are location volume (every update is a Postgres write and an event through the single outbox publisher), the matching query in dense areas, contention on the nearest drivers, the single primary database, polling reads, the broker's partition count and rematch sweeps during driver shortages. The evolution path replaces the batched relay with CDC, moves live locations into Redis GEO / H3 cells, partitions Kafka and dispatch workers by geography, adds batch matching, shards the store by city, replaces polling with push, backs off rematching when no driver is near, and splits `driver` and `dispatch` into services. Full analysis with load numbers and trade-offs: [docs/scalability.md](docs/scalability.md).

### Known limitations

- Single-node Postgres, Kafka and Redis; no replication or multi-region.
- Greedy nearest-driver matching on straight-line distance; no ETA, roads or batch optimisation.
- No authentication or authorisation: any caller may cancel, complete, accept or reject (out of scope).
- Ride status and the driver API's OFFERED/BUSY are eventually consistent with dispatch (typically milliseconds). A driver who goes offline at the exact moment dispatch claims them keeps that offer until it times out (ADR-002).
- Pickup and driver locations travel in Kafka events and stay in the `outbox` table until a day after they are published (`EventRetention`, `events.retention.*`). Every topic and DLT keeps records for 7 days (`Topics.RETENTION_DAYS`), so a dead-lettered record must be looked at within a week. `idempotency_keys` is not pruned yet (GDPR).
- No push to drivers or passengers; clients poll `GET /rides/{id}`.
- No cap on rematch attempts and no expiry for a ride that waits in `MATCHING` indefinitely; the sweeper keeps retrying.
- Outbox delivery is at-least-once and polled (≈ 50–200 ms latency); a relay failure mid-batch re-sends that whole batch (de-duplicated by consumers). The `outbox` table is shared by all modules. One relay publishes at a time per database (advisory lock), so relaying does not scale out with app instances; CDC is the next step ([docs/scalability.md](docs/scalability.md) B5, S5).
- Metrics are exposed through Actuator only: no Prometheus registry, dashboards or alerts are configured, and no trace context travels through Kafka.
- A dead-lettered event for ride's or driver's copies (for example DriverAssignmentChanged) leaves that copy stale until the next event for the same ride or driver; dispatch's own state stays correct, and the record waits on the DLT for replay.
- A ride's cached read may lag by up to its 30 s TTL if an after-commit cache write fails.
- Idempotency keys are not yet purged after 24 h (the TTL is enforced in Redis only).
- No dedicated race tests yet for accept vs. timeout and accept vs. cancel; both serialise on dispatch's request-row lock (ADR-006).
- Seeded passengers and drivers only; registration is out of scope.

## AI Tools Used

- Claude (Opus 5.5)
- Gemini (3.1 Pro)