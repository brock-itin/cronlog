package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry records a single auditable event for a cron job execution.
type Entry struct {
	Timestamp time.Time         `json:"timestamp"`
	Job       string            `json:"job"`
	Event     string            `json:"event"`
	ExitCode  int               `json:"exit_code,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// Auditor appends structured audit entries to a file.
type Auditor struct {
	mu   sync.Mutex
	f    *os.File
	path string
	now  func() time.Time
}

// New opens (or creates) the audit log at path.
func New(path string) (*Auditor, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("audit: open %s: %w", path, err)
	}
	return &Auditor{f: f, path: path, now: time.Now}, nil
}

// Record writes a single audit entry as a JSON line.
func (a *Auditor) Record(job, event string, exitCode int, meta map[string]string) error {
	entry := Entry{
		Timestamp: a.now().UTC(),
		Job:       job,
		Event:     event,
		ExitCode:  exitCode,
		Meta:      meta,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_, err = fmt.Fprintf(a.f, "%s\n", data)
	return err
}

// Close flushes and closes the underlying file.
func (a *Auditor) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.f.Close()
}
