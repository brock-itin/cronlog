// Package alert implements threshold-based alerting for cronlog job runs.
//
// An Evaluator is configured with one or more thresholds:
//
//   - MaxConsecutiveFailures: triggers an Alert when the trailing N job
//     runs all returned a non-zero exit code.
//
//   - MaxFailureRate / FailureRateWindow: triggers an Alert when the
//     proportion of failed runs within the given time window exceeds the
//     configured rate (expressed as a value between 0.0 and 1.0).
//
// Typical usage:
//
//	ev := alert.New(alert.Config{
//	    MaxConsecutiveFailures: 3,
//	    FailureRateWindow:      time.Hour,
//	    MaxFailureRate:         0.5,
//	})
//
//	if a := ev.Evaluate(jobName, entries); a != nil {
//	    log.Printf("alert: %s – %s", a.JobName, a.Reason)
//	}
package alert
