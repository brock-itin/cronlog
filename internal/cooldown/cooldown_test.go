package cooldown_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/owner/cronlog/internal/checkpoint"
	"github.com/owner/cronlog/internal/cooldown"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

type fakeRunner struct {
	code int
	err  error
	calls int
}

func (f *fakeRunner) Run(_ string, _ []string) (int, error) {
	f.calls++
	return f.code, f.err
}

func TestNew_RejectsNilRunner(t *testing.T) {
	cp, _ := checkpoint.New(tempPath(t))
	_, err := cooldown.New(nil, cp, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsNilCheckpoint(t *testing.T) {
	_, err := cooldown.New(&fakeRunner{}, nil, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil checkpoint")
	}
}

func TestNew_RejectsNonPositiveInterval(t *testing.T) {
	cp, _ := checkpoint.New(tempPath(t))
	_, err := cooldown.New(&fakeRunner{}, cp, 0)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestRun_NoCheckpoint_ExecutesJob(t *testing.T) {
	cp, _ := checkpoint.New(tempPath(t))
	runner := &fakeRunner{code: 0}
	cd, _ := cooldown.New(runner, cp, time.Minute)

	code, err := cd.Run("test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if runner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", runner.calls)
	}
}

func TestRun_WithinCooldown_Skips(t *testing.T) {
	cp, _ := checkpoint.New(tempPath(t))
	_ = cp.Save(time.Now()) // simulate a very recent run

	runner := &fakeRunner{code: 0}
	cd, _ := cooldown.New(runner, cp, time.Hour)

	code, err := cd.Run("test", nil)
	if err == nil {
		t.Fatal("expected cooldown skip error")
	}
	if code != 0 {
		t.Fatalf("expected exit 0 on skip, got %d", code)
	}
	if runner.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", runner.calls)
	}
}

func TestRun_AfterCooldown_ExecutesJob(t *testing.T) {
	cp, _ := checkpoint.New(tempPath(t))
	_ = cp.Save(time.Now().Add(-2 * time.Hour)) // old enough

	runner := &fakeRunner{code: 0}
	cd, _ := cooldown.New(runner, cp, time.Hour)

	_, err := cd.Run("test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", runner.calls)
	}
}

func TestRun_FailedJob_StillRecordsCheckpoint(t *testing.T) {
	path := tempPath(t)
	cp, _ := checkpoint.New(path)
	runner := &fakeRunner{code: 1, err: errors.New("boom")}
	cd, _ := cooldown.New(runner, cp, time.Minute)

	cd.Run("test", nil) //nolint:errcheck

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected checkpoint file to exist after failed run")
	}

	// A second run should be blocked by cooldown
	runner2 := &fakeRunner{code: 0}
	cp2, _ := checkpoint.New(path)
	cd2, _ := cooldown.New(runner2, cp2, time.Minute)
	_, err := cd2.Run("test", nil)
	if err == nil {
		t.Fatal("expected cooldown to block second run after failure")
	}
}
