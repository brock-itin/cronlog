// Package oncall provides a runner decorator that skips execution
// outside of a defined time-of-day window. This is useful for cron jobs
// that should only run during business hours or on-call periods.
package oncall

import (
	"context"
	"fmt"
	"time"
)

// Runner is the interface expected by New.
type Runner interface {
	Run(ctx context.Context, name string, args []string) (int, error)
}

// Oncall wraps a Runner and skips execution when the current time
// falls outside the configured window.
type Oncall struct {
	runner Runner
	start  int // hour (0-23), inclusive
	end    int // hour (0-23), exclusive
	clock  func() time.Time
}

// New returns an Oncall decorator. start and end are wall-clock hours
// in 24-hour format (e.g. start=9, end=17 means 09:00–17:00 local time).
// start must be strictly less than end.
func New(runner Runner, start, end int) (*Oncall, error) {
	if runner == nil {
		return nil, fmt.Errorf("oncall: runner must not be nil")
	}
	if start < 0 || start > 23 || end < 0 || end > 23 {
		return nil, fmt.Errorf("oncall: start and end must be in range 0-23")
	}
	if start >= end {
		return nil, fmt.Errorf("oncall: start (%d) must be less than end (%d)", start, end)
	}
	return &Oncall{
		runner: runner,
		start:  start,
		end:    end,
		clock:  time.Now,
	}, nil
}

// Run delegates to the wrapped runner only when the current hour falls
// within [start, end). Otherwise it returns (0, nil) silently.
func (o *Oncall) Run(ctx context.Context, name string, args []string) (int, error) {
	hour := o.clock().Hour()
	if hour < o.start || hour >= o.end {
		return 0, nil
	}
	return o.runner.Run(ctx, name, args)
}
