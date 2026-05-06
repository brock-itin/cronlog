package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain_NoArgs verifies that running without arguments exits non-zero.
func TestMain_NoArgs(t *testing.T) {
	if os.Getenv("CRONLOG_TEST_SUBPROCESS") == "1" {
		run() //nolint:errcheck
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMain_NoArgs")
	cmd.Env = append(os.Environ(), "CRONLOG_TEST_SUBPROCESS=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit, got nil")
	}
}

// TestMain_MissingJob verifies that omitting --job returns an error.
func TestMain_MissingJob(t *testing.T) {
	configPath := writeTempConfig(t)

	os.Args = []string{"cronlog", "--config", configPath, "--", "echo", "hello"}
	err := run()
	if err == nil {
		t.Fatal("expected error for missing --job flag")
	}
}

// TestMain_SuccessfulRun verifies a successful end-to-end execution.
func TestMain_SuccessfulRun(t *testing.T) {
	logDir := t.TempDir()
	configPath := writeTempConfigWithDir(t, logDir)

	os.Args = []string{"cronlog", "--config", configPath, "--job", "test-job", "--", "echo", "hello"}
	if err := run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := filepath.Glob(filepath.Join(logDir, "test-job", "*.log"))
	if err != nil {
		t.Fatalf("glob error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one log file to be created")
	}
}

func writeTempConfig(t *testing.T) string {
	t.Helper()
	return writeTempConfigWithDir(t, t.TempDir())
}

func writeTempConfigWithDir(t *testing.T, logDir string) string {
	t.Helper()
	content := []byte("log_dir: " + logDir + "\nmax_files: 5\n")
	p := filepath.Join(t.TempDir(), "cronlog.yaml")
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return p
}
