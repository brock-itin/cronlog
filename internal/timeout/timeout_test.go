package timeout_test

import (
	"testing"
	"time"

	"github.com/user/cronlog/internal/runner"
	"github.com/user/cronlog/internal/timeout"
)

func newRunner(t *testing.T) *runner.Runner {
	t.Helper()
	r, err := runner.New()
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	return r
}

func TestRun_CompletesWithinTimeout(t *testing.T) {
	r := timeout.New(newRunner(t), 5*time.Second)
	code, out, err := r.Run("echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if len(out) == 0 {
		t.Fatal("expected output, got none")
	}
}

func TestRun_ExceedsTimeout(t *testing.T) {
	r := timeout.New(newRunner(t), 100*time.Millisecond)
	code, _, err := r.Run("sleep", "5")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if code != -1 {
		t.Fatalf("expected exit code -1, got %d", code)
	}
}

func TestRun_ZeroTimeout_NoLimit(t *testing.T) {
	r := timeout.New(newRunner(t), 0)
	code, _, err := r.Run("true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestRun_NegativeTimeout_NoLimit(t *testing.T) {
	r := timeout.New(newRunner(t), -1*time.Second)
	code, _, err := r.Run("true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}
