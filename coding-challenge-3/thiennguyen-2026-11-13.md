# Coding Challenge #3 Review: Thien Nguyen

**Review Date:** 2026-11-13
**Submission Path / PR URL:** https://github.com/twvn-dev-community-coding-challenge/coding-challenges/pull/18
**Reviewed commit:** `21e1dd3a29eb731d8a82327a3c1d59fef594bf8b` (PR head; same commit as the 2026-11-02 and 2026-11-12 reviews)
**Submission directory:** `coding-challenge-3/thiennguyen`
**Reviewer:** AI Code Reviewer (Hanh Tran), using `.agents/skills/review-coding-challenge-3` from PR #6 (`bf701e6`) and `openspec/specs/challenge-3-review-scoring/spec.md`
**Tech Stack:** Java 25, Spring Boot 4, Spring Modulith 2, PostgreSQL 18, Kafka 4 (KRaft), Redis 8, Gradle 9 (Kotlin DSL), JUnit 5, Testcontainers 2, Docker Compose.

All citations are relative to the submission directory at the reviewed commit. The brief used is `requirement-submission/challenge-3/ride-hailing-dispatch/README.md` at `origin/main`, including its Judge Scoring Guide.

---

## Verification Summary

| Evidence type | Scope / command | Outcome |
| --- | --- | --- |
| Execution in this review | None. The submission needs JDK 25 for Gradle, and Docker for its integration tests (Testcontainers) and for `make up`. Neither was run in this review environment. | **Not run.** Nothing in this report is described as passing or failing at runtime. |
| Inherited execution (2026-11-12 review, not reproduced) | `./gradlew test`; `./gradlew :app:integrationTest` | That review reports **124 unit/architecture/MVC tests passing**, and the integration task **could not start: no Docker environment**. Its evidence files (`results/coding-challenge-3/evidence/thiennguyen-2026-11-12/`) are not committed to this repository. |
| Static inspection | All of `modules/{ride,driver,dispatch,common,contracts}/src/main`, all 10 SQL migrations, `app/src/main/resources`, the 76 unit-test and 80 integration-test methods (`modules/*/src/test`, `app/src/test`, `app/src/integrationTest`), `README.md`, `docs/` (architecture, sequences, scalability, testing strategy, 8 ADRs), `scripts/demo.sh`, `compose.yaml`, `Makefile` | All eight required flows traced. Reservation boundary, event boundary and failure paths mapped below. |
| Searches for absence | `.github/` in the submission; replay assertions in `AC01CreateRideAcceptanceIT`; race tests for accept vs timeout and accept vs cancel | No `.github/` directory, though `docs/testing-strategy.md:119–121` describes a CI workflow there. The replay test asserts the same ID only, not the original state. No accept-vs-timeout or accept-vs-cancel race test, which `README.md:267` acknowledges. |

**Evidence limitations:** This is a static review. Postgres compare-and-set semantics, Kafka delivery, DLT routing and the barrier-synchronised race tests were read, not run. Behaviour that depends on runtime is labelled **static inspection**. All tests, including the integration suite, should be run on a machine with Docker before these scores are treated as final. The integration suite is especially important here, because it is the authored proof for AC-09 and Testing.

---

## Final Evaluation

> **Exceptional — 91/100**

**Judging rationale:** Every acceptance criterion is implemented. Dispatch owns all of its decisions in one transaction per command:
- request-row lock, then a supply-row compare-and-set
- partial unique indexes as a schema backstop against double assignment
- a durable per-ride exclusion table
- an offer expiry that is SQL-guarded
- a transactional outbox to Kafka with `(consumer, eventId)` de-duplication

The authored tests match the brief almost one-to-one, including a real `CyclicBarrier` race repeated 40 times. It loses points on three things:
- the driver module decides availability from a lagging copy (F1)
- two failure paths through the dead-letter topic can leave state permanently wrong (F2, F3)
- a replay returns the ride's current state rather than the original response (F4)

It also loses points for the distributed-copy complexity itself, for documentation that overstates recovery and CI, and for an incomplete demo.

**Engineering assessment:** The core dispatch decision is unusually rigorous. Correctness rests on SQL compare-and-set and constraints rather than application checks, and module boundaries are enforced by Spring Modulith and ArchUnit. The cost is choreography: ride and driver each keep an eventually consistent copy of dispatch's state. That design is the direct source of F1–F3, none of which a single state owner would have. All three are narrow (a race window or a dead-letter path), documented in part, and not observed at runtime.

---

## Score Summary

| Category | Score | Max |
| --- | ---: | ---: |
| Acceptance Criteria | 33 | 35 |
| System Architecture | 19 | 20 |
| Distributed Systems & Concurrency | 12 | 15 |
| Testing & Validation | 14 | 15 |
| Scalability & Performance Design | 5 | 5 |
| Observability | 5 | 5 |
| Documentation & Communication | 3 | 5 |
| **Total** | **91** | **100** |

| Score | Interpretation |
| --- | --- |
| 90–100 | Exceptional |
| 80–89 | Strong |
| 70–79 | Good |
| 60–69 | Meets expectations |
| 50–59 | Needs improvement |
| <50 | Does not meet expectations |

---

## Critical Issues

| Issue | Severity | Evidence | Score impact |
| --- | --- | --- | --- |
| AC-09 concurrent driver assignment: **Safe (static inspection)** | None | The claim is a single conditional `UPDATE driver_supply … WHERE driver_id = :id AND assignment IN ('NONE') AND online AND latitude IS NOT NULL AND current_ride_id IS NULL RETURNING …` (`modules/dispatch/src/main/java/com/ridehailing/dispatch/adapters/persistence/JdbcDriverSupplyRepository.java:62–83`). It runs inside the dispatch transaction after the request row is locked (`DispatchService.java:69–100`). Under Postgres READ COMMITTED, a second concurrent UPDATE waits on the row lock, re-evaluates the `WHERE` clause against the committed row and matches zero rows. Partial unique indexes enforce the same invariant even if code regresses: `dispatch_requests_active_driver_uq` and `driver_supply_current_ride_uq` (`V20261026_0600__dispatch_requests_and_driver_supply.sql:22–23, 47`), and `ride_offers_pending_driver_uq` / `_ride_uq` (`V20261026_0500__offers_and_exclusions.sql:14–15`). | AC-09 4/4. Concurrency row 4/5 because the authored race test could not be executed (calibration: static inspection does not prove concurrency behaviour). No unsafe-race penalty. |
| F1: Driver can go OFFLINE while dispatch holds them | Major | `DriverService.changeAvailability` decides on the driver module's own copy of the assignment (`modules/driver/…/application/DriverService.java:46–58`), which lags dispatch's authoritative `driver_supply`. The Javadoc states that going offline "still succeeds" if dispatch has just made an offer. `SupplyUpdates.availabilityChanged` then expires a pending offer, or, for a trip, "keeps the trip and just stops receiving offers" (`SupplyUpdates.java:20–22, 42–45`). | AC-02 −1. |
| F2: A dead-lettered predecessor poisons the ride projection | Medium | `DispatchProjection.apply` marks the event processed (`DispatchProjection.java:36`) and then ignores an invalid transition with only a log line (`:67–71`). If `RideOffered` is dead-lettered, the following `RideAccepted` hits a `MATCHING` ride, is ignored and consumed, and a later replay of `RideOffered` cannot complete the sequence. The ride then shows `MATCHING` or `DRIVER_ASSIGNED` while dispatch has `ACCEPTED`, and `complete` is rejected. The 2026-11-12 review reproduced this at handler level. | Event processing semantics −1. |
| F3: A dead-lettered `RideCancelled` keeps offering a cancelled ride | Medium | If `RideCancelled` fails three times for an `OFFERED` request, `RideOutcomes.rideEnded` never runs (`RideOutcomes.java:40–72`). The pending offer expires (`OfferService.java:105–127`), the request returns to `WAITING`, and `MatchingSweeper` re-dispatches it every 5 s (`MatchingSweeper.java:27–33`). Dispatch cannot see the ride's status. A driver who accepts becomes `BUSY`, while the ride projection ignores the acceptance on a `CANCELLED` ride and `complete` is refused, so the driver is never released. `README.md:193` and `ADR-005:49` say periodic jobs mean a dead-lettered message "never strands a ride". | Failure handling −1; README −1 (together with the CI claim). |

No path to double assignment was found.

---

## Business Flow and Evidence Ledger

Paths below are under `modules/<module>/src/main/java/com/ridehailing/<module>/` unless shown in full.

| Flow / objective | Component and state owner; reads/writes; event/atomic boundary; invariant | Evidence state and verification mode | Evidence / gaps / finding ID / scoring rows |
| --- | --- | --- | --- |
| Create → event → match → offer | `POST /rides` → `RideCreationService` (ride). In one transaction: the passenger check, the `idempotency_keys` claim (key + SHA-256 of the canonical command), the ride insert, history and the `RideRequested` outbox row (`RideCreationService.java:49–73`). `OutboxRelay` publishes to `ride-events`, keyed by ride ID. `DispatchEventProcessor.process` de-duplicates on `(dispatch, eventId)` in the same transaction (`DispatchEventProcessor.java:41–46`), opens the request and runs `dispatch()`: lock the request, query free drivers in a bounding box, rank them with `NearestDriverMatcher`, compare-and-set the first one, insert the offer, and emit `RideOffered` and `DriverAssignmentChanged` (`DispatchService.java:54–119`). | Implemented and verified by static inspection; authored integration test not executed | `app/src/integrationTest/…/acceptance/AC10EventDrivenDispatchAcceptanceIT.java:44` (through the real broker). AC-01, AC-04, AC-10. |
| Acceptance | `OfferService.accept` locks the request, checks the holder through `DispatchRequest.evaluate`, then runs `offers.accept` with `expires_at > :now` (`JdbcOfferRepository.java:45–48`), request `ACCEPT`, and supply `START_TRIP` → `BUSY`. It emits `RideAccepted`. The ride projection applies `ACCEPT` asynchronously. | Implemented and verified by static inspection | `OfferService.java:48–75`; `AC05RideOfferAcceptanceIT.java:61`. A duplicate accept returns the offer with no side effects (`:52–55`; `AC08IdempotencyAcceptanceIT.java:55`). AC-05, AC-08. |
| Rejection → rematch | `OfferService.reject` resolves the offer `REJECTED`, then `returnToWaiting` adds an exclusion row (primary key ride + driver), moves the request to `WAITING` and releases supply. It emits `OfferRejected`, which dispatch consumes from its own topic to re-match (`OfferService.java:78–98, 137–145`; `DispatchEventProcessor.java:52`). Matching skips excluded drivers. | Implemented and verified by static inspection | Exclusion is durable and per ride (`V20261026_0500…:19–25`); `AC07RematchingAcceptanceIT.java:27, 40, 60`. No manual offer route exists. AC-07. |
| Timeout → rematch | `OfferExpiryScheduler.expireDueOffers` runs every second (`OfferExpiryScheduler.java:25`). `OfferService.expire` locks the request, uses `expireIfDue` (`expires_at <= :now`), returns to waiting with an exclusion, and emits `OfferExpired`, which leads to a re-match. | Implemented and verified by static inspection | `OfferService.java:105–127`; `AC05RideOfferAcceptanceIT.java:84` (mutable test clock). AC-05, AC-06, AC-07. |
| Cancellation → release | `RideCommandService.cancel` checks the ride's own status (all four cancellable states, `RideTransition.java:21`) and emits `RideCancelled`. Dispatch's `RideOutcomes.rideEnded` locks the request, cancels a pending offer, releases `OFFERED` or `BUSY` supply, and moves the request to `CANCELLED`. If the ride ended before dispatch opened it, dispatch records it as ended so a late `RideRequested` is ignored. | Implemented with weaknesses (static inspection) | `RideCommandService.java:34–51`; `RideOutcomes.java:40–72`; `AC05RideOfferAcceptanceIT.java:116, 128`; `AC10…:98`. **F3** when `RideCancelled` is dead-lettered. AC-05, Failure handling. |
| Completion → release | `RideCommandService.complete` requires `DRIVER_ACCEPTED` in the ride's copy and emits `RideCompleted`. Dispatch releases `BUSY` → `NONE`. | Implemented and verified by static inspection | `RideCommandService.java:54–71`; `AC05RideOfferAcceptanceIT.java:139`; `AC02…:86`. If the ride copy lags, completion returns 409 until the projection catches up (documented eventual consistency). |
| Duplicate request/event | Creation: database key claim in the ride transaction, with concurrent duplicates blocking on the primary key and then replaying. A hash mismatch gives 422. Redis replay cache sits in front but never decides. Commands: already-applied outcomes return current state with no new event. Events: `processed_events` insert in the consumer's transaction. Driver copies are applied only if their version is newer. | Implemented with weaknesses (static inspection) | `RideCreationService.java:40–87`; `AC08IdempotencyAcceptanceIT.java:39, 55, 73, 84`; `AC01CreateRideAcceptanceIT.java:106–140`. **F4:** a replay returns the ride's current state (`RideCreationService.java:80–82`). AC-01, AC-08. |
| Two rides compete for one driver | Reservation point: the supply-row compare-and-set inside the dispatch transaction. The loser walks to the next ranked candidate or returns `NO_DRIVER`, and its request stays `WAITING`. Locks are always taken request row first, then supply row, so there is no deadlock. Unique indexes back this up. | Implemented and verified by static inspection; authored barrier test not executed | `AC09ConcurrentAssignmentAcceptanceIT.java:38–56` (20 repetitions, exactly one `OFFERED` and one `NO_DRIVER`, supply, request and pending-offer assertions), `:58–72` (loser falls through to the next driver), `:74–88` (accept vs reject); `ConcurrentCalls.java:19–24` (`CyclicBarrier`). AC-09, Concurrency, Testing. |
| Competing accept / cancel / timeout | Accept, reject and expiry all lock the dispatch request row first, and `RideOutcomes` does too, so they serialise. Cancellation wins over a concurrent acceptance, because the later `RideCancelled` releases the `BUSY` driver, and the ride projection ignores `RideAccepted` on a `CANCELLED` ride. | Implemented and verified by static inspection; no authored test | `OfferService.java:31–33` (contract comment), `RideOutcomes.java:24–25`. `README.md:267` acknowledges the missing race tests. Testing → Concurrency & idempotency tests −1. |

## Key Technical Findings

| ID | Severity | Problem and evidence | Impact | Recommendation | Official row(s), distinct score impact or advisory |
| --- | --- | --- | --- | --- | --- |
| F1 | High | Availability is decided on the driver module's copy of the assignment (`DriverService.java:46–58`), not on dispatch's authoritative `driver_supply`. Documented as a trade-off (`README.md:258`; `SupplyUpdates.java:20–22`). `AC05RideOfferAcceptanceIT.java:98` deliberately exercises the "offline while offered" path. | In the window between dispatch claiming a driver and the driver module seeing `DriverAssignmentChanged`, `PUT /availability OFFLINE` succeeds for an `OFFERED` or `BUSY` driver. That breaks the brief's fixed rule that `OFFERED` and `BUSY` drivers cannot switch to `OFFLINE`. The offered case self-heals (the offer expires). The busy case leaves an offline driver on a trip. | Route the availability decision through dispatch's supply row (one compare-and-set on `online` guarded by `assignment = 'NONE'`), or reject `OFFLINE` there and compensate. | AC-02 −1. |
| F2 | Medium | The projection marks an event processed before applying it, and ignores invalid transitions (`DispatchProjection.java:35–71`). Per-ride Kafka ordering normally guarantees the order, but a dead-lettered predecessor breaks it, and replaying the dead letter cannot repair the ride. | After any dead-lettered `RideOffered` or `MatchingStarted`, the ride copy can be stuck permanently out of step with dispatch. A stuck ride also blocks `complete`. Only reachable after three failed attempts. | Don't mark a skipped prerequisite as processed (retry it or park it), or rebuild the ride status from dispatch's current state instead of applying event deltas. | Event processing semantics −1. |
| F3 | Medium | A dead-lettered `RideCancelled` or `RideCompleted` never reaches `RideOutcomes` (`RideOutcomes.java:40–72`). `MatchingSweeper` keeps re-matching the `WAITING` request (`MatchingSweeper.java:27–33`). | A cancelled ride keeps being offered to drivers, and one who accepts is held `BUSY` with no way to complete. This contradicts the recovery claims in `README.md:193` and `ADR-005:49`. Static risk, not observed. | Let the sweeper confirm with ride before re-matching (through a ride-status copy in dispatch), or treat terminal events as must-not-dead-letter with a repair job. | Failure handling −1. Documentation accuracy is counted under README. |
| F4 | Medium | A replay returns `rideCache.get(rideId).or(() -> rides.find(rideId))`, the ride's current state (`RideCreationService.java:75–87`), with `Idempotent-Replayed: true`. `AC01CreateRideAcceptanceIT.java:106–117` asserts only the ID. | Once dispatch has run, a retried create returns `MATCHING` or `DRIVER_ASSIGNED` instead of the original `REQUESTED` response. The brief requires the original result. (The fingerprint itself is canonical, so key order doesn't matter, unlike the PR #19 submission.) | Store the original response body or a snapshot with the idempotency key, and add a replay-after-dispatch test. | AC-01 −1. Applied the same way as in the Jessica Nguyen review, for consistency across submissions. |
| F5 | Medium | Three copies of dispatch state (dispatch's `driver_supply`, driver's `drivers.assignment`, ride's projected status) plus an outbox, Kafka, Redis, sweepers and eight ADRs, in a 4–6 hour challenge. | F1–F3 all exist only because state is copied across modules. A single dispatch-owned state with synchronous reads by ride and driver would avoid all three. The complexity has a measurable correctness cost here, beyond its learning and extraction value. | Keep the module boundaries but make dispatch the only owner of assignment, and let other modules read it rather than copy it, until extraction is actually needed. | System Architecture → Architecture structure −1. |
| F6 | Low | No authored race test for accept vs timeout or accept vs cancel (`README.md:267`). There is also no fault-injection test for a dead-lettered predecessor (F2) or a dead-lettered terminal event (F3). | The lock-ordering argument is sound but unproven for the two most likely lifecycle races, and both dead-letter paths are untested. | Add `ConcurrentCalls` tests for those two pairs, and a DLT replay test. | Testing → Concurrency & idempotency tests −1. |
| F7 | Low | `docs/testing-strategy.md:3, 119–121` says CI runs `.github/workflows/ci.yml` "on every push and pull request". No `.github/` exists in the submission, and the PR has no check runs (per the 2026-11-12 review). The recovery claim in F3 is also overstated. | Readers can't rely on the documented CI or recovery guarantees. | Add the workflow or mark it as planned, and narrow the recovery claims. | README 1/2. |
| F8 | Low | `scripts/demo.sh` plays create → match → reject → rematch → accept → complete plus a trace (`README.md:40–53`), but not duplicate requests, concurrent assignment or the test run. | Partial demo evidence. | Extend the script with a key replay, a two-ride race and `make it` output. | Demo/communication 0/1 (re-score after Demo Day). |

---

## Acceptance Criteria — 33/35

| AC | Result | Score | Evidence and lost-points rationale |
| --- | --- | ---: | --- |
| AC-01 Create Ride | Partial | 3/4 | **Implemented with weaknesses; static inspection (MVC and fingerprint tests executed per the 2026-11-12 review).** Seeded passengers, bean-validated bounded coordinates, a UUID per ride, `REQUESTED`, `GET /rides/{id}`, a required `Idempotency-Key`, a database-transactional key claim with a canonical SHA-256 fingerprint, and 422 on reuse with a different request (`RideCreationService.java:40–73`; `AC01CreateRideAcceptanceIT.java:35–140`). **−1 (F4):** a replay returns the current state, not the original result. |
| AC-02 Driver Availability | Partial | 2/3 | **Implemented with weaknesses; static inspection.** Full lifecycle (`OFFLINE` ↔ `AVAILABLE`, claim → `OFFERED`, accept → `BUSY`, release on reject, expiry, cancel and complete) and rejection of invalid changes on the normal path (`AC02DriverAvailabilityAcceptanceIT.java:27–109`). **−1 (F1):** going offline is decided on a lagging copy, so in a race window an `OFFERED` or `BUSY` driver can go `OFFLINE`, against the brief's fixed rule. |
| AC-03 Driver Location | Demonstrated | 2/2 | **Implemented and verified by static inspection.** Validated update, versioned `DriverLocationUpdated`, and a supply copy applied only if newer (`JdbcDriverSupplyRepository.java:102–118`). Matching reads the latest applied location. |
| AC-04 Driver Matching | Demonstrated | 4/4 | **Implemented and verified by static inspection; matcher unit and property tests executed per the 2026-11-12 review.** Indexed bounding-box pre-filter on online, unassigned drivers (`JdbcDriverSupplyRepository.java:46–60`), Haversine ranking, inclusive 2 km default, lexicographic `driverId` tie-break, exclusions, empty result (`NearestDriverMatcher`; `NearestDriverMatcherTest`). |
| AC-05 Ride Offer | Demonstrated | 4/4 | **Implemented and verified by static inspection.** Pending offer with the ride `DRIVER_ASSIGNED`, holder-only responses (`OFFER_NOT_HELD`), accept guarded by `expires_at > now`, 30 s timeout treated as rejection, and cancellation releasing both `OFFERED` and `BUSY` drivers (`OfferService.java:48–127`; `RideOutcomes.java:57–66`; `AC05RideOfferAcceptanceIT.java:39–139`). |
| AC-06 Ride State Machine | Demonstrated | 4/4 | **Implemented and verified by static inspection; transition tests executed per the 2026-11-12 review.** An enum transition table that generates the SQL compare-and-set guards, including `DRIVER_ASSIGNED → MATCHING`, all four cancellations, terminal states, clearing the assigned driver on return to matching, duplicate no-ops, and persisted history (`RideTransition.java:16–21`; `RideTransitionTest`). |
| AC-07 Rematching | Demonstrated | 3/3 | **Implemented and verified by static inspection.** Rejection and expiry add a durable `(ride_id, driver_id)` exclusion and re-match through dispatch's own event. A ride with no eligible driver stays `WAITING`/`MATCHING` with no driver (`OfferService.java:137–145`; `AC07RematchingAcceptanceIT.java:27–60`). |
| AC-08 Idempotency | Demonstrated | 4/4 | **Implemented and verified by static inspection.** Creation (transactional key, including concurrent duplicates), acceptance, rejection, cancel, complete, and Kafka redelivery (`processed_events` in the same transaction). The strategy is documented (`README.md:174–183`; `AC08IdempotencyAcceptanceIT.java:39–84`). The replay-content gap (F4) is charged under AC-01. |
| AC-09 Concurrent Driver Assignment | Demonstrated | 4/4 | **Implemented and verified by static inspection.** A database compare-and-set claim with request-then-supply lock ordering and four partial unique indexes as a backstop. The losing ride falls through or stays `WAITING`. A barrier-synchronised authored test races two rides 20 times and asserts exactly one offer (`AC09ConcurrentAssignmentAcceptanceIT.java:38–72`). The strategy is documented (`ADR-006`). The test was not executed here; see Concurrency row. |
| AC-10 Event-Driven Dispatch | Demonstrated | 3/3 | **Implemented and verified by static inspection.** Transactional outbox → Kafka → per-module listeners. Creation never waits for matching. Duplicate delivery is de-duplicated. Failures retry (0.5/1/2 s) and then go to the DLT. The README explains why async was chosen, duplicate handling and failure handling (`README.md:185–200`; `AC10EventDrivenDispatchAcceptanceIT.java:44–98`). F2 and F3 affect copies after a dead-letter and are charged under Distributed Systems, not here: the Ride Created → dispatch → matching flow itself is intact. |
| **Total** |  | **33/35** | |

---

## Detailed Scoring

### System Architecture — 19/20

Each business module (`ride`, `driver`, `dispatch`) is its own Gradle project with its own migrations and no cross-module foreign keys. Allowed dependencies are declared to Spring Modulith and enforced by `ModularityTest` and the ArchUnit rules (domain is plain Java, layers point inwards, only adapters touch SQL, Kafka or Redis). `dispatch` depends only on `common` and `contracts`. Change cost is concrete: matching changes stay inside `NearestDriverMatcher` plus one SQL pre-filter, and extraction is a new Boot app plus its own database. The cost is F5: the choreography copies dispatch's state into two other modules, and that copying is what makes F1–F3 possible.

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Domain boundaries & responsibilities | 4 | 4 | Clear ownership: ride owns passengers, rides and history; driver owns availability and location; dispatch owns requests, supply, offers, exclusions and the claim. Each lifecycle is an enum transition table (`RideTransition`, `DispatchRequestTransition`, `AssignmentTransition`, `OfferTransition`). |
| Architecture structure | 3 | 4 | Clear, enforced layering and dependencies (`app/src/test/…/ModularityTest.java`, `ArchitectureRules.java`). **−1 (F5):** complexity is out of proportion to the brief, with a measurable correctness cost. Three copies of assignment state plus Kafka choreography produce F1–F3, which a single owner would not have. This scores the proportionality objective only; each defect is charged once elsewhere. |
| API design | 3 | 3 | Resource-oriented endpoints, bean validation, RFC 9457 problem bodies with stable codes (400/404/409/422), `Location` and `Idempotent-Replayed` headers, holder checks (`README.md:55–75`; `ApiExceptionHandler`). Eventual-consistency responses are documented. The F1 state restriction is charged under AC-02. |
| Event-driven architecture | 3 | 3 | Owned topics, versioned payload records in `contracts`, an envelope with `eventId`, keyed partitioning, outbox relay under an advisory lock to keep per-key order, and a documented event catalogue (`docs/architecture.md`). The dead-letter recovery gaps are charged under Failure handling and Event semantics. |
| Data architecture | 2 | 2 | CHECK constraints for the status/driver relations, partial unique indexes, an indexed free-driver location, versioned copies, and a durable exclusion table (`V20261026_0500…`, `V20261026_0600…`, `V20261029_0200…`). |
| Architecture trade-offs | 2 | 2 | Eight ADRs with alternatives and consequences: PostGIS/Redis GEO/H3/batch for matching, compare-and-set vs optimistic or distributed locks, choreography vs calls, cache-only Redis (`docs/adr/`). |
| Architecture evolution | 2 | 2 | A staged 1K → 100K → 1M plan with a load model (`docs/scalability.md:1–48`): CDC replacing the relay, Redis GEO/H3 for live locations, geo-partitioned Kafka and workers, batch matching, sharding by city, push instead of polling. |
| **Total** | **19** | **20** | |

### Distributed Systems & Concurrency — 12/15

Reservation is decided entirely in dispatch's tables, in one transaction per command, with a consistent lock order. That is correct for multiple instances as well as one, not just within a process. Delivery is at least once, with per-key ordering preserved by a single relay and consumer de-duplication. The weaknesses are on the dead-letter path: copies apply event deltas and assume no predecessor was lost (F2), and dispatch has no way to learn about a terminal ride it missed (F3).

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Concurrency / race conditions | 4 | 5 | The race is identified, the mechanism is correct (SQL compare-and-set plus unique indexes plus lock ordering), the trade-offs are explained (`ADR-006`), and an authored overlapping race test exists. **−1:** that test could not be executed here, and static inspection does not prove concurrency behaviour (calibration). Raise to 5/5 once `make ac AC=09` is run and passes. **AC-09 is safe, so the mandatory 0/5 does not apply.** |
| Idempotency | 3 | 3 | Actual protection in code for creation (transactional key and canonical hash), acceptance and transitions (already-applied outcomes), and events (`processed_events` primary key in the same transaction). |
| Event processing semantics | 2 | 3 | At-least-once delivery, per-key ordering, retry with back-off, DLT, versioned copies and retention are coherently explained and implemented. **−1 (F2):** the projection's ordering assumption fails after a dead-lettered predecessor, which is marked consumed and cannot be repaired by replay. |
| Failure handling | 1 | 2 | Transaction rollback with no offset commit, relay retries, three sweepers, dead-letter metrics, and `recordEnded` for a ride cancelled before dispatch saw it. **−1 (F3):** a dead-lettered terminal event leaves dispatch re-matching a cancelled ride and can strand a driver `BUSY`. |
| Consistency | 2 | 2 | Within dispatch, request, supply, offer and outbox rows change together under CHECK and unique constraints. The other modules' copies converge in the normal path, and that eventual consistency is documented (`README.md:258`). The F1–F3 divergences are charged once in their own rows. |
| **Total** | **12** | **15** | |

### Testing & Validation — 14/15

There are 76 unit, architecture and MVC test methods and 80 integration test methods: one acceptance class per AC, plus cache, outbox, metrics, observability, retention and persistence-constraint tests. Integration tests share one Spring context and one set of containers, isolate themselves on separate map patches (`TestCity`), and check global invariants after each test (`SystemInvariants`). Scoring is based on what the assertions prove. Only the unit tests have been executed (2026-11-12 review). Integration tests are unrun in both reviews.

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Unit tests | 4 | 4 | Matcher with property tests, geo and bounding box, location bounds, every transition table, request and offer domain, fingerprint, assignment projection, controller/MVC (`NearestDriverMatcherTest`, `RideTransitionTest`, `DispatchTransitionsTest`, `RequestFingerprintTest`, `RideControllerTest`). |
| Integration tests | 3 | 3 | Real API → Postgres → outbox → Kafka → dispatch → offer through Testcontainers (`AC10EventDrivenDispatchAcceptanceIT.java:44`), plus a dead-lettered unprocessable event, sweeper recovery, an ended-before-opened ride, Redis paused (`CacheIT`), outbox ordering (`OutboxRelayIT`) and database constraints (`PersistenceConstraintsIT`). Authored, not executed here. |
| Edge-case testing | 3 | 3 | No driver, out of radius, offline driver, rejection and permanent exclusion, timeout, offline while offered, cancellation pending and accepted, invalid transitions, duplicate requests and events (`AC02`–`AC08`). |
| Concurrency & idempotency tests | 2 | 3 | Strong: a barrier-synchronised two-ride race (repeated 20 + 10 times), an accept-vs-reject race, concurrent duplicate creates and real event redelivery (`AC09…:38–88`; `AC08…:39, 84`). **−1 (F6):** no accept-vs-timeout or accept-vs-cancel race (acknowledged in `README.md:267`) and no dead-letter replay or fault test, even though the design depends on those paths. |
| Test quality | 2 | 2 | Deterministic mutable clock, isolated map patches, invariant checks after each test, readable behaviour names, and AC tags so one AC can be run on its own (`make ac AC=09`). |
| **Total** | **14** | **15** | |

### Scalability & Performance Design — 5/5

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Bottleneck identification | 2 | 2 | Location write volume through the single outbox, the matching query in dense areas, contention on the nearest drivers, the single primary database, polling reads, partition count and rematch sweeps during shortages (`docs/scalability.md:15–27`). |
| Scaling strategy | 2 | 2 | CDC, Redis GEO/H3, geo-partitioned Kafka and workers, batch matching, city sharding, push, back-off, and service extraction, staged by driver count (`docs/scalability.md:28–47`). |
| Performance considerations | 1 | 1 | Explicit trade-offs with a load model (`docs/scalability.md:5–14, 48+`). No load test required. |
| **Total** | **5** | **5** | |

### Observability — 5/5

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Structured logging | 2 | 2 | Fixed-width console layout with timestamp, level, class and `[rideId eventId driverId]` on every line, filled from MDC by the HTTP filter, Kafka listeners, outbox and services (`app/src/main/resources/logback-spring.xml`; `README.md:206–231`). |
| Correlation / tracing | 1 | 1 | `rideId`, `eventId` and `driverId` travel in the event envelope and go into MDC per handler. |
| Metrics | 1 | 1 | Outbox backlog and age, event latency, dead letters, republished rides, dispatch attempts by outcome, lost claims, time to offer, offer outcomes, cache hit rates, all recorded after commit (`DispatchMetrics.java:25–56`; `README.md:233–245`; `MetricsIT`). |
| Troubleshooting capability | 1 | 1 | `make trace RIDE=<id>`, `GET /rides/{id}/history`, and `ObservabilityIT` asserting a complete trace. |
| **Total** | **5** | **5** | |

### Documentation & Communication — 3/5

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| README | 1 | 2 | Run, test, API, error codes, matching, concurrency, idempotency, events, caching, observability, scale and a candid limitations list. **−1 (F7, F3):** the recovery claim ("never strands a ride", `README.md:193`) is contradicted by F3, and the CI pipeline described in `docs/testing-strategy.md:119–121` does not exist in the submission. One point for documentation accuracy. |
| Architecture diagrams | 1 | 1 | C4 views, module rules, state machines, data model and event catalogue (`docs/architecture.md`), plus sequences for create, reject/rematch/accept/complete, timeout and offline, concurrent assignment, cancel vs accept, and failure (`docs/sequence-diagram.md`). |
| ADRs / design decisions | 1 | 1 | Eight meaningful ADRs, including matching (ADR-004) and concurrency (ADR-006), each with context, alternatives and consequences. |
| Demo & communication | 0 | 1 | **−1 (F8):** `scripts/demo.sh` covers the happy path with rematch and a trace, but not duplicates, concurrent assignment or tests. Re-score after Demo Day. |
| **Total** | **3** | **5** | |

---

## 🧹 Clean Code & Principles Audit

Challenge 3 has no separate Code Quality score. The impact column points to points already included above.

| Finding | Evidence | Challenge 3 mapping | Score impact already included |
| --- | --- | --- | --- |
| SRP ✅: one use-case class per command, with domain transition enums holding the rules and SQL guards generated from them | `OfferService`, `DispatchService`, `RideOutcomes`, `RideTransition` | Architecture → Domain boundaries | None (4/4). |
| OCP / DIP ✅: ports in `application/port`, adapters in `adapters/*`, enforced by ArchUnit | `ArchitectureRules.java` | Architecture → Structure | None. |
| KISS ⚠️: choreography with three copies of assignment state, plus outbox, Kafka, Redis and sweepers, for a 4–6 hour scope | F5 | Architecture → Structure | −1, already included. |
| YAGNI ⚠️: Redis read and replay cache and layer-coloured log output are beyond the brief, but optional and isolated (Redis failure falls back to Postgres, `CacheIT`) | `ADR-008`; `LayerColorConverter.java` | Advisory — no applicable rubric impact | None. |
| Error handling ✅: typed `ConflictException`, `NotFoundException`, `ValidationException` and `UnprocessableException` with stable codes and problem details | `ApiExceptionHandler` | Architecture → API design | None (3/3). |
| No hidden side effects ⚠️: a replay returns current state (F4) | `RideCreationService.java:80–82` | AC-01 | −1, already included. |
| Type safety ✅: records, enums and the Java 25 toolchain; Lombok limited to constructors and loggers | `lombok.config` | Advisory | None. |
| Debug artifacts ✅: logging goes through SLF4J with MDC; no `System.out` found | `src/main` | Advisory | None. |

**No separate Code Quality score or extra total deduction applies.**

---

## Strengths

- A database-level reservation (compare-and-set plus partial unique indexes plus lock ordering) that stays correct across multiple instances, not only in one process.
- A transactional outbox with consumer de-duplication in the same transaction: no dual write, and duplicates are harmless.
- An authored test for almost every acceptance bullet, including a real barrier-synchronised race and global invariants checked after every integration test.
- Real metrics and a one-command ride trace, plus a staged scalability plan with a load model.

## Key Risks

- **F1:** in a race window, a driver can go offline while holding an offer or a trip.
- **F2, F3:** after a dead-letter, copies can drift permanently, a cancelled ride can keep being offered, and a driver can be stranded `BUSY`. The README says this cannot happen.
- **F5:** every new event type adds to the choreography, which is where F1–F3 came from.
- None of the integration tests, including AC-09, have been executed by either reviewer.

## Improvements

- **Challenge correctness:**
  - **F1:** decide availability against dispatch's supply row.
  - **F4:** store and replay the original creation response.
  - **F3:** have the sweeper confirm the ride's status before re-matching.
- **Maintainability:**
  - **F2:** make projections recover from a lost predecessor, by not consuming a skipped transition or by rebuilding from dispatch state.
  - **F5:** keep dispatch as the single assignment owner and let other modules read it.
  - **F7:** correct the CI and recovery documentation.
- **Future scale / proof:**
  - **F6:** add accept-vs-timeout and accept-vs-cancel races plus a dead-letter replay test.
  - **F8:** extend the demo to cover duplicates and the concurrent race.
  - Run `make verify` on a machine with Docker and attach the output.

## Final Verdict

**Exceptional — 91/100.** The dispatch core is rigorous: reservation is enforced by SQL and schema constraints, with authored tests that map almost one-to-one to the brief. Points are lost for the availability decision made on a lagging copy (F1), dead-letter paths that can leave state permanently wrong (F2, F3), replay content (F4), complexity out of proportion to the scope (F5), and overstated documentation. This score comes from static inspection. AC-09, the Concurrency row and the Testing rows depend on an integration suite that neither review could run. Run `make verify` with Docker before the score is final.

### Comparison with the 2026-11-12 review (87/100)

Same commit and rubric. The two reviews find the same defects (F1, F2, unverified integration suite, CI claim, demo gap), but count them differently:

| Area | 2026-11-12 | This review | Reason |
| --- | ---: | ---: | --- |
| Acceptance Criteria | 32 | 33 | AC-01 −1 here for replay content (F4, applied the same way as for Jessica Nguyen); not flagged on 11-12. AC-09 4 instead of 3: the earlier review deducted for "unverified at runtime", but AC rows score required behaviour, and the mechanism is statically certain (proof is scored under Concurrency and Testing). AC-10 3 instead of 2: F2 affects the ride copy after a dead-letter, not the creation → dispatch → matching flow, and is scored once under Event semantics. |
| System Architecture | 18 | 19 | API and Event architecture full here (F1 and F2 already scored elsewhere). Architecture structure −1 here for disproportionate complexity (F5). |
| Distributed Systems | 11 | 12 | Consistency 2 instead of 1: F1 and F2 are not charged again here. Failure handling −1 here is for the new F3 (dead-lettered terminal event), a different defect from F2. |
| Testing & Validation | 13 | 14 | Integration 3 instead of 2: scored on what the authored tests prove, as for other submissions; the earlier review deducted because Docker was unavailable. Concurrency & idempotency tests 2 in both, for different reasons: here, the missing lifecycle races and dead-letter tests (F6). |
| Scalability / Observability / Docs | 5 / 5 / 3 | 5 / 5 / 3 | Same. |

The earlier review links its probe and validation files under `results/coding-challenge-3/evidence/thiennguyen-2026-11-12/`; they are not committed.
