package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	mu           sync.Mutex
	f            *os.File
	path         string
	maxSizeBytes int64
}

type entry struct {
	Ts    string `json:"ts"`
	Level string `json:"level"`
	Cat   string `json:"cat"`
	Msg   string `json:"msg"`
}

const defaultMaxSize = 10 * 1024 * 1024 // 10MB

func New(path string) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	return &Logger{f: f, path: path, maxSizeBytes: defaultMaxSize}, nil
}

func NewWithMaxSize(path string, maxSize int64) (*Logger, error) {
	l, err := New(path)
	if err != nil {
		return nil, err
	}
	l.maxSizeBytes = maxSize
	return l, nil
}

func (l *Logger) Path() string { return l.path }

func (l *Logger) Log(level, cat, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rotate()
	e := entry{
		Ts:    time.Now().Format(time.RFC3339Nano),
		Level: level,
		Cat:   cat,
		Msg:   msg,
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	data = append(data, '\n')
	_, _ = l.f.Write(data)
}

func (l *Logger) Info(cat, msg string)  { l.Log("info", cat, msg) }
func (l *Logger) Warn(cat, msg string)  { l.Log("warn", cat, msg) }
func (l *Logger) Error(cat, msg string) { l.Log("error", cat, msg) }

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f != nil {
		return l.f.Close()
	}
	return nil
}

func (l *Logger) rotate() {
	fi, err := l.f.Stat()
	if err != nil || fi.Size() < l.maxSizeBytes {
		return
	}
	_ = l.f.Close()
	_ = os.Rename(l.path, l.path+".1")
	l.f, _ = os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// Nop returns a no-op logger that discards all output.
func Nop() *Logger {
	return &Logger{}
}

var defaultLogger = Nop()

// Init initializes the global default logger at the given path.
func Init(path string) error {
	l, err := New(path)
	if err != nil {
		return err
	}
	defaultLogger = l
	return nil
}

// Default returns the global logger instance.
func Default() *Logger { return defaultLogger }

// Shutdown closes the global logger.
func Shutdown() { defaultLogger.Close() }
