package env

import (
	"os"
	"strings"
	"testing"
)

func findKey(env []string, key string) (string, bool) {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix), true
		}
	}
	return "", false
}

func TestResolve_InheritsProcessEnv(t *testing.T) {
	t.Setenv("CRONLOG_TEST_VAR", "hello")

	r := New(nil, nil)
	env := r.Resolve()

	v, ok := findKey(env, "CRONLOG_TEST_VAR")
	if !ok {
		t.Fatal("expected CRONLOG_TEST_VAR to be present")
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %q", v)
	}
}

func TestResolve_ExtrasOverrideInherited(t *testing.T) {
	t.Setenv("CRONLOG_TEST_VAR", "original")

	r := New(map[string]string{"CRONLOG_TEST_VAR": "overridden"}, nil)
	env := r.Resolve()

	v, ok := findKey(env, "CRONLOG_TEST_VAR")
	if !ok {
		t.Fatal("expected CRONLOG_TEST_VAR to be present")
	}
	if v != "overridden" {
		t.Fatalf("expected 'overridden', got %q", v)
	}
}

func TestResolve_MaskedKeyIsEmpty(t *testing.T) {
	t.Setenv("SECRET_TOKEN", "super-secret")

	r := New(nil, []string{"SECRET_TOKEN"})
	env := r.Resolve()

	v, ok := findKey(env, "SECRET_TOKEN")
	if !ok {
		t.Fatal("expected SECRET_TOKEN to be present but empty")
	}
	if v != "" {
		t.Fatalf("expected empty value for masked key, got %q", v)
	}
}

func TestResolve_ExtraKeyAddedToEnv(t *testing.T) {
	os.Unsetenv("CRONLOG_BRAND_NEW")

	r := New(map[string]string{"CRONLOG_BRAND_NEW": "42"}, nil)
	env := r.Resolve()

	v, ok := findKey(env, "CRONLOG_BRAND_NEW")
	if !ok {
		t.Fatal("expected CRONLOG_BRAND_NEW to be present")
	}
	if v != "42" {
		t.Fatalf("expected '42', got %q", v)
	}
}

func TestResolve_MaskedAndExtraInteraction(t *testing.T) {
	r := New(
		map[string]string{"CRONLOG_SECRET": "injected-secret"},
		[]string{"CRONLOG_SECRET"},
	)
	env := r.Resolve()

	v, ok := findKey(env, "CRONLOG_SECRET")
	if !ok {
		t.Fatal("expected CRONLOG_SECRET to be present")
	}
	if v != "" {
		t.Fatalf("masking should win over injection, got %q", v)
	}
}
