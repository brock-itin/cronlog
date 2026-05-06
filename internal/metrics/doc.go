// Package metrics implements lightweight, file-backed statistics collection
// for cronlog job executions.
//
// A Collector is created with a path to a JSON file that serves as persistent
// storage. On each job completion, Record updates the in-memory stats and
// atomically rewrites the file so that data survives process restarts.
//
// Typical usage:
//
//	col, err := metrics.New("/var/log/cronlog/metrics.json")
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := col.Record(jobName, exitCode, duration); err != nil {
//		log.Printf("metrics: %v", err)
//	}
//
// Stats are keyed by job name and include total runs, failed runs,
// last exit code, last run timestamp, and cumulative runtime.
package metrics
