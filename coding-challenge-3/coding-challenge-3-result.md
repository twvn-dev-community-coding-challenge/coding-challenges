# Leaderboard: Coding Challenge #3 — All 6 Submissions Reviewed

## Summary Table

The total scores include the submission review, presentation, questions and answers. The additional points below are the difference between the supplied total and the submission review score; an individual allocation across presentation, questions and answers was not supplied.

| Rank | Team | Submission review | Additional points | Total score |
|------|------|------------------:|------------------:|------------:|
| 1 | KVP — Khoa Phan, Victor Nguyen, Duy Phuong Nguyen (`kvp`) | 92 | 22 | 114 |
| 2 | Grace Phan (`grace-phan`) | 98 | 14 | 112 |
| 3 | Thien Nguyen (`thiennguyen`) | 91 | 10 | 101 |
| 4 | Jessica Nguyen (`jessicanguyen`) | 86 | 12 | 98 |
| 5 | Duy Lu (`jameslu`) | 78 | 0 | 78 |
| 6 | Light Nguyen (`lightnguyen`) | 54 | 12 | 66 |

Duy Lu is matched here to the `jameslu` review based on the supplied score; that review names James Lu and specifies 78 for the deadline revision; its reviewed later head scores 79. Thien's review score is provisional pending execution of the integration suite. The supplied totals are shown as given.

---

## Submission Review Score Breakdown by Category

These categories total 100 points and do not include the additional presentation, question and answer points.

| Category (Max) | Grace | KVP | Thien | Jessica | Duy (`jameslu`) | Light |
|----------------|------:|----:|------:|--------:|-------:|------:|
| Acceptance Criteria (35) | 35 | 32 | 33 | 31 | 28 | 23 |
| System Architecture (20) | 19 | 19 | 19 | 17 | 18 | 12 |
| Distributed Systems & Concurrency (15) | 15 | 12 | 12 | 13 | 7 | 6 |
| Testing & Validation (15) | 15 | 14 | 14 | 12 | 12 | 10 |
| Scalability & Performance Design (5) | 5 | 5 | 5 | 5 | 5 | 0 |
| Observability (5) | 4 | 5 | 5 | 4 | 4 | 1 |
| Documentation & Communication (5) | 5 | 5 | 3 | 4 | 4 | 2 |
| **Total (100)** | **98** | **92** | **91** | **86** | **78** | **54** |

For the submission review scores only, the review guide interprets 90–100 as Exceptional, 80–89 as Strong, 70–79 as Good, 60–69 as Meets expectations, 50–59 as Needs improvement, and below 50 as Does not meet expectations.

---

## 1. KVP (`kvp`) — 114 total (92 submission review)

**Tech stack:** Go, PostgreSQL, Redis, Redpanda, REST/gRPC, OpenTelemetry/Grafana.

**Submission review summary:** KVP combines clear service boundaries, atomic database-backed driver claims, transactional messaging, extensive tests and detailed engineering documentation. The review identifies two material reliability gaps: failed dead-letter publication can still advance the source offset, and valid polar coordinates can leave the Redis GEO projection partially applied.

**Key strengths:**
- Authoritative PostgreSQL reservation exercised with concurrent transactions.
- Transactional inbox/outbox and substantial offer, timeout and rematching coverage.
- Concrete scale analysis, ADRs, tracing and metrics.

**Key risks:**
- Failed dead-letter publication can lose a message from the recovery path.
- Accepted polar coordinates may not be usable by Redis GEO matching.
- Some recovery needs manual replay; scaling the relay requires additional coordination.

[Full review](kvp-2026-11-02.md)

---

## 2. Grace Phan (`grace-phan`) — 112 total (98 submission review)

**Tech stack:** Elixir/Phoenix, OTP GenServer/Supervisor, in-memory Maps, ExUnit.

**Submission review summary:** Grace demonstrated all ten acceptance criteria within the declared single-process scope. A single authoritative owner makes reservation and related state changes atomic, while immutable replay results and meaningful lifecycle and concurrency tests support correctness. The review reports 321 passing tests and a fresh external API probe, including a competing-request check and the actual offer timeout.

**Key strengths:**
- Atomic reservation and cross-record transitions, with tested winner and loser outcomes.
- Original-response replay, payload conflict handling, and safe lifecycle replays.
- Controlled-clock and worker-recovery tests; concrete scaling trade-offs.

**Key risks:**
- Centralized validation and accumulated history increase maintenance and serialization costs.
- State is volatile on owner loss; this limitation is disclosed and permitted by the challenge.
- No active business metrics for detecting operational trends.

[Full review](grace-phan-2026-11-08.md)

---

## 3. Thien Nguyen (`thiennguyen`) — 101 total (91 submission review, provisional)

**Tech stack:** Java 25, Spring Boot/Modulith, PostgreSQL, Kafka, Redis, Gradle, JUnit and Testcontainers.

**Submission review summary:** The review finds a rigorous dispatch core with SQL compare-and-set claims, unique constraints, durable exclusions and a transactional outbox. Authored tests cover the acceptance criteria extensively. The latest review is static, however, and the integration suite—including its assignment race test—has not been executed by either reviewer.

**Key strengths:**
- Database-level reservation, lock ordering and schema constraints.
- Transactional outbox and consumer deduplication.
- Extensive authored acceptance tests, metrics and a staged scaling plan.

**Key risks:**
- Availability decisions based on a lagging copy can permit an invalid transition.
- Dead-letter paths can leave copies inconsistent and strand a driver or continue matching a cancelled ride.
- Creation replay returns current state instead of the original response; documentation overstates some recovery and CI claims.
- Run `make verify` with Docker before treating the score as final.

[Full review](thiennguyen-2026-11-13.md)

---

## 4. Jessica Nguyen (`jessicanguyen`) — 98 total (86 submission review)

**Tech stack:** TypeScript, Node.js, Express, in-memory repositories and queue, Vitest and Supertest.

**Submission review summary:** Jessica delivers a proportionate design with a real asynchronous consumer and a safe reservation within its single-process scope. The review reports 81 passing tests and overlapping HTTP probes. A public manual-offer path bypasses eligibility rules, and request fingerprints depend on JSON member order.

**Key strengths:**
- Clear domain and use-case boundaries with an asynchronous handoff.
- Reviewer-verified process-local driver exclusivity and timeout/rematching behavior.
- Useful structured logs and passing service, HTTP and queue tests.

**Key risks:**
- Manual offers can bypass the radius and permanent rejection exclusion.
- Equivalent reordered JSON can conflict under the same idempotency key.
- The participant test labelled concurrent is sequential; some acceptance paths lack test coverage.

[Full review](jessicanguyen-2026-11-12.md)

---

## 5. Duy Lu (`jameslu`) — 78 total (78 submission review at deadline)

**Tech stack:** Go, `net/http`, `log/slog`, in-memory stores and a partitioned goroutine event bus.

**Submission review summary:** James provides encapsulated state machines, asynchronous dispatch, broad tests and a concrete scaling plan. A reviewer reproduced a duplicate-matching and acceptance interleaving that can release a live reservation and result in two accepted rides for one driver. The default bus serializes same-ride handlers, which narrows the defect's reach, but the reviewed command-level guarantee remains unsafe.

**Key strengths:**
- Explicit ride and driver state machines with table-driven tests.
- Asynchronous event handling with retry and dead-letter tests.
- Race-detector execution and a reasoned future scaling design.

**Key risks:**
- Reservation compensation lacks attempt ownership, allowing double assignment under concurrent matching calls.
- A delayed timer can allow acceptance after the recorded offer deadline.
- The deadline revision can strand an offer during cancellation; the later head fixes this issue and scores 79/100.

[Full review](jameslu-2026-11-12.md)

---

## 6. Light Nguyen (`lightnguyen`) — 66 total (54 submission review)

**Tech stack:** .NET 10, ASP.NET Core, EF Core/SQLite, RabbitMQ and xUnit.

**Submission review summary:** Matching, rejection exclusions and the database reservation for competing rides have useful test and probe evidence. The decisive issue is an acceptance-versus-timeout race that can leave an accepted ride with an available driver, allowing another offer. The review also finds create-contract, availability and event-recovery defects.

**Key strengths:**
- Matching covers radius and deterministic tie-breaking.
- Database-enforced competing-ride reservation held in an overlapping probe.
- Asynchronous RabbitMQ boundary and documented design decisions.

**Key risks:**
- Acceptance and expiry can race into double assignment.
- Standard `longitude` input is mishandled; availability can bypass transition rules.
- Stale events and a save-then-publish gap can leave services out of sync.
- No scale plan, diagrams or substantial operational logging.

[Full review](lightnguyen-2026-11-12.md)

---
