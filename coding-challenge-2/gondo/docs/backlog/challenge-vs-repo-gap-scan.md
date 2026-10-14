# Challenge brief vs `gondo` codebase — gap scan

**Source brief:** [`coding-challenge-2.md`](../../../coding-challenge-2.md) (inside `coding-challenge-2/`, sibling folder of `gondo/`).  
**Last reviewed:** 2026-04-18 — `notification-service` (`cqrs/`, `kpis.py`, `serialization.py`), optional **`X-Calling-Domain`**, KPI **`by_calling_domain`**, **`from`/`to`** on KPIs.

## Story coverage (US1–US6 + expectations)

| US / area | Brief asks for | In repo | Status |
|-----------|----------------|---------|--------|
| **US1** | messageId, country, phone, message; carrier; one provider; lifecycle to success; estimate/actual cost | `POST /notifications`, DB carrier prefixes, gRPC `SelectProvider`, dispatch/retry, **`cost_story`**, charging | **Met** — US1 **illustrative** VN/TH/SG routing tables ≠ **US3** authoritative PH/VN rows; **`as_of`** + migrations |
| **US2** | Same SMS behavior for Customer / Booking / ERP; no separate implementations per domain | One HTTP API + OpenAPI/`ts-core`; optional **`X-Calling-Domain`** → `channel_payload.calling_domain`, KPI **`by_calling_domain`** | **Met** |
| **US3** | PH carriers + updated VN routing | Seeds + [`routing-vs-challenge-brief.md`](../routing-vs-challenge-brief.md) | **Met** *(re-verify whenever `0002_seed_data` / routing migrations change)* |
| **US4** | Listed states/transitions; traceability; carrier-rejected vs send-failed | [`models.py`](../../apps/notification-service/models.py), NATS/callbacks, tests; MNO reject sim | **Met** — brief diagram **`Queue → Carrier-rejected`** is shorthand; **`Queue → Send-to-carrier → Carrier-rejected`** matches MNO rejection (see repo [`diagrams/notification-state-machine.md`](../../../../diagrams/notification-state-machine.md)); retries **`Carrier-rejected → Send-to-provider`** |
| **US5** | Est/actual; totals per provider & country; volume; success/failure | Row-level costs + **`GET /notifications/kpis`** (`from`/`to` on `created_at`) + **`by_calling_domain`** + `/kpis` UI | **Met** — aggregates are **in-memory process lifetime** (restart clears; not historical BI) |
| **US6** | Simulated async callbacks (messageId, provider, newState, actualCost) | `POST /provider-callbacks` + state machine | **Met** |
| **Tech § simplicité** | In-memory, simulated integrations, CLI/simple triggers OK | Postgres/NATS/React used **on purpose** for platform narrative | **Not a functional gap** — see **`SUBMISSION.md`** trade-offs + [stretch-brief-vs-repo-scope.md](stretch-brief-vs-repo-scope.md) (**Partial**) |
| **Out of scope** | Real SMS/CPaaS, live prod webhooks, cost **optimization** as product work | WireMock/sim paths; routing **selection** vs **business** optimizer | **Respected** |
| **Deliverables** | Code, README, scenarios | Root **`README.md`**, **`SUBMISSION.md`**, **`docs/`**, repo **`diagrams/`** | **Met** |

---

## Remaining gaps (anything still “open”)

These are **not** missing core user stories; they are **submission hygiene**, **stretch**, or **voluntary hardening**.

| Gap | Kind | Notes / owner |
|-----|------|----------------|
| **`SUBMISSION.md` → “AI tools used”** empty | **Minor (program)** | Fill when applicable — organizer disclosure |
| **Numbered demo script** (happy path + failure) | **Stretch [ST-02](stretch-demo-script.md)** | Not required for story acceptance; helps Demo Day / judges |
| **Durable notification store** | **Stretch [ST-01](stretch-notification-persistence.md)** | Brief allows in-memory; persistence is optional lift |
| **TWVN program bonuses** (TDD commits, peer review, etc.) | **Stretch [ST-03](stretch-program-bonuses.md)** | Process scoring, not domain features |
| **Explicit “CLI/minimal vs our stack” heading** | **Stretch [ST-04](stretch-brief-vs-repo-scope.md)** (**Partial**) | Trade-offs already in `SUBMISSION.md`; optional rename for skimmers |
| **Seed drift vs challenge tables** | **Hygiene** | After any Alembic/seed edit, reconcile with [`routing-vs-challenge-brief.md`](../routing-vs-challenge-brief.md) |

**No open P1/P2 product tickets** in [`README.md`](README.md) — optional polish items (`X-Calling-Domain`, KPI time filters) are implemented.

---

## Nuances (looks like a gap, isn’t)

| Topic | Detail |
|-------|--------|
| **OTP generation “out of scope” in brief** | Repo may still use **otp-service** for **delivery** demos; brief means *you are not graded on building OTP from scratch*. |
| **Cost “optimization”** | Brief mentions optimizing *as a business concern*; challenge scope = **track** estimate/actual and route by rules — not a solver for cheapest path in prod. |
| **`Implementation Simplicity` § DB/queue** | Official text suggests minimal infra; **`gondo`** chose Postgres + optional NATS + UI **deliberately** — document that choice for reviewers (see ST-04 / `SUBMISSION.md`). |

---

## Suggested follow-up (backlog cards)

| ID | Link |
|----|------|
| **ST-01** | [stretch-notification-persistence.md](stretch-notification-persistence.md) |
| **ST-02** | [stretch-demo-script.md](stretch-demo-script.md) |
| **ST-03** | [stretch-program-bonuses.md](stretch-program-bonuses.md) |
| **ST-04** | [stretch-brief-vs-repo-scope.md](stretch-brief-vs-repo-scope.md) — **Partial** |

---

*Update this file when the challenge text or `gondo` implementation changes materially.*
