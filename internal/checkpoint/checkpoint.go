// Package checkpoint persists the last successful run timestamp for a job,
// allowing cronlog to detect and report missed or overdue executions.
package checkpoint

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// Entry holds the persisted state for a single job checkpoint.
type Entry struct {
	Job       string    `json:"job"`
	LastOK    time.Time `json:"last_ok"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Checkpoint manages reading and writing job checkpoint files.
type Checkpoint struct {
	path string
}

// New returns a Checkpoint backed by the file at path.
func New(path string) *Checkpoint {
	return &Checkpoint{path: path}
}

// Load reads the checkpoint entry from disk.
// Returns a zero-value Entry and no error when the file does not exist.
func (c *Checkpoint) Load() (Entry, error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Entry{}, nil
		}
		return Entry{}, err
	}
	var e Entry
	if err := json.Unmarshal(data, &e); err != nil {
		return Entry{}, err
	}
	return e, nil
}

// Save writes the entry to disk, creating or truncating the file.
func (c *Checkpoint) Save(e Entry) error {
	e.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}

// Overdue returns true when the last successful run is older than the given
// interval. It always returns false when no checkpoint has been saved yet.
func (c *Checkpoint) Overdue(job string, interval time.Duration) (bool, error) {
	e, err := c.Load()
	if err != nil {
		return false, err
	}
	if e.LastOK.IsZero() {
		return false, nil
	}
	return time.Since(e.LastOK) > interval, nil
}
