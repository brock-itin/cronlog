package throttle

import (
	"fmt"
	"time"

	"github.com/user/cronlog/internal/history"
)

// Runner is the interface expected by Throttle to execute a job.
type Runner interface {
	Run() (int, error)
}

// Throttle prevents a job from running too frequently by enforcing a minimum
// interval between successful or any executions.
type Throttle struct {
	runner   Runner
	history  *history.History
	minGap   time.Duration
	jobName  string
}

// New returns a Throttle that wraps runner and refuses to run the job if the
// last recorded execution is more recent than minGap.
func New(r Runner, h *history.History, jobName string, minGap time.Duration) *Throttle {
	return &Throttle{
		runner:  r,
		history: h,
		minGap:  minGap,
		jobName: jobName,
	}
}

// ErrThrottled is returned when the job is suppressed due to the minimum gap
// not having elapsed since the last run.
type ErrThrottled struct {
	JobName   string
	LastRun   time.Time
	MinGap    time.Duration
	Remaining time.Duration
}

func (e *ErrThrottled) Error() string {
	return fmt.Sprintf(
		"throttle: job %q ran %s ago, minimum gap is %s (%s remaining)",
		e.JobName,
		time.Since(e.LastRun).Round(time.Second),
		e.MinGap,
		e.Remaining.Round(time.Second),
	)
}

// Run checks the job history and either delegates to the wrapped runner or
// returns ErrThrottled without executing the job.
func (t *Throttle) Run() (int, error) {
	if t.minGap <= 0 {
		return t.runner.Run()
	}

	entries := t.history.All()
	if len(entries) > 0 {
		last := entries[len(entries)-1]
		elapsed := time.Since(last.StartedAt)
		if elapsed < t.minGap {
			return -1, &ErrThrottled{
				JobName:   t.jobName,
				LastRun:   last.StartedAt,
				MinGap:    t.minGap,
				Remaining: t.minGap - elapsed,
			}
		}
	}

	return t.runner.Run()
}
