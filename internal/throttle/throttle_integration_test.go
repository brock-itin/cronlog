package throttle_test

import (
	"errors"
	"testing"
	"time"

	"github.com/user/cronlog/internal/history"
	"github.com/user/cronlog/internal/throttle"
)

// TestThrottle_RunTwiceInSuccession verifies that two back-to-back calls with a
// non-zero gap result in the second call being throttled, while the first
// succeeds, and that the history file is properly updated between runs.
func TestThrottle_RunTwiceInSuccession(t *testing.T) {
	p := t.TempDir() + "/hist.json"
	h, err := history.New(p, 50)
	if err != nil {
		t.Fatalf("history.New: %v", err)
	}

	r := &fakeRunner{exitCode: 0}
	th := throttle.New(r, h, "integration-job", 10*time.Minute)

	// First run — should succeed.
	code, err := th.Run()
	if err != nil {
		t.Fatalf("first run: unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("first run: expected exit 0, got %d", code)
	}

	// Record the first run in history so the throttle can see it.
	if addErr := h.Add(history.Entry{
		StartedAt: time.Now(),
		ExitCode:  0,
		Duration:  500 * time.Millisecond,
	}); addErr != nil {
		t.Fatalf("h.Add: %v", addErr)
	}

	// Second run — should be throttled.
	_, err = th.Run()
	var te *throttle.ErrThrottled
	if !errors.As(err, &te) {
		t.Fatalf("second run: expected ErrThrottled, got %T: %v", err, err)
	}
	if te.Remaining <= 0 {
		t.Errorf("expected positive remaining duration, got %s", te.Remaining)
	}
	if r.calls != 1 {
		t.Errorf("runner should have been called exactly once, got %d", r.calls)
	}
}
