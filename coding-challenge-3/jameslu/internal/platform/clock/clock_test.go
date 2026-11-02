package clock_test

import (
	"testing"
	"time"

	"ridedispatch/internal/platform/clock"
)

func TestFakeClockFiresDueTimersInOrder(t *testing.T) {
	start := time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)
	c := clock.NewFake(start)

	var fired []string
	c.AfterFunc(30*time.Second, func() { fired = append(fired, "30s") })
	c.AfterFunc(10*time.Second, func() { fired = append(fired, "10s") })
	stopped := c.AfterFunc(20*time.Second, func() { fired = append(fired, "20s") })
	if !stopped.Stop() {
		t.Fatal("Stop on a pending timer should return true")
	}

	c.Advance(29 * time.Second)
	if len(fired) != 1 || fired[0] != "10s" {
		t.Fatalf("after 29s fired = %v", fired)
	}
	if !c.Now().Equal(start.Add(29 * time.Second)) {
		t.Fatalf("Now = %v", c.Now())
	}

	c.Advance(time.Second)
	if len(fired) != 2 || fired[1] != "30s" {
		t.Fatalf("after 30s fired = %v", fired)
	}
	if stopped.Stop() {
		t.Fatal("Stop on a stopped timer should return false")
	}
}

func TestFakeClockTimerCanScheduleAnother(t *testing.T) {
	c := clock.NewFake(time.Unix(0, 0))
	count := 0
	c.AfterFunc(time.Second, func() {
		count++
		c.AfterFunc(time.Second, func() { count++ })
	})
	c.Advance(2 * time.Second)
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}
