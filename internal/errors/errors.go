package errors

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ParseError struct {
	File    string
	Line    int
	Message string
	Err     error
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error in %s at line %d: %s", e.File, e.Line, e.Message)
	}
	return fmt.Sprintf("parse error in %s: %s", e.File, e.Message)
}

func (e *ParseError) Unwrap() error { return e.Err }

type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error [%s]: %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error { return e.Err }

type ConflictError struct {
	SourceItem string
	TargetItem string
	Message    string
	Err        error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict between '%s' and '%s': %s", e.SourceItem, e.TargetItem, e.Message)
}

func (e *ConflictError) Unwrap() error { return e.Err }

type APIError struct {
	Service    string
	StatusCode int
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error [%s] (HTTP %d): %s", e.Service, e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error { return e.Err }

func (e *APIError) IsRateLimit() bool { return e.StatusCode == 429 }

type PermissionError struct {
	Resource string
	Message  string
	Err      error
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied for %s: %s", e.Resource, e.Message)
}

func (e *PermissionError) Unwrap() error { return e.Err }

type ErrorCollector struct {
	Errors   []error
	Warnings []string
}

func NewErrorCollector() *ErrorCollector {
	return &ErrorCollector{}
}

func (c *ErrorCollector) AddError(err error) {
	c.Errors = append(c.Errors, err)
}

func (c *ErrorCollector) AddWarning(msg string) {
	c.Warnings = append(c.Warnings, msg)
}

func (c *ErrorCollector) HasErrors() bool {
	return len(c.Errors) > 0
}

func (c *ErrorCollector) Summary() string {
	return fmt.Sprintf("%d errors, %d warnings", len(c.Errors), len(c.Warnings))
}

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
}

func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
	}
}

func Retry(cfg *RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if apiErr, ok := lastErr.(*APIError); ok && !apiErr.IsRateLimit() {
			return lastErr
		}

		if attempt < cfg.MaxRetries {
			delay := cfg.BaseDelay * time.Duration(math.Pow(2, float64(attempt)))
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(delay + jitter)
		}
	}
	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, lastErr)
}

type SignalHandler struct {
	cleanup func()
}

func NewSignalHandler(cleanup func()) *SignalHandler {
	return &SignalHandler{cleanup: cleanup}
}

func (h *SignalHandler) Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Fprintln(os.Stderr, "\nInterrupted. Cleaning up...")
		if h.cleanup != nil {
			h.cleanup()
		}
		os.Exit(1)
	}()
}
