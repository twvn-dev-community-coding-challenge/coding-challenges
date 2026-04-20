# Specs-Kit: Test Cases Matrix

This document is the canonical testcase matrix for the SMS OTP project. It complements [SPECS_KIT_REGISTRATION_USER_JOURNEY.md](./SPECS_KIT_REGISTRATION_USER_JOURNEY.md).

## Implementation notes (current backend)

| Area | Implemented endpoint / behavior |
| --- | --- |
| Registration | `POST /api/register` with `username`, `password`, `phone_number`, `country`, optional `otp_ttl_seconds` |
| Verify OTP | `POST /api/verify` with `username`, `otp_code` |
| SMS | `POST /api/sms/send`, `POST /api/sms/callback`, `GET /api/sms/messages`, `GET /api/sms/stats` |
| Test-only SMS simulation | Header `x-test-sms-mode`: `transient-failure-once`, `always-fail`, `always-success` |
| OTP expiry | Env `OTP_TTL_SECONDS` default; per-request `otp_ttl_seconds` (bounded) |

Journey doc sections that reference `POST /api/register/start` and `POST /api/register/verify-otp` describe the **target contract**; map matrix rows below to the **implemented** routes when executing tests.

---

## 1. Registration and OTP verification (API)

| ID | Layer | Scenario | Preconditions | Steps | Expected HTTP / outcome | Automation ref |
| --- | --- | --- | --- | --- | --- | --- |
| REG-001 | API / E2E | Happy path: register then verify with valid OTP | Unique username and phone; country+carrier resolvable (e.g. Philippines `63917…`) | `POST /api/register` → obtain OTP (test hook or `GET /api/sms/messages`) → `POST /api/verify` | `201` on register; `200` on verify; user `active` | Playwright `REG-001` |
| REG-002 | API | Missing required field | None | `POST /api/register` with empty `username` or missing `country` / `password` / `phone_number` | `400` | Playwright `REG-002` |
| REG-003 | API | Duplicate username | User A exists | Register twice with same `username`, different phones | First `201`; second `409` | Playwright `REG-003` |
| REG-004 | API | Duplicate phone | User A exists | Register twice with same `phone_number`, different usernames | First `201`; second `409` | Playwright `REG-004` |
| REG-005 | API / E2E | SMS transient simulation then success | Header `x-test-sms-mode: transient-failure-once` | `POST /api/register` with header → verify with OTP from messages | `201`; verify `200`; logs may show transient marker | Playwright `REG-005` |
| REG-006 | API | SMS hard failure | Header `x-test-sms-mode: always-fail` | `POST /api/register` with header | `500`; user may exist pending (strict error path) | Playwright `REG-006` |
| REG-007 | API | Incorrect OTP | Pending user after `201` | `POST /api/verify` with wrong `otp_code` | `401` | Playwright `REG-007` |
| REG-008 | API | Expired OTP | Short TTL e.g. `otp_ttl_seconds: 1` | Register → wait past TTL → verify with real OTP | `410` with OTP expired semantics | Playwright `REG-008` |
| REG-009 | API | Reuse OTP after success | User verified once | Verify again with same OTP | `400` (already active) or equivalent | Playwright `REG-009` |
| REG-010 | API / policy | SMS rate limit per phone (10/hour) | Policy implemented | Same phone, many registers | `429` or policy status when enforced; else document as future | Playwright `REG-010` (adaptive if not implemented) |
| REG-011 | API / policy | Throttle by IP | Policy implemented | Same `X-Forwarded-For`, many registers | `429` when enforced | Playwright `REG-011` |
| REG-012 | API / policy | Throttle by device fingerprint | Policy + header supported | Same device header, many registers | `429` when enforced | Playwright `REG-012` |
| REG-013 | API | Verify after session consumed | After successful verify | Second verify with wrong OTP | `400` / `401` as implemented | Playwright `REG-013` |
| REG-014 | API / data | Data integrity after verify | Successful flow | Assert DB: unique fields; password hashed; no plain password in API | Pass | Playwright `REG-014` |
| REG-015 | Observability | Audit / metrics | Optional `PW_TEST_AUDIT_ENDPOINT` or `GET /api/sms/stats` | Run success path | Events or stats available | Playwright `REG-015` |

### Priority (registration)

| Tier | IDs |
| --- | --- |
| P0 | REG-001, REG-002, REG-003, REG-004, REG-007, REG-008 |
| P1 | REG-005, REG-006, REG-009, REG-013, REG-014 |
| P2 | REG-010, REG-011, REG-012, REG-015 |

---

## 2. SMS core (`internal/sms`)

| ID | Layer | Scenario | Preconditions | Steps | Expected outcome |
| --- | --- | --- | --- | --- | --- |
| SMS-LC-001 | Unit | Valid lifecycle happy path | In-mem repo + mock provider | `SendSMS` for resolvable number | `New` → `Send-to-provider` → `Queue` |
| SMS-LC-002 | Unit | Carrier unknown | Number not matching resolver | `SendSMS` | Fails; message ends in failed path per service |
| SMS-LC-003 | Unit | No routing rule | Country+carrier not in router | `SendSMS` | Routing error; failed transition |
| SMS-LC-004 | Unit | Callback order valid | Message in `Queue` | `HandleCallback` with `Send-to-carrier` then success states per design | Transitions succeed |
| SMS-LC-005 | Unit | Callback skip step | Message in `Queue` | `HandleCallback` directly to `Send-success` | `updateStatus` error; **no** invalid persist |
| SMS-LC-006 | Unit | Callback after terminal | Message `Send-success` | Any non-identity callback | Invalid transition error |
| SMS-LC-007 | Integration | `x-test-sms-mode: always-fail` on send | Auth or `POST /api/sms/send` with header | Send | Provider path forced failure |
| SMS-LC-008 | Integration | `x-test-sms-mode: always-success` | Provider returns error in test | Send with header | Message reaches `Queue` via bypass log path |

### Priority (SMS core)

| Tier | IDs |
| --- | --- |
| P0 | SMS-LC-001, SMS-LC-002, SMS-LC-005, SMS-LC-006 |
| P1 | SMS-LC-003, SMS-LC-004, SMS-LC-007, SMS-LC-008 |

---

## 3. Message status transitions (`IsValidTransition`)

**Summary (scenarios and recovery):** see [SPECS_KIT_SMS_INVALID_TRANSITIONS_AND_RECOVERY.md](./SPECS_KIT_SMS_INVALID_TRANSITIONS_AND_RECOVERY.md).

| ID | Scenario | From → To (invalid) | Expected system behavior | Recovery / mitigation |
| --- | --- | --- | --- | --- |
| FSM-INV-001 | Skip carrier stage | `Queue` → `Send-success` | `HandleCallback` returns error; DB unchanged | Map webhook: enqueue intermediate state or ignore out-of-order event |
| FSM-INV-002 | Regression from carrier | `Send-to-carrier` → `Queue` | Invalid transition error | Reject callback; reconcile from provider timeline |
| FSM-INV-003 | Touch after success | `Send-success` → any other | Invalid transition error | Idempotent success handling; ignore duplicates |
| FSM-INV-004 | Wrong early state | `New` → `Send-success` | Invalid | Only `SendSMS` or allowed callbacks from current state |

Reference: `internal/sms/models.go` (`IsValidTransition`).

---

## 4. Cross-reference: Playwright vs matrix

| Playwright file | Covered IDs |
| --- | --- |
| `web/tests/e2e/registration/registration-success.spec.ts` | REG-001, REG-014 |
| `web/tests/e2e/registration/registration-validation.spec.ts` | REG-002, REG-003, REG-004 |
| `web/tests/e2e/registration/registration-otp.spec.ts` | REG-005, REG-006, REG-007, REG-008, REG-009, REG-013 |
| `web/tests/e2e/registration/registration-rate-limit.spec.ts` | REG-010, REG-011, REG-012, REG-015 |

---

## 5. Test data conventions

- **Philippines Globe-style numbers for resolver**: `63917` + 7 digits (see `PrefixCarrierResolver`).
- **Isolation**: unique `username` per case; unique `phone` unless testing duplicates.
- **OTP in automation**: `PW_TEST_OTP_ENDPOINT` optional; fallback `GET /api/sms/messages` parse `Your registration OTP is: (\d{6})`.
- **Single listener on `:8080`**: avoid duplicate `main` processes during E2E.

---

## 6. Revision history

| Version | Date | Notes |
| --- | --- | --- |
| 1.0 | 2026-04-10 | Initial matrix: registration, SMS core, FSM invalid paths, Playwright map |
| 1.1 | 2026-04-10 | Link to SPECS_KIT_SMS_INVALID_TRANSITIONS_AND_RECOVERY.md from Section 3 |
