// Package cooldown provides a runner decorator that enforces a minimum
// interval between consecutive executions of a job, regardless of outcome.
// Unlike throttle (which skips based on last success), cooldown prevents
// rapid re-runs after any execution — useful for avoiding thundering-herd
// scenarios when a cron fires more frequently than intended.
package cooldown

import (
	"fmt"
	"time"

	"github.com/owner/cronlog/internal/checkpoint"
)

// Runner is the interface satisfied by any executable unit.
type Runner interface {
	Run(name string, args []string) (int, error)
}

// Cooldown wraps a Runner and enforces a minimum wait between runs.
type Cooldown struct {
	runner   Runner
	cp       *checkpoint.Checkpoint
	interval time.Duration
	now      func() time.Time
}

// New creates a Cooldown decorator.
// interval must be positive; runner and cp must be non-nil.
func New(runner Runner, cp *checkpoint.Checkpoint, interval time.Duration) (*Cooldown, error) {
	if runner == nil {
		return nil, fmt.Errorf("cooldown: runner must not be nil")
	}
	if cp == nil {
		return nil, fmt.Errorf("cooldown: checkpoint must not be nil")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("cooldown: interval must be positive, got %v", interval)
	}
	return &Cooldown{
		runner:   runner,
		cp:       cp,
		interval: interval,
		now:      time.Now,
	}, nil
}

// Run executes the wrapped runner only if the cooldown interval has elapsed
// since the last recorded run. If the job is still cooling down, Run returns
// exit code 0 and a descriptive error indicating the skip.
func (c *Cooldown) Run(name string, args []string) (int, error) {
	last, ok := c.cp.Load()
	if ok {
		elapsed := c.now().Sub(last)
		if elapsed < c.interval {
			remaining := c.interval - elapsed
			return 0, fmt.Errorf("cooldown: job %q is cooling down, %v remaining", name, remaining.Truncate(time.Second))
		}
	}

	code, err := c.runner.Run(name, args)

	// Record the run time regardless of outcome so the cooldown applies
	// even after failures.
	if saveErr := c.cp.Save(c.now()); saveErr != nil {
		if err == nil {
			err = fmt.Errorf("cooldown: failed to save checkpoint: %w", saveErr)
		}
	}

	return code, err
}
