// Package quota enforces a maximum number of cron job runs within a
// rolling time window, rejecting executions that exceed the configured limit.
package quota

import (
	"context"
	"fmt"
	"time"

	"github.com/cronlog/internal/history"
)

// Runner is the interface satisfied by anything that can execute a cron job.
type Runner interface {
	Run(ctx context.Context, args []string) (int, error)
}

// Quota wraps a Runner and enforces a maximum execution count per window.
type Quota struct {
	runner  Runner
	history *history.History
	max     int
	window  time.Duration
	now     func() time.Time
}

// New returns a Quota that allows at most max runs within window.
// It returns an error if runner or history is nil, max is zero, or window is
// non-positive.
func New(runner Runner, h *history.History, max int, window time.Duration) (*Quota, error) {
	if runner == nil {
		return nil, fmt.Errorf("quota: runner must not be nil")
	}
	if h == nil {
		return nil, fmt.Errorf("quota: history must not be nil")
	}
	if max <= 0 {
		return nil, fmt.Errorf("quota: max must be greater than zero, got %d", max)
	}
	if window <= 0 {
		return nil, fmt.Errorf("quota: window must be positive, got %s", window)
	}
	return &Quota{
		runner:  runner,
		history: h,
		max:     max,
		window:  window,
		now:     time.Now,
	}, nil
}

// Run executes the wrapped runner only if the number of runs recorded in the
// history within the rolling window is below the configured maximum. If the
// quota is exceeded it returns (0, ErrQuotaExceeded) without running the job.
func (q *Quota) Run(ctx context.Context, args []string) (int, error) {
	cutoff := q.now().Add(-q.window)
	entries := q.history.Since(cutoff)
	if len(entries) >= q.max {
		return 0, fmt.Errorf("quota: limit of %d runs per %s exceeded (%d recorded)",
			q.max, q.window, len(entries))
	}
	return q.runner.Run(ctx, args)
}
