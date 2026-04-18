# Product & engineering backlog

Structured backlog cards for **`gondo`**. Priority bands match [system-vs-program-requirements.md](../system-vs-program-requirements.md).

### Mapping to [`coding-challenge-2.md`](../../../coding-challenge-2.md)

| Challenge area | Backlog coverage |
|----------------|------------------|
| US1 Membership SMS + routing (VN/TH/SG) | P1-03 seeds, P1-01 cost; OTP beyond “SMS only” narrative → P2-01 |
| US2 Cross-domain API | P2-02 OpenAPI |
| US3 PH + updated VN | P1-03 |
| US4 Lifecycle / transitions | Partially covered by tests — **gap:** explicit **Carrier-rejected** path → **P1-US4** |
| US5 Aggregate KPIs (cost/volume/rates **by provider & country**) | **Gap:** per-notification cost exists; aggregates not built → **P1-US5** |
| US6 Async callbacks | Implemented; no separate ticket |
| Brief “CLI / in‑memory only” vs this repo | **stretch-brief-vs-repo-scope.md** |

| Priority | Meaning |
|----------|---------|
| **P0** | Submission / reviewer experience — do before merge when possible |
| **P1** | Core challenge fit — correctness, tests, cost semantics |
| **P2** | Enhancements — often implemented or partially implemented; track hardening |
| **Stretch** | Explicitly optional — persistence, demos, bonuses |

**Status legend:** **Done** · **Done (minor gaps)** · **Partial** · **Open** — last reviewed against the repo in this folder.

## Done

| ID | Status | File | Summary |
|----|--------|------|---------|
| P0-01 | **Done (minor gaps)** | [p0-submission-readme.md](p0-submission-readme.md) | Template + links; optional **AI tools** line |
| P0-02 | **Done** | [p0-judge-quickstart-limitations.md](p0-judge-quickstart-limitations.md) | Quickstart + limitations in README |
| P1-01 | **Done** | [p1-cost-story.md](p1-cost-story.md) | Charging estimate/actual + tests |
| P1-02 | **Done** | [p1-resilience-tests.md](p1-resilience-tests.md) | Extended dispatch/retry/OTP/otp-API tests |
| P1-03 | **Done** | [p1-routing-seeds-user-story-3.md](p1-routing-seeds-user-story-3.md) | Story 3 vs seeds in routing doc |
| P2-01 | **Done** | [p2-otp-service-hardening.md](p2-otp-service-hardening.md) | Policy + rate limits + README operator table |
| P2-02 | **Done** | [p2-openapi-platform-follow-ups.md](p2-openapi-platform-follow-ups.md) | Verify + catalog doc + openapi-typescript + CI workflow |

## Not done

| ID | Status | File | Summary |
|----|--------|------|---------|
| P1-04 | **Open** | [p1-us5-cost-aggregates-kpis.md](p1-us5-cost-aggregates-kpis.md) | **User Story 5:** totals per provider/country, volume, success/failure rates |
| P1-05 | **Open** | [p1-us4-carrier-rejected-path.md](p1-us4-carrier-rejected-path.md) | **User Story 4:** Queue → Carrier-rejected → retry test/trace |
| ST-01 | **Open** | [stretch-notification-persistence.md](stretch-notification-persistence.md) | Persist notifications (replace in-memory store) |
| ST-02 | **Open** | [stretch-demo-script.md](stretch-demo-script.md) | Numbered demo: happy path + failure path (~15 min) |
| ST-03 | **Open** | [stretch-program-bonuses.md](stretch-program-bonuses.md) | Optional program bonuses (TDD commits, peer review, creative note, …) |
| ST-04 | **Open** | [stretch-brief-vs-repo-scope.md](stretch-brief-vs-repo-scope.md) | Doc: brief (CLI/in-memory) vs REST/Postgres/NATS stack |

### Stretch quick reference *(all **Open**)*

| Ticket | Focus |
|--------|--------|
| **ST-01** | Postgres (or other) persistence for notifications; migration story |
| **ST-02** | Live/reviewer script in `SUBMISSION.md` or `docs/` |
| **ST-03** | TWVN extras: TDD (+5), peer review (+5), junior-led demo (+5), creative (+10) |
| **ST-04** | README/SUBMISSION note reconciling **`coding-challenge-2.md`** minimal deliverable vs full stack |

## Related docs

- [system-vs-program-requirements.md](../system-vs-program-requirements.md) — master mapping & table backlog
- [platform-sms-openapi.md](../platform-sms-openapi.md) — OpenAPI export story
- [routing-vs-challenge-brief.md](../routing-vs-challenge-brief.md) — User Story 3
