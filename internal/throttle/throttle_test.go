package throttle_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/cronlog/internal/history"
	"github.com/user/cronlog/internal/throttle"
)

// fakeRunner records calls and returns preset values.
type fakeRunner struct {
	calls   int
	exitCode int
	err     error
}

func (f *fakeRunner) Run() (int, error) {
	f.calls++
	return f.exitCode, f.err
}

func tempHistory(t *testing.T) *history.History {
	t.Helper()
	p := filepath.Join(t.TempDir(), "history.json")
	h, err := history.New(p, 50)
	if err != nil {
		t.Fatalf("history.New: %v", err)
	}
	return h
}

func TestRun_NoHistory_ExecutesJob(t *testing.T) {
	h := tempHistory(t)
	r := &fakeRunner{exitCode: 0}
	th := throttle.New(r, h, "myjob", 5*time.Minute)

	code, err := th.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if r.calls != 1 {
		t.Errorf("expected 1 call, got %d", r.calls)
	}
}

func TestRun_RecentHistory_Throttled(t *testing.T) {
	h := tempHistory(t)
	_ = h.Add(history.Entry{StartedAt: time.Now(), ExitCode: 0, Duration: time.Second})

	r := &fakeRunner{exitCode: 0}
	th := throttle.New(r, h, "myjob", 10*time.Minute)

	code, err := th.Run()
	if code != -1 {
		t.Errorf("expected -1 exit code, got %d", code)
	}
	var te *throttle.ErrThrottled
	if !errors.As(err, &te) {
		t.Fatalf("expected ErrThrottled, got %T: %v", err, err)
	}
	if te.JobName != "myjob" {
		t.Errorf("expected job name 'myjob', got %q", te.JobName)
	}
	if r.calls != 0 {
		t.Errorf("runner should not have been called")
	}
}

func TestRun_OldHistory_ExecutesJob(t *testing.T) {
	h := tempHistory(t)
	_ = h.Add(history.Entry{StartedAt: time.Now().Add(-30 * time.Minute), ExitCode: 0, Duration: time.Second})

	r := &fakeRunner{exitCode: 0}
	th := throttle.New(r, h, "myjob", 5*time.Minute)

	_, err := th.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.calls != 1 {
		t.Errorf("expected 1 call, got %d", r.calls)
	}
}

func TestRun_ZeroGap_AlwaysExecutes(t *testing.T) {
	h := tempHistory(t)
	_ = h.Add(history.Entry{StartedAt: time.Now(), ExitCode: 0, Duration: time.Second})

	r := &fakeRunner{exitCode: 0}
	th := throttle.New(r, h, "myjob", 0)

	_, err := th.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.calls != 1 {
		t.Errorf("expected 1 call, got %d", r.calls)
	}
}

func TestErrThrottled_ErrorString(t *testing.T) {
	err := &throttle.ErrThrottled{
		JobName:   "backup",
		LastRun:   time.Now().Add(-2 * time.Minute),
		MinGap:    10 * time.Minute,
		Remaining: 8 * time.Minute,
	}
	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
	_ = os.DevNull // suppress unused import
}
