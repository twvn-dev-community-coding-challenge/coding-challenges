# Go & Do (gondo) — Challenge submission notes

Use this file for **PR descriptions** and **Demo Day** talking points. Extend sections as you finalize the solution.

## Team information

| | |
|--|--|
| **Team name** | Go & Do (`gondo`) |
| **Members** | Hien Huynh (Team Lead / Kaluza Pod 4); Thien Nguyen (Developer / Kaluza Pod 3) |

## Solution overview

Distributed SMS orchestration for membership (and reusable notification entry): **notification-service** exposes HTTP; **provider-service** resolves routing from PostgreSQL + YAML registry and publishes **NATS** dispatch events; **carrier-service** simulates outbound send + outcomes; **charging-service** stores **provider × carrier × country** rates in PostgreSQL and exposes gRPC cost APIs. Optional **frontend** exercises create + dispatch + tracking.

*(Tighten to 2–3 sentences for the official PR.)*

## Key design decisions

- **Why multiple services:** Mirror bounded contexts (customer notification vs provider routing vs carrier execution vs billing) aligned with the challenge’s platform-SMS narrative; not all boundaries would survive a production rewrite unchanged.
- **Why NATS:** Async handoff between provider selection and carrier work; supports `sms.dispatch.received` / `outcome` without synchronous HTTP coupling.
- **Why YAML registries:** Provider/carrier metadata and `api_endpoint` simulation without embedding huge config in SQL.
- **Trade-off:** Notification persistence is **in-memory** for the challenge simulation; production would use a real store and outbox.

## Technologies and tools

| Layer | Choices |
|-------|---------|
| Monorepo / UI | NX, React, Vite, MUI |
| Services | Python 3.13, FastAPI, uvicorn |
| Integration | gRPC (`libs/grpc-contracts`), optional NATS |
| Data | PostgreSQL, Alembic per service schema |
| Docs | Mermaid diagrams in repo `diagrams/` |

## How to run

See **[README.md](README.md)** — especially **For reviewers (judge quickstart)** and **Tests**.

## How to test

```bash
cd coding-challenge-2/gondo   # from repo root
source .venv/bin/activate     # after creating venv per README
yarn nx run-many -t test      # Postgres required for DB-backed tests
```

Charging/provider tests expect `DATABASE_URL` pointing at migrated PostgreSQL (`yarn db:migrate`).

## Trade-offs and limitations

- **No real CPaaS:** Provider HTTP is simulated (WireMock / skeleton GET in carrier-service).
- **In-memory notifications:** Restart clears state; not multi-instance safe.
- **OTP (membership demo):** **otp-service** issues codes (TTL, hashed at rest); notification substitutes `{{OTP}}` when `issue_server_otp` is true. Demo may return **`otp_plaintext`** for the UI (`OTP_EXPOSE_PLAINTEXT_TO_CLIENT` — turn **off** in production).
- **Cost at lifecycle steps:** **EstimateCost** runs on dispatch/retry after provider selection; **RecordActualCost** runs on success/failed callbacks (requires prior dispatch so `selected_provider_id` is set).

## With more time, we would

- Persist notifications across restarts; tighten OTP exposure policy in non-demo environments.
- Call charging on dispatch / callback consistently; expose estimate/actual on the resource.
- Broaden resilience and contract tests; publish the **notification-service** OpenAPI (`docs/openapi/notification-service.openapi.json`) to an internal catalog or generate a typed client SDK.

## AI tools used (if any)

*(Document tools and how they were used — e.g. Cursor, Copilot — for organizer bonus eligibility.)*

- 

## Links for reviewers

- [Challenge fit & backlog](docs/system-vs-program-requirements.md)
- [Platform SMS OpenAPI export (User Story 2)](docs/platform-sms-openapi.md)
- [Architecture diagrams](../../diagrams/README.md) *(repo root `diagrams/` — two levels up from `gondo/`)*
