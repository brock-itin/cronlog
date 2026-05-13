package audit

import (
	"bufio"
	"os"
	"sync"
	"testing"
)

func TestRecord_ConcurrentSafe(t *testing.T) {
	p := tempPath(t)
	a, err := New(p)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer a.Close()

	const workers = 10
	const perWorker = 5

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				if err := a.Record("job", "tick", 0, nil); err != nil {
					t.Errorf("Record: %v", err)
				}
			}
		}()
	}
	wg.Wait()
	a.Close()

	f, _ := os.Open(p)
	defer f.Close()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	expected := workers * perWorker
	if count != expected {
		t.Errorf("lines = %d, want %d", count, expected)
	}
}
