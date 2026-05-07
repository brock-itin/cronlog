// Package summary provides end-of-run report generation for cron job executions.
package summary

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Report holds the collected data for a single cron job run.
type Report struct {
	JobName   string
	Command   string
	StartedAt time.Time
	Duration  time.Duration
	ExitCode  int
	Lines     int
	Filtered  int
	LogFile   string
}

// Writer renders a summary report to an io.Writer.
type Writer struct {
	out io.Writer
}

// New creates a new summary Writer that writes to out.
func New(out io.Writer) *Writer {
	return &Writer{out: out}
}

// Write renders the report as a human-readable block.
func (w *Writer) Write(r Report) error {
	status := "OK"
	if r.ExitCode != 0 {
		status = fmt.Sprintf("FAILED (exit %d)", r.ExitCode)
	}

	lines := []string{
		"=== cronlog summary ===",
		fmt.Sprintf("  job      : %s", r.JobName),
		fmt.Sprintf("  command  : %s", r.Command),
		fmt.Sprintf("  started  : %s", r.StartedAt.Format(time.RFC3339)),
		fmt.Sprintf("  duration : %s", r.Duration.Round(time.Millisecond)),
		fmt.Sprintf("  status   : %s", status),
		fmt.Sprintf("  output   : %d lines (%d filtered)", r.Lines, r.Filtered),
	}
	if r.LogFile != "" {
		lines = append(lines, fmt.Sprintf("  log file : %s", r.LogFile))
	}
	lines = append(lines, "=======================")

	_, err := fmt.Fprintln(w.out, strings.Join(lines, "\n"))
	return err
}
