// Package main is the entry point for the cronlog command-line tool.
//
// cronlog wraps cron job commands and captures their output to structured,
// rotating log files. It is intended to be used as a thin wrapper in crontab
// entries:
//
//	* * * * * cronlog --job my-job --config /etc/cronlog/cronlog.yaml -- /usr/local/bin/my-script.sh
//
// Flags:
//
//	--config  Path to the YAML configuration file (default: /etc/cronlog/cronlog.yaml)
//	--job     Name of the cron job; used to derive the log subdirectory and file prefix
//
// All output (stdout and stderr) from the wrapped command is captured and
// written as newline-delimited JSON log entries. Log files are automatically
// rotated and pruned according to the max_files setting in the config.
package main
