# P2 — OpenAPI platform follow-ups (User Story 2)

| Field | Value |
|-------|-------|
| **Priority** | P2 |
| **Status** | **Partial** — **Done:** dual JSON export + [`platform-sms-openapi.md`](../platform-sms-openapi.md). **Open:** CI drift check, portal publish, typed SDK |
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
- [ ] CI optional check that checked-in OpenAPI matches running app (`openapi.json` drift).
- [ ] Optional: publish JSON to an internal portal or attach to releases.
- [ ] Optional: generate a **typed client** or publish package name + workflow in README.

## References

- [platform-sms-openapi.md](../platform-sms-openapi.md)
