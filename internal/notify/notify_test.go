package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/cronlog/internal/notify"
)

func TestNotify_NoOp_WhenURLEmpty(t *testing.T) {
	n := notify.New("")
	if err := n.Notify("backup", 1); err != nil {
		t.Fatalf("expected no error with empty URL, got: %v", err)
	}
}

func TestNotify_NoOp_WhenExitCodeZero(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer ts.Close()

	n := notify.New(ts.URL)
	if err := n.Notify("backup", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("webhook should not be called for exit code 0")
	}
}

func TestNotify_PostsPayload_OnFailure(t *testing.T) {
	var received notify.Payload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected content-type: %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.New(ts.URL)
	if err := n.Notify("cleanup", 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Job != "cleanup" {
		t.Errorf("expected job=cleanup, got %q", received.Job)
	}
	if received.ExitCode != 2 {
		t.Errorf("expected exit_code=2, got %d", received.ExitCode)
	}
	if received.Message == "" {
		t.Error("expected non-empty message")
	}
	if received.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestNotify_ReturnsError_OnBadStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.New(ts.URL)
	if err := n.Notify("job", 1); err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}
