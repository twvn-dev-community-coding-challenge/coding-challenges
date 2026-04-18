# P2 — OTP service hardening (post-implementation)

| Field | Value |
|-------|-------|
| **Priority** | P2 |
| **Status** | **Partial** — **`apps/otp-service`** + notification **`issue_server_otp`** shipped; **hardening** acceptance (prod flags, UI without plaintext) still **open** |
| **Area** | Security narrative |

## Context

**otp-service** issues server-side codes (TTL, hashed storage, verify API). Notification can substitute `{{OTP}}` when `issue_server_otp` is true.

## Problem

Demo ergonomics may expose **`otp_plaintext`** to browsers; production deployments should not treat that as normal.

## Value

Clear separation between **demo UX** and **production security** posture.

## Acceptance criteria

- [x] Production-style guidance documented: **`OTP_EXPOSE_PLAINTEXT_TO_CLIENT=false`** (main `README.md` — Scope and limitations / OTP bullet).
- [ ] UI / API path validated end-to-end with **plaintext disabled** (membership UX may still assume demo responses).
- [ ] **`OTP_HASH_PEPPER`** highlighted in operator README / runbook for non-dev (today: `docker-compose.yml` + `apps/otp-service/settings.py` defaults).
- [ ] Optional: rate limiting / abuse controls on issue & verify endpoints (stretch).

## References

- `apps/otp-service/`
- Main `README.md` — OTP scope note
- [system-vs-program-requirements.md § Prioritized backlog](../system-vs-program-requirements.md#prioritized-backlog-suggested-order)
