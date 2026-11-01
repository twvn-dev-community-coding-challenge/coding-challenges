# ADR-003 — DDD and Clean Architecture inside each bounded context: domain, application, adapters, frameworks

- **Status:** Accepted
- **Relates to:** [ADR-001](ADR-001-module-structure.md) (one Gradle project per module), [ADR-002](ADR-002-decoupled-contexts.md) (ownership and events), [ADR-006](ADR-006-concurrency.md) (compare-and-set)

## Context

ADR-001 and ADR-002 give the system its strategic design: `ride`, `driver` and `dispatch` are bounded contexts with their own tables, talking only through versioned events in `contracts`. Inside each context, a review found the tactical design thin:

- Each module had a root package and an `internal` package holding domain records, JDBC repositories, Kafka listeners and schedulers side by side. Nothing kept frameworks out of the business rules.
- Lifecycle rules lived in two places. `RideTransition` listed the legal "from" statuses while the repository repeated them in SQL. Dispatch's request, offer and assignment lifecycles existed **only** as SQL `WHERE` clauses, with no Java model and no unit tests.
- Domain records used the wire type `contracts.AssignmentState`, the shared `Location` value threw the HTTP-flavoured `ValidationException`, and an HTTP DTO sat in the ride's public domain package.

## Decision

### 1. Contexts and the shared kernel

| Context | Kind | Owns | Relationship |
|---|---|---|---|
| **dispatch** | core domain | dispatch requests, offers, exclusions, the driver assignment, matching | Upstream for assignments: publishes them as the published language (`contracts`); ride and driver *conform* by keeping versioned copies. Downstream of driver for availability and location (copied into `driver_supply`). |
| **ride** | supporting | passengers (seeded), ride lifecycle, history, idempotency | Upstream of dispatch for ride requests and endings. |
| **driver** | supporting | availability, location | Upstream of dispatch for supply; conforms to dispatch's assignment. |
| `common.kernel` | **shared kernel** | `Location`, `GeoDistance`, `BoundingBox`, `StateTransition`, `TransitionOutcome`, `InvalidValueException` | Plain Java, used by every context's entities. |
| rest of `common` | frameworks & drivers | outbox, de-duplication, retention, cache, Kafka listener support, errors, logging; `common.model.LocationDto` for the web models | Used by the application, adapters and frameworks; never by the domain. |

**passengers** are part of ride: they are seeded records with no behaviour, and ride is the only context that needs them.

The vocabulary each context uses is in the [glossary](../architecture.md#ubiquitous-language).

### 2. Four layers in every context

```
com.ridehailing.<context>
├── domain            entities, value objects, lifecycle transitions, domain services (plain Java)
├── application       use cases: their transactions, event handlers, schedulers
│   └── port          the interfaces the use cases need: repositories, caches, passengers
├── adapters          translate between the outside world and the application
│   ├── web           REST controllers
│   │   └── model     their request and response models
│   ├── messaging     the module's @KafkaListener
│   ├── persistence   JDBC implementations of the ports
│   └── cache         Redis implementations of the ports
└── frameworks        Spring wiring: topics the module owns, configuration properties, beans
```

| Package | Clean Architecture ring | DDD layer |
|---|---|---|
| `domain` | Entities | Domain |
| `application` (+ `port`) | Use cases | Application |
| `adapters.*` | Interface adapters | Infrastructure (and presentation, for `web`) |
| `frameworks` | Frameworks & drivers | Infrastructure (configuration) |

**The dependency rule:** source code depends only inwards: `frameworks` → `adapters` → `application` → `domain`. Nothing depends on an adapter except the frameworks that wire it, and nothing depends on `frameworks`. Adapters of different kinds never use each other; they meet only in the application. Request and response models belong to the web adapter, and translate to application input (`CreateRideRequest.toCommand()`) and from domain objects (`RideResponse.from(ride)`).

- **The domain is plain Java.** It depends only on `java..`, the domain packages and the shared kernel (Modulith keeps one context out of another's): no Spring, SQL, Kafka, Jackson, Lombok or `contracts`.
- **The application owns the transactions** (`@Transactional`) and translates between domain objects and integration events (`DispatchEvents`, `DispatchProjection`, `AssignmentProjection`). It never touches SQL, Kafka or Redis.
- **Ports are the application's output interfaces**, in `application.port`: `RideRepository`, `IdempotencyRepository`, `IdempotencyCache`, `RideReadCache`, `Passengers` (ride); `DriverRepository`, `DriverReadCache` (driver); `DispatchRequestRepository`, `OfferRepository`, `ExclusionRepository`, `DriverSupplyRepository` (dispatch). They take and return domain objects.
- **Adapters implement the ports or call the application:** `adapters.persistence` holds the `Jdbc*` classes, `adapters.cache` the `Redis*` caches (the read models through `common.cache.VersionedJsonCache`, the idempotency replay through `VersionedCache` directly), `adapters.web` the controllers and their models (`adapters.web.model`), and `adapters.messaging` the listener.
- **`common` is reached through interfaces too.** `OutboxWriter`, `ProcessedEvents` and `EventCodec` are interfaces in `common.events`, implemented by `JdbcOutboxWriter`, `JdbcProcessedEvents` and `JacksonEventCodec`. The application never sees JDBC or Jackson, and a use case can be unit-tested with plain fakes (`AssignmentProjectionTest`).
- **Exceptions say what went wrong; the web adapter says how to report it.** Use cases throw `common.error` exceptions (`NotFoundException`, `ConflictException`, `ValidationException`, `UnprocessableException`), a sealed hierarchy under `ApplicationException` carrying a `code` and details but no HTTP status. `common.web.ApiExceptionHandler` maps each kind to its status in one exhaustive `switch`, so a new kind does not compile without one, and a Kafka listener that hits the same exception never sees HTTP.
- **Listeners are thin.** Decoding, filtering by event type and the logging context live in `common.events.InboundEvents`, so each listener is one call into the application.
- **Lombok writes the Spring beans' boilerplate and nothing else:** `@RequiredArgsConstructor` for constructor injection and `@Slf4j` for loggers, in the application, the adapters' beans, frameworks and `common`. The domain, the web request and response models, contracts and the shared kernel are records or plain classes without Lombok. `lombok.config` marks generated members `@lombok.Generated`, so JaCoCo skips them and the architecture check can see where Lombok is used.

### 3. Aggregates and where their invariants live

| Aggregate (root) | Context | Invariants |
|---|---|---|
| **Ride** (`Ride`) with its history | ride | The ride lifecycle (`RideTransition`); only the assigned driver's offer outcome moves an assigned ride; a terminal ride never changes. |
| **Driver** (`Driver`) | driver | Availability only OFFLINE ⇄ AVAILABLE (`DriverTransition`), never while the assignment copy is OFFERED/BUSY. |
| **Dispatch request** (`DispatchRequest`) with its offers and exclusions | dispatch | One driver at a time; only the holder may accept, reject or let the offer expire (`DispatchRequest.evaluate`); an excluded driver is never re-offered; an ended request never changes (`DispatchRequestTransition`, `OfferTransition`). |
| **Supply driver** (`SupplyDriver`) | dispatch | A driver is held for at most one ride; only a free, online, located driver can be claimed (`AssignmentTransition`). |

Every lifecycle is an enum implementing the kernel's `StateTransition<S>`: the statuses it starts from, its target, and any extra rule such as "requires the holder" or what happens to the held driver. `evaluate(current)` answers `APPLY`, `ALREADY_APPLIED` or `REJECTED`.

**The transition is the single source of the rule.** The persistence adapters build each compare-and-set from it: `status IN (:from)` from `fromNames()`, the holder condition from `requiresHolder()`/`requiresAssignedDriver()`, and the driver column from `heldDriver()`/`assignedDriver()`. A new or changed edge is one enum constant, covered by unit tests (`RideTransitionTest`, `DriverTransitionTest`, `DispatchTransitionsTest`, `DispatchRequestTest`).

### 4. Two deliberate deviations from the textbooks

1. **Immutable records plus compare-and-set, not load–modify–save.** Domain objects are records whose methods *decide* (`evaluate`, `isHeldBy`, `isFree`); the persistence adapter *applies* the decision as one conditional `UPDATE`. A mutable aggregate saved later would need optimistic locking and retries to keep AC-09 safe; the compare-and-set gets the same safety in one statement ([ADR-006](ADR-006-concurrency.md)).
2. **One transaction changes two dispatch aggregates.** Claiming a driver updates both the dispatch request and the supply driver, because "a driver holds at most one ride" (AC-09) spans both. They live in the same context and database, the request is always locked first, and partial unique indexes back the invariant. Splitting it into two transactions with compensation would trade a simple guarantee for a saga ([ADR-006](ADR-006-concurrency.md) § Consequences).

There are no input-port interfaces: controllers and listeners call the application's classes directly, because each use case has exactly one implementation and one caller per delivery mechanism. `application.port` holds output ports only.

### 5. Events

There is no in-process domain-event bus. The application writes **integration events** (`contracts.*V1`) to the outbox in the same transaction as the change ([ADR-005](ADR-005-event-driven-dispatch.md)), mapping domain values to the published language (for example `dispatch.domain.Assignment` → `contracts.AssignmentState`). Each context has its own `Assignment` enum, so a change to the wire format never reaches the domain.

### 6. Enforcement

`ModularityTest` runs the ArchUnit rules of `ArchitectureRules` inside Spring Modulith's `verify()` ([ADR-001](ADR-001-module-structure.md) decision 7), and fails the build when:

- a layer depends outwards (`DEPENDENCIES_POINT_INWARDS`);
- one kind of adapter uses another, for example a controller using a JDBC class (`ADAPTERS_DO_NOT_DEPEND_ON_EACH_OTHER`);
- a top-level type in `application.port` is not an interface (`PORTS_ARE_INTERFACES`);
- the domain uses anything but `java..`, the domain packages and the shared kernel, so no Spring, SQL, Kafka, Redis, HTTP, Jackson, Lombok or `contracts` (`DOMAIN_IS_PLAIN_JAVA`); the shared kernel or `contracts` uses anything but `java..` and itself (`SHARED_KERNEL_IS_PLAIN_JAVA`, `CONTRACTS_ARE_PLAIN_JAVA`);
- the application or a `common.error` exception uses Spring's HTTP or web classes (`APPLICATION_AND_ERRORS_KNOW_NOTHING_OF_HTTP`);
- the application uses a class in `common.events` that is not an interface, or uses Jackson (`APPLICATION_USES_COMMON_EVENTS_THROUGH_INTERFACES`);
- the application or the web adapter uses JDBC or Kafka (`APPLICATION_AND_WEB_DO_NOT_USE_SQL_OR_KAFKA`);
- anything but `OutboxRelay` and `MessagingConfig` uses `KafkaTemplate`, or anything outside `common.cache` uses Redis (`ONLY_THE_OUTBOX_RELAY_PUBLISHES_TO_KAFKA`, `ONLY_THE_COMMON_CACHE_TALKS_TO_REDIS`);
- a web request or response model (`adapters.web.model`) has Lombok-generated members (`NO_LOMBOK_IN_WEB_MODELS`).

Which modules may depend on which (ride, driver and dispatch only on `common` and `contracts`) is declared in each module's `package-info.java` as `@ApplicationModule(allowedDependencies = …)` and checked by the same `verify()`.

## Consequences

- Four packages per context, one per ring, so the dependency rule reads straight off the tree; inside `adapters`, the sub-package names the technology.
- The dispatch lifecycle is a tested Java model, not only SQL, and changing a lifecycle edge is one enum change: the guard in SQL follows it.
- Ports add one interface per implementation (eleven in all, plus three in `common.events`), which is boilerplate at this size; it is what keeps the application free of SQL, Jackson and Redis, and testable without them.
- Lombok removes about 36 hand-written constructors and 20 logger fields, at the cost of an annotation processor in every module, IDE support, and a Lombok release that supports each new JDK (1.18.42 for Java 25, managed by the Spring Boot BOM).
- The domain-purity rule caught one leak while the layers were introduced: `Location` threw the web-layer `ValidationException`. It now throws the kernel's `InvalidValueException`, which `ApiExceptionHandler` maps to the same `400` and code.
- The SQL that applies each transition lives in the persistence adapters. That is persistence of a decision the domain made, and the database constraints stay as the last line of defence.

## Alternatives considered

| Option | Why not |
|---|---|
| `domain` / `application` / `infrastructure` / `web` (the first version) | Worked, but `infrastructure` mixed inbound and outbound adapters, and repository interfaces sat in the domain rather than with the use cases that need them. |
| `entity` / `usecase` / `controller` / `model` / `listener` / `gateway` / `config` (the second version) | Seven role names for four rings: the rings were not visible in the tree, and the models had a package apart from the controllers they belong to. |
| Adapters by direction, then technology (`adapters.in.web`, `adapters.out.persistence`) | One more level for no new information: an adapter's technology already tells its direction here, and the ports show what the application needs. |
| Rich mutable aggregates with JPA and `@Version` | Load–modify–save under contention turns AC-09 into optimistic-lock failures and retries, and needs an ORM for little gain over records. |
| A Gradle project per layer (`ride-domain`, `ride-application`, …) | Compile-time enforcement, but six times as many projects; ArchUnit gives the same guarantee in one test. |
| Lombok everywhere, including `@Value`/`@Builder` on domain objects and models | Records already give constructors, accessors, `equals` and `toString`; the domain stays plain Java. |
| No Lombok | Explicit constructors are the most transparent; dropped because the 36 bean constructors only repeated their fields. |
| An in-process domain-event bus (Spring application events) | Two event mechanisms to explain; every event that matters already crosses a context and must be durable, so it goes to the outbox. |
| A separate passenger module | One class and one query; ride is its only user. Splitting it out again is mechanical if passengers ever gain behaviour. |
