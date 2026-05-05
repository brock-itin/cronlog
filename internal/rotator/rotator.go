package rotator

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config holds configuration for log rotation.
type Config struct {
	Dir        string
	BaseName   string
	MaxSizeMB  int64
	MaxFiles   int
}

// Rotator manages a rotating log file.
type Rotator struct {
	cfg     Config
	mu      sync.Mutex
	file    *os.File
	current string
}

// New creates a new Rotator with the given config.
func New(cfg Config) (*Rotator, error) {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("rotator: create dir: %w", err)
	}
	r := &Rotator{cfg: cfg}
	if err := r.openNew(); err != nil {
		return nil, err
	}
	return r, nil
}

// Write implements io.Writer, rotating the file if necessary.
func (r *Rotator) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.rotateIfNeeded(); err != nil {
		return 0, err
	}
	return r.file.Write(p)
}

// Close closes the underlying log file.
func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

func (r *Rotator) rotateIfNeeded() error {
	if r.cfg.MaxSizeMB <= 0 {
		return nil
	}
	info, err := r.file.Stat()
	if err != nil {
		return fmt.Errorf("rotator: stat: %w", err)
	}
	if info.Size() >= r.cfg.MaxSizeMB*1024*1024 {
		if err := r.file.Close(); err != nil {
			return err
		}
		if err := r.pruneOld(); err != nil {
			return err
		}
		return r.openNew()
	}
	return nil
}

func (r *Rotator) openNew() error {
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("%s-%s.log", r.cfg.BaseName, timestamp)
	path := filepath.Join(r.cfg.Dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("rotator: open file: %w", err)
	}
	r.file = f
	r.current = path
	return nil
}

func (r *Rotator) pruneOld() error {
	if r.cfg.MaxFiles <= 0 {
		return nil
	}
	pattern := filepath.Join(r.cfg.Dir, r.cfg.BaseName+"-*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("rotator: glob: %w", err)
	}
	for len(matches) >= r.cfg.MaxFiles {
		if err := os.Remove(matches[0]); err != nil {
			return fmt.Errorf("rotator: remove old log: %w", err)
		}
		matches = matches[1:]
	}
	return nil
}
