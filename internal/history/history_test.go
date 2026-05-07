package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "history.json")
}

func TestNew_MissingFileIsOK(t *testing.T) {
	h, err := New(tempPath(t), 10)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(h.Entries()) != 0 {
		t.Fatalf("expected empty entries")
	}
}

func TestAdd_PersistsEntry(t *testing.T) {
	path := tempPath(t)
	h, _ := New(path, 10)
	e := Entry{Job: "backup", StartedAt: time.Now(), ExitCode: 0, LogFile: "/var/log/backup.log"}
	if err := h.Add(e); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
	h2, _ := New(path, 10)
	if len(h2.Entries()) != 1 {
		t.Fatalf("expected 1 entry after reload, got %d", len(h2.Entries()))
	}
}

func TestAdd_PrunesOldEntries(t *testing.T) {
	h, _ := New(tempPath(t), 3)
	base := time.Now()
	for i := 0; i < 5; i++ {
		_ = h.Add(Entry{Job: "job", StartedAt: base.Add(time.Duration(i) * time.Minute)})
	}
	if len(h.Entries()) != 3 {
		t.Fatalf("expected 3 entries after pruning, got %d", len(h.Entries()))
	}
}

func TestLastFailure_ReturnsLatestNonZero(t *testing.T) {
	h, _ := New(tempPath(t), 10)
	base := time.Now()
	_ = h.Add(Entry{Job: "job", StartedAt: base, ExitCode: 1})
	_ = h.Add(Entry{Job: "job", StartedAt: base.Add(time.Minute), ExitCode: 0})
	_ = h.Add(Entry{Job: "job", StartedAt: base.Add(2 * time.Minute), ExitCode: 2})

	f := h.LastFailure()
	if f == nil {
		t.Fatal("expected a failure entry")
	}
	if f.ExitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", f.ExitCode)
	}
}

func TestLastFailure_NilWhenNoFailures(t *testing.T) {
	h, _ := New(tempPath(t), 10)
	_ = h.Add(Entry{Job: "job", StartedAt: time.Now(), ExitCode: 0})
	if h.LastFailure() != nil {
		t.Fatal("expected nil when no failures")
	}
}
