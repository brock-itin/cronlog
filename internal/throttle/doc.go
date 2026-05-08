// Package throttle provides a Runner wrapper that enforces a minimum interval
// between job executions.
//
// When a cron job is configured with a min_gap duration, the Throttle wrapper
// inspects the job's execution history before delegating to the real runner.
// If the most recent entry is more recent than min_gap, the job is skipped and
// an ErrThrottled error is returned instead of executing the underlying command.
//
// This is useful for jobs that are scheduled more frequently than they should
// actually run under normal circumstances, or for preventing rapid re-execution
// after a manual trigger.
//
// Example usage:
//
//	th := throttle.New(runner, hist, "backup", 30*time.Minute)
//	code, err := th.Run()
//	var te *throttle.ErrThrottled
//	if errors.As(err, &te) {
//		// job was skipped — log and exit cleanly
//	}
package throttle
