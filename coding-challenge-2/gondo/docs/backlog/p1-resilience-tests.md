# P1 — Resilience & illegal-transition tests

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Mostly done** — `notification-service/tests/test_dispatch.py`: **503** (bus publish failure), **409** (double dispatch), **`CHARGING_RATE_NOT_FOUND`**; **`test_notifications_api`**: **NOT_FOUND**. Further edge cases optional. |
| **Area** | Tests & reliability (+5 edge cases in program scoring) |

## Problem

Production-like behavior includes **dependency failures** (e.g. 503), **illegal state transitions**, and **idempotency / conflict** responses (e.g. double dispatch → 409).

## Value

Higher confidence in lifecycle correctness and clearer failure semantics for API consumers.

## Acceptance criteria

- [x] Dispatch **503** when dispatch event not published (`test_dispatch_returns_503_when_dispatch_event_not_published`).
- [x] Illegal second dispatch → **409** (`test_dispatch_rejects_when_notification_not_in_new_state`).
- [x] Charging / gRPC failure paths (**NOT_FOUND**, rate not found) covered in dispatch tests + API tests.
- [x] `yarn nx run-many -t test` + Postgres / `DATABASE_URL` documented (`README.md`, `SUBMISSION.md`).
- [ ] Optional stretch: more illegal lifecycle transitions beyond double-dispatch if product scope expands.

## References

- [system-vs-program-requirements.md § 3](../system-vs-program-requirements.md#3-program-scoring-lenses-tests--docs)
