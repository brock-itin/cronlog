// Package config handles loading, parsing, and validating cronlog configuration
// from YAML files.
//
// A minimal configuration file looks like:
//
//	log_dir: /var/log/cronlog
//	max_files: 7
//	max_size_mb: 50
//	timeout: 30m
//	job_name: my-job
//
// All fields are optional; sensible defaults are applied automatically.
// Use config.Load to obtain a validated *Config ready for use by other
// cronlog components such as the rotator and runner.
package config
