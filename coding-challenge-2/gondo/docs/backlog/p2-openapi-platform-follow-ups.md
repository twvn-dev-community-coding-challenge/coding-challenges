# P2 — OpenAPI platform follow-ups (User Story 2)

| Field | Value |
|-------|-------|
| **Priority** | P2 |
| **Status** | **Done** — Drift **verify** (`verify_openapi.py`, `yarn verify-openapi`, pytest); **catalog** [`openapi/README.md`](../openapi/README.md); **typed** [`openapi-typescript`](https://github.com/openapi-ts/openapi-typescript) → `libs/ts-core/src/notification-api/openapi.generated.ts`; CI: **`.github/workflows/openapi-verify.yml`** |
| **Area** | Cross-domain SMS API |

## Context

`yarn nx run notification-service:generate-openapi` writes:

- `apps/notification-service/openapi.json`
- `docs/openapi/notification-service.openapi.json`

See [platform-sms-openapi.md](../platform-sms-openapi.md).

## Problem

Generating the spec is not the same as **adopting** it: versioning, discovery, and consumer ergonomics still matter for a “platform SMS API” story.

## Value

Other domains integrate against a **stable, reviewable contract** rather than tribal knowledge.

## Acceptance criteria

- [x] Regeneration steps documented (`yarn nx run notification-service:generate-openapi`; dual output paths in **`generate_openapi.py`**).
- [x] Drift check: **`verify_openapi.py`**, **`yarn verify-openapi`**, pytest `test_committed_openapi_json_matches_live_app`, GitHub Actions workflow.
- [x] Catalog / discovery: **`docs/openapi/README.md`** as the checked-in artifact index (substitute internal portal upload from this folder if required).
- [x] Typed API surface: **`yarn nx run ts-core:generate-openapi-types`** + **`openapi-typescript`** dependency; README §7 documents the workflow.

## References

- [platform-sms-openapi.md](../platform-sms-openapi.md)
