// Package clock abstracts time so that offer timeouts can be tested by
// advancing a fake clock instead of sleeping.
package clock

import (
	"sort"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
	AfterFunc(d time.Duration, f func()) Timer
}

type Timer interface {
	// Stop prevents the timer from firing. It returns false if the timer has
	// already fired or been stopped.
	Stop() bool
}

// Real is the wall clock.
type Real struct{}

func (Real) Now() time.Time { return time.Now() }

func (Real) AfterFunc(d time.Duration, f func()) Timer { return time.AfterFunc(d, f) }

// Fake is a manually advanced clock. Timers fire synchronously inside Advance,
// in due-time order.
type Fake struct {
	mu     sync.Mutex
	now    time.Time
	timers []*fakeTimer
}

type fakeTimer struct {
	clock *Fake
	due   time.Time
	f     func()
	done  bool
}

func NewFake(now time.Time) *Fake { return &Fake{now: now} }

func (c *Fake) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *Fake) AfterFunc(d time.Duration, f func()) Timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := &fakeTimer{clock: c, due: c.now.Add(d), f: f}
	c.timers = append(c.timers, t)
	return t
}

// Advance moves time forward by d and runs every timer that becomes due,
// including timers scheduled by timers that fire during the advance.
func (c *Fake) Advance(d time.Duration) {
	c.mu.Lock()
	target := c.now.Add(d)
	c.mu.Unlock()

	for {
		c.mu.Lock()
		next := c.nextDue(target)
		if next == nil {
			c.now = target
			c.mu.Unlock()
			return
		}
		next.done = true
		c.now = next.due
		c.mu.Unlock()
		next.f()
	}
}

// nextDue returns the earliest pending timer due at or before target.
// Callers hold c.mu.
func (c *Fake) nextDue(target time.Time) *fakeTimer {
	pending := c.timers[:0]
	for _, t := range c.timers {
		if !t.done {
			pending = append(pending, t)
		}
	}
	c.timers = pending
	sort.SliceStable(c.timers, func(i, j int) bool { return c.timers[i].due.Before(c.timers[j].due) })
	if len(c.timers) == 0 || c.timers[0].due.After(target) {
		return nil
	}
	return c.timers[0]
}

func (t *fakeTimer) Stop() bool {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	if t.done {
		return false
	}
	t.done = true
	return true
}
