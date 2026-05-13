// Package stagger wraps a Runner to introduce a random delay before job
// execution. This is useful when multiple cron jobs share the same schedule
// and would otherwise hammer downstream resources simultaneously.
//
// # Usage
//
//	s, err := stagger.New(runner, 5*time.Minute, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	exitCode, err := s.Run(ctx, "backup.sh", nil, nil)
//
// The actual delay is chosen uniformly at random from [0, max). If the
// context is cancelled while sleeping, Run returns immediately with the
// context error and the underlying runner is never invoked.
package stagger
