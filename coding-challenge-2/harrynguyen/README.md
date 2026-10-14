# SMS OTP

A production-ready SMS-based One-Time Password (OTP) service built with Go and React. It handles user registration, phone verification, and two-factor login by dispatching SMS messages through a pluggable multi-provider routing layer backed by a Redis/BullMQ job queue.

## Introduction

SMS OTP provides a complete auth flow:

1. **Register** — A user submits their username, password, phone number, and country. The API creates a pending account and fires a 6-digit OTP via SMS.
2. **Verify** — The user submits the OTP to activate their account.


![Architecture Diagram](./docs/sms-otp-architecture.png)


### Key features

- **Multi-provider SMS routing** — Twilio is wired in; Vonage, Infobip, AWS SNS, Telnyx, MessageBird, and Sinch are available as mock/drop-in adapters.
- **Async delivery via BullMQ** — SMS jobs are enqueued in Redis and processed by an in-process worker, keeping HTTP response times low.
- **Rate limiting** — Per-IP and per-device sliding-window counters (backed by Redis) protect registration and login endpoints.
- **Observability stack** — Prometheus metrics, Grafana dashboards, Loki log aggregation, and Promtail log shipping are all included in `docker-compose.yml`.
- **React frontend** — A Vite + React SPA served by Nginx acts as a demo UI.

### Tech stack

| Layer | Technology |
|---|---|
| API server | Go 1.25 · Echo v4 |
| Database | PostgreSQL 15 (GORM) |
| Queue | Redis 7 · BullMQ (gobullmq) |
| Frontend | React · Vite · Tailwind CSS |
| Metrics | Prometheus · Grafana |
| Logging | slog (structured) · Loki · Promtail |

---

## How to run locally with docker-compose

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) ≥ 24
- [Docker Compose](https://docs.docker.com/compose/install/) v2 (bundled with Docker Desktop)

### 1. Clone the repository

```bash
git clone https://github.com/dotdak/sms-otp.git
cd sms-otp
```

### 2. Configure environment variables

Copy the example file and adjust any values you need:

```bash
cp .env.example .env
```

The defaults work out of the box for a local stack. Relevant variables:

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://user:password@localhost:5432/sms_otp?sslmode=disable` | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis connection string |
| `QUEUE_NAME` | `sms_queue` | BullMQ queue name |
| `SMS_INTERNAL_API_KEY` | `dev-local-sms-key` | Bearer / `X-Internal-Key` for internal SMS endpoints |
| `ALLOW_TEST_SMS_HEADERS` | `true` | Enable `x-test-sms-mode` header for local testing |
| `LOG_REGISTRATION_OTP` | `true` | Print OTP codes to stdout (disable in production) |

### 3. Build and start all services

```bash
docker compose up --build
```

The first build compiles the Go binary and the React SPA — subsequent starts skip the compile step.

To run in detached mode:

```bash
docker compose up --build -d
```

### 4. Access the services

| Service | URL | Notes |
|---|---|---|
| **Frontend (SPA)** | http://localhost:5173 | React demo UI |
| **API** | http://localhost:8080 | Go REST API |
| **Grafana** | http://localhost:3000 | Login: `admin` / `admin` |
| **Prometheus** | http://localhost:9090 | Metrics scraper |
| **pgweb** | http://localhost:8081 | PostgreSQL browser UI |
| **Redis Insight** | http://localhost:5540 | Redis browser UI |
| **Loki** | http://localhost:3100 | Log aggregation (used by Grafana) |

### 5. Try the API

Because `LOG_REGISTRATION_OTP=true` is set by default, the generated OTP is printed to stdout — check `docker compose logs api`.

**Register a new user**

```bash
curl -s -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "secret",
    "phone_number": "09171234567",
    "country": "PH"
  }' | jq
```

**Verify the OTP** (grab the code from `docker compose logs api`)

```bash
curl -s -X POST http://localhost:8080/api/verify \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "otp_code": "123456"}' | jq
```

**Login**

```bash
curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret"}' | jq
```

**Complete login with OTP**

```bash
curl -s -X POST http://localhost:8080/api/login/verify \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "otp_code": "654321"}' | jq
```

### 6. Stop the stack

```bash
docker compose down
```

To also remove all persistent volumes (database, Redis, Grafana data):

```bash
docker compose down -v
```

---

## Stress tests
`go run cmd/stressworker/main.go` — a long-running worker that fires endless mixed scenarios (register/verify flows, internal SMS sends, chaotic callback sequences, callback storms, and garbage callbacks) against the running API stack. Requires `docker compose up` and `ALLOW_TEST_SMS_HEADERS=true`. Run with: `SMS_INTERNAL_API_KEY=dev-local-sms-key go run ./cmd/stressworker -base http://127.0.0.1:8080`

### Prerequisites

- Go 1.25+
- No running Docker stack required (Redis is replaced by [miniredis](https://github.com/alicebob/miniredis) for rate-limit tests)

This executes:
1. Resilience + load functional tests
2. Race-detector pass on the concurrent SendSMS test
3. Callback recovery invariant tests
4. A 500ms parallel throughput benchmark

### Run individual test groups

**Resilience and load tests (verbose)**

```bash
go test ./internal/sms/... -run '^(TestResilience|TestLoad)_' -count=1 -v
```

**Race-detector check (concurrent 256-goroutine SendSMS)**

```bash
go test ./internal/sms/... -race -run '^TestResilience_ConcurrentSendSMS$' -count=1
```

**Callback recovery invariants**

```bash
go test ./internal/sms/... -run '^TestApplyStatusWithRecovery' -count=1 -v
```

**Skip long-running tests**

Pass `-short` to skip all concurrency and load tests and only run the fast unit tests:

```bash
go test ./internal/sms/... -short
```

### Benchmarks

**Quick sample (500ms wall time)**

```bash
go test ./internal/sms/... -run '^$' -bench '^BenchmarkSendSMS_Parallel$' -benchmem -benchtime=500ms -count=1
```

**Stricter multi-run benchmark**

```bash
go test ./internal/sms/... -bench '^BenchmarkSendSMS' -benchmem -count=5
```

### What each test proves

| Test | What it verifies |
|---|---|
| `TestResilience_FlakyProviderEventualSuccess` | After N consecutive provider failures the service recovers and succeeds on the next calls |
| `TestResilience_FailedSendLeavesRecoverableState` | A failed provider call leaves the message in a retryable `Send-to-provider` state |
| `TestResilience_ConcurrentSendSMS` | 256 goroutines sending in parallel produce zero errors and no data races |
| `TestResilience_ConcurrentDeliveryCallbacks` | 64 duplicate delivery webhooks fired concurrently reach a consistent `Send-success` final state without panics |
| `TestResilience_RedisRateLimitUnderConcurrency` | Per-phone rate limits are enforced exactly (no over- or under-counting) under parallel load |
| `TestLoad_SustainedThroughput` | 400 sequential sends complete without error; throughput (msg/s) is logged |
| `BenchmarkSendSMS_Parallel` | Steady-state parallel throughput and allocations-per-op reported by `testing.B` |

### Interpreting the output

Running `./scripts/prove-resilience.sh` produces structured evidence across four phases. Here is how to read each one.

#### Phase 1 — Resilience + load tests

A healthy run looks like this:

```
=== RUN   TestResilience_FlakyProviderEventualSuccess
--- PASS: TestResilience_FlakyProviderEventualSuccess (0.00s)
=== RUN   TestResilience_ConcurrentSendSMS
--- PASS: TestResilience_ConcurrentSendSMS (0.00s)
=== RUN   TestResilience_RedisRateLimitUnderConcurrency
--- PASS: TestResilience_RedisRateLimitUnderConcurrency (0.11s)
=== RUN   TestLoad_SustainedThroughput
    resilience_test.go:252: 400 sequential SendSMS in 776.416µs (515188 msg/s)
--- PASS: TestLoad_SustainedThroughput (0.00s)
```

Key signals:

- **`TestResilience_FlakyProviderEventualSuccess` PASS** — the service survived N injected provider failures and resumed normal operation on subsequent calls. No sends are silently dropped.
- **`TestResilience_ConcurrentSendSMS` PASS** — all 256 goroutines completed without error and every message reached `Queue` state. The repository has no hidden data races under concurrent writes.
- **`TestResilience_RedisRateLimitUnderConcurrency` PASS (≈ 0.11s)** — the miniredis-backed limiter allowed exactly 10 sends for the shared phone and rejected the remaining 4. No over-counting (which would allow too many) and no under-counting (which would block legitimate sends).
- **`515 188 msg/s` log line** — sequential throughput baseline. A significant drop vs this number after a code change is a performance regression signal.

**`TestResilience_ConcurrentDeliveryCallbacks`** generates intentional log noise. You will see a mix of:

```
level=INFO  msg=sms_callback_dispatch   step=dispatch  previous_status=Queue  callback_status=Send-success
level=INFO  msg=sms_recovery_event      recovery_kind=recovered_queue_to_delivered
level=INFO  msg=sms_callback_applied    step=applied   final_status=Send-success
level=ERROR msg=sms_callback_apply_failed  err="invalid status transition from Send-success to Send-to-carrier"
```

This is **expected and correct**. The sequence proves:

1. The first callback to arrive transitions the message from `Queue` → `Send-success` (logged as `recovered_queue_to_delivered`).
2. All subsequent duplicate callbacks hit the already-terminal `Send-success` state and are rejected with an `invalid status transition` error — they do not corrupt the final state.
3. The test asserts the final status is still `Send-success`. A PASS confirms idempotent callback handling.

#### Phase 2 — Race-detector pass

```
ok  github.com/dotdak/sms-otp/internal/sms  1.554s
```

No output beyond the `ok` line means Go's race detector found **zero data races** across all 256 concurrent goroutines. Any data race would print a `DATA RACE` block with goroutine stack traces and cause a FAIL.

#### Phase 3 — Callback recovery invariants

Each test name describes a specific out-of-order or invalid webhook scenario:

| Test suffix | Scenario | Recovery kind logged | Expected outcome |
|---|---|---|---|
| `QueueSkipsSendToCarrier` | `Send-success` arrives while message is still in `Queue` | `recovered_queue_to_delivered` | Final status `Send-success` |
| `SendToProviderSkipsToDelivered` | `Send-success` arrives while in `Send-to-provider` | `recovered_provider_to_delivered` | Final status `Send-success` |
| `MitigateLateFailureAfterSuccess` | `Send-failed` arrives after `Send-success` | `mitigation` | Final status `Send-success` (late failure ignored) |
| `CarrierRejectedAtSendToProvider` | `Carrier-rejected` arrives at `Send-to-provider` | `mapped_carrier_reject` | Status stays `Send-to-provider` (logged for retry) |
| `SendToProviderMissingMessageID` | Recovery impossible — no provider message ID | `recovery_failed_missing_provider_id` | Error returned, no silent data loss |
| `UnrecoverableReturnsError` | Transition `New` → `Send-success` with no recovery path | `unrecoverable` | Error returned explicitly |

All six tests passing confirms the state machine correctly handles every real-world webhook edge case without panicking or silently corrupting state.

#### Phase 4 — Benchmark

```
BenchmarkSendSMS_Parallel-14   147354   3774 ns/op   1808 B/op   42 allocs/op
```

- **`147 354` iterations** completed in 500ms wall time on an Apple M4 Pro (14 goroutines).
- **`3 774 ns/op`** — each `SendSMS` round-trip (resolve carrier → route provider → persist → enqueue) completes in under 4 µs.
- **`42 allocs/op`** — allocations per call; a useful baseline to catch allocation regressions in hot-path changes.
