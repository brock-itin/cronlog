package alert_test

import (
	"strings"
	"testing"
	"time"

	"github.com/cronlog/cronlog/internal/alert"
)

// TestEvaluate_BothThresholds_ConsecutiveTakesPrecedence verifies that when
// both thresholds are configured, a consecutive-failure alert fires before
// the rate threshold is evaluated for the same entry set.
func TestEvaluate_BothThresholds_ConsecutiveTakesPrecedence(t *testing.T) {
	ev := alert.New(alert.Config{
		MaxConsecutiveFailures: 2,
		FailureRateWindow:      time.Hour,
		MaxFailureRate:         0.9, // high, would not trigger on its own
	})
	// last two are failures → consecutive threshold hit
	entries := makeEntries([]int{0, 0, 1, 1}, time.Hour)
	got := ev.Evaluate("backup", entries)
	if got == nil {
		t.Fatal("expected alert, got nil")
	}
	if !strings.Contains(got.Reason, "consecutive") {
		t.Errorf("expected consecutive reason, got: %s", got.Reason)
	}
}

// TestEvaluate_WindowFiltersOldEntries ensures entries outside the window
// are not counted toward the failure rate.
func TestEvaluate_WindowFiltersOldEntries(t *testing.T) {
	ev := alert.New(alert.Config{
		FailureRateWindow: 10 * time.Minute,
		MaxFailureRate:    0.5,
	})
	now := time.Now()
	entries := []alert.Entry{
		{JobName: "sync", ExitCode: 1, RunAt: now.Add(-2 * time.Hour)}, // outside window
		{JobName: "sync", ExitCode: 1, RunAt: now.Add(-3 * time.Hour)}, // outside window
		{JobName: "sync", ExitCode: 0, RunAt: now.Add(-1 * time.Minute)}, // inside, success
	}
	if got := ev.Evaluate("sync", entries); got != nil {
		t.Fatalf("expected no alert (old failures outside window), got: %s", got.Reason)
	}
}
