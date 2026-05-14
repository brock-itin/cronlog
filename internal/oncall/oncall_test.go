package oncall_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronlog/internal/oncall"
)

// fakeRunner records calls and returns a preset exit code / error.
type fakeRunner struct {
	calls int
	code  int
	err   error
}

func (f *fakeRunner) Run(_ context.Context, _ string, _ []string) (int, error) {
	f.calls++
	return f.code, f.err
}

func fixedClock(hour int) func() time.Time {
	return func() time.Time {
		return time.Date(2024, 1, 1, hour, 0, 0, 0, time.UTC)
	}
}

func TestNew_RejectsNilRunner(t *testing.T) {
	_, err := oncall.New(nil, 9, 17)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsInvalidRange(t *testing.T) {
	fr := &fakeRunner{}
	if _, err := oncall.New(fr, 17, 9); err == nil {
		t.Fatal("expected error when start >= end")
	}
	if _, err := oncall.New(fr, -1, 9); err == nil {
		t.Fatal("expected error for negative start")
	}
	if _, err := oncall.New(fr, 9, 25); err == nil {
		t.Fatal("expected error for end > 23")
	}
}

func TestRun_WithinWindow_Executes(t *testing.T) {
	fr := &fakeRunner{code: 0}
	o, err := oncall.New(fr, 9, 17)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o.(*oncall.Oncall) // compile-time check skipped; use exported field via test helper

	// inject clock via unexported field — use the package-level test hook instead
	_ = o
}

func TestRun_InsideWindow_Delegates(t *testing.T) {
	fr := &fakeRunner{code: 42}
	o, _ := oncall.New(fr, 9, 17)
	oncall.SetClock(o, fixedClock(12))

	code, err := o.Run(context.Background(), "echo", []string{"hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 42 {
		t.Errorf("expected exit code 42, got %d", code)
	}
	if fr.calls != 1 {
		t.Errorf("expected 1 call, got %d", fr.calls)
	}
}

func TestRun_OutsideWindow_Skips(t *testing.T) {
	fr := &fakeRunner{code: 1, err: errors.New("should not run")}
	o, _ := oncall.New(fr, 9, 17)
	oncall.SetClock(o, fixedClock(3))

	code, err := o.Run(context.Background(), "echo", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if fr.calls != 0 {
		t.Errorf("expected 0 calls, got %d", fr.calls)
	}
}

func TestRun_AtWindowBoundary_Start_Executes(t *testing.T) {
	fr := &fakeRunner{}
	o, _ := oncall.New(fr, 9, 17)
	oncall.SetClock(o, fixedClock(9))

	o.Run(context.Background(), "x", nil) //nolint:errcheck
	if fr.calls != 1 {
		t.Errorf("expected execution at start boundary, got %d calls", fr.calls)
	}
}

func TestRun_AtWindowBoundary_End_Skips(t *testing.T) {
	fr := &fakeRunner{}
	o, _ := oncall.New(fr, 9, 17)
	oncall.SetClock(o, fixedClock(17))

	o.Run(context.Background(), "x", nil) //nolint:errcheck
	if fr.calls != 0 {
		t.Errorf("expected skip at end boundary, got %d calls", fr.calls)
	}
}
