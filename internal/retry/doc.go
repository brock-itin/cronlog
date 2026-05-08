// Package retry provides a simple, configurable retry runner for use within
// cronlog when transient failures should be tolerated before marking a job
// as failed.
//
// Usage:
//
//	r := retry.New(retry.Policy{
//		MaxAttempts: 3,
//		Delay:       2 * time.Second,
//		Backoff:     2.0, // exponential: 2s, 4s
//	})
//
//	res := r.Run(ctx, func() error {
//		return runMyJob()
//	})
//	if res.Err != nil {
//		log.Printf("job failed after %d attempts: %v", res.Attempts, res.Err)
//	}
//
// A Backoff of 1.0 produces a constant delay between attempts. Context
// cancellation is checked before each attempt so callers can abort early.
package retry
