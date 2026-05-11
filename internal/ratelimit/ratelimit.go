// Package ratelimit provides a token-bucket style rate limiter that
// prevents a cron job from executing more than N times within a rolling
// time window.  Unlike throttle (which checks the last successful run)
// ratelimit counts every attempt — including failures — so a repeatedly
// crashing job cannot flood downstream systems.
package ratelimit

import (
	"fmt"
	"time"
)

// Runner is the interface satisfied by any execution back-end.
type Runner interface {
	Run() (int, error)
}

// Limiter wraps a Runner and enforces a maximum call rate.
type Limiter struct {
	runner   Runner
	max      int
	window   time.Duration
	now      func() time.Time
	timestamps []time.Time
}

// New returns a Limiter that allows at most maxRuns executions within
// window.  Pass a non-nil nowFn to override the clock in tests; pass nil
// to use time.Now.
func New(r Runner, maxRuns int, window time.Duration, nowFn func() time.Time) (*Limiter, error) {
	if r == nil {
		return nil, fmt.Errorf("ratelimit: runner must not be nil")
	}
	if maxRuns <= 0 {
		return nil, fmt.Errorf("ratelimit: maxRuns must be > 0, got %d", maxRuns)
	}
	if window <= 0 {
		return nil, fmt.Errorf("ratelimit: window must be > 0, got %s", window)
	}
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Limiter{
		runner: r,
		max:    maxRuns,
		window: window,
		now:    nowFn,
	}, nil
}

// Run executes the wrapped runner if the rate limit has not been reached.
// It returns (0, ErrRateLimited) when the limit is exceeded without
// calling the underlying runner.
func (l *Limiter) Run() (int, error) {
	now := l.now()
	cutoff := now.Add(-l.window)

	// Prune timestamps outside the current window.
	active := l.timestamps[:0]
	for _, ts := range l.timestamps {
		if ts.After(cutoff) {
			active = append(active, ts)
		}
	}
	l.timestamps = active

	if len(l.timestamps) >= l.max {
		return 0, fmt.Errorf("ratelimit: limit of %d runs per %s exceeded", l.max, l.window)
	}

	l.timestamps = append(l.timestamps, now)
	return l.runner.Run()
}
