package summary_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/cronlog/internal/summary"
)

func baseReport() summary.Report {
	return summary.Report{
		JobName:   "backup",
		Command:   "/usr/bin/backup.sh",
		StartedAt: time.Date(2024, 1, 15, 3, 0, 0, 0, time.UTC),
		Duration:  2*time.Second + 345*time.Millisecond,
		ExitCode:  0,
		Lines:     42,
		Filtered:  3,
		LogFile:   "/var/log/cronlog/backup.log",
	}
}

func TestWrite_SuccessfulRun(t *testing.T) {
	var buf bytes.Buffer
	w := summary.New(&buf)

	if err := w.Write(baseReport()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"cronlog summary",
		"backup",
		"/usr/bin/backup.sh",
		"2.345s",
		"OK",
		"42 lines (3 filtered)",
		"/var/log/cronlog/backup.log",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestWrite_FailedRun(t *testing.T) {
	var buf bytes.Buffer
	w := summary.New(&buf)

	r := baseReport()
	r.ExitCode = 2

	if err := w.Write(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "FAILED (exit 2)") {
		t.Errorf("expected failure status in output, got:\n%s", out)
	}
}

func TestWrite_NoLogFile(t *testing.T) {
	var buf bytes.Buffer
	w := summary.New(&buf)

	r := baseReport()
	r.LogFile = ""

	if err := w.Write(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(buf.String(), "log file") {
		t.Errorf("log file line should be absent when LogFile is empty")
	}
}
