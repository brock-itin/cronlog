// Package logger provides structured JSON logging for cron job execution.
package logger

import (
	"encoding/json"
	"io"
	"time"
)

// Level represents the severity of a log entry.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelError Level = "ERROR"
	LevelDebug Level = "DEBUG"
)

// Entry represents a single structured log record.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Level     Level  `json:"level"`
	Job       string `json:"job"`
	Message   string `json:"message"`
	ExitCode  *int   `json:"exit_code,omitempty"`
	Duration  string `json:"duration,omitempty"`
}

// Logger writes structured JSON log entries to an io.Writer.
type Logger struct {
	writer io.Writer
	job    string
}

// New creates a new Logger that writes to w for the given job name.
func New(w io.Writer, job string) *Logger {
	return &Logger{writer: w, job: job}
}

// Info logs an informational message.
func (l *Logger) Info(msg string) error {
	return l.write(LevelInfo, msg, nil, "")
}

// Error logs an error message with an optional exit code.
func (l *Logger) Error(msg string, exitCode *int) error {
	return l.write(LevelError, msg, exitCode, "")
}

// Done logs job completion with duration and exit code.
func (l *Logger) Done(exitCode int, duration time.Duration) error {
	return l.write(LevelInfo, "job finished", &exitCode, duration.String())
}

func (l *Logger) write(level Level, msg string, exitCode *int, duration string) error {
	entry := Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Job:       l.job,
		Message:   msg,
		ExitCode:  exitCode,
		Duration:  duration,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = l.writer.Write(data)
	return err
}
