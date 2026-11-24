package idle

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestIdleAfterTimeout(t *testing.T) {
	clock := newFakeClock()
	enters, exits := 0, 0
	c := newController(true, 10*time.Minute, testLogger(), func() { enters++ }, func() { exits++ }, clock.Now)

	c.maybeIdle()
	if c.State().Idle {
		t.Fatal("idled before timeout")
	}

	clock.Advance(9*time.Minute + 59*time.Second)
	c.maybeIdle()
	if c.State().Idle {
		t.Fatal("idled before timeout")
	}

	clock.Advance(time.Second)
	c.maybeIdle()
	state := c.State()
	if !state.Idle {
		t.Fatal("did not idle after timeout")
	}
	if enters != 1 || exits != 0 {
		t.Fatalf("hooks = %d/%d, want 1/0", enters, exits)
	}
	if state.IdleSince.IsZero() {
		t.Fatal("idle since is zero")
	}

	c.maybeIdle()
	if enters != 1 {
		t.Fatalf("entered idle %d times", enters)
	}
}

func TestBeginReactivates(t *testing.T) {
	clock := newFakeClock()
	enters, exits := 0, 0
	c := newController(true, time.Minute, testLogger(), func() { enters++ }, func() { exits++ }, clock.Now)

	clock.Advance(time.Minute)
	c.maybeIdle()
	if !c.State().Idle {
		t.Fatal("expected idle")
	}

	done := c.Begin()
	state := c.State()
	if state.Idle {
		t.Fatal("Begin did not reactivate")
	}
	if state.InFlight != 1 {
		t.Fatalf("in flight = %d", state.InFlight)
	}
	if exits != 1 {
		t.Fatalf("exits = %d", exits)
	}

	done()
	if got := c.State().InFlight; got != 0 {
		t.Fatalf("in flight after done = %d", got)
	}
	if c.State().Idle {
		t.Fatal("unexpected idle after done")
	}
	if enters != 1 {
		t.Fatalf("enters = %d", enters)
	}
}

func TestInFlightPreventsIdle(t *testing.T) {
	clock := newFakeClock()
	c := newController(true, time.Minute, testLogger(), nil, nil, clock.Now)

	done := c.Begin()
	clock.Advance(time.Hour)
	c.maybeIdle()
	if c.State().Idle {
		t.Fatal("idled with a request in flight")
	}

	done()
	clock.Advance(time.Minute)
	c.maybeIdle()
	if !c.State().Idle {
		t.Fatal("did not idle after the request completed")
	}
}

func TestDisabledNeverIdles(t *testing.T) {
	clock := newFakeClock()
	c := newController(false, time.Minute, testLogger(), nil, nil, clock.Now)

	clock.Advance(time.Hour)
	c.maybeIdle()
	state := c.State()
	if state.Enabled {
		t.Fatal("enabled should be false")
	}
	if state.Idle {
		t.Fatal("disabled controller idled")
	}

	done := c.Begin()
	done()
	if c.State().Idle {
		t.Fatal("disabled controller idled after Begin")
	}
}

func TestZeroTimeoutDisables(t *testing.T) {
	c := New(true, 0, testLogger(), nil, nil)
	if c.State().Enabled {
		t.Fatal("zero timeout should disable idle")
	}
}

func TestConcurrentBeginDone(t *testing.T) {
	c := newController(true, time.Minute, testLogger(), nil, nil, time.Now)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			done := c.Begin()
			done()
		}()
	}
	wg.Wait()

	if got := c.State().InFlight; got != 0 {
		t.Fatalf("in flight = %d, want 0", got)
	}
}
