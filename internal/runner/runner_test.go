package runner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/cronlog/internal/rotator"
	"github.com/user/cronlog/internal/runner"
)

func newTestRunner(t *testing.T) (*runner.Runner, string) {
	t.Helper()
	dir := t.TempDir()
	rot, err := rotator.New(rotator.Config{
		Dir:      dir,
		Prefix:   "test",
		MaxFiles: 5,
	})
	if err != nil {
		t.Fatalf("rotator.New: %v", err)
	}
	return runner.New(rot), dir
}

func TestRun_SuccessfulCommand(t *testing.T) {
	r, dir := newTestRunner(t)

	res, err := r.Run("echo", "hello cronlog")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
	if res.Duration <= 0 {
		t.Errorf("expected positive duration, got %s", res.Duration)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one log file to be created")
	}

	data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "hello cronlog") {
		t.Errorf("log does not contain command output; got:\n%s", body)
	}
	if !strings.Contains(body, "[cronlog] start") {
		t.Errorf("log missing start header; got:\n%s", body)
	}
	if !strings.Contains(body, "[cronlog] done") {
		t.Errorf("log missing done footer; got:\n%s", body)
	}
}

func TestRun_NonZeroExit(t *testing.T) {
	r, _ := newTestRunner(t)

	res, err := r.Run("sh", "-c", "exit 42")
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}
	if res.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", res.ExitCode)
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	r, _ := newTestRunner(t)

	_, err := r.Run("__no_such_binary_exists__")
	if err == nil {
		t.Fatal("expected error for non-existent command, got nil")
	}
}
