package debounce_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronlog/internal/checkpoint"
	"github.com/cronlog/internal/debounce"
)

// stubRunner records how many times it was called and what exit code to return.
type stubRunner struct {
	calls    int
	exitCode int
}

func (s *stubRunner) Run(_ []string) (int, error) {
	s.calls++
	return s.exitCode, nil
}

func tempCheckpoint(t *testing.T) *checkpoint.Checkpoint {
	t.Helper()
	dir := t.TempDir()
	cp, err := checkpoint.New(filepath.Join(dir, "cp.json"))
	if err != nil {
		t.Fatalf("checkpoint.New: %v", err)
	}
	return cp
}

func TestNew_RejectsNilRunner(t *testing.T) {
	cp := tempCheckpoint(t)
	_, err := debounce.New(nil, time.Second, cp)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsNilCheckpoint(t *testing.T) {
	_, err := debounce.New(&stubRunner{}, time.Second, nil)
	if err == nil {
		t.Fatal("expected error for nil checkpoint")
	}
}

func TestNew_RejectsZeroInterval(t *testing.T) {
	cp := tempCheckpoint(t)
	_, err := debounce.New(&stubRunner{}, 0, cp)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestRun_ExecutesWhenNoCheckpoint(t *testing.T) {
	cp := tempCheckpoint(t)
	stub := &stubRunner{}
	d, _ := debounce.New(stub, time.Minute, cp)

	code, err := d.Run(nil)
	if err != nil || code != 0 {
		t.Fatalf("unexpected result: code=%d err=%v", code, err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls)
	}
}

func TestRun_SuppressedWithinInterval(t *testing.T) {
	cp := tempCheckpoint(t)
	_ = cp.Save(time.Now()) // mark as just-run

	stub := &stubRunner{}
	d, _ := debounce.New(stub, time.Minute, cp)

	code, err := d.Run(nil)
	if err != nil || code != 0 {
		t.Fatalf("unexpected result: code=%d err=%v", code, err)
	}
	if stub.calls != 0 {
		t.Fatalf("expected 0 calls (suppressed), got %d", stub.calls)
	}
}

func TestRun_ExecutesAfterIntervalExpires(t *testing.T) {
	dir := t.TempDir()
	cp, _ := checkpoint.New(filepath.Join(dir, "cp.json"))
	// Persist a timestamp well in the past.
	_ = cp.Save(time.Now().Add(-2 * time.Minute))

	// Reload so the in-memory state reflects the persisted value.
	cp2, _ := checkpoint.New(filepath.Join(dir, "cp.json"))

	stub := &stubRunner{}
	d, _ := debounce.New(stub, time.Minute, cp2)

	_, _ = d.Run(nil)
	if stub.calls != 1 {
		t.Fatalf("expected 1 call after interval, got %d", stub.calls)
	}
	_ = os.RemoveAll(dir)
}
