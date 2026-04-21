# Current system vs program & challenge expectations

This doc does three things:

1. Align **TWVN program** expectations (submission + scoring mindset) with the **`gondo`** codebase.  
2. Map the **challenge brief** ([`coding-challenge-2.md`](../../coding-challenge-2.md)) epics and user stories to what is implemented vs missing.  
3. Provide a **prioritized improvement backlog** and **reviewer quickstart** so you can close gaps deliberately (not randomly).

**Related:** [Program overview (process only)](coding-challenge-program-context.md).

**Stack recap:** NX monorepo — FastAPI services (notification, provider, charging, carrier, message-bus gateway), React frontend, PostgreSQL, optional NATS, YAML registries (`infra/provider-registry`, `infra/carrier-registry`), gRPC (`libs/grpc-contracts`), diagrams under repo [`diagrams/`](../../../diagrams/).

---

## Executive summary

| Area | Strong fit | Main gap |
|------|------------|----------|
| **Program / PR hygiene** | Branch + folder layout; architecture README + `SUBMISSION.md` | Fill **AI tools** in `SUBMISSION.md` if applicable |
| **Challenge — routing & lifecycle** | Provider routing DB + YAML; NATS dispatch; notification states; **`cost_story`**; **`GET /notifications/kpis`** (`from`/`to`) + KPI UI | Stretch: durable store (**ST-01**) |
| **Challenge — multi-domain SMS** | Single HTTP API entry (`notification-service`) reusable in principle | OpenAPI + `ts-core`; optional **`X-Calling-Domain`** on create → KPI **`by_calling_domain`** |
| **Challenge — PH + VN rule updates** | Seed/migration data can encode rules | Verify seeds match **User Story 3** tables; document how ops would update rules |
| **Tests (15 pts)** | Pytest across services; resilience paths covered in service tests | Keep CI green; extend edge cases if gaps appear |
| **Docs (15 pts)** | Deep technical README; judge quickstart + limitations | **Demo script** still stretch (**ST-02**); see [challenge-vs-repo-gap-scan.md](backlog/challenge-vs-repo-gap-scan.md) |

---

## 1. Program scenario: fork, branch, folder, PR

| Expectation | Today | Improvement |
|-------------|-------|-------------|
| Code under `coding-challenge-2/<team>/` | Matches `gondo` layout | Keep root-level changes minimal; `.gitignore` etc. justified in PR if needed |
| README template (team) | Root [`README.md`](../README.md) is architecture-heavy | Add **`SUBMISSION.md`** or a top section **“Challenge submission checklist”** with: Approach, Design decisions, Challenges, Learned, AI tools — link [`diagrams/`](../../../diagrams/) + this `docs/` folder |
| PR description | Manual | Copy organizer PR template; include **network diagram link** + **how to run tests** |

**PR description snippet (paste & fill):**

```markdown
## Solution overview
## Key design decisions (3 bullets)
## How to run (copy from docs/system-vs-program-requirements.md § Judge quickstart)
## How to test (`yarn nx run-many -t test` + Postgres note)
## Trade-offs & limitations (simulated SMS, in-memory notifications, …)
## AI tools used (if any)
```

---

## 2. Challenge brief ↔ system mapping

The brief assumes OTP generation **out of scope** but requires **SMS capability**, shared platform usage, routing by **country + carrier**, **pricing over time**, and lifecycle/cost semantics. Below is an honest mapping.

### User Story 1 — Membership OTP SMS

| Requirement | System behavior | Gap / improvement |
|-------------|-------------------|-------------------|
| SMS with messageId, country, phone, message | `POST /notifications` + dispatch | ✅ Core path exists |
| Carrier from phone (simulation OK) | `derive_carrier` + prefixes | ✅ Document derivation rules vs brief examples |
| Exactly one provider | `SelectProvider` + routing rules | ✅ Keep seeds aligned with brief |
| Lifecycle through Send-success | States + NATS + callbacks | ✅ **`cost_story`** + charging gRPC — estimate at dispatch/retry (**Send-to-provider**); actual from callback (**Send-success** / failed) — see §4 |
| OTP in SMS | UI generates OTP client-side today | ⚠️ Brief says OTP gen out of scope *for them* — your **delivery** story is still stronger if optional **server-side OTP** service documents verification separately |

### User Story 2 — Cross-domain SMS

| Requirement | System behavior | Gap / improvement |
|-------------|-------------------|-------------------|
| Any domain triggers send | HTTP API on notification-service | ✅ |
| Consistent routing/lifecycle/cost | Shared backend | ✅ OpenAPI narrative ([`platform-sms-openapi.md`](platform-sms-openapi.md)); optional **`X-Calling-Domain`** header on create (see same doc) |

### User Story 3 — PH market + VN rule updates

| Requirement | System behavior | Gap / improvement |
|-------------|-------------------|-------------------|
| PH + updated VN routing | PostgreSQL `routing_rules` + seeds | ✅ Reconcile **every** row with [`coding-challenge-2.md`](../../coding-challenge-2.md) § User Story 3 |
| Operational change over time | Alembic + `as_of` on routing | ✅ Explain in README how ops rolls out new rules |

### User Story 4 — Lifecycle management

| Requirement | System behavior | Gap / improvement |
|-------------|-------------------|-------------------|
| Listed states & retries | `models.py` + callbacks + tests (carrier-rejected sim path, retries) | ✅ Core coverage; extend if new edge cases found |

---

## 3. Program scoring lenses (tests & docs)

### B. Tests & reliability (15 pts)

| Lens | Today | Improvement |
|------|-------|-------------|
| Core coverage | Multiple pytest suites | Run `yarn nx run-many -t test` before PR; fix flakes |
| Edge cases (+5) | Partial | Add: dispatch **503** path, **invalid state** dispatch, **gRPC NOT_FOUND** routing |

### C. Documentation & communication (15 pts)

| Lens | Today | Improvement |
|------|-------|-------------|
| Setup (+5) | Long README | **Judge quickstart** (below) at top of README or `docs/` |
| Explain decisions (+10) | Diagrams | 5 bullets: **why microservices**, **why NATS**, **why YAML registries**, **why charging schema** |

---

## Judge quickstart (copy-paste)

From repo **`coding-challenge-2/gondo/`** (adjust if paths differ):

```bash
nvm use
yarn install
python3 -m venv .venv && source .venv/bin/activate
pip install -r apps/notification-service/requirements.txt
# Postgres (Docker): yarn db:start   # from gondo/package.json
yarn db:migrate
yarn start
# UI: yarn nx run frontend:serve   # optional; membership flow uses notification HTTP
```

Smoke API (notification-service on **8001**):

```bash
curl -s http://localhost:8001/health
```

If tests need Postgres:

```bash
export DATABASE_URL=postgresql://gondo:gondo@localhost:5432/gondo
yarn nx run-many -t test
```

---

## 4. Cost story (challenge: estimate vs actual)

**Implementation (notification-service):** `EstimateCost` runs on **`POST …/dispatch`** and **`POST …/retry`** after **`SelectProvider`**, before the notification enters **Send-to-provider**; estimate fields are persisted on the resource. **`RecordActualCost`** runs on **`POST /provider-callbacks`** for **Send-success** / **Send-failed** (when `actual_cost` is supplied for Send-failed). Each **`GET /notifications/{id}`** (and dispatch/retry responses) includes a read-only **`cost_story`** object mapping these semantics for reviewers.

| Moment (brief) | Expected semantics | Status |
|----------------|-------------------|--------|
| **Send-to-provider** | Estimated cost available | **Done** — `estimated_cost`, `charging_estimate_id`, … after charging gRPC during dispatch/retry |
| **Send-success** | Actual cost available | **Done** — `last_actual_cost`, `charging_actual_cost_id` from `RecordActualCost` after callback (estimate fallback if `actual_cost` omitted) |

Pipeline rows: **`state.Send-to-provider`** includes an estimate snapshot; **`charging.RecordActualCost.ok`** records the actual billing outcome.

---

## 5. Bonuses (process + optional engineering)

| Bonus | What to do |
|-------|------------|
| **TDD (+5)** | Small feature with red-green-refactor **commits** |
| **Junior-led demo (+5)** | Rotate presenter; rehearse |
| **Peer review (+5)** | Constructive review on another fork |
| **Creative (+10)** | e.g. structured error codes on outcomes, OTEL-ready logging — **one paragraph** in README on intent |

---

## 6. Anti–over-engineering guardrails

| Risk | Mitigation |
|------|------------|
| Many services vs brief | README: **“Minimal deployment”** = which processes can merge for a monolith thought experiment |
| Dead code | Delete or mark `# stretch` |

---

## Prioritized backlog

Shipped baseline (routing, lifecycle, costs, KPIs, OpenAPI, OTP service, tests) is summarized in §2–§4 above and [`docs/backlog/challenge-vs-repo-gap-scan.md`](backlog/challenge-vs-repo-gap-scan.md).

**Active backlog cards:** [`docs/backlog/README.md`](backlog/README.md).

| Priority | Item | Helps |
|----------|------|-------|
| **P0** *(minor)* | Complete **`AI tools used`** in `SUBMISSION.md` when applicable | Program disclosure |

### Stretch backlog *(optional — not implemented)*

| Priority | Item | Card |
|----------|------|------|
| **Stretch** | Durable notification store | [`docs/backlog/stretch-notification-persistence.md`](backlog/stretch-notification-persistence.md) |
| **Stretch** | Demo script (happy path + failure path) | [`docs/backlog/stretch-demo-script.md`](backlog/stretch-demo-script.md) |
| **Stretch** | Program bonuses (TDD, peer review, …) | [`docs/backlog/stretch-program-bonuses.md`](backlog/stretch-program-bonuses.md) |
| **Stretch** | Brief vs repo (CLI/in-memory vs REST/Postgres) — reviewer note | [`docs/backlog/stretch-brief-vs-repo-scope.md`](backlog/stretch-brief-vs-repo-scope.md) *(**Partial** — `SUBMISSION.md` trade-offs; optional heading tweak)* |

---

## One-page submission checklist

- [ ] PR uses organizer template + linked diagrams/docs  
- [ ] `yarn start` (or Docker) works from clean clone after `db:migrate`  
- [ ] Tests command documented; Postgres noted where required  
- [ ] Trade-offs + simulated vs production behavior stated  
- [ ] Demo script: happy path + one failure path  
- [ ] AI tools section (if used)  

---

*Living document — trim or extend as the challenge brief and `gondo` implementation evolve.*
