// Package lock provides a simple file-based locking mechanism to prevent
// concurrent execution of the same cron job.
package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Lock represents a file-based process lock.
type Lock struct {
	path string
}

// New creates a new Lock for the given job name under the specified directory.
func New(dir, jobName string) *Lock {
	safe := strings.NewReplacer("/", "_", " ", "_").Replace(jobName)
	return &Lock{
		path: filepath.Join(dir, fmt.Sprintf(".cronlog_%s.lock", safe)),
	}
}

// Acquire attempts to acquire the lock. It returns an error if the lock is
// already held by a running process.
func (l *Lock) Acquire() error {
	if data, err := os.ReadFile(l.path); err == nil {
		pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
		if parseErr == nil && processExists(pid) {
			return fmt.Errorf("job already running with pid %d (lock: %s)", pid, l.path)
		}
		// Stale lock — remove it.
		_ = os.Remove(l.path)
	}

	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return fmt.Errorf("lock: create directory: %w", err)
	}

	content := []byte(strconv.Itoa(os.Getpid()))
	if err := os.WriteFile(l.path, content, 0o644); err != nil {
		return fmt.Errorf("lock: write pid file: %w", err)
	}
	return nil
}

// Release removes the lock file.
func (l *Lock) Release() error {
	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("lock: remove pid file: %w", err)
	}
	return nil
}

// processExists returns true if a process with the given pid is running.
func processExists(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds; signal 0 checks existence.
	err = proc.Signal(os.Signal(nil))
	return err == nil
}
