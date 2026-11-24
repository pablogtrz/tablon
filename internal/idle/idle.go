// Package idle tracks meaningful API activity and performs best-effort
// resource release when Tablon has been quiet for the configured timeout.
package idle

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	maxCheckInterval = 30 * time.Second
	minCheckInterval = 100 * time.Millisecond
)

// State is a point-in-time snapshot of the idle controller.
type State struct {
	Enabled      bool
	Idle         bool
	LastActivity time.Time
	IdleSince    time.Time
	Timeout      time.Duration
	InFlight     int
}

// Name returns the active/idle state used by the status endpoint.
func (s State) Name() string {
	if s.Idle {
		return "idle"
	}
	return "active"
}

// Controller is shared by the HTTP middleware and future background workers.
type Controller struct {
	enabled  bool
	timeout  time.Duration
	interval time.Duration
	now      func() time.Time
	logger   *slog.Logger
	onEnter  func()
	onExit   func()

	// transitionMu serializes enter/exit hooks with state transitions so an
	// onEnter can never run after a concurrent Begin has already exited idle.
	transitionMu sync.Mutex

	mu           sync.Mutex
	idle         bool
	idleSince    time.Time
	lastActivity time.Time
	inFlight     int
}

func New(enabled bool, timeout time.Duration, logger *slog.Logger, onEnter, onExit func()) *Controller {
	return newController(enabled, timeout, logger, onEnter, onExit, time.Now)
}

func newController(
	enabled bool,
	timeout time.Duration,
	logger *slog.Logger,
	onEnter, onExit func(),
	now func() time.Time,
) *Controller {
	if timeout <= 0 {
		enabled = false
	}
	interval := timeout / 4
	if interval < minCheckInterval {
		interval = minCheckInterval
	}
	if interval > maxCheckInterval {
		interval = maxCheckInterval
	}
	return &Controller{
		enabled:      enabled,
		timeout:      timeout,
		interval:     interval,
		now:          now,
		logger:       logger,
		onEnter:      onEnter,
		onExit:       onExit,
		lastActivity: now(),
	}
}

// Run checks for inactivity until ctx is cancelled. It is a no-op when idle
// mode is disabled.
func (c *Controller) Run(ctx context.Context) {
	if !c.enabled {
		return
	}
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.maybeIdle()
		}
	}
}

// Begin marks a meaningful API request as in flight and reactivates the
// controller before the handler runs. The returned function must be called
// once the request completes.
func (c *Controller) Begin() func() {
	c.transitionMu.Lock()
	c.mu.Lock()
	now := c.now()
	c.inFlight++
	c.lastActivity = now
	wasIdle := c.idle
	var idleFor time.Duration
	if wasIdle {
		c.idle = false
		idleFor = now.Sub(c.idleSince)
		c.idleSince = time.Time{}
	}
	c.mu.Unlock()

	if wasIdle {
		if c.logger != nil {
			c.logger.Info("leaving idle mode", "idle_for", idleFor)
		}
		if c.onExit != nil {
			c.onExit()
		}
	}
	c.transitionMu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			if c.inFlight > 0 {
				c.inFlight--
			}
			c.lastActivity = c.now()
			c.mu.Unlock()
		})
	}
}

// State returns a consistent snapshot for the status endpoint.
func (c *Controller) State() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	return State{
		Enabled:      c.enabled,
		Idle:         c.idle,
		LastActivity: c.lastActivity,
		IdleSince:    c.idleSince,
		Timeout:      c.timeout,
		InFlight:     c.inFlight,
	}
}

// Active reports whether background work may run. v0.2 GitHub polling must
// consult this before doing work.
func (c *Controller) Active() bool {
	return !c.State().Idle
}

func (c *Controller) maybeIdle() {
	c.transitionMu.Lock()
	defer c.transitionMu.Unlock()

	c.mu.Lock()
	if !c.enabled || c.idle || c.inFlight > 0 {
		c.mu.Unlock()
		return
	}
	now := c.now()
	idleFor := now.Sub(c.lastActivity)
	if idleFor < c.timeout {
		c.mu.Unlock()
		return
	}
	c.idle = true
	c.idleSince = now
	c.mu.Unlock()

	if c.logger != nil {
		c.logger.Info("entering idle mode", "timeout", c.timeout, "idle_for", idleFor)
	}
	if c.onEnter != nil {
		c.onEnter()
	}
}
