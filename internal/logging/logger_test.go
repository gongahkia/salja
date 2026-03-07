package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestLogWritesJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	l, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	l.Info("interaction", "user opened convert")
	l.Error("error", "file not found")
	l.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}
	for _, line := range lines {
		var e entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("invalid JSON line: %s", line)
		}
		if e.Ts == "" || e.Level == "" || e.Cat == "" || e.Msg == "" {
			t.Fatalf("missing fields in log entry: %+v", e)
		}
	}
}

func TestRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	l, err := NewWithMaxSize(path, 50) // 50 bytes threshold
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		l.Info("test", "some log message that is fairly long")
	}
	l.Close()
	if _, err := os.Stat(path + ".1"); os.IsNotExist(err) {
		t.Fatal("expected rotated log file")
	}
}

func TestConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")
	l, err := New(path)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.Info("test", "concurrent write")
		}()
	}
	wg.Wait()
	l.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 50 {
		t.Fatalf("expected 50 lines, got %d", len(lines))
	}
}

func TestNopLogger(t *testing.T) {
	l := Nop()
	l.Info("test", "should not panic") // should not panic
	l.Close()
}
