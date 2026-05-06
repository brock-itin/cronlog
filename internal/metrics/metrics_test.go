package metrics_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronlog/internal/metrics"
)

func tempPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "metrics.json")
}

func TestRecord_CreatesStatsEntry(t *testing.T) {
	c, err := metrics.New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := c.Record("backup", 0, 2*time.Second); err != nil {
		t.Fatalf("Record: %v", err)
	}

	s, ok := c.Get("backup")
	if !ok {
		t.Fatal("expected stats entry for 'backup'")
	}
	if s.TotalRuns != 1 {
		t.Errorf("TotalRuns = %d, want 1", s.TotalRuns)
	}
	if s.FailedRuns != 0 {
		t.Errorf("FailedRuns = %d, want 0", s.FailedRuns)
	}
}

func TestRecord_TracksFailures(t *testing.T) {
	c, err := metrics.New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_ = c.Record("sync", 0, time.Second)
	_ = c.Record("sync", 1, time.Second)
	_ = c.Record("sync", 2, time.Second)

	s, _ := c.Get("sync")
	if s.TotalRuns != 3 {
		t.Errorf("TotalRuns = %d, want 3", s.TotalRuns)
	}
	if s.FailedRuns != 2 {
		t.Errorf("FailedRuns = %d, want 2", s.FailedRuns)
	}
	if s.LastExitCode != 2 {
		t.Errorf("LastExitCode = %d, want 2", s.LastExitCode)
	}
}

func TestNew_LoadsExistingFile(t *testing.T) {
	path := tempPath(t)

	c1, _ := metrics.New(path)
	_ = c1.Record("cleanup", 0, 500*time.Millisecond)

	c2, err := metrics.New(path)
	if err != nil {
		t.Fatalf("second New: %v", err)
	}
	s, ok := c2.Get("cleanup")
	if !ok {
		t.Fatal("expected persisted stats to be loaded")
	}
	if s.TotalRuns != 1 {
		t.Errorf("TotalRuns = %d, want 1", s.TotalRuns)
	}
}

func TestNew_MissingFileIsOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does_not_exist.json")
	_, err := metrics.New(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}

func TestNew_CorruptFileReturnsError(t *testing.T) {
	path := tempPath(t)
	if err := os.WriteFile(path, []byte("not json{"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := metrics.New(path)
	if err == nil {
		t.Fatal("expected error for corrupt metrics file")
	}
}
