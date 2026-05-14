package watchdog_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cronlog/internal/watchdog"
)

// --- fakes ---

type fakeRunner struct {
	code int
	err  error
}

func (f *fakeRunner) Run(_ string, _ ...string) (int, error) { return f.code, f.err }

type fakeCheckpoint struct {
	data map[string]time.Time
}

func newFakeCheckpoint() *fakeCheckpoint {
	return &fakeCheckpoint{data: make(map[string]time.Time)}
}
func (c *fakeCheckpoint) Load(k string) (time.Time, bool) { v, ok := c.data[k]; return v, ok }
func (c *fakeCheckpoint) Save(k string, t time.Time) error { c.data[k] = t; return nil }

type fakeNotifier struct{ calls []string }

func (n *fakeNotifier) Notify(_ string, msg string) error { n.calls = append(n.calls, msg); return nil }

// --- tests ---

func TestNew_RejectsNilRunner(t *testing.T) {
	_, err := watchdog.New(nil, newFakeCheckpoint(), nil, time.Hour)
	if err == nil {
		t.Fatal("expected error for nil runner")
	}
}

func TestNew_RejectsNilCheckpoint(t *testing.T) {
	_, err := watchdog.New(&fakeRunner{}, nil, nil, time.Hour)
	if err == nil {
		t.Fatal("expected error for nil checkpoint")
	}
}

func TestNew_RejectsZeroInterval(t *testing.T) {
	_, err := watchdog.New(&fakeRunner{}, newFakeCheckpoint(), nil, 0)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestRun_SavesCheckpointOnSuccess(t *testing.T) {
	cp := newFakeCheckpoint()
	wd, _ := watchdog.New(&fakeRunner{code: 0}, cp, nil, time.Hour)
	code, err := wd.Run("backup")
	if err != nil || code != 0 {
		t.Fatalf("unexpected result: code=%d err=%v", code, err)
	}
	if _, ok := cp.Load("backup"); !ok {
		t.Fatal("checkpoint was not saved")
	}
}

func TestRun_NoCheckpoint_OnNonZeroExit(t *testing.T) {
	cp := newFakeCheckpoint()
	wd, _ := watchdog.New(&fakeRunner{code: 1}, cp, nil, time.Hour)
	wd.Run("backup")
	if _, ok := cp.Load("backup"); ok {
		t.Fatal("checkpoint should not be saved on failure")
	}
}

func TestRun_OverdueCheckpoint_TriggersNotifier(t *testing.T) {
	cp := newFakeCheckpoint()
	cp.Save("backup", time.Now().Add(-2*time.Hour))
	n := &fakeNotifier{}
	wd, _ := watchdog.New(&fakeRunner{code: 0}, cp, n, time.Hour)
	wd.Run("backup")
	if len(n.calls) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(n.calls))
	}
}

func TestRun_RecentCheckpoint_NoNotification(t *testing.T) {
	cp := newFakeCheckpoint()
	cp.Save("backup", time.Now().Add(-30*time.Minute))
	n := &fakeNotifier{}
	wd, _ := watchdog.New(&fakeRunner{code: 0}, cp, n, time.Hour)
	wd.Run("backup")
	if len(n.calls) != 0 {
		t.Fatalf("expected no notifications, got %d", len(n.calls))
	}
}

func TestRun_PropagatesRunnerError(t *testing.T) {
	cp := newFakeCheckpoint()
	expected := errors.New("exec failed")
	wd, _ := watchdog.New(&fakeRunner{code: -1, err: expected}, cp, nil, time.Hour)
	_, err := wd.Run("backup")
	if !errors.Is(err, expected) {
		t.Fatalf("expected runner error, got %v", err)
	}
}
