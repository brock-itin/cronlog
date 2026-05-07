package lock_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/yourorg/cronlog/internal/lock"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "cronlog-lock-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestAcquire_CreatesLockFile(t *testing.T) {
	dir := tempDir(t)
	l := lock.New(dir, "backup-job")

	if err := l.Acquire(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer l.Release() //nolint:errcheck

	matches, _ := filepath.Glob(filepath.Join(dir, ".cronlog_*.lock"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 lock file, found %d", len(matches))
	}
}

func TestAcquire_FailsWhenAlreadyLocked(t *testing.T) {
	dir := tempDir(t)
	l := lock.New(dir, "backup-job")

	if err := l.Acquire(); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	defer l.Release() //nolint:errcheck

	l2 := lock.New(dir, "backup-job")
	if err := l2.Acquire(); err == nil {
		l2.Release() //nolint:errcheck
		t.Fatal("expected error on second acquire, got nil")
	}
}

func TestAcquire_ClearsStateLock(t *testing.T) {
	dir := tempDir(t)
	l := lock.New(dir, "stale-job")

	// Write a lock file with a non-existent PID.
	lockPath := filepath.Join(dir, ".cronlog_stale-job.lock")
	_ = os.WriteFile(lockPath, []byte(strconv.Itoa(99999999)), 0o644)

	if err := l.Acquire(); err != nil {
		t.Fatalf("expected stale lock to be cleared, got %v", err)
	}
	defer l.Release() //nolint:errcheck
}

func TestRelease_RemovesLockFile(t *testing.T) {
	dir := tempDir(t)
	l := lock.New(dir, "cleanup-job")

	if err := l.Acquire(); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}

	matches, _ := filepath.Glob(filepath.Join(dir, ".cronlog_*.lock"))
	if len(matches) != 0 {
		t.Fatalf("expected lock file to be removed, found %d", len(matches))
	}
}
