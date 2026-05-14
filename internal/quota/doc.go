// Package quota provides a Runner decorator that enforces a maximum number of
// cron job executions within a rolling time window.
//
// It reads recent run history from a [history.History] store and compares the
// count of entries that fall within the configured window against the
// configured maximum. If the limit has already been reached the job is skipped
// and an error is returned — no side-effects are produced by the wrapped
// runner.
//
// Typical usage:
//
//	q, err := quota.New(runner, hist, 10, 24*time.Hour)
//	if err != nil {
//		log.Fatal(err)
//	}
//	exitCode, err := q.Run(ctx, args)
//
// A zero or negative max, a non-positive window, a nil runner, or a nil
// history will cause New to return an error immediately.
package quota
