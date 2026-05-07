package alert_test

import (
	"testing"
	"time"

	"github.com/cronlog/cronlog/internal/alert"
)

func makeEntries(codes []int, window time.Duration) []alert.Entry {
	now := time.Now()
	entries := make([]alert.Entry, len(codes))
	for i, c := range codes {
		entries[i] = alert.Entry{
			JobName:  "testjob",
			ExitCode: c,
			RunAt:    now.Add(-window + time.Duration(i)*time.Minute),
		}
	}
	return entries
}

func TestEvaluate_NoAlert_OnSuccess(t *testing.T) {
	ev := alert.New(alert.Config{MaxConsecutiveFailures: 3})
	entries := makeEntries([]int{0, 0, 0}, time.Hour)
	if got := ev.Evaluate("testjob", entries); got != nil {
		t.Fatalf("expected no alert, got: %v", got.Reason)
	}
}

func TestEvaluate_ConsecutiveFailures_TriggersAlert(t *testing.T) {
	ev := alert.New(alert.Config{MaxConsecutiveFailures: 3})
	entries := makeEntries([]int{0, 1, 1, 1}, time.Hour)
	got := ev.Evaluate("testjob", entries)
	if got == nil {
		t.Fatal("expected alert, got nil")
	}
	if got.JobName != "testjob" {
		t.Errorf("expected job name testjob, got %s", got.JobName)
	}
}

func TestEvaluate_ConsecutiveFailures_BelowThreshold(t *testing.T) {
	ev := alert.New(alert.Config{MaxConsecutiveFailures: 4})
	entries := makeEntries([]int{1, 1, 1}, time.Hour)
	if got := ev.Evaluate("testjob", entries); got != nil {
		t.Fatalf("expected no alert, got: %v", got.Reason)
	}
}

func TestEvaluate_FailureRate_TriggersAlert(t *testing.T) {
	ev := alert.New(alert.Config{
		FailureRateWindow: time.Hour,
		MaxFailureRate:    0.5,
	})
	// 3 out of 4 failures = 75%
	entries := makeEntries([]int{1, 1, 1, 0}, time.Hour)
	got := ev.Evaluate("testjob", entries)
	if got == nil {
		t.Fatal("expected alert for high failure rate, got nil")
	}
}

func TestEvaluate_FailureRate_BelowThreshold(t *testing.T) {
	ev := alert.New(alert.Config{
		FailureRateWindow: time.Hour,
		MaxFailureRate:    0.8,
	})
	// 1 out of 4 = 25%
	entries := makeEntries([]int{1, 0, 0, 0}, time.Hour)
	if got := ev.Evaluate("testjob", entries); got != nil {
		t.Fatalf("expected no alert, got: %v", got.Reason)
	}
}

func TestEvaluate_EmptyEntries_ReturnsNil(t *testing.T) {
	ev := alert.New(alert.Config{MaxConsecutiveFailures: 1, MaxFailureRate: 0.1})
	if got := ev.Evaluate("testjob", nil); got != nil {
		t.Fatalf("expected nil for empty entries, got: %v", got)
	}
}
