// Package stagger provides a runner wrapper that introduces a random delay
// before executing the wrapped job. This helps distribute load when many cron
// jobs are scheduled at the same time (e.g. all at midnight).
package stagger

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Runner is the interface satisfied by any job executor.
type Runner interface {
	Run(ctx context.Context, name string, args []string, env []string) (int, error)
}

// Stagger wraps a Runner and sleeps for a random duration in [0, max) before
// delegating to the underlying runner.
type Stagger struct {
	runner Runner
	max    time.Duration
	sleep  func(context.Context, time.Duration) error
}

// New returns a Stagger that will delay up to max before running the job.
// max must be positive. sleep is injectable for testing; pass nil to use the
// default context-aware sleep.
func New(r Runner, max time.Duration, sleep func(context.Context, time.Duration) error) (*Stagger, error) {
	if r == nil {
		return nil, fmt.Errorf("stagger: runner must not be nil")
	}
	if max <= 0 {
		return nil, fmt.Errorf("stagger: max delay must be positive, got %s", max)
	}
	if sleep == nil {
		sleep = contextSleep
	}
	return &Stagger{runner: r, max: max, sleep: sleep}, nil
}

// Run sleeps for a random duration in [0, max) then delegates to the wrapped
// runner. If the context is cancelled during the sleep the function returns
// immediately with the context error.
func (s *Stagger) Run(ctx context.Context, name string, args []string, env []string) (int, error) {
	//nolint:gosec // non-cryptographic random is intentional here
	delay := time.Duration(rand.Int63n(int64(s.max)))
	if err := s.sleep(ctx, delay); err != nil {
		return 0, fmt.Errorf("stagger: context cancelled during delay: %w", err)
	}
	return s.runner.Run(ctx, name, args, env)
}

// contextSleep sleeps for d or until ctx is done, whichever comes first.
func contextSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
