package retry

import (
	"context"
	"fmt"
	"time"
)

// Policy defines how retries are attempted.
type Policy struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     float64 // multiplier applied to delay after each attempt
}

// Result holds the outcome of a retry sequence.
type Result struct {
	Attempts int
	Err      error
}

// Runner executes a function according to a retry policy.
type Runner struct {
	policy Policy
	sleep  func(time.Duration)
}

// New returns a Runner with the given policy.
func New(p Policy) *Runner {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}
	if p.Backoff <= 0 {
		p.Backoff = 1.0
	}
	return &Runner{
		policy: p,
		sleep:  time.Sleep,
	}
}

// Run calls fn up to MaxAttempts times, stopping on nil error or context
// cancellation. It returns a Result describing how many attempts were made.
func (r *Runner) Run(ctx context.Context, fn func() error) Result {
	delay := r.policy.Delay
	var lastErr error

	for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return Result{Attempts: attempt - 1, Err: fmt.Errorf("context cancelled before attempt %d: %w", attempt, err)}
		}

		lastErr = fn()
		if lastErr == nil {
			return Result{Attempts: attempt}
		}

		if attempt < r.policy.MaxAttempts {
			r.sleep(delay)
			delay = time.Duration(float64(delay) * r.policy.Backoff)
		}
	}

	return Result{Attempts: r.policy.MaxAttempts, Err: lastErr}
}
