#!/usr/bin/env bash
# Evidence suite: flaky external provider, concurrent load, Redis-backed rate limits, and parallel throughput.
# Usage: from repo root, ./scripts/prove-resilience.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== 1) Resilience + load tests (functional proof) =="
go test ./internal/sms/... -run '^(TestResilience|TestLoad)_' -count=1 -v

echo ""
echo "== 2) Data races on concurrent SendSMS (sample) =="
go test ./internal/sms/... -race -run '^TestResilience_ConcurrentSendSMS$' -count=1

echo ""
echo "== 3) Callback recovery invariants (status_recovery) =="
go test ./internal/sms/... -run '^TestApplyStatusWithRecovery' -count=1 -v

echo ""
echo "== 4) Parallel throughput benchmark (sample, 500ms) =="
go test ./internal/sms/... -run '^$' -bench '^BenchmarkSendSMS_Parallel$' -benchmem -benchtime=500ms -count=1

echo ""
echo "Done. For stricter benchmarks: go test ./internal/sms/... -bench '^BenchmarkSendSMS' -benchmem -count=5"
