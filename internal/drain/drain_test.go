package drain_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/cronlog/internal/drain"
)

func TestNew_PanicsOnNilHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil handler")
		}
	}()
	drain.New(nil)
}

func TestWrite_DispatchesCompleteLines(t *testing.T) {
	var got []string
	d := drain.New(func(line string) { got = append(got, line) })

	_, err := d.Write([]byte("hello\nworld\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(got))
	}
	if got[0] != "hello" || got[1] != "world" {
		t.Errorf("unexpected lines: %v", got)
	}
}

func TestWrite_BuffersPartialLines(t *testing.T) {
	var got []string
	d := drain.New(func(line string) { got = append(got, line) })

	d.Write([]byte("par"))
	if len(got) != 0 {
		t.Fatal("expected no lines dispatched for partial write")
	}

	d.Write([]byte("tial\n"))
	if len(got) != 1 || got[0] != "partial" {
		t.Errorf("expected 'partial', got %v", got)
	}
}

func TestClose_FlushesRemainingData(t *testing.T) {
	var got []string
	d := drain.New(func(line string) { got = append(got, line) })

	d.Write([]byte("no newline"))
	if len(got) != 0 {
		t.Fatal("line should not be dispatched before Close")
	}

	d.Close()
	if len(got) != 1 || got[0] != "no newline" {
		t.Errorf("expected 'no newline', got %v", got)
	}
}

func TestWrite_ConcurrentSafe(t *testing.T) {
	var mu sync.Mutex
	var got []string
	d := drain.New(func(line string) {
		mu.Lock()
		got = append(got, line)
		mu.Unlock()
	})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Write([]byte("line\n"))
		}()
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 20 {
		t.Errorf("expected 20 lines, got %d", len(got))
	}
}

func TestWriterTo_ImplementsWriteCloser(t *testing.T) {
	d := drain.New(func(string) {})
	wc := d.WriterTo()
	wc.Write([]byte(strings.Repeat("x", 10) + "\n"))
	if err := wc.Close(); err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}
