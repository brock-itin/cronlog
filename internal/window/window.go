// Package window provides a sliding-window runner that suppresses execution
// when the job has already run successfully within a configurable duration.
package window

import (
	"context"
	"fmt"
	"time"

	"github.com/example/cronlog/internal/history"
)

// Runner is the interface satisfied by any runnable unit.
type Runner interface {
	Run(ctx context.Context, args []string) (int, error)
}

// Recorder persists and retrieves run history.
type Recorder interface {
	Add(entry history.Entry) error
	Entries() ([]history.Entry, error)
}

// Window wraps a Runner and skips execution when a successful run already
// exists within the configured look-back window.
type Window struct {
	runner   Runner
	recorder Recorder
	window   time.Duration
	now      func() time.Time
}

// New creates a Window. window must be positive.
func New(r Runner, rec Recorder, window time.Duration) (*Window, error) {
	if r == nil {
		return nil, fmt.Errorf("window: runner must not be nil")
	}
	if rec == nil {
		return nil, fmt.Errorf("window: recorder must not be nil")
	}
	if window <= 0 {
		return nil, fmt.Errorf("window: window must be positive, got %s", window)
	}
	return &Window{runner: r, recorder: rec, window: window, now: time.Now}, nil
}

// Run delegates to the underlying runner only when no successful execution
// has been recorded within the configured window. When suppressed it returns
// exit code 0 and a nil error.
func (w *Window) Run(ctx context.Context, args []string) (int, error) {
	entries, err := w.recorder.Entries()
	if err != nil {
		return 0, fmt.Errorf("window: reading history: %w", err)
	}

	cutoff := w.now().Add(-w.window)
	for _, e := range entries {
		if e.ExitCode == 0 && e.StartedAt.After(cutoff) {
			// A recent successful run exists — skip.
			return 0, nil
		}
	}

	return w.runner.Run(ctx, args)
}
