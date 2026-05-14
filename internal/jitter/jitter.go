// Package jitter wraps a Runner and introduces a random delay before
// each execution, spreading load when many cron jobs fire simultaneously.
package jitter

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Runner is the interface expected by Jitter.
type Runner interface {
	Run(ctx context.Context, name string, args []string) (int, error)
}

// Jitter holds configuration for the random pre-execution delay.
type Jitter struct {
	runner Runner
	max    time.Duration
	sleep  func(context.Context, time.Duration) error
}

// New returns a Jitter that waits a random duration in [0, max) before
// delegating to runner. max must be positive.
func New(runner Runner, max time.Duration) (*Jitter, error) {
	if runner == nil {
		return nil, fmt.Errorf("jitter: runner must not be nil")
	}
	if max <= 0 {
		return nil, fmt.Errorf("jitter: max must be positive, got %s", max)
	}
	return &Jitter{
		runner: runner,
		max:    max,
		sleep:  contextSleep,
	}, nil
}

// Run waits a random duration up to j.max, then delegates to the wrapped runner.
// If the context is cancelled during the delay, Run returns immediately with the
// context error.
func (j *Jitter) Run(ctx context.Context, name string, args []string) (int, error) {
	delay := time.Duration(rand.Int63n(int64(j.max)))
	if err := j.sleep(ctx, delay); err != nil {
		return 0, fmt.Errorf("jitter: context cancelled during delay: %w", err)
	}
	return j.runner.Run(ctx, name, args)
}

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
