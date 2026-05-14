package quota_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronlog/internal/history"
	"github.com/cronlog/internal/quota"
)

// fakeRunner records calls and returns a preset exit code.
type fakeRunner struct {
	calls int
	code  int
}

func (f *fakeRunner) Run(_ context.Context, _ []string) (int, error) {
	f.calls++
	return f.code, nil
}

func tempHistory(t *testing.T) *history.History {
	t.Helper()
	p := filepath.Join(t.TempDir(), "history.json")
	h, err := history.New(p, 100)
	if err != nil {
		t.Fatalf("history.New: %v", err)
	}
	return h
}

func TestNew_RejectsNilRunner(t *testing.T) {
	h := tempHistory(t)
	_, err := quota.New(nil, h, 5, time.Hour)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsNilHistory(t *testing.T) {
	r := &fakeRunner{}
	_, err := quota.New(r, nil, 5, time.Hour)
	if err == nil {
		t.Fatal("expected error for nil history")
	}
}

func TestNew_RejectsZeroMax(t *testing.T) {
	h := tempHistory(t)
	r := &fakeRunner{}
	_, err := quota.New(r, h, 0, time.Hour)
	if err == nil {
		t.Fatal("expected error for zero max")
	}
}

func TestNew_RejectsNonPositiveWindow(t *testing.T) {
	h := tempHistory(t)
	r := &fakeRunner{}
	_, err := quota.New(r, h, 3, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestRun_BelowLimit_ExecutesJob(t *testing.T) {
	h := tempHistory(t)
	r := &fakeRunner{code: 0}
	q, err := quota.New(r, h, 3, time.Hour)
	if err != nil {
		t.Fatalf("quota.New: %v", err)
	}

	code, err := q.Run(context.Background(), []string{"echo", "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if r.calls != 1 {
		t.Fatalf("expected 1 call, got %d", r.calls)
	}
}

func TestRun_AtLimit_Blocked(t *testing.T) {
	p := filepath.Join(t.TempDir(), "history.json")
	h, _ := history.New(p, 100)

	// Pre-populate history with 2 recent entries.
	now := time.Now()
	_ = h.Add(history.Entry{StartedAt: now.Add(-10 * time.Minute), ExitCode: 0})
	_ = h.Add(history.Entry{StartedAt: now.Add(-5 * time.Minute), ExitCode: 0})

	r := &fakeRunner{}
	q, err := quota.New(r, h, 2, time.Hour)
	if err != nil {
		t.Fatalf("quota.New: %v", err)
	}

	_, err = q.Run(context.Background(), []string{"echo"})
	if err == nil {
		t.Fatal("expected quota exceeded error")
	}
	if r.calls != 0 {
		t.Fatalf("runner should not have been called, got %d calls", r.calls)
	}
	_ = os.Remove(p)
}

func TestRun_OldEntriesIgnored_ExecutesJob(t *testing.T) {
	p := filepath.Join(t.TempDir(), "history.json")
	h, _ := history.New(p, 100)

	// Entry older than the window should not count.
	old := time.Now().Add(-3 * time.Hour)
	_ = h.Add(history.Entry{StartedAt: old, ExitCode: 0})
	_ = h.Add(history.Entry{StartedAt: old, ExitCode: 0})

	r := &fakeRunner{}
	q, err := quota.New(r, h, 2, time.Hour)
	if err != nil {
		t.Fatalf("quota.New: %v", err)
	}

	_, err = q.Run(context.Background(), []string{"echo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.calls != 1 {
		t.Fatalf("expected 1 call, got %d", r.calls)
	}
}
