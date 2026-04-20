# P1 — Resilience & illegal-transition tests

| Field | Value |
|-------|-------|
| **Priority** | P1 |
| **Status** | **Done** — Dispatch/retry **404**; retry **409** when not retryable; duplicate **`message_id`** **422**; OTP issue **503** / expose policy tests; otp-service verify outcome → HTTP status matrix + invalid UUID **422**. |
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
- [x] Extra coverage: unknown ID on dispatch/retry (`test_dispatch.py`); retry from **New** → `RETRY_NOT_ALLOWED` (`test_dispatch.py`); duplicate create (`test_notifications_api.py`); OTP path (`test_notification_otp_policy.py`); otp-service verify errors (`tests/test_api.py`).

## References

- [system-vs-program-requirements.md § 3](../system-vs-program-requirements.md#3-program-scoring-lenses-tests--docs)
