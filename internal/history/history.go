package history

import (
	"encoding/json"
	"os"
	"sort"
	"time"
)

// Entry represents a single historical run record for a cron job.
type Entry struct {
	Job       string        `json:"job"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ms"`
	ExitCode  int           `json:"exit_code"`
	LogFile   string        `json:"log_file"`
}

// History manages a persistent list of run entries for a job.
type History struct {
	path    string
	maxSize int
	entries []Entry
}

// New loads (or creates) a history file at the given path.
// maxSize controls how many entries are retained.
func New(path string, maxSize int) (*History, error) {
	h := &History{path: path, maxSize: maxSize}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return h, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &h.entries); err != nil {
		return nil, err
	}
	return h, nil
}

// Add appends a new entry and prunes old ones beyond maxSize.
func (h *History) Add(e Entry) error {
	h.entries = append(h.entries, e)
	sort.Slice(h.entries, func(i, j int) bool {
		return h.entries[i].StartedAt.Before(h.entries[j].StartedAt)
	})
	if h.maxSize > 0 && len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
	return h.save()
}

// Entries returns a copy of all stored entries.
func (h *History) Entries() []Entry {
	out := make([]Entry, len(h.entries))
	copy(out, h.entries)
	return out
}

// LastFailure returns the most recent failed entry, or nil if none.
func (h *History) LastFailure() *Entry {
	for i := len(h.entries) - 1; i >= 0; i-- {
		if h.entries[i].ExitCode != 0 {
			e := h.entries[i]
			return &e
		}
	}
	return nil
}

func (h *History) save() error {
	data, err := json.MarshalIndent(h.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0o644)
}
