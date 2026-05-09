// Package tag provides utilities for attaching static key-value metadata
// (tags) to every structured log entry emitted by cronlog.
package tag

import "strings"

// Tagger holds a resolved set of key=value tags parsed from configuration.
type Tagger struct {
	tags map[string]string
}

// New parses a slice of "key=value" strings and returns a Tagger.
// Entries that do not contain "=" or have an empty key are silently ignored.
func New(raw []string) *Tagger {
	tags := make(map[string]string, len(raw))
	for _, entry := range raw {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		tags[key] = val
	}
	return &Tagger{tags: tags}
}

// Apply merges the tagger's static tags into dst, preferring existing keys in
// dst so that per-run fields are never overwritten by static metadata.
func (t *Tagger) Apply(dst map[string]string) map[string]string {
	if dst == nil {
		dst = make(map[string]string, len(t.tags))
	}
	for k, v := range t.tags {
		if _, exists := dst[k]; !exists {
			dst[k] = v
		}
	}
	return dst
}

// Tags returns a copy of the resolved tag map.
func (t *Tagger) Tags() map[string]string {
	copy := make(map[string]string, len(t.tags))
	for k, v := range t.tags {
		copy[k] = v
	}
	return copy
}

// Empty reports whether no tags have been configured.
func (t *Tagger) Empty() bool {
	return len(t.tags) == 0
}
