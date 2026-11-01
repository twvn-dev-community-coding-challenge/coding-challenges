# Automated testing strategy

How the Ride-Hailing Dispatch Service is tested, from a single function up to the whole application against real infrastructure. Every test runs unattended, locally with one command and in CI on every push.

Sections 1–9 describe what exists in `src/test` and `src/integrationTest` today. Section 10 lists what is planned but not built.

## 1. Principles

1. **Outside-in.** Each acceptance criterion has its own *acceptance test* class exercising the public HTTP API; unit tests cover the rules underneath it.
2. **Real infrastructure where behaviour depends on it.** Concurrency, constraints, the outbox, Kafka redelivery and Redis behaviour are tested against real Postgres, Kafka and Redis (Testcontainers). Mocks only replace things the code under test does not own.
3. **Deterministic by construction.** No `Thread.sleep`: time comes from an injected `Clock` (`MutableClock` in tests, which a test can advance past an offer timeout), and asynchronous outcomes are awaited with Awaitility on an observable condition.
4. **Tests own their data.** Every test creates its own passengers, drivers and rides with unique IDs, in its own patch of the map (§ 6). No test depends on another test, on seed data, or on run order.
5. **Invariants over examples.** Besides example-based tests, `SystemInvariants` asserts global safety properties after every acceptance test (§ 5).
6. **Every acceptance criterion is traceable** through JUnit tags (`@Tag("AC-09")`).

## 2. Test levels

| Level | Gradle suite | What | Tools | Docker |
|---|---|---|---|---|
| Unit | `test` | shared kernel (`Location`, `GeoDistance` (Haversine), `BoundingBox`); every lifecycle (`RideTransition`, `DriverTransition`, `DispatchRequest` and `DispatchRequestTransition`, `OfferTransition`, `AssignmentTransition`); `NearestDriverMatcher`; factories (`RideTest`, `RideOfferTest`, `EventEnvelopeTest`); `RequestFingerprint`; `EventType`; `EventRetentionProperties`; `LayerColorConverter`; the web error mapping (`ApiExceptionHandlerTest`); a use case on fakes (`AssignmentProjectionTest`) | JUnit Jupiter, AssertJ | no |
| Architecture | `test` | module boundaries, the Clean Architecture dependency rule and domain purity, in one Spring Modulith `verify()` (§ 4) | Spring Modulith `ApplicationModules.verify()`, ArchUnit | no |
| Slice | `test` | `RideController`: validation, status codes, `ProblemDetail` bodies, `Location` and `Idempotent-Replayed` headers | `@WebMvcTest`, MockMvc, Mockito | no |
| Integration | `integrationTest` | partial unique indexes, Redis cache behaviour and fail-open, log contents, the outbox relay's lock and ordering, metrics | Testcontainers (Postgres, Kafka, Redis), `@ServiceConnection`, Awaitility | yes |
| Acceptance | `integrationTest` | one class per AC, through the HTTP API of the whole app; includes the AC-09 concurrency races | `@SpringBootTest(webEnvironment = RANDOM_PORT)` + Testcontainers | yes |

`NearestDriverMatcherTest.rankingIsSortedEligibleAndStartsWithTheTrueNearest` and `GeoDistanceTest.isSymmetricAndSatisfiesTheTriangleInequality` are seeded-random property checks written in plain JUnit.

## 3. Acceptance tests per criterion

| Class | AC | Key scenarios |
|---|---|---|
| `AC01CreateRideAcceptanceIT` | AC-01 | create → `201` + `Location`, starts `REQUESTED`, retrievable; unique IDs; invalid lat/lng, missing or unknown passenger rejected; missing key → `400`; same key + same body → same ride; same key + different body → `422`; replay correct when the cache entry is lost |
| `AC02DriverAvailabilityAcceptanceIT` | AC-02 | `OFFLINE ⇄ AVAILABLE`; repeats are harmless; only `AVAILABLE` receives offers; `OFFERED`/`BUSY` cannot go `OFFLINE`; completion returns `BUSY` to `AVAILABLE`; the API cannot set `OFFERED`/`BUSY`; unknown statuses and drivers are rejected |
| `AC03DriverLocationAcceptanceIT` | AC-03 | invalid coordinates rejected; latest location stored and used by matching |
| `AC04MatchingAcceptanceIT` | AC-04 | nearest within the radius; all outside the radius → stays `MATCHING`; tie → smallest `driverId`; busy drivers skipped |
| `AC05RideOfferAcceptanceIT` | AC-05 | offer sets `DRIVER_ASSIGNED` + `OFFERED`; only the holder can respond; accept → `DRIVER_ACCEPTED`/`BUSY`; reject → `MATCHING`/`AVAILABLE`; timeout (advance `MutableClock` past 30 s); offline while offered expires the offer; cancel before and after acceptance, and completion, release the driver |
| `AC06StateMachineAcceptanceIT` | AC-06 | completing an unaccepted ride → `409`; terminal states reject further transitions; duplicates answered without change; history persisted in order |
| `AC07RematchingAcceptanceIT` | AC-07 | reject → next driver offered → accepts; a driver who rejected is never re-offered that ride but is still offered others |
| `AC08IdempotencyAcceptanceIT` | AC-08 | concurrent duplicate creates → one ride; duplicate accept and reject harmless; a re-published Kafka record processed once |
| `AC09ConcurrentAssignmentAcceptanceIT` | AC-09 | two rides, one driver, released together → exactly one offer (×20); the loser falls through to the next driver (×10); accept and reject racing on one offer → one winner (×10) |
| `AC10EventDrivenDispatchAcceptanceIT` | AC-10 | create → RideRequested on `ride-events` → RideOffered on `dispatch-events`, keyed by `rideId`; poison record → DLT without blocking dispatch; `MatchingSweeper` recovers a waiting request; `UndispatchedRideSweeper` republishes a RideRequested that never reached dispatch; a ride cancelled before dispatch opened it is never offered |

## 4. Architecture tests

One check, `ModularityTest.modulesRespectTheirBoundariesAndLayering`, runs Spring Modulith's `ApplicationModules.verify()` with the ArchUnit rules of `ArchitectureRules` added to the same pass ([ADR-001](adr/ADR-001-module-structure.md) decision 7).

- Spring Modulith, the module graph of [ADR-002](adr/ADR-002-decoupled-contexts.md):
  - no cycles, no access to another module's sub-packages;
  - only the `allowedDependencies` each module declares in its `package-info.java`: `dispatch → common, contracts`, `ride → common, contracts`, `driver → common, contracts`.
- `ArchitectureRules` (ArchUnit), the layering of [ADR-003](adr/ADR-003-clean-architecture.md):
  - the dependency rule: `frameworks` → `adapters` → `application` → `domain`; nothing depends on `frameworks`, and only `frameworks` on an adapter;
  - the web, messaging, persistence and cache adapters never use each other;
  - every top-level type in `application.port` is an interface;
  - the domain uses only `java..`, the domain packages and `common.kernel`, which rules out Spring, SQL, Kafka, Redis, HTTP, Jackson and Lombok in one rule; `common.kernel` and `contracts` use only `java..` and themselves;
  - the application and the `common.error` exceptions use no Spring HTTP or web classes;
  - the application uses `common.events` only through interfaces, and never Jackson;
  - the application and `adapters.web` never use JDBC or Kafka classes;
  - the web models (`adapters.web.model`) have no Lombok-generated members;
- `ArchitectureRules`, technology boundaries:
  - only `OutboxRelay` (and `MessagingConfig`) use `KafkaTemplate`;
  - only `common.cache` uses Spring Data Redis.

The Gradle project boundaries add a compile-time check on top of these ([ADR-001](adr/ADR-001-module-structure.md)).

Two more Spring Modulith uses:

- `ModularityTest.writesModuleDocumentation` (`Documenter`) writes C4 component diagrams and a canvas per module to `app/build/spring-modulith-docs`.
- `DispatchModuleIT` (`@ApplicationModuleTest(mode = STANDALONE)`, integration suite) boots dispatch with only `common` and `contracts`, asserts that no ride or driver bean exists, and offers a ride from a `RideRequested` event alone: dispatch's extractability, tested.

## 5. Invariants and concurrency

**`SystemInvariants`** queries Postgres and asserts:

| # | Invariant |
|---|---|
| I1 | No driver holds more than one `dispatch_requests` row in `OFFERED`/`ACCEPTED`. |
| I2 | At most one `PENDING` offer per ride and per driver. |
| I3 | A driver's `driver_supply.assignment` agrees with the request it holds: `OFFERED` ⇔ request `OFFERED`; `BUSY` ⇔ request `ACCEPTED`. |
| I4 | A pending offer belongs to the request's offered driver. |
| I5 | No excluded driver ever received a later offer for the same ride. |
| I6 | Every `ride_state_history` row is a legal edge of the state machine. |
| I7 | *(eventual)* Every ride's status agrees with its dispatch request (ride is a projection, [ADR-002](adr/ADR-002-decoupled-contexts.md)). |
| I8 | *(eventual)* Every driver's assignment and availability agree with dispatch's supply copy. |

I1–I6 hold in every committed state (`assertHold`). I7–I8 are polled until they converge (`assertProjectionsConverged`). Acceptance classes call both through `assertInvariants()` in an `@AfterEach`.

**Use cases on fakes.** `AssignmentProjectionTest` runs a use case with hand-written fakes of its ports and of `ProcessedEvents` and `EventCodec`: no database, Kafka or Spring. `ApiExceptionHandlerTest` checks the HTTP status, title, code and details the web adapter gives each exception.

**Concurrency.** `ConcurrentCalls` starts every call at the same instant on its own virtual thread. It drives the AC-09 races and the concurrent duplicate create of AC-08. `PersistenceConstraintsIT` bypasses the services and writes conflicting rows directly, proving the partial unique indexes and check constraints reject a double assignment, or an offered driver with no ride, even if the code regresses. `ApplicationIT` boots the whole application against real infrastructure and checks `/actuator/health` and `/actuator/modulith`.

**Resilience.** `CacheIT.everyFlowStillWorksWhileRedisIsUnavailable` pauses the Redis container and runs create, dispatch, accept and read.

**Relay.** `OutboxRelayIT` takes the relay's advisory lock on a second connection, as another instance would, and checks that nothing is published while it is held, that the backlog gauges report the waiting rows, and that the rows are published in order once it is released. It also writes a backlog of more than two batches for one key and reads it back from Kafka in recorded order. Its rows go to a topic of their own, so no module consumes them.

**Metrics.** `MetricsIT` drives a first offer, a rejection and rematch, an acceptance, a timeout, a cancellation, a ride with no driver nearby, a dead-lettered record and a republished ride, and checks that each meter moves. Meters are shared by the whole context, so it compares counts before and after its own actions, and asserts "at least" where the sweepers can also move a count.

**Retention.** `EventRetentionIT` writes outbox rows and de-duplication markers of different ages, runs `EventRetention.prune()` and checks what survives, including a backlog larger than one batch. Like `PersistenceConstraintsIT`, it works in a rolled-back transaction, so the relay never publishes its rows. `KafkaTopicsIT` checks that every topic and DLT has its 7-day `retention.ms` set on the topic, not inherited from the broker. `EventRetentionPropertiesTest` checks that startup fails when markers would be kept no longer than the topics.

## 6. Test isolation and fixtures

- **One application context and one set of containers** for every integration and acceptance test (`@IntegrationTest`: `@SpringBootTest` + `TestcontainersConfiguration` with `@ServiceConnection` for Postgres 18, Kafka 4.1 and Redis 8.2; tags match `compose.yaml`). Containers start in parallel.
- **`TestCity`** gives each test its own patch of the map, at least 50 km from every other test's patch and from the seed data, so matching never sees another test's drivers and no clean-up is needed.
- **`Fixtures`** inserts passengers and idle drivers (in the driver module and dispatch's supply copy), and the states some tests start from: a `WAITING` dispatch request, or a ride whose RideRequested never reached dispatch. Offers and trips are always created through the API.
- **`MutableClock`** replaces the system clock; tests advance it past an offer timeout instead of waiting. Background jobs run on faster intervals in `src/integrationTest/resources/application.properties`, with staleness thresholds long enough that they never race a test's own flow.
- **Kafka**: tests publish or re-publish records with `KafkaTemplate` and read topics with a throwaway consumer.
- **Log capture** (`OutputCaptureExtension`) in `ObservabilityIT`: every lifecycle log line of a ride carries `rideId`, plus `driverId` and `eventId` where they apply.

## 7. Commands

```bash
./gradlew test               # unit, architecture, slice — no Docker, seconds
./gradlew integrationTest    # Testcontainers: integration and acceptance
./gradlew build              # compile + unit/architecture/slice tests + jars; integration tests are NOT part of build
./gradlew integrationTest --tests 'com.ridehailing.acceptance.AC09*'   # one acceptance criterion
```

`make build`, `make test`, `make it`, `make acceptance`, `make ac AC=09` and `make verify` wrap the same tasks; `make build` never needs Docker. Running Gradle needs JDK 25 and, for `integrationTest`, Docker.

## 8. CI pipeline (GitHub Actions)

One job, `.github/workflows/ci.yml`, on every push and pull request:

1. `./gradlew build`: compile, unit, architecture and slice tests.
2. `./gradlew integrationTest`: integration and acceptance tests against Testcontainers.
3. Upload the JUnit HTML reports.
4. `docker build`: the image builds.

JaCoCo is applied to every project and produces coverage reports; no coverage threshold is enforced.

## 9. Mapping to the brief's minimum coverage

| Brief | Unit | Integration / acceptance |
|---|---|---|
| Matching: nearest, outside radius, offline, busy, none | `NearestDriverMatcherTest` | `AC04MatchingAcceptanceIT` |
| Driver status: offer, reject/cancel, accept/complete, invalid | `DriverTransitionTest` (availability, status), `DispatchTransitionsTest` (assignment) | `AC02DriverAvailabilityAcceptanceIT`, `AC05RideOfferAcceptanceIT` |
| State machine: valid, invalid, duplicate, reject clears driver | `RideTransitionTest`, `DispatchRequestTest`, `DispatchTransitionsTest` | `AC06StateMachineAcceptanceIT` |
| Rematching: reject, second driver, exclusion | `NearestDriverMatcherTest.skipsDriversExcludedForThisRide` | `AC07RematchingAcceptanceIT` |
| Concurrency: two rides, same driver | — | `AC09ConcurrentAssignmentAcceptanceIT`, `PersistenceConstraintsIT` (+ invariants) |
| Idempotency: duplicate creation, duplicate acceptance | `RequestFingerprintTest` | `AC01CreateRideAcceptanceIT`, `AC08IdempotencyAcceptanceIT` |

## 10. Planned, not built

| Item | Purpose |
|---|---|
| Property-based tests (jqwik) | jqwik generators in place of today's seeded-random checks (Haversine symmetry, sorted and eligible rankings), plus generated command sequences for the state machines. Waiting on verified jqwik support for JUnit 6. |
| Dedicated race tests: accept vs. offer expiry, accept vs. cancel, reject vs. cancel | Today these are covered one after the other, not concurrently ([ADR-006](adr/ADR-006-concurrency.md) § Verification). |
| Quality gates: JaCoCo thresholds, PIT mutation score on the domain | Fail the build on coverage or mutation regressions. |
| `e2eTest` suite | Black-box tests against the built Docker image started from `compose.yaml`, mirroring `scripts/demo.sh`. |
| Chaos tests (Toxiproxy) | Kafka unreachable, Redis latency, app restart mid-dispatch. |
| Soak test | Randomised commands on many drivers and rides for a fixed time, ending with all invariants. |
| OpenAPI drift test | Fail when the generated OpenAPI document differs from a committed one. |
