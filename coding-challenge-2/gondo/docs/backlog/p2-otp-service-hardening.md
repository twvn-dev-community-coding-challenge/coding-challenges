# P2 — OTP service hardening (post-implementation)

| Field | Value |
|-------|-------|
| **Priority** | P2 |
| **Status** | **Done** — Per-IP rolling-minute limits on **`POST /v1/challenges`** and **`POST /v1/verify`** (`OTP_ISSUE_REQUESTS_PER_MINUTE`, `OTP_VERIFY_REQUESTS_PER_MINUTE`; **`0`** = off); tests in **`tests/test_rate_limit.py`**. |
| **Area** | Security narrative |

## Context

**otp-service** issues server-side codes (TTL, hashed storage, verify API). Notification can substitute `{{OTP}}` when `issue_server_otp` is true.

## Problem

Demo ergonomics may expose **`otp_plaintext`** to browsers; production deployments should not treat that as normal.

## Value

Clear separation between **demo UX** and **production security** posture.

## Acceptance criteria

- [x] Production-style guidance documented: **`OTP_EXPOSE_PLAINTEXT_TO_CLIENT=false`** (main `README.md` — Scope and limitations / OTP bullet).
- [x] **notification-service** create path with **`issue_server_otp`** tested with expose **off**: response has no **`otp_plaintext`**; **`content`** still substituted (`tests/test_notification_otp_policy.py`). *(Frontend membership still targets demo defaults; prod clients rely on SMS channel, not API plaintext.)*
- [x] **`OTP_HASH_PEPPER`** and related vars in **README** operator table (`OTP_SERVICE_BASE_URL`, `OTP_TTL_SECONDS`, …).
- [x] Rate limiting on issue / verify (**`rate_limit.py`**, **`429`** with `issue_rate_limit_exceeded` / `verify_rate_limit_exceeded`).

## References

- `apps/otp-service/`
- Main `README.md` — OTP scope note
- [system-vs-program-requirements.md § Prioritized backlog](../system-vs-program-requirements.md#prioritized-backlog-suggested-order)
