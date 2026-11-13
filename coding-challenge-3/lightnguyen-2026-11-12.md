# Coding Challenge #3 Review: Light Nguyen

**Review Date:** 2026-11-12
**Submission Path / PR URL:** [PR #22](https://github.com/twvn-dev-community-coding-challenge/coding-challenges/pull/22), `tw-lightnguyen`, `coding-challenge-3/lightnguyen`
**Reviewed head:** `c471fa8b660f76f8b6e6d77fa5692b7fa9d77e75`
**Reviewer:** AI Code Reviewer
**Tech Stack:** .NET 10 (SDK 10.0.403), ASP.NET Core controllers, EF Core 10 + SQLite, RabbitMQ.Client 6.8.1, xUnit 2.9 (EF InMemory plus one SQLite test)

This is a fresh review using the revised Challenge 3 skill (reviewer repo `main` at `b8462c1`) and the local Ride-Hailing Dispatch Service Judge Scoring Guide. All `path:line` references are relative to [`coding-challenge-3/lightnguyen` at the reviewed head](https://github.com/tw-lightnguyen/coding-challenges/tree/c471fa8b660f76f8b6e6d77fa5692b7fa9d77e75/coding-challenge-3/lightnguyen). Earlier reports for this PR (`lightnguyen-2026-11-02.md`, `lightnguyen-2026-11-10.md`) are not evidence for this review. Every runtime claim below was re-executed on 2026-11-12.

---

## Verification Summary

| Evidence type | Scope / command | Outcome |
| --- | --- | --- |
| Submission identity | `gh pr view 22 --json ...`; paginated `pulls/22/files`; `git rev-parse HEAD` in a fresh worktree of `refs/pull/22/head` | Head `c471fa8` confirmed. 71 files, all under `coding-challenge-3/lightnguyen`. No CI checks reported. Participant source not edited. |
| Build-file safety check | Search for `Directory.Build.*`, `*.targets`, `*.props`, `NuGet.config`, `<Target>`/`<Exec>` | None present; only standard nuget.org packages. |
| Submitted driver suite | `dotnet test CodingChallenge.DriverManagement.Tests/CodingChallenge.DriverManagement.Tests.csproj` | **Passed: 26, Failed: 0, Skipped: 0** |
| Submitted ride suite | `dotnet test CodingChallenge.RideManagement.Tests/CodingChallenge.RideManagement.Tests.csproj` | **Passed: 13, Failed: 0, Skipped: 0** |
| Reviewer probes P1-P11 | A reviewer-only .NET console harness (not committed): actual services, actual consumer handlers (via reflection), real SQLite files, recording `IEventPublisher` in place of RabbitMQ; P5/P6 force overlap with an EF `SaveChangesInterceptor` barrier | All outcomes reproduced from the previous run; steps and observed output are under [Reviewer Probe Reproduction](#reviewer-probe-reproduction). Reviewer probes are **not** participant tests and earn no Testing credit. |
| Live HTTP (broker unreachable) | Both APIs via `dotnet run` in Development, RabbitMQ port pointed at a closed port; `curl`. Executed 2026-11-10 on the same head and same SDK; no code change since. | Status codes listed under F2/F3/F10. |
| Static inspection | Every `.cs`, `.csproj`, `appsettings.json`, `compose.yaml`, both `.http` files, README, `adr/0001-0008`; all `ILogger` call sites (`grep -rnE "_logger\|logger\.Log"`) | Evidence mapped below. |
| Not executed | `docker compose up --build`, RabbitMQ transport (publisher confirms, DLQ routing, redelivery), multi-instance runs, load | Broker semantics are assessed by static inspection only. Probes prove handler/database behaviour, not broker scheduling or how often a race occurs. |

**Evidence limitations:** No end-to-end run through RabbitMQ. P5/P6 show that an interleaving is possible and what state it produces; they do not measure its probability.

---

## Final Evaluation

> **Needs improvement - 54/100**

**Judging rationale:** Matching, per-ride rejection exclusion, rematching, cancellation release and the competing-rides database constraint are implemented and verified by the submitted tests and reviewer probes. However, acceptance and the offer-timeout scanner write the same offer and driver rows without a version or conditional predicate. Overlapping them leaves a ride `DRIVER_ACCEPTED` while its driver is `AVAILABLE`, and the next ride is offered to that driver (F1, P5). That is a double-assignment defect under the Judge Scoring Guide. The other main losses come from:
- a defective ride-create contract (F2);
- an availability API that bypasses the driver's own transition rules (F3);
- stale and duplicate event side effects (F4, F5);
- no scale plan, no diagrams and almost no operational logging.

**Engineering assessment:** The two-service split is clean, and the `Driver` model guards its own lifecycle. The partial unique index on active offers per driver is the right database-level arbiter, and the RabbitMQ boundary is genuinely asynchronous, with durable queues, manual ack, DLQ and publisher confirms. The weaknesses are concentrated in cross-service coordination:
- Events carry no offer or event identity, so the ride side cannot tell stale messages from current ones (F4).
- State is saved before publication with no retry (F7).
- Writes outside the reservation path are unconditional read-modify-write (F1, F3).

Tests are readable and cover matching thoroughly, but they stay inside one service with EF InMemory. The only "concurrency" test runs sequentially (F8). All of the above applies to the reviewed head in a single-process-per-service deployment. Multi-instance deployment is not claimed and is not assumed.

---

## Score Summary

| Category | Score | Max |
| --- | ---: | ---: |
| Acceptance Criteria | 23 | 35 |
| System Architecture | 12 | 20 |
| Distributed Systems & Concurrency | 6 | 15 |
| Testing & Validation | 10 | 15 |
| Scalability & Performance Design | 0 | 5 |
| Observability | 1 | 5 |
| Documentation & Communication | 2 | 5 |
| **Total** | **54** | **100** |

| Score | Interpretation |
| --- | --- |
| 90-100 | Exceptional |
| 80-89 | Strong |
| 70-79 | Good |
| 60-69 | Meets expectations |
| 50-59 | Needs improvement |
| <50 | Does not meet expectations |

---

## Critical Issues

| Issue | Severity | Evidence | Score impact |
| --- | --- | --- | --- |
| AC-09 concurrent driver assignment: **Unsafe** (acceptance vs offer timeout) | Critical | F1: `DriverService.cs:169-183` and `DriverService.cs:285-311` both load and then unconditionally save the same `Offer`/`Driver` rows; P5 ends with R1 `DRIVER_ACCEPTED`/D1 and R2 `DRIVER_ASSIGNED`/D1. The competing-rides path itself is correct (P6). | AC-09 **0/4** (exclusivity requirement broken; Judge Scoring Guide AC-09 rule); Distributed → Concurrency **0/5** (Judge Scoring Guide rule); Distributed → Consistency 1/2 (cross-record coherence). Testing → Concurrency/idempotency tests 2/3 is scored separately, on the submitted test's missing overlap (F8). |
| Ride-create contract defects | Major | F2 | AC-01 2/4; API design 1/3 |
| Availability API bypasses driver transitions | Major | F3 | AC-02 2/3; API design (shared with F2) |
| Stale `DriverMatched` re-assigns a rejected driver and wedges the ride | Major | F4 | AC-06 3/4; Event semantics 2/3; Consistency 1/2 |

---

## Business Flow and Evidence Ledger

Evidence states: **IV** = implemented and verified; **IW** = implemented with weaknesses; **DP** = documented design/future proposal; **NI** = not implemented; **NE** = not demonstrated/insufficient evidence. Verification modes: **E** execution (submitted tests, reviewer probes or HTTP), **S** static inspection, **D** documentation. IV/E does not imply exhaustive correctness.

| Flow / objective | Component and state owner; reads/writes; event/atomic boundary; invariant | Evidence state and verification mode | Evidence / gaps / finding ID / scoring rows |
| --- | --- | --- | --- |
| Create → event → match → offer | **Ride Mgmt** owns `Ride` + `IdempotencyKey`. `RideService.CreateRide` writes both in one `SaveChangesAsync`, then publishes persistent `RideCreated` (`RideService.cs:25-73`). **Driver Mgmt** (`RideCreatedConsumer.cs:23-54`) reads the active offer for the ride, else finds the nearest eligible driver, then `SendOfferToDriver` writes driver `OFFERED` + offer `PENDING` in one save, guarded by the partial unique index (`DriverService.cs:90-133`, `ApplicationDbContext.cs:27-30`). It publishes `DriverMatched`, and the ride becomes `DRIVER_ASSIGNED` (`RideService.cs:80-115`). Invariant: one active offer per driver. | IW / E+S | Flow verified by P8 and the submitted tests. Gaps: the brief's `longitude` key is lost and the passenger is unchecked (F2). Save-then-publish has no outbox; a broker failure returns 500 after the ride is saved, and the README documents client retry (F7). The ride goes from `REQUESTED` straight to `DRIVER_ASSIGNED` without persisting `MATCHING` first (advisory). Rows: AC-01, AC-04, AC-10, Event architecture. |
| Acceptance | Ride API checks `DRIVER_ASSIGNED` and `AssignedDriverId == driverId` (`RideService.cs:117-133`), then publishes `RideDriverAction(Accept)`. Driver consumer runs `AcceptOffer`: offer `ACCEPTED` + driver `BUSY` in one save (`DriverService.cs:169-183`), then publishes the completion. The ride becomes `DRIVER_ACCEPTED` (`RideService.cs:186-189`). | IW / E+S | Ownership is enforced on both sides (submitted tests `RideServiceTest.cs:155-184`, `DriverServiceTest.cs:432-504`). Gaps: no deadline check (F6, P4); no version predicate against the expiry scanner (F1, P5); a repeated API accept after projection returns 409 (safe, not idempotent). Rows: AC-05, AC-08, AC-09. |
| Rejection → rematch | Driver consumer `RejectOffer` writes offer `REJECTED` + driver `AVAILABLE` (`DriverService.cs:185-205`). The ride completion handler clears the assignment, sets `MATCHING`, saves, then republishes `RideCreated` (`RideService.cs:190-200`). Matching excludes `REJECTED`/`TIMED_OUT` drivers for that ride (`DriverMatchingService.cs:35-43`). Invariant: a rejected driver is never re-offered the same ride. | IW / E+S | Exclusion and the unmatched state are verified (`DriverServiceTest.cs:506-554`, `RideServiceTest.cs:186-213,295-316`). Gap: a stale `DriverMatched` can still re-assign the rejected driver on the ride side (F4, P9). Rows: AC-05, AC-07, AC-06. |
| Timeout → rematch | `OfferExpirationService` ticks every 1 s and calls `ExpirePendingOffers`, which writes `TIMED_OUT` + driver `AVAILABLE` for `PENDING` offers older than the configured timeout (default 30 s). It then publishes a `Reject` completion per offer (`OfferExpirationService.cs:18-36`, `DriverService.cs:285-311`), and the ride follows the rejection path. | IW / E+S | The expiry write is verified (`DriverServiceTest.cs:556-585`). Gaps: no conditional update, so it overwrites a concurrent acceptance (F1, P5); acceptance does not check the deadline itself (F6, P4); a publish failure after the save loses the ride update (F7). Rows: AC-05, AC-06, AC-09, Concurrency, Consistency. |
| Cancellation → release | `CancelRide` accepts `REQUESTED`/`MATCHING`/`DRIVER_ASSIGNED`/`DRIVER_ACCEPTED` (`RideService.cs:135-161`). With an assigned driver it publishes `Cancel`: the driver side releases `OFFERED` or `BUSY` (`DriverService.cs:207-232`) and the ride becomes `CANCELLED` with the assignment cleared (`RideService.cs:201-206`). Without one, it saves `CANCELLED` and sends `RideCreated(IsCancelled)` to cancel any in-flight pending offer (`RideCreatedConsumer.cs:33-37`, `DriverService.cs:262-283`). | IV / E+S | P11 (accepted ride → driver `AVAILABLE`); submitted tests `DriverServiceTest.cs:587-652`, `RideServiceTest.cs:215-256`. Advisory: cancellation is carried by an overloaded `RideCreated` event (F10). Rows: AC-05, AC-06. |
| Completion → release | `CompleteRide` requires `DRIVER_ACCEPTED` (`RideService.cs:163-174`). The driver side writes offer `COMPLETED` + driver `AVAILABLE` (`DriverService.cs:234-260`), and the ride becomes `COMPLETED`. | IV / E+S | P8 full flow; `RideServiceTest.cs:258-293`, `DriverServiceTest.cs:587-621`. Rows: AC-02, AC-06. |
| Duplicate request/event | Creation: key PK + SHA-256 request hash + stored response; a PK conflict reloads the winner (`RideService.cs:27-69`). Acceptance replay returns the accepted offer (`DriverService.cs:171-175`). Ride completion handlers act only from their expected source state (`RideService.cs:184-211`). A duplicate `RideCreated` reuses an active offer (`RideCreatedConsumer.cs:39-42`). No event IDs are stored (ADR-0007). | IW / E+S | Positives: `RideServiceTest.cs:62-125,373-401`, `DriverServiceTest.cs:481-504`, P10. Gaps: replaying the create after completion republishes and reserves the driver again (F5, P8); a stale `DriverMatched` is not idempotent (F4, P9); key reuse is rejected but surfaces as 500 (F2). Rows: AC-08, Idempotency, Event semantics. |
| Two rides compete for one driver | The reservation point is `SendOfferToDriver`'s single `SaveChangesAsync`. The SQLite partial unique index `Offers(DriverId) WHERE Status IN (PENDING, ACCEPTED)` arbitrates; error 2067 maps to `false` (`DriverService.cs:117-132`), and the consumer publishes `DriverMatched(null)`, so the losing ride stays `MATCHING`. Scope: the DB file per Driver Mgmt deployment, so it is valid across processes sharing that DB. | IW / E+S | P6: both workers passed the save barrier, exactly one `PENDING` offer, loser `MATCHING`/null. Gaps: the submitted test does not overlap (F8, P7); exclusivity is lost through the expiry path (F1, P5). Rows: AC-09, Concurrency, Testing. |

## Reviewer Probe Reproduction

Each probe calls the submission's own services and consumer handlers against a fresh SQLite file, with RabbitMQ replaced by an in-memory recording publisher. "Overlap" means both workers load their entities before either calls `SaveChangesAsync`, enforced by a two-party barrier in an EF `SaveChangesInterceptor`.

| Probe | Steps | Observed outcome |
| --- | --- | --- |
| P1 | Deserialize the brief's create payload (`longitude` key) with ASP.NET web JSON defaults; then build a `Coordinate` with lat=100 | `pickup.Longtitude=0`, `destination.Longtitude=0`; lat=100 throws `ArgumentOutOfRangeException` (HTTP 500) |
| P2 | `CreateRide` with `PassengerId=Guid.Empty` and an empty Passengers table | Ride created, `REQUESTED` |
| P3 | Driver with a `PENDING` offer → `UpdateDriverAvailability(false)`; then accept. Separately: accepted (`BUSY`) driver → availability false, then true | `OFFERED`→`OFFLINE`, then accept fails "Driver must have an offer to accept"; `BUSY`→`OFFLINE`→`AVAILABLE` while the offer is still `ACCEPTED` |
| P4 | Offer created 120 s ago (timeout 30 s); `AcceptOffer` before the expiry pass | Offer `ACCEPTED`, driver `BUSY` |
| **P5** | D1 has a `PENDING` offer for R1 older than the timeout. Overlap `AcceptOffer(R1,D1)` with `ExpirePendingOffers()`. Feed both completion events (Accept, then Reject) to the ride service. Then dispatch a second ride R2 through the real `RideCreated` handler | Driver side: D1 `AVAILABLE`, R1 offer `TIMED_OUT`. Ride side: R1 `DRIVER_ACCEPTED`/D1 (the Reject is ignored because R1 is no longer `DRIVER_ASSIGNED`). R2 dispatch: D1 `OFFERED` with a new `PENDING` offer, R2 `DRIVER_ASSIGNED`/D1. **D1 is held by two rides (F1).** |
| P6 | Two rides, one eligible driver; overlap both `RideCreated` handlers | Both passed the barrier; exactly one `PENDING` offer; loser `MATCHING`, unassigned |
| P7 | Reproduce the submitted test's `Task.WhenAll(service1.SendOfferToDriver(...), service2.SendOfferToDriver(...))` | First task already completed before the second is invoked; results `True, False` (no overlap, F8) |
| **P8** | Run R1 through create → match → accept → complete on both sides, then repeat the original create request with the same idempotency key and feed the published event to the real `RideCreated` handler | Replay republishes `RideCreated`; R1 stays `COMPLETED` while D1 becomes `OFFERED` with a new `PENDING` offer for R1 (F5) |
| **P9** | `DriverMatched(R1,D1)` → D1 rejects → R1 `MATCHING`; redeliver the old `DriverMatched(R1,D1)`, then deliver the real rematch `DriverMatched(R1,D2)` | Stale event re-assigns R1 to D1 (`DRIVER_ASSIGNED`); the real D2 match is ignored and R1 stays on D1 (F4) |
| P10 | Duplicate `RideCreated` while an offer is pending | Both deliveries report D1; D1 has one `PENDING` offer; D2 untouched |
| P11 | Driver-side Cancel on an `ACCEPTED` offer | Driver `AVAILABLE`, offer `CANCELLED` |

## Key Technical Findings

| ID | Severity | Problem and evidence | Impact | Recommendation | Official row(s), distinct score impact or advisory |
| --- | --- | --- | --- | --- | --- |
| F1 | Critical | Acceptance (`DriverService.cs:169-183`) and expiry (`DriverService.cs:285-311`) each read `Offer`/`Driver`, mutate in memory and save without a concurrency token or `WHERE Status = PENDING`. P5: both load, accept saves `ACCEPTED`/`BUSY`, expiry then saves `TIMED_OUT`/`AVAILABLE`. Both publish completions, and in Accept-then-Reject order the ride stays `DRIVER_ACCEPTED`/D1. The overwritten offer leaves the partial index, so R2 is offered to D1. | A driver holds an accepted ride and a new offer at the same time. The window is small but recurs at every 30 s boundary, and every service instance runs its own 1 s scanner. | Add a row-version concurrency token to `Offer`/`Driver`, or use a conditional `UPDATE ... WHERE Id=@id AND Status=0` on both paths and act only when one row changed. Check the deadline inside `AcceptOffer`. | AC-09 (exclusivity) 0/4; Concurrency 0/5 (guide rule); Consistency 1/2 (cross-record coherence). Separate objectives; the defect is the same. |
| F2 | High | `Coordinate` exposes `Longtitude` (`Coordinate.cs:6,27`), so P1 turns the brief's `"longitude"` into `0`. No passenger lookup (`RideService.cs:36-42`); P2 creates a ride with `PassengerId=Guid.Empty` and an empty `Passengers` table. HTTP: `latitude:100` → **500** `ArgumentOutOfRangeException`; same key with a different body → **500** `InvalidOperationException` (`RidesController.cs:26-37` catches nothing). | Rides created through the documented contract are matched at the wrong location. Invalid input is reported as a server fault. | Rename to `Longitude` (or `[JsonPropertyName("longitude")]`); validate the passenger against seeded rows; map validation to 400 and key reuse to 409/422. | AC-01 2/4; API design 1/3. |
| F3 | High | `UpdateDriverAvailability` assigns `driver.Status` directly (`DriverService.cs:72`), bypassing `Driver.GoOffline` (`Driver.cs:13-21`). P3: `OFFERED` → `OFFLINE` with the offer still `PENDING`, after which accept throws and the message is dead-lettered. `BUSY` → `OFFLINE` → `AVAILABLE` also succeeds. Unknown driver → **500** `System.Exception` (`DriverService.cs:69`). | Breaks the explicit availability rules and strands offers. The index still blocks a second active offer for the driver. | Route through guarded `GoOffline`/`GoOnline`; return 409 for `OFFERED`/`BUSY` and 404 for unknown drivers. | AC-02 2/3; API design (shared with F2). |
| F4 | High | `HandleDriverMatched` accepts any match while the ride is `REQUESTED`/`MATCHING` and unassigned (`RideService.cs:80-115`); events carry no offer or attempt identity (`DriverMatchedEvent.cs:3`). P9: after D1's rejection, a redelivered `DriverMatched(R1,D1)` sets `DRIVER_ASSIGNED`/D1, and the genuine rematch `DriverMatched(R1,D2)` is ignored. | The ride is wedged on a driver whose offer is `REJECTED`. D2 is `OFFERED` for a ride that does not reference them. Expiry only targets `PENDING` offers, so nothing recovers it. The README claim at `README.md:105-106` is inaccurate. | Include `offerId` in `DriverMatched`/completions and accept only the current offer; or have the ride side query/compare an offer version. | AC-06 3/4 (duplicate transition); AC-08 2/4 (with F5); Event semantics 2/3 (ordering); Consistency 1/2 (shared with F1, distinct path). |
| F5 | Medium | Replaying the original create republishes `RideCreated` unless the ride is `CANCELLED` (`RideService.cs:221-227`); the consumer creates a new offer (`RideCreatedConsumer.cs:39-48`). P8: a `COMPLETED` ride gains a `PENDING` offer, and its driver becomes `OFFERED`. | A repeated logical side effect: the driver is reserved for about 30 s for a finished ride. | Republish only while the ride is `REQUESTED`/`MATCHING`; the consumer should also check a terminal marker. | AC-08 2/4 (with F4); AC-10 2/3 (duplicate delivery not harmless); Idempotency 2/3. |
| F6 | Medium | Acceptance never compares `CreatedDateTime` with the timeout (`DriverService.cs:169-183`). P4: an offer created 120 s earlier is accepted. | Late acceptance is honoured until the next scan; combined with F1 it creates the race. | Enforce the deadline in `AcceptOffer`, using the same predicate as expiry. | AC-05 3/4. |
| F7 | Medium (static risk) | Every handler saves and then publishes: `RideService.cs:56-71,193-199`, `RideDriverActionConsumer.cs:30-51`, `OfferExpirationService.cs:25-34`. A publish failure after the save dead-letters the inbound message with no retry (`RabbitMqConsumer.cs:91-95`). | Driver and ride state can diverge permanently after a broker blip. Observed only by inspection. | Transactional outbox, or a retry with backoff before dead-lettering (only retry/backoff is listed as future work, `README.md:146-153`; an outbox is not mentioned). | Failure handling 1/2. |
| F8 | Medium | The submitted "race" test (`DriverServiceTest.cs:349-400`) passes two already-started async calls to `Task.WhenAll`. Microsoft.Data.Sqlite completes synchronously, so P7 shows the first task complete before the second is invoked. | The test proves stale-context constraint handling, not overlapping dispatch, and the acceptance/timeout race is untested. | Use `Task.Run` plus a barrier (or an interceptor) to overlap saves; add an accept-vs-expiry test. | Concurrency/idempotency tests 2/3; test quality 1/2 (overclaiming name). |
| F9 | Medium | Only four log statements exist: validation warnings (`Program.cs:35` in each API) and consumer failures (`RabbitMqConsumer.cs:42,93`). No `rideId`/`driverId`/`eventId`, no metrics, no diagrams, no scale section (search scope: `README.md`, `adr/`, all `.cs`). | A ride's history cannot be reconstructed from logs; the required design artifacts are absent. | Structured lifecycle logs, an `eventId` on every event, counters and latency metrics, a `docs/` folder with diagrams and a 1K/100K/1M plan. | Observability 1/5; Scalability 0/5; Evolution 0/2; Diagrams 0/1. |
| F10 | Low | Inconsistent roots (`/Rides` vs `/api/drivers`); README `POST /api/Driver` → **404** (route is `/api/Drivers`); `DriverManagement.http` is the `weatherforecast` template; `RideCreatedEvent.IsCancelled` overloads one event as create/rematch/cancel (`RideCreatedEvent.cs:7`); generic `throw new Exception` (`DriverService.cs:69,104,140,157`). | Harder to consume, document and evolve the contracts. | Consistent `/api/rides`/`/api/drivers`; separate `RideCancelled`/`RematchRequested` events; domain exceptions mapped to status codes. | API design 1/3; Event architecture 2/3; README 1/2; Demo 0/1. |
| F11 | Low (future scale) | Matching loads every `AVAILABLE` located driver and computes Haversine in memory per request (`DriverMatchingService.cs:45-63`); each publish opens a new RabbitMQ connection (`RabbitMqEventPublisher.cs:25-34`). | Linear cost per dispatch and connection churn. Fine at 1K drivers; not addressed beyond that. | Bounding-box/geohash prefilter or a geo index; a shared long-lived connection or channel pool. | Advisory; scale rows are already 0 for absent design reasoning. |

---

## Acceptance Criteria - 23/35

| AC | Result | Score | Evidence and lost-points rationale |
| --- | --- | ---: | --- |
| AC-01 Create Ride | Partial | 2/4 | **IW / E+S.** Required: passenger, bounds, unique id, `REQUESTED`, retrieval, replay, conflict. Met: v7 GUID (`BaseModel.cs:8`), `REQUESTED` (`RideService.cs:41`), retrieval and replay (submitted tests `RideServiceTest.cs:29-153`), conflict refused. Gaps: passenger never validated, `longitude` key dropped (F2, P1/P2), bounds and conflict surface as 500. **Anchor:** fits 2 ("happy path; major validation gap") rather than 3, because two required validations (passenger, coordinate contract) fail rather than one edge. |
| AC-02 Driver Availability | Partial | 2/3 | **IW / E+S.** Met: guarded `AVAILABLE`→`OFFERED`→`BUSY` and all three releases (`Driver.cs:23-71`; `DriverServiceTest.cs:316-347,432-479,587-621`; P11). Gap: the availability API permits `OFFERED`/`BUSY`→`OFFLINE` and `BUSY`→`AVAILABLE` (F3, P3). **Anchor:** 2 ("main lifecycle works with material transition gap"), not 3, because invalid availability changes are not rejected. |
| AC-03 Driver Location | Demonstrated | 2/2 | **IV / E+S.** `[Range]` gives 400 (`UpdateDriverLocationRequest.cs:7-11`, HTTP); domain guard `Coordinate.cs:19,32`; latest value persisted (`DriverServiceTest.cs:228-263`); matching reads the stored coordinate (`DriverMatchingService.cs:45-57`). |
| AC-04 Driver Matching | Demonstrated | 4/4 | **IV / E+S.** `AVAILABLE`-only located drivers, per-ride exclusions, Haversine at 6,371 km, default 2 km configurable (`DriverMatchingOptions.cs:5`, `appsettings.json:17-19`), ordinal tie-break on the canonical `driverId` string (`DriverMatchingService.cs:61`), null when no candidate. All exercised by submitted tests (`DriverServiceTest.cs:48-183`). |
| AC-05 Ride Offer | Partial | 3/4 | **IW / E+S.** Met: pending offer → `DRIVER_ASSIGNED`/`assignedDriverId`/`OFFERED`; ownership on both sides; accept → `BUSY`/`DRIVER_ACCEPTED`; reject → `MATCHING`/cleared/`AVAILABLE`; permanent exclusion; configurable 30 s timeout as rejection; cancellation releases `OFFERED` and `BUSY` (P11). Gap: acceptance ignores the deadline (F6, P4). **Anchor:** 3 ("one significant edge missing"). The F1 race is scored under AC-09, not here. |
| AC-06 Ride State Machine | Partial | 3/4 | **IW / E+S.** Met: source-state-guarded transitions (`RideService.cs:184-211`), all four cancellations, terminal states reject commands (409) and ignore events (`RideServiceTest.cs:318-371`), clearing persisted (`:186-213,295-316`), duplicate completions no-op (`:373-401`). Gap: a duplicate/stale `DriverMatched` produces an invalid re-assignment (F4, P9). **Anchor:** 3, not 4, because duplicate transitions are not uniformly safe. |
| AC-07 Rematching | Demonstrated | 3/3 | **IV / E+S.** Rejection and timeout republish `RideCreated` (`RideService.cs:190-200`, `OfferExpirationService.cs:28-33`); permanent per-ride exclusion that does not leak to other rides, and the no-candidate `MATCHING`/null state (`DriverServiceTest.cs:506-554`, `RideServiceTest.cs:295-316`). F4 concerns stale event ordering, scored in AC-06. |
| AC-08 Idempotency | Partial | 2/4 | **IW / E+S.** Met: creation key with hash conflict detection, acceptance replay, guarded transitions, duplicate `RideCreated` while an offer is active (P10); documented in ADR-0007. Gaps: the brief requires duplicate events "without invalid side effects", and two duplicate-event paths have them: a late `RideCreated` re-reserves a driver for a completed ride (F5, P8), and a redelivered `DriverMatched` re-assigns a rejected driver (F4, P9). **Anchor:** 2 ("some protection; material repeat effects remain"), not 3, because the duplicate-event path fails in two independent places rather than one edge. |
| AC-09 Concurrent Driver Assignment | Failed | 0/4 | **IW / E+S.** Met: race identified (ADR-0008); database-enforced reservation; loser stays `MATCHING`; verified with a real overlap by a reviewer probe (P6, which earns AC evidence but no Testing credit). Failed: "the same driver cannot hold two simultaneous offers or accepted rides" is demonstrably broken via the expiry path (F1, P5). **Score:** 0. The brief's Judge Scoring Guide requires a significant AC-09 penalty when "the same driver [can] be assigned to two active rides concurrently", and the calibration scores a broken exclusivity requirement 0/4. P5 ends with R1 `DRIVER_ACCEPTED`/D1 and R2 `DRIVER_ASSIGNED`/D1, so the requirement is broken. The correct competing-rides reservation (P6) does not offset this: the requirement is driver exclusivity, whichever write path breaks it. |
| AC-10 Event-Driven Dispatch | Partial | 2/3 | **IW / E+S+D.** A real async boundary (persistent publish with confirms `RabbitMqEventPublisher.cs:36-44` → durable queue → `RideCreatedConsumer` → matching → `DriverMatched`). README explains why async, duplicate delivery and failure (nack → `.dead`, no retry) (`README.md:97-109`). Gap: duplicate `RideCreated` delivery is not harmless. The consumer dedupes only against an active offer (`RideCreatedConsumer.cs:39-42`) and never checks ride state, so a late redelivery for a completed ride creates a new offer (F5, P8 fed the event to the real handler). The calibration's "duplicate event is not harmless" case deducts from both AC-08 and AC-10. **Anchor:** 2 ("async dispatch works but duplicate behavior materially weak"). Broker transport is static only. |
| **Total** |  | **23/35** | |

---

## Detailed Scoring

### System Architecture - 12/20

Ride Management owns ride lifecycle and creation idempotency; Driver Management owns drivers, matching, offers and timeouts (dispatch lives there). Each has its own SQLite file, and they communicate only through four routing keys on one topic exchange (`EventRoutingKeys.cs`). Dependency direction is clean: both services depend on `CodingChallenge.Share` for messaging and base types, and nothing depends on a service. Change cost:
- **Matching:** swapping in a geo index touches only `DriverMatchingService`.
- **Storage:** the Postgres move is mostly provider configuration, but the 2067 error-code check (`DriverService.cs:123-127`) and the raw SQL index (`Program.cs:68-69`) are SQLite-specific.
- **Workers:** they are already separate hosted services.

The main coupling is the shared `Coordinate` used as HTTP DTO, EF owned type and domain value at once, which is why the `Longtitude` typo leaks across every boundary (F2). `Driver` encapsulates its transitions; `Ride` is mutated directly by the service, and `Passenger` is unused.

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Domain boundaries & responsibilities | 3 | 4 | Ride/Driver/Dispatch ownership is clear and deliberate; meaningful coupling remains (anemic `Ride`, unused `Passenger`, shared `Coordinate`). Anchor 3: generally sound with meaningful coupling. |
| Architecture structure | 3 | 4 | Two services + shared lib + hosted consumers; proportionate. The material coupling is one type spanning wire, persistence and domain. |
| API design | 1 | 3 | Major consistency and validation gaps: inconsistent roots, 500s for validation, key reuse and unknown drivers, misspelled contract, raw `true` body on an unconstrained `{driverId}` route (`DriversController.cs:45-46`) (F2, F3, F10). Not 2, because the main create endpoint itself is not coherent with the brief's contract. |
| Event-driven architecture | 2 | 3 | Real async flow, clear producer/consumer ownership, durable queues with DLQ. Material contract gap: no event/offer identity, and `RideCreated` overloaded for rematch and cancel (F4, F10). |
| Data architecture | 1 | 2 | The partial unique index is well chosen. Significant consistency gap: no offer version across services, no `Offers.RideId` index, no one-active-offer-per-ride guarantee. |
| Architecture trade-offs | 2 | 2 | Choices, alternatives and consequences are explained: async vs synchronous request path (clients must poll), SQLite vs an installed server (ADR-0003, production reconsideration), Haversine vs a geospatial DB/ellipsoidal geodesic (ADR-0006), constraint as final arbiter with explicit startup-failure consequences (ADR-0008). |
| Architecture evolution | 0 | 2 | No 1K → 100K → 1M discussion (search scope: README, `adr/`; only "SQLite may need reconsideration"). Not 1: the README's "With More Time" list (retry, end-to-end tests, tracing/metrics) is not a scale path, and "SQLite may need reconsideration" names no stage, load or boundary change. |
| **Total** | **12** | **20** | |

### Distributed Systems & Concurrency - 6/15

The reservation is a single EF save of offer + driver under a database uniqueness constraint. That is atomic for its scope and valid across processes sharing the DB file (P6). Every other state change is an unguarded read-modify-write: acceptance, rejection, expiry, cancellation and the availability API. The acceptance/expiry pair is a demonstrated defect (F1). The availability API weakens transitions but cannot create a second active offer (F3). Delivery is at-least-once with manual ack, prefetch 1, persistent messages, publisher confirms and per-queue DLQ without retry (`RabbitMqConsumer.cs:59-100`, `RabbitMqTopology.cs:24-44`; static inspection). Ordering is not handled: events carry no identity, so a stale event can win (F4). Persistence and publication are separate steps (F7, a static risk); the README documents this as future work, which earns no implementation credit.

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Concurrency / race conditions | 0 | 5 | **Double assignment is possible (F1, P5) → 0/5 per the Judge Scoring Guide**, regardless of the correct competing-rides constraint (P6). |
| Idempotency | 2 | 3 | Actual protection for creation, acceptance, transitions and active-ride duplicate events; a material replay side effect remains (F5). |
| Event processing semantics | 2 | 3 | At-least-once, ack/nack, DLQ and confirms are coherent and documented. Material omission: ordering/staleness (F4); the README's stale-event claim is wrong. |
| Failure handling | 1 | 2 | Processing failures are dead-lettered and create retries republish, but the save-then-publish gap and no retry leave unrecovered divergence (F7). |
| Consistency | 1 | 2 | Most transitions are coherent (cancellation, completion, rejection), but ride/driver/offer views diverge in P5, P8 and P9. Anchor 1: some coherent transitions with material inconsistency. |
| **Total** | **6** | **15** | |

### Testing & Validation - 10/15

Only the 39 participant-authored tests are scored here; reviewer probes P1-P11 are excluded. The tests use AAA-style arrangements with explicit behavioural assertions and cover matching thoroughly. The ride suite isolates each test with a GUID-named InMemory database, but its create tests use `RandomFloat` coordinates (`RideServiceTest.cs:38-39,413-416`). Several driver tests share one named InMemory database `"DriverTest"` (`DriverServiceTest.cs:17-19`); they avoid collisions only through fresh GUIDs. The only real-provider test is the SQLite "race" test, and it does not overlap (F8, P7). There are no controller/HTTP, consumer-handler or broker tests, and none cross Ride ↔ Driver. Missing scenarios: invalid availability transitions, late acceptance, acceptance vs timeout, stale events, unknown passenger, invalid ride coordinates.

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Unit tests | 3 | 4 | Broad: matching (nearest, offline/busy, radius, configured radius, tie, missing location), driver status transitions, coordinate validation, rematch exclusion, ride transitions. Meaningful missing area: availability transition rules and the timeout/acceptance edge. |
| Integration tests | 1 | 3 | One narrow real-SQLite interaction (`DriverServiceTest.cs:349-400`); everything else uses EF InMemory and a fake publisher. |
| Edge-case testing | 3 | 3 | All five required edges are meaningfully asserted: no driver (`:132-154`), out of radius (`:48-95,156-183`), rejection (`:506-554`, `RideServiceTest.cs:186-213`), invalid/terminal transitions (`DriverServiceTest.cs:402-430`, `RideServiceTest.cs:318-371`), duplicate request/event (`RideServiceTest.cs:62-125,373-401`, `DriverServiceTest.cs:481-504`). |
| Concurrency & idempotency tests | 2 | 3 | Duplicate logical-result proof is strong (creation ×3 → one ride; repeated acceptance; duplicate completions). The overlapping two-ride proof is absent (F8). Anchor 2: one of the two is strong. |
| Test quality | 1 | 2 | Readable and behaviour-focused, but nondeterministic inputs, a shared named database and an overclaiming race-test name. |
| **Total** | **10** | **15** | |

### Scalability & Performance Design - 0/5

Current cost: every dispatch loads all `AVAILABLE` located drivers and ranks them in memory, which is O(n) per match. Location updates are single-row writes to SQLite, and each publish opens a broker connection (F11). The submission does not discuss any of this. There are no bottlenecks, strategies, trade-offs or 1K/100K/1M evolution in the README or ADRs (search scope: `README.md`, `adr/*.md`; terms scal/bottleneck/partition/cache).

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Bottleneck identification | 0 | 2 | None identified. |
| Scaling strategy | 0 | 2 | None described. |
| Performance considerations | 0 | 1 | No workload-linked trade-offs. |
| **Total** | **0** | **5** | |

### Observability - 1/5

Logs cannot answer "what happened to ride R001". No lifecycle step is logged, no identifier is attached to any log, and events have no `eventId`. The only operational trail is database rows and dead-letter queues (F9).

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| Structured logging | 1 | 2 | Structured templates exist, but only for validation failures and consumer failures (`{QueueName}`, `{Path}`, `{@Errors}`); significant gaps. |
| Correlation / tracing | 0 | 1 | No `rideId`/`driverId`/`eventId` in any log or event envelope. |
| Metrics | 0 | 1 | None implemented. |
| Troubleshooting capability | 0 | 1 | A ride's path cannot be reconstructed from logs. |
| **Total** | **1** | **5** | |

### Documentation & Communication - 2/5

The README is accurate about async dispatch, at-least-once delivery, DLQ behaviour and the lack of retry. It is inaccurate in two places: it claims stale matching events are ignored (contradicted by P9), and its driver-create step returns 404. It documents the `longtitude` typo instead of fixing it. Architecture, matching and concurrency strategies are delegated to ADRs, and there is no scalability section or `docs/` folder. The ADRs are concise and match the code (0006 Haversine, 0007 idempotency, 0008 the unique index, including its startup migration at `Program.cs:68-69`).

| Subcriterion | Score | Max | Earned/lost evidence |
| --- | ---: | ---: | --- |
| README | 1 | 2 | Run/test, approach, duplicates, failure and trade-offs present; omissions (architecture, scale, limitations list) and inaccuracies (stale-event claim, 404 route). |
| Architecture diagrams | 0 | 1 | None in the submission tree. |
| ADRs / design decisions | 1 | 1 | Eight ADRs, including meaningful matching (0006), idempotency (0007) and concurrency (0008) decisions. |
| Demo & communication | 0 | 1 | A textual HTTP walkthrough covers create/match/accept/reject only; no duplicate, concurrency or test-run evidence. `DriverManagement.http` is the `weatherforecast` template. |
| **Total** | **2** | **5** | |

---

## 🧹 Clean Code & Principles Audit

### SOLID Compliance

| Principle | Status | Evidence | Deduction (raw, shared skill) |
| --- | --- | --- | --- |
| SRP | ⚠️ | `RideService` mixes creation/idempotency, event projection and command publishing (`RideService.cs:25-238`); `DriverService` mixes CRUD, offers, lifecycle and expiry. Proportionate to the timebox. | -1 |
| OCP | ✅ | Radius and timeout via options; consumers share `RabbitMqConsumer<T>`. | 0 |
| LSP | ✅ | Consumer subclasses honour the template contract. | 0 |
| ISP | ⚠️ | `IDriverService` carries duplicate accept/reject overloads by offer ID that only tests use (`IDriverService.cs:13-16`). | -1 |
| DIP | ✅ | Constructor-injected `IEventPublisher`/`IDriverService`. | 0 |

### DRY / KISS / YAGNI

| Principle | Violation | Location | Fix / Alternative | Deduction |
| --- | --- | --- | --- | --- |
| DRY | Identical `InvalidModelStateResponseFactory` blocks; `ConnectionFactory` built twice | Both `Program.cs:16-41`; `RabbitMqEventPublisher.cs:25-31`, `RabbitMqConsumer.cs:61-69` | Shared extension / factory | -1 |
| KISS | `RideCreatedEvent` doubles as rematch and cancel command | `RideCreatedEvent.cs:7`, `RideService.cs:154-160`, `RideCreatedConsumer.cs:33-37` | Separate event types | -1 |
| YAGNI | Unused `Passenger`, always-zero `Offer.Price`, unused `IsDeleted`, template `.http` | `Passenger.cs`, `RideCreatedConsumer.cs:44`, `BaseModel.cs:9`, `DriverManagement.http` | Remove, or use `Passenger` for validation | -1 |

### Additional Clean Code

| Check | Status | Evidence | Deduction |
| --- | --- | --- | --- |
| Meaningful names | ⚠️ | `Longtitude` propagates into the API contract | -1 |
| Small functions | ✅ | Most methods under 25 lines; `HandleDriverActionCompleted` ~36 lines but flat | 0 |
| No side effects | ⚠️ | Create replay republishes on every repeat (`RideService.cs:224-227`) | -1 |
| Error handling | ❌ | `throw new Exception("Driver not found.")` (`DriverService.cs:69,104,140,157`); domain exceptions unmapped → 500 | -1 |
| Consistent style | ✅ | Consistent formatting | 0 |
| No debug artifacts | ⚠️ | Redundant dead branches (`RideService.cs:89-102`); template `weatherforecast` request | -1 |
| Immutability | ✅ | No input mutation found | 0 |
| Type safety | ✅ | Nullable enabled | 0 |

**Raw shared-skill deductions:** -10 (informational; not subtracted from the Challenge 3 total).

| Finding | Evidence | Challenge 3 mapping | Score impact already included |
| --- | --- | --- | --- |
| Generic/unmapped exceptions → 500 | F2, F3, F10 | System Architecture → API design | API design 1/3 |
| `Longtitude` leaks into the wire contract | F2 | AC-01; Architecture structure | AC-01 2/4; structure 3/4 |
| Overloaded `RideCreatedEvent` | F10 | Event-driven architecture | 2/3 |
| Direct status assignment bypassing guards | F3 | AC-02 | 2/3 |
| Replay side effect | F5 | AC-08; Idempotency | 3/4; 2/3 |
| Nondeterministic inputs, shared InMemory DB | Testing narrative | Testing → Test quality | 1/2 |
| SRP breadth, ISP overloads, DRY, YAGNI leftovers, dead branches | Above | Advisory - no applicable rubric impact | None |

**No separate Code Quality score or extra total deduction applies.**

---

## Strengths

- Correct, well-tested matching, including configurable radius, missing-location handling and lexicographic `driverId` tie-breaking (`DriverMatchingService.cs:31-64`, `DriverServiceTest.cs:48-183`).
- A database-enforced reservation for competing rides that held under a genuinely overlapped probe, with the losing ride returned to `MATCHING` (P6).
- Permanent per-ride exclusion of rejected and timed-out drivers; cancellation releases both `OFFERED` and `BUSY` drivers (P11).
- A genuine asynchronous RabbitMQ boundary with durable queues, publisher confirms, manual ack and DLQ.
- Creation idempotency stores a request hash and response atomically with the ride and resolves concurrent creates by primary-key conflict.
- Eight short ADRs that match the implementation.

## Key Risks

- Double assignment through the acceptance/expiry overlap (F1). It recurs at every timeout boundary, and each service instance runs its own scanner.
- Rides created with the standard `longitude` key are matched at longitude 0 (F2).
- Stale or redelivered events wedge rides with no recovery path (F4).
- The save-then-publish gap without retry can desynchronise services after a broker blip (F7; static).
- O(n) in-memory matching and per-publish connections limit throughput (F11).

## Improvements

- **Challenge correctness:**
  - F1/F6: make acceptance and expiry conditional (`WHERE Status = PENDING` or a row version) and enforce the deadline in `AcceptOffer`.
  - F2/F3: fix the `longitude` contract, validate passengers, route availability through guarded domain methods and map errors to 4xx.
  - F4/F5: add `offerId`/`eventId` to events and accept only the current offer; stop republishing `RideCreated` for non-`REQUESTED` rides.
- **Maintainability:**
  - F8: add a truly overlapping two-ride test, an accept-vs-timeout test and at least one API → consumer integration test.
  - F10: split `RideCreatedEvent` into explicit cancel and rematch events; unify routes; remove template and dead code.
  - F9: add structured lifecycle logs keyed by `rideId`/`driverId`/`eventId`.
- **Future scale:**
  - F7/F11/F9: transactional outbox plus retry with backoff; a geohash prefilter or geo index for candidate search; a pooled broker connection; a documented 1K → 100K → 1M plan covering location-write load, partitioned dispatch and broker throughput, with diagrams.

## Final Verdict

**Needs improvement - 54/100.** The core dispatch path works and is backed by passing tests and probes: matching, rematch exclusion, cancellation and completion release, and a correct database arbiter for competing rides. The decisive defect is F1. Acceptance and offer expiry overwrite each other, so a driver can hold an accepted ride and a new offer at the same time (P5). The Judge Scoring Guide therefore sets Distributed → Concurrency to 0/5 and AC-09 to 0/4. The rest of the score reflects the create-contract and availability defects, stale/duplicate event side effects, tests confined to single services, and the absence of a scale plan, diagrams and operational logging.
