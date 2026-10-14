# prove-resilience.sh — What value does this script provide?

`prove-resilience.sh` is an evidence suite, not a standard unit-test run.
Its purpose is to demonstrate that the SMS delivery pipeline holds its correctness and
performance guarantees under the four failure modes that matter most in production:
flaky upstream providers, concurrent goroutine access, out-of-order delivery callbacks,
and Redis-backed rate limiting under contention.

Run it from the repo root:

```bash
./scripts/prove-resilience.sh
```

No external services are required. Redis is replaced by an in-process miniredis instance
and all SMS providers use a zero-delay mock.

---

## The four proofs

### 1. Functional resilience + load (`TestResilience_*`, `TestLoad_*`)

**What problem it guards against:** a provider outage, a slow provider, or high send volume
silently dropping messages or leaving them in an unrecoverable state.

| Test | Claim it proves |
|---|---|
| `TestResilience_FlakyProviderEventualSuccess` | After N consecutive upstream errors the service self-recovers; no sends are silently swallowed |
| `TestResilience_FailedSendLeavesRecoverableState` | A provider failure moves the message to `Send-to-provider` (retryable), never to a terminal error state |
| `TestResilience_ConcurrentSendSMS` | 256 parallel goroutines all succeed; no deadlocks, no lost writes |
| `TestResilience_ConcurrentDeliveryCallbacks` | 64 duplicate delivery webhooks for the same message converge to a single `Send-success`; subsequent duplicates are rejected with a transition error, not by corrupting state |
| `TestResilience_RedisRateLimitUnderConcurrency` | The sliding-window rate limiter allows exactly the configured quota and rejects the remainder — no over- or under-counting under concurrent access |
| `TestLoad_SustainedThroughput` | 400 sequential sends complete without error; the `msg/s` figure is logged as a performance regression baseline |

**Expected output:** all tests `PASS`; the sustained-throughput line logs a msg/s figure
(e.g. `515 188 msg/s` on an M4 Pro). A significant drop in that number after a change
signals a hot-path regression.

---

### 2. Race-detector pass (`-race`)

**What problem it guards against:** unsynchronised concurrent reads/writes to shared
state (repository, observer registry, cost tracker) that only appear under load and are
invisible to normal tests.

The race detector instruments every memory access. A passing run with `-race` is proof
that no goroutine pair accesses shared data without proper synchronisation.

**Expected output:** the `ok` line with no `DATA RACE` block. Any race prints a full
goroutine stack trace and fails the test.

---

### 3. Callback recovery invariants (`TestApplyStatusWithRecovery_*`)

**What problem it guards against:** real SMS providers send webhooks out of order, late,
or multiple times. Without explicit recovery rules the state machine would reject
legitimate transitions and leave messages stuck.

Each test covers one edge case that occurs regularly in production:

| Test suffix | Scenario | Recovery action | Outcome |
|---|---|---|---|
| `QueueSkipsSendToCarrier` | Delivery receipt arrives before carrier-accepted event | Skip intermediate state | `Send-success` |
| `SendToProviderSkipsToDelivered` | Delivery receipt arrives while still waiting for provider ACK | Skip intermediate state | `Send-success` |
| `MitigateLateFailureAfterSuccess` | `Send-failed` webhook arrives after `Send-success` already recorded | Ignore the late failure | `Send-success` preserved |
| `CarrierRejectedAtSendToProvider` | Carrier rejection arrives at `Send-to-provider` | Map to `Send-failed` for retry routing | `Send-to-provider` logged |
| `SendToProviderMissingMessageID` | Recovery impossible — no provider message ID to correlate | Return explicit error, no silent data loss | Error |
| `UnrecoverableReturnsError` | No recovery path exists for this transition | Return explicit error | Error |

**Expected output:** all six tests `PASS`. The verbose logs will show `sms_recovery_event`
lines with a `recovery_kind` field naming the recovery path taken — these are not errors,
they are the recovery logic working as designed.

---

### 4. Parallel throughput benchmark (`BenchmarkSendSMS_Parallel`)

**What problem it guards against:** unnoticed performance regressions on the hot path
(carrier resolution → provider routing → repository write → queue publish).

The benchmark runs for 500ms using `b.RunParallel` and reports:

- **`ns/op`** — latency per `SendSMS` call end-to-end
- **`B/op`** — heap bytes allocated per call
- **`allocs/op`** — number of allocations per call

**Expected output (Apple M4 Pro, 14 cores):**

```
BenchmarkSendSMS_Parallel-14   147354   3774 ns/op   1808 B/op   42 allocs/op
```

Use `count=5` for a statistically stable comparison across branches:

```bash
go test ./internal/sms/... -bench '^BenchmarkSendSMS' -benchmem -count=5
```

---

## When to re-run this script

- Before merging any change to `internal/sms/`, `internal/ratelimit/`, or `repository/`
- After updating provider adapters or adding a new SMS status transition
- When investigating a production incident involving dropped messages, duplicate deliveries, or rate-limit drift
