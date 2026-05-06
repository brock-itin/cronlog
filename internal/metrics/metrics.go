// Package metrics provides lightweight job execution statistics
// collection for cronlog, tracking run counts, durations, and failure rates.
package metrics

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Stats holds cumulative execution statistics for a named cron job.
type Stats struct {
	JobName      string        `json:"job_name"`
	TotalRuns    int64         `json:"total_runs"`
	FailedRuns   int64         `json:"failed_runs"`
	LastExitCode int           `json:"last_exit_code"`
	LastRunAt    time.Time     `json:"last_run_at"`
	TotalRuntime time.Duration `json:"total_runtime_ns"`
}

// Collector accumulates job stats and persists them to a JSON file.
type Collector struct {
	mu      sync.Mutex
	stats   map[string]*Stats
	path    string
}

// New creates a Collector that persists state to path.
// Existing data is loaded if the file is present.
func New(path string) (*Collector, error) {
	c := &Collector{
		stats: make(map[string]*Stats),
		path:  path,
	}
	if err := c.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return c, nil
}

// Record updates statistics for jobName with the result of a completed run.
func (c *Collector) Record(jobName string, exitCode int, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.stats[jobName]
	if !ok {
		s = &Stats{JobName: jobName}
		c.stats[jobName] = s
	}

	s.TotalRuns++
	s.LastExitCode = exitCode
	s.LastRunAt = time.Now().UTC()
	s.TotalRuntime += duration
	if exitCode != 0 {
		s.FailedRuns++
	}

	return c.save()
}

// Get returns a copy of the stats for jobName, and whether they exist.
func (c *Collector) Get(jobName string) (Stats, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.stats[jobName]
	if !ok {
		return Stats{}, false
	}
	return *s, true
}

func (c *Collector) load() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &c.stats)
}

func (c *Collector) save() error {
	data, err := json.MarshalIndent(c.stats, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0644)
}
