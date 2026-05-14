package checkpoint_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronlog/internal/checkpoint"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

func TestLoad_MissingFileIsOK(t *testing.T) {
	cp := checkpoint.New(tempPath(t))
	e, err := cp.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !e.LastOK.IsZero() {
		t.Errorf("expected zero time, got %v", e.LastOK)
	}
}

func TestSave_PersistsEntry(t *testing.T) {
	cp := checkpoint.New(tempPath(t))
	now := time.Now().UTC().Truncate(time.Second)
	err := cp.Save(checkpoint.Entry{Job: "backup", LastOK: now})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	e, err := cp.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if e.Job != "backup" {
		t.Errorf("job: want backup, got %q", e.Job)
	}
	if !e.LastOK.Equal(now) {
		t.Errorf("LastOK: want %v, got %v", now, e.LastOK)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set by Save")
	}
}

func TestOverdue_NoCheckpoint_ReturnsFalse(t *testing.T) {
	cp := checkpoint.New(tempPath(t))
	overdue, err := cp.Overdue("sync", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overdue {
		t.Error("expected false when no checkpoint exists")
	}
}

func TestOverdue_RecentRun_ReturnsFalse(t *testing.T) {
	cp := checkpoint.New(tempPath(t))
	_ = cp.Save(checkpoint.Entry{Job: "sync", LastOK: time.Now().UTC()})
	overdue, err := cp.Overdue("sync", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overdue {
		t.Error("expected false for a recent checkpoint")
	}
}

func TestOverdue_OldRun_ReturnsTrue(t *testing.T) {
	cp := checkpoint.New(tempPath(t))
	old := time.Now().UTC().Add(-2 * time.Hour)
	_ = cp.Save(checkpoint.Entry{Job: "sync", LastOK: old})
	overdue, err := cp.Overdue("sync", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !overdue {
		t.Error("expected true for an old checkpoint")
	}
}

func TestLoad_CorruptFile_ReturnsError(t *testing.T) {
	p := tempPath(t)
	_ = os.WriteFile(p, []byte("not-json{"), 0o644)
	cp := checkpoint.New(p)
	_, err := cp.Load()
	if err == nil {
		t.Error("expected error for corrupt checkpoint file")
	}
}

func TestSave_OverwritesPreviousEntry(t *testing.T) {
	cp := checkpoint.New(tempPath(t))
	first := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	_ = cp.Save(checkpoint.Entry{Job: "backup", LastOK: first})

	second := time.Now().UTC().Truncate(time.Second)
	if err := cp.Save(checkpoint.Entry{Job: "backup", LastOK: second}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	e, err := cp.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !e.LastOK.Equal(second) {
		t.Errorf("LastOK: want %v, got %v", second, e.LastOK)
	}
}
