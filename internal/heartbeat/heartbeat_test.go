package heartbeat_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cronlog/internal/heartbeat"
)

func TestPing_NoOp_WhenURLEmpty(t *testing.T) {
	p := heartbeat.New("", 0)
	if err := p.Ping(0); err != nil {
		t.Fatalf("expected no error for empty URL, got %v", err)
	}
}

func TestPing_Success_OnZeroExit(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := heartbeat.New(srv.URL+"/ping/abc123", time.Second)
	if err := p.Ping(0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/ping/abc123" {
		t.Errorf("expected path /ping/abc123, got %s", gotPath)
	}
}

func TestPing_AppendsFail_OnNonZeroExit(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := heartbeat.New(srv.URL+"/ping/abc123", time.Second)
	if err := p.Ping(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(gotPath, "/fail") {
		t.Errorf("expected path to end with /fail, got %s", gotPath)
	}
}

func TestPing_ReturnsError_OnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	p := heartbeat.New(srv.URL+"/ping/abc123", time.Second)
	if err := p.Ping(0); err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestPing_ReturnsError_OnUnreachableHost(t *testing.T) {
	p := heartbeat.New("http://127.0.0.1:19999/ping/xyz", 100*time.Millisecond)
	if err := p.Ping(0); err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}
