# Current system vs program & challenge expectations

This doc does three things:

1. Align **TWVN program** expectations (submission + scoring mindset) with the **`gondo`** codebase.  
2. Map the **challenge brief** ([`coding-challenge-2.md`](../coding-challenge-2.md)) epics and user stories to what is implemented vs missing.  
3. Provide a **prioritized improvement backlog** and **reviewer quickstart** so you can close gaps deliberately (not randomly).

**Related:** [Program overview (process only)](coding-challenge-program-context.md).

**Stack recap:** NX monorepo — FastAPI services (notification, provider, charging, carrier, message-bus gateway), React frontend, PostgreSQL, optional NATS, YAML registries (`infra/provider-registry`, `infra/carrier-registry`), gRPC (`libs/grpc-contracts`), diagrams under repo [`diagrams/`](../../../diagrams/).

---

## Executive summary

| Area | Strong fit | Main gap |
|------|------------|----------|
| **Program / PR hygiene** | Branch + folder layout; architecture README | Template sections + explicit trade-offs + AI disclosure in one place (`SUBMISSION.md` or README sections) |
| **Challenge — routing & lifecycle** | Provider routing DB + YAML; NATS dispatch; notification states; diagrams | Surface **cost at Send-to-provider** and **actual at Send-success** end-to-end in API behavior + tests |
| **Challenge — multi-domain SMS** | Single HTTP API entry (`notification-service`) reusable in principle | Document **domain-agnostic contract**; optional shared “client SDK” or OpenAPI narrative |
| **Challenge — PH + VN rule updates** | Seed/migration data can encode rules | Verify seeds match **User Story 3** tables; document how ops would update rules |
| **Tests (15 pts)** | Pytest across services | Add resilience tests + `run-many -t test` prerequisites doc |
| **Docs (15 pts)** | Deep technical README | **Judge quickstart** + **limitations** paragraph + **demo script** |

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
| Lifecycle through Send-success | States + NATS + callbacks | ⚠️ Ensure **cost estimate** appears when entering Send-to-provider; **actual cost** when reaching Send-success (see §4) |
| OTP in SMS | UI generates OTP client-side today | ⚠️ Brief says OTP gen out of scope *for them* — your **delivery** story is still stronger if optional **server-side OTP** service documents verification separately |

### User Story 2 — Cross-domain SMS

| Requirement | System behavior | Gap / improvement |
|-------------|-------------------|-------------------|
| Any domain triggers send | HTTP API on notification-service | ✅ |
| Consistent routing/lifecycle/cost | Shared backend | ⚠️ Document **one contract** (OpenAPI) as the “platform API”; mention **caller identity / domain header** if you add it |

### User Story 3 — PH market + VN rule updates

| Requirement | System behavior | Gap / improvement |
|-------------|-------------------|-------------------|
| PH + updated VN routing | PostgreSQL `routing_rules` + seeds | ✅ Reconcile **every** row with [`coding-challenge-2.md`](../coding-challenge-2.md) § User Story 3 |
| Operational change over time | Alembic + `as_of` on routing | ✅ Explain in README how ops rolls out new rules |

### User Story 4 — Lifecycle management

| Requirement | System behavior | Gap / improvement |
|-------------|-------------------|-------------------|
| Listed states & retries | `models.py` + callbacks | ✅ Add tests for **illegal transitions** + **retry** |

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

| Moment (brief) | Expected semantics | Suggested improvement |
|----------------|-------------------|------------------------|
| **Send-to-provider** | Estimated cost available | Wire **charging `EstimateCost`** during dispatch (or immediately after provider selection) and persist estimate ref on notification — **or** document precisely why estimate is deferred and mock it in tests |
| **Send-success** | Actual cost available | On provider callback / terminal success path, call **`RecordActualCost`** (charging already exposes RPC) or document simulation |

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

## Prioritized backlog (suggested order)

**Detailed backlog cards (one Markdown file per item):** [`docs/backlog/README.md`](backlog/README.md).

| Priority | Item | Helps |
|----------|------|-------|
| **P0** | `SUBMISSION.md` + README reviewer/limitations/cost/minimal-deployment sections | Docs/scoring *(started — see repo `SUBMISSION.md` & main `README.md`)* |
| **P0** | Judge quickstart + “limitations” (simulation, in-memory store) | Docs/scoring *(see README § For reviewers)* |
| **P1** | Cost estimate/actual wired or explicitly documented gap | Challenge § cost *(**dispatch/retry** call `EstimateCost`; **Send-success** / **Send-failed** callback → `RecordActualCost` — see root README)* |
| **P1** | Resilience tests (503, invalid transition) | Tests +5 *(503 existed; double-dispatch → 409 added)* |
| **P1** | Verify routing seeds = User Story 3 | Core correctness *(see `docs/routing-vs-challenge-brief.md`)* |
| **P1** | **Gap:** User Story 5 aggregate KPIs (totals per provider/country, volume, success rates) | Challenge § US5 — [`docs/backlog/p1-us5-cost-aggregates-kpis.md`](backlog/p1-us5-cost-aggregates-kpis.md) |
| **P1** | **Gap:** User Story 4 **Carrier-rejected** transition path tested end-to-end | Challenge § US4 — [`docs/backlog/p1-us4-carrier-rejected-path.md`](backlog/p1-us4-carrier-rejected-path.md) |
| **P2** | Server-side OTP + TTL | **Implemented** — `apps/otp-service/` (hashed codes, `expires_at`, `OTP_TTL_SECONDS`, verify endpoint). Notification integrates via `issue_server_otp`; **disable** `OTP_EXPOSE_PLAINTEXT_TO_CLIENT` in production so browsers do not receive plaintext OTP in API responses. |
| **P2** | OpenAPI export for “platform SMS API” | **Implemented** — generate (`yarn nx run notification-service:generate-openapi`), verify (`yarn verify-openapi`), catalog [`docs/openapi/README.md`](openapi/README.md), TS types (`yarn nx run ts-core:generate-openapi-types`). Narrative: [`platform-sms-openapi.md`](platform-sms-openapi.md). Details: [`docs/backlog/p2-openapi-platform-follow-ups.md`](backlog/p2-openapi-platform-follow-ups.md). |

### Stretch backlog *(optional — not implemented)*

| Priority | Item | Card |
|----------|------|------|
| **Stretch** | Durable notification store | [`docs/backlog/stretch-notification-persistence.md`](backlog/stretch-notification-persistence.md) |
| **Stretch** | Demo script (happy path + failure path) | [`docs/backlog/stretch-demo-script.md`](backlog/stretch-demo-script.md) |
| **Stretch** | Program bonuses (TDD, peer review, …) | [`docs/backlog/stretch-program-bonuses.md`](backlog/stretch-program-bonuses.md) |
| **Stretch** | Brief vs repo (CLI/in-memory vs REST/Postgres) — reviewer note | [`docs/backlog/stretch-brief-vs-repo-scope.md`](backlog/stretch-brief-vs-repo-scope.md) |

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
