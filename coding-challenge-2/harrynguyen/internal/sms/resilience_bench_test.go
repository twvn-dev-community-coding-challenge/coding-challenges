package sms

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/dotdak/sms-otp/internal/providers"
)

// BenchmarkSendSMS_Parallel measures steady-state throughput with a zero-delay mock provider.
// Run: go test ./internal/sms/... -bench '^BenchmarkSendSMS' -benchmem -count=5
func BenchmarkSendSMS_Parallel(b *testing.B) {
	ctx := context.Background()
	svc, _ := newResilienceService(b, fastMock(providers.ProviderMessageBird), nil)
	var seq atomic.Int64

	var nErr atomic.Int32
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := int(seq.Add(1))
			if _, err := svc.SendSMS(ctx, "PH", phGlobePhone(30_000+n), "bench"); err != nil {
				nErr.Add(1)
			}
		}
	})
	b.StopTimer()
	if nErr.Load() != 0 {
		b.Fatalf("SendSMS errors under parallel load: %d", nErr.Load())
	}
}
