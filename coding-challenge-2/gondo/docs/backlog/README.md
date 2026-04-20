# Product & engineering backlog

Active backlog cards for **`gondo`** (completed work is summarized in [challenge-vs-repo-gap-scan.md](challenge-vs-repo-gap-scan.md)). Priority bands match [system-vs-program-requirements.md](../system-vs-program-requirements.md).

### Mapping to [`coding-challenge-2.md`](../../../coding-challenge-2.md)

**Story-by-story scan:** [challenge-vs-repo-gap-scan.md](challenge-vs-repo-gap-scan.md).

| Challenge area | Notes |
|----------------|-------|
| US1 Membership SMS + routing | Implemented — gap scan + [routing-vs-challenge-brief.md](../routing-vs-challenge-brief.md); charging / `cost_story` in repo |
| US2 Cross-domain API | OpenAPI + `ts-core`; optional **`X-Calling-Domain`** on create → `channel_payload.calling_domain`, KPI **`by_calling_domain`** |
| US3 PH + updated VN | Implemented — seeds vs brief in routing doc |
| US4 Lifecycle / transitions | Implemented — callbacks, retries, carrier-rejected sim path |
| US5 Aggregate KPIs | Implemented — `GET /notifications/kpis` (`from` / `to` on `created_at`) + `/kpis` UI |
| US6 Async callbacks | Implemented |
| Brief “CLI / in‑memory only” vs this repo | **stretch-brief-vs-repo-scope.md** (**Partial**) |

| Priority | Meaning |
|----------|---------|
| **Stretch** | Explicitly optional — persistence, demos, bonuses |

**Status legend:** **Open** · **Partial**

## Not done (stretch)

| ID | Status | File | Summary |
|----|--------|------|---------|
| ST-01 | **Open** | [stretch-notification-persistence.md](stretch-notification-persistence.md) | Persist notifications (replace in-memory store) |
| ST-02 | **Open** | [stretch-demo-script.md](stretch-demo-script.md) | Numbered demo: happy path + failure path (~15 min) |
| ST-03 | **Open** | [stretch-program-bonuses.md](stretch-program-bonuses.md) | Optional program bonuses (TDD commits, peer review, creative note, …) |
| ST-04 | **Partial** | [stretch-brief-vs-repo-scope.md](stretch-brief-vs-repo-scope.md) | Brief (CLI/in-memory) vs REST/Postgres/NATS — narrative in `SUBMISSION.md`; optional heading rename |

### Stretch quick reference

| Ticket | Status | Focus |
|--------|--------|-------|
| **ST-01** | **Open** | Postgres (or other) persistence for notifications; migration story |
| **ST-02** | **Open** | Live/reviewer script in `SUBMISSION.md` or `docs/` |
| **ST-03** | **Open** | TWVN extras: TDD (+5), peer review (+5), junior-led demo (+5), creative (+10) |
| **ST-04** | **Partial** | README/SUBMISSION already explain stack; optional “Challenge brief vs this repository” heading |

## Related docs

- [challenge-vs-repo-gap-scan.md](challenge-vs-repo-gap-scan.md) — **US1–US6 vs code** snapshot
- [system-vs-program-requirements.md](../system-vs-program-requirements.md) — master mapping & reviewer quickstart
- [platform-sms-openapi.md](../platform-sms-openapi.md) — OpenAPI export story
- [routing-vs-challenge-brief.md](../routing-vs-challenge-brief.md) — User Story 3
