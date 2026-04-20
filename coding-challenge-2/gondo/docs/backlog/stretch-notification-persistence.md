# Stretch — Durable notification store

| Field | Value |
|-------|-------|
| **Priority** | Stretch |
| **Status** | Open |
| **Area** | Architecture / resilience |

## Problem

Notifications are intentionally **in-memory** in notification-service for the challenge simulation; restarts lose state and multi-instance behavior is undefined.

## Value

Closer to production orchestration and easier debugging across deploys.

## Acceptance criteria

- [ ] Persist notification aggregates (Postgres or other) with clear migration story.
- [ ] Document trade-offs vs challenge scope (avoid over-engineering beyond team narrative).

## References

- Main `README.md` — Scope and limitations
- [SUBMISSION.md](../../SUBMISSION.md) — With more time
