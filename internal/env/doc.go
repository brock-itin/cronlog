// Package env manages environment variable resolution for cron job subprocesses.
//
// It merges the current process environment with any extra key=value pairs
// defined in the job configuration, allowing operators to inject runtime
// variables (e.g. PATH overrides, credentials) without modifying the system
// crontab.
//
// Sensitive keys can be listed under the "masked" configuration option; their
// values are replaced with an empty string in the child process environment so
// that secrets are never forwarded unintentionally.
//
// Usage:
//
//	r := env.New(
//		map[string]string{"APP_ENV": "production"},
//		[]string{"AWS_SECRET_ACCESS_KEY"},
//	)
//	cmd.Env = r.Resolve()
package env
