# ADR-001 — Modular monolith: one Gradle project per module, assembled by :app

- **Status:** Accepted
- **Followed by:** [ADR-003](ADR-003-clean-architecture.md) — the packages inside each module (domain, application with port, adapters, frameworks)
- **Relates to:** [architecture.md § Module rules](../architecture.md#module-rules-verified-by-applicationmodulesverify), [scalability.md S8](../scalability.md#scaling-strategies)

## Context

The brief accepts a modular monolith but asks for boundaries that could evolve into separate services. This is the first decision every other ADR builds on: what the deployable is and how its modules are cut.

The first version kept all modules as packages of one Gradle project. Spring Modulith's `verify()` caught boundary violations, but only when tests ran, and nothing stopped one module's code from compiling against another module's tables or internals. To extract `ride`, `driver` or `dispatch` into a service later, each module needs its own build unit, its own dependencies and its own schema.

## Decision

1. **Gradle multi-project build.**

   ```
   build.gradle.kts    shared settings for every project: Java 25 toolchain, Spring Boot + Modulith BOMs, tests
   modules/contracts/  library, plain Java: topic names, event wire names, envelope, versioned payload records
   modules/common/     library: shared kernel (common.kernel: geo, StateTransition) and platform: errors and their
                       HTTP mapping, outbox + de-duplication + retention, Redis cache, logging (+ testFixtures)
   modules/driver/     library
   modules/ride/       library
   modules/dispatch/   library
   app/                Spring Boot application: assembles every module, config, seed data, cross-module tests
   ```

2. **Dependencies follow the module graph and are enforced at compile time.** `common → contracts`, `driver → common`, `ride → common`, `dispatch → common, contracts`. `common` exposes `contracts` as an `api` dependency, so `ride` and `driver` compile against `contracts` through `common`; that is the `→ contracts` edge in decision 7 and in [architecture.md](../architecture.md#module-rules-verified-by-applicationmodulesverify). Neither `ride` nor `driver` nor `dispatch` depends on another business module; they talk only through events ([ADR-002](ADR-002-decoupled-contexts.md)). A missing edge in Gradle means the code does not compile. The same graph is declared to Spring Modulith (decision 7), which adds the finer rules: no cycles and no access to another module's sub-packages.
3. **Each module owns its migrations** under `src/main/resources/db/migration/<module>/`. Versions are timestamps (`V20261026_0400__rides.sql`), so migrations from different modules interleave without collisions. `:app` lists every module's location in `spring.flyway.locations` and adds seed data under `db/seed`.
4. **No foreign keys across module boundaries.** `rides.assigned_driver_id`, `drivers.current_ride_id`, `dispatch_requests.ride_id/driver_id`, `driver_supply.driver_id/current_ride_id`, `ride_offers.ride_id/driver_id` and `ride_driver_exclusions.*` are validated by the owning module's code and protected by dispatch's compare-and-set updates and partial unique indexes ([ADR-006](ADR-006-concurrency.md)). Foreign keys inside a module stay (`idempotency_keys → rides`, `ride_state_history → rides`).
5. **Unit and slice tests live in their module.** Tests that need every module (module boundaries, architecture rules, acceptance tests, Testcontainers) live in `:app`. Shared test helpers are published as `testFixtures` of `:common`.
6. **Shared build settings live in the root `build.gradle.kts`** (one `subprojects {}` block). Each module's own build file only applies `java-library` and lists its dependencies. This is deliberately the simplest setup; if the number of deployable services grows, the block can move into convention plugins without changing any module.
7. **Spring Modulith is the model of the modules: verified, documented, booted one by one and visible at runtime.** Each top-level package under `com.ridehailing` is an application module.

   | Feature | How it is used |
   |---|---|
   | `@ApplicationModule(displayName, allowedDependencies)` in each `package-info.java` | `dispatch → common, contracts`; `ride → common, contracts`; `driver → common, contracts`; `common` (type `OPEN`) `→ contracts`. ride, driver and dispatch keep nothing in their base package, so they expose no API: other modules can reach them only through events. |
   | `ApplicationModules.verify()` | `ModularityTest` verifies cycles, sub-package access and the allowed dependencies, and runs the layering rules of `ArchitectureRules` (ArchUnit) in the same pass (`VerificationOptions.andAdditionalVerifications`), so there is one architecture check ([ADR-003](ADR-003-clean-architecture.md) § 6). |
   | `Documenter` | `ModularityTest.writesModuleDocumentation` writes a C4 component diagram of all modules, one diagram per module and a canvas per module (beans, dependencies) to `app/build/spring-modulith-docs` on every `./gradlew test`. The diagrams show only code dependencies: ride, driver and dispatch link only to `common` and `contracts`, because they talk through Kafka. |
   | `@ApplicationModuleTest(mode = STANDALONE)` | `DispatchModuleIT` boots dispatch with only `common` and `contracts`, checks that no ride or driver bean exists, and serves a ride from a `RideRequested` event alone: the extraction path below, tested. Its automatic verification is off because the integration-test classpath has test-only packages that Modulith would treat as modules; `ModularityTest` covers the production classes. |
   | `spring-modulith-actuator` | `GET /actuator/modulith` shows the module structure of the running application. |

## Extracting a module into a service

1. Add `services/<name>-service/` applying `java` and `org.springframework.boot` (like `:app`), depending on `:<name>` and `:common`, with its own `application.yml` listing only `db/migration/common` and `db/migration/<name>`.
2. Give it its own database. Its migrations already have no cross-module foreign keys.
3. Give it its own `outbox` and `processed_events` tables (today one shared copy lives in `common`'s migrations) and its own consumer group. No in-process calls need replacing: modules already communicate only through events on their owner's topic ([ADR-002](ADR-002-decoupled-contexts.md)).
4. Move the module's acceptance tests from `:app` into the new service, keeping cross-service flows as end-to-end tests.

## Consequences

- Module boundaries fail the build at compile time, before any test runs.
- The module graph is stated three times — Gradle dependencies, `allowedDependencies`, and the event-only integration of ADR-002 — but each is checked by exactly one tool: the compiler, `ModularityTest`, and `DispatchModuleIT`.
- The generated Modulith diagrams complement, not replace, the hand-drawn views in [architecture.md](../architecture.md): only those show the event flow.
- Losing cross-module foreign keys means a bug could store an offer for a ride or driver that does not exist; each module validates what it receives, and the safety-critical rules (one active ride per driver, one pending offer per driver) remain database constraints.
- Slice tests in a module need a small test-only `@SpringBootApplication` (for example `RideTestApplication`), because the real application class lives in `:app`.
- Changing the migration layout means an existing local database must be recreated: `make down` (which removes volumes), then `make up`.

## Alternatives considered

| Option | Why not |
|---|---|
| Keep one project, rely on Spring Modulith only | Boundaries are checked only at test time; no per-module dependencies or schema ownership. |
| Separate Postgres schemas per module (`ride.rides`, `driver.drivers`) | A good next step before extraction. Deferred because it changes every SQL statement; the migration layout already makes it a mechanical change. |
| `api`/`impl` sub-projects per module | Stricter, but doubles the number of projects for little gain while each module has a single small public API package. |
| Java modules (JPMS) | Spring and Testcontainers interplay adds friction; Gradle project boundaries plus Modulith give the same protection here. |
