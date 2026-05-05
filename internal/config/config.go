// Package config provides configuration loading and validation for cronlog.
package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level cronlog configuration.
type Config struct {
	LogDir     string        `yaml:"log_dir"`
	MaxFiles   int           `yaml:"max_files"`
	MaxSizeMB  int           `yaml:"max_size_mb"`
	Timeout    time.Duration `yaml:"timeout"`
	JobName    string        `yaml:"job_name"`
}

// Defaults applied when fields are zero-valued.
const (
	DefaultLogDir    = "/var/log/cronlog"
	DefaultMaxFiles  = 7
	DefaultMaxSizeMB = 50
	DefaultTimeout   = 30 * time.Minute
)

// Load reads a YAML config file from path and returns a validated Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// applyDefaults fills in zero-valued fields with sensible defaults.
func applyDefaults(cfg *Config) {
	if cfg.LogDir == "" {
		cfg.LogDir = DefaultLogDir
	}
	if cfg.MaxFiles == 0 {
		cfg.MaxFiles = DefaultMaxFiles
	}
	if cfg.MaxSizeMB == 0 {
		cfg.MaxSizeMB = DefaultMaxSizeMB
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}
}

// validate returns an error if the config contains invalid values.
func validate(cfg *Config) error {
	if cfg.MaxFiles < 1 {
		return errors.New("config: max_files must be at least 1")
	}
	if cfg.MaxSizeMB < 1 {
		return errors.New("config: max_size_mb must be at least 1")
	}
	if cfg.Timeout < time.Second {
		return errors.New("config: timeout must be at least 1 second")
	}
	return nil
}
