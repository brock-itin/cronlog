package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourorg/cronlog/internal/config"
	"github.com/yourorg/cronlog/internal/logger"
	"github.com/yourorg/cronlog/internal/rotator"
	"github.com/yourorg/cronlog/internal/runner"
)

const defaultConfigPath = "/etc/cronlog/cronlog.yaml"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "cronlog: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", defaultConfigPath, "path to config file")
	jobName := flag.String("job", "", "name of the cron job (used in log file naming)")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("usage: cronlog [flags] -- <command> [args...]")
	}

	if *jobName == "" {
		return fmt.Errorf("--job flag is required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	rot, err := rotator.New(cfg.LogDir, *jobName, cfg.MaxFiles)
	if err != nil {
		return fmt.Errorf("creating rotator: %w", err)
	}

	w, err := rot.Open()
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	defer w.Close()

	log := logger.New(w, *jobName)
	log.Info("starting", args)

	r := runner.New(args[0], args[1:]...)
	result, err := r.Run()
	if err != nil {
		log.Error("command error", err, result.ExitCode)
		return nil
	}

	log.Done(result.Duration, result.ExitCode)

	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}
