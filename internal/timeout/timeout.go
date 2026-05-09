// Package timeout provides a wrapper for enforcing a maximum execution
// duration on cron job commands. If the command exceeds the deadline,
// it is killed and an error is returned.
package timeout

import (
	"context"
	"fmt"
	"time"

	"github.com/user/cronlog/internal/runner"
)

// Runner wraps a runner.Runner with a configurable timeout.
type Runner struct {
	inner    *runner.Runner
	duration time.Duration
}

// New returns a Runner that cancels execution after d.
// If d is zero or negative, no timeout is applied.
func New(r *runner.Runner, d time.Duration) *Runner {
	return &Runner{inner: r, duration: d}
}

// Duration returns the configured timeout duration.
// A value of zero or negative means no timeout is applied.
func (t *Runner) Duration() time.Duration {
	return t.duration
}

// Run executes the command, killing it if the timeout elapses.
// It returns the exit code, combined output, and any error.
func (t *Runner) Run(name string, args ...string) (int, []byte, error) {
	if t.duration <= 0 {
		return t.inner.Run(name, args...)
	}

	type result struct {
		code int
		out  []byte
		err  error
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.duration)
	defer cancel()

	ch := make(chan result, 1)
	go func() {
		code, out, err := t.inner.Run(name, args...)
		ch <- result{code, out, err}
	}()

	select {
	case res := <-ch:
		return res.code, res.out, res.err
	case <-ctx.Done():
		return -1, nil, fmt.Errorf("command timed out after %s", t.duration)
	}
}
