// Package alert provides threshold-based alerting for cron job metrics.
// It evaluates recent job history and emits alerts when failure rates or
// consecutive failure counts exceed configured thresholds.
package alert

import (
	"fmt"
	"time"
)

// Config holds alerting thresholds.
type Config struct {
	// MaxConsecutiveFailures triggers an alert after N consecutive failures.
	MaxConsecutiveFailures int
	// FailureRateWindow is the duration over which failure rate is evaluated.
	FailureRateWindow time.Duration
	// MaxFailureRate is the maximum tolerated failure rate (0.0–1.0).
	MaxFailureRate float64
}

// Entry represents a single job run result used for evaluation.
type Entry struct {
	JobName  string
	ExitCode int
	RunAt    time.Time
}

// Alert describes a triggered alert condition.
type Alert struct {
	JobName string
	Reason  string
	At      time.Time
}

// Evaluator checks job history entries against configured thresholds.
type Evaluator struct {
	cfg Config
}

// New creates a new Evaluator with the given Config.
func New(cfg Config) *Evaluator {
	return &Evaluator{cfg: cfg}
}

// Evaluate inspects entries for the named job and returns any triggered Alert.
// Returns nil if no threshold is breached.
func (e *Evaluator) Evaluate(jobName string, entries []Entry) *Alert {
	if len(entries) == 0 {
		return nil
	}

	if e.cfg.MaxConsecutiveFailures > 0 {
		count := 0
		for i := len(entries) - 1; i >= 0; i-- {
			if entries[i].ExitCode != 0 {
				count++
			} else {
				break
			}
		}
		if count >= e.cfg.MaxConsecutiveFailures {
			return &Alert{
				JobName: jobName,
				Reason:  fmt.Sprintf("%d consecutive failures", count),
				At:      time.Now(),
			}
		}
	}

	if e.cfg.MaxFailureRate > 0 && e.cfg.FailureRateWindow > 0 {
		cutoff := time.Now().Add(-e.cfg.FailureRateWindow)
		var total, failed int
		for _, en := range entries {
			if en.RunAt.After(cutoff) {
				total++
				if en.ExitCode != 0 {
					failed++
				}
			}
		}
		if total > 0 {
			rate := float64(failed) / float64(total)
			if rate > e.cfg.MaxFailureRate {
				return &Alert{
					JobName: jobName,
					Reason:  fmt.Sprintf("failure rate %.0f%% exceeds threshold %.0f%%", rate*100, e.cfg.MaxFailureRate*100),
					At:      time.Now(),
				}
			}
		}
	}

	return nil
}
