package watchdog

import (
	"fmt"
	"time"
)

// Runner is the interface for executing a cron job.
type Runner interface {
	Run(name string, args ...string) (int, error)
}

// Notifier is called when the watchdog detects a missed or overdue run.
type Notifier interface {
	Notify(jobName string, msg string) error
}

// Checkpoint tracks the last known successful run time.
type Checkpoint interface {
	Load(key string) (time.Time, bool)
	Save(key string, t time.Time) error
}

// Watchdog wraps a Runner and alerts when a job has not run within the
// expected interval. It is intended to catch silent cron failures where
// the job is never scheduled rather than failing with a non-zero exit.
type Watchdog struct {
	runner     Runner
	checkpoint Checkpoint
	notifier   Notifier
	interval   time.Duration
	clock      func() time.Time
}

// New creates a Watchdog. interval is the maximum acceptable gap between
// successful runs. notifier may be nil to disable alerting.
func New(r Runner, cp Checkpoint, n Notifier, interval time.Duration) (*Watchdog, error) {
	if r == nil {
		return nil, fmt.Errorf("watchdog: runner must not be nil")
	}
	if cp == nil {
		return nil, fmt.Errorf("watchdog: checkpoint must not be nil")
	}
	if interval <= 0 {
		return nil, fmt.Errorf("watchdog: interval must be positive")
	}
	return &Watchdog{
		runner:     r,
		checkpoint: cp,
		notifier:   n,
		interval:   interval,
		clock:      time.Now,
	}, nil
}

// Run executes the job identified by name+args, records a checkpoint on
// success, and fires the notifier if the previous checkpoint is overdue.
func (w *Watchdog) Run(name string, args ...string) (int, error) {
	now := w.clock()

	if last, ok := w.checkpoint.Load(name); ok {
		if now.Sub(last) > w.interval && w.notifier != nil {
			msg := fmt.Sprintf("job %q last ran %s ago (threshold %s)",
				name, now.Sub(last).Round(time.Second), w.interval)
			_ = w.notifier.Notify(name, msg)
		}
	}

	code, err := w.runner.Run(name, args...)
	if err != nil {
		return code, err
	}
	if code == 0 {
		_ = w.checkpoint.Save(name, now)
	}
	return code, nil
}
