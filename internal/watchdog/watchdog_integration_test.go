package watchdog_test

import (
	"testing"
	"time"

	"github.com/cronlog/internal/checkpoint"
	"github.com/cronlog/internal/watchdog"
)

func tempCheckpointPath(t *testing.T) string {
	t.Helper()
	return t.TempDir() + "/checkpoint.json"
}

func TestWatchdog_IntegrationWithCheckpoint(t *testing.T) {
	path := tempCheckpointPath(t)
	cp, err := checkpoint.New(path)
	if err != nil {
		t.Fatalf("checkpoint.New: %v", err)
	}

	n := &fakeNotifier{}
	wd, err := watchdog.New(&fakeRunner{code: 0}, cp, n, time.Hour)
	if err != nil {
		t.Fatalf("watchdog.New: %v", err)
	}

	// First run — no prior checkpoint, no notification.
	if _, err := wd.Run("sync"); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if len(n.calls) != 0 {
		t.Fatalf("expected no alert on first run, got %d", len(n.calls))
	}

	// Second run within interval — still no notification.
	if _, err := wd.Run("sync"); err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if len(n.calls) != 0 {
		t.Fatalf("expected no alert on timely run, got %d", len(n.calls))
	}
}
