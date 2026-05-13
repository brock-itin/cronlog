// Package debounce suppresses repeated job executions within a minimum
// interval, preventing alert storms when a flapping cron fires too frequently.
package debounce

import (
	"fmt"
	"time"

	"github.com/cronlog/internal/checkpoint"
)

// Runner is the interface satisfied by any runnable job.
type Runner interface {
	Run(args []string) (int, error)
}

// Debouncer wraps a Runner and skips execution if the job ran too recently.
type Debouncer struct {
	runner   Runner
	interval time.Duration
	cp       *checkpoint.Checkpoint
}

// New returns a Debouncer that enforces a minimum interval between runs.
// cp must be a non-nil Checkpoint used to persist the last execution time.
func New(r Runner, interval time.Duration, cp *checkpoint.Checkpoint) (*Debouncer, error) {
	if r == nil {
		return nil, fmt.Errorf("debounce: runner must not be nil")
	}
	if cp == nil {
		return nil, fmt.Errorf("debounce: checkpoint must not be nil")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("debounce: interval must be positive")
	}
	return &Debouncer{runner: r, interval: interval, cp: cp}, nil
}

// Run executes the wrapped job only if at least d.interval has elapsed since
// the last successful run. If the job is suppressed, Run returns (0, nil).
func (d *Debouncer) Run(args []string) (int, error) {
	if d.cp.Overdue(d.interval) {
		code, err := d.runner.Run(args)
		if err == nil || code != 0 {
			// Persist timestamp regardless of exit code so a failing job
			// does not spam retries within the debounce window.
			_ = d.cp.Save(time.Now())
		}
		return code, err
	}
	return 0, nil
}
