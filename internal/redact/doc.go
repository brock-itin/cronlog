// Package redact scrubs sensitive values from cron job output before the
// text reaches log files, notification webhooks, or any other sink.
//
// Usage:
//
//	r := redact.New(
//		[]string{"my-api-key", "db-password"},   // literal strings
//		[]string{`(?i)token=\S+`},               // regex patterns
//	)
//	clean := r.Apply(rawOutput)
//
// Both literal replacement and regex substitution replace matched text with
// the fixed placeholder "[REDACTED]". Literal matching is case-sensitive;
// use a regex pattern when case-insensitive matching is required.
//
// Invalid regex patterns are silently ignored so that a misconfigured
// pattern does not prevent the rest of the pipeline from running.
package redact
