package logger_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/yourorg/cronlog/internal/logger"
)

// parseEntry is a helper that unmarshals JSON from buf into a logger.Entry,
// failing the test immediately if parsing fails.
func parseEntry(t *testing.T, buf *bytes.Buffer) logger.Entry {
	t.Helper()
	var entry logger.Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	return entry
}

func TestInfo_WritesJSONEntry(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "backup")

	if err := l.Info("starting job"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := parseEntry(t, &buf)

	if entry.Level != logger.LevelInfo {
		t.Errorf("expected level INFO, got %q", entry.Level)
	}
	if entry.Job != "backup" {
		t.Errorf("expected job 'backup', got %q", entry.Job)
	}
	if entry.Message != "starting job" {
		t.Errorf("expected message 'starting job', got %q", entry.Message)
	}
	if entry.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestError_IncludesExitCode(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "cleanup")
	code := 1

	if err := l.Error("command failed", &code); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := parseEntry(t, &buf)

	if entry.Level != logger.LevelError {
		t.Errorf("expected level ERROR, got %q", entry.Level)
	}
	if entry.ExitCode == nil || *entry.ExitCode != 1 {
		t.Errorf("expected exit_code 1, got %v", entry.ExitCode)
	}
}

func TestDone_IncludesDurationAndExitCode(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "sync")

	if err := l.Done(0, 2*time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := parseEntry(t, &buf)

	if entry.ExitCode == nil || *entry.ExitCode != 0 {
		t.Errorf("expected exit_code 0, got %v", entry.ExitCode)
	}
	if entry.Duration == "" {
		t.Error("expected non-empty duration")
	}
	if entry.Message != "job finished" {
		t.Errorf("expected message 'job finished', got %q", entry.Message)
	}
}

func TestNew_SetsJobName(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, "myjob")

	if err := l.Info("test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := parseEntry(t, &buf)

	if entry.Job != "myjob" {
		t.Errorf("expected job name 'myjob', got %q", entry.Job)
	}
}
