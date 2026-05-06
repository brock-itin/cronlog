// Package filter provides line-level output filtering for cron job logs.
//
// A Filter is constructed with a set of exact strings and regular expression
// patterns. Any output line that matches either an exact string or a compiled
// pattern will be suppressed before being forwarded to the structured logger
// or the log rotator.
//
// Usage:
//
//	f := filter.New(
//		[]string{`^DEBUG:`, `\bskipping\b`}, // regex patterns
//		[]string{"nothing to do"},            // exact strings
//	)
//
//	for _, line := range outputLines {
//		if !f.Suppress(line) {
//			logger.Info(line)
//		}
//	}
//
// Patterns that fail to compile are silently ignored so that a single
// malformed entry does not disable all filtering.
package filter
