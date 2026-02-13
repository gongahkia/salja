package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMissingFileDefaults(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("expected no error for missing config, got: %v", err)
	}
	if cfg.PreferredMode != "file" {
		t.Errorf("expected default preferred_mode 'file', got '%s'", cfg.PreferredMode)
	}
	if cfg.DefaultTimezone != "UTC" {
		t.Errorf("expected default timezone 'UTC', got '%s'", cfg.DefaultTimezone)
	}
	if cfg.ConflictStrategy != "ask" {
		t.Errorf("expected default conflict_strategy 'ask', got '%s'", cfg.ConflictStrategy)
	}
}

func TestPartialConfigMerge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	_ = os.WriteFile(path, []byte(`default_timezone = "America/New_York"`), 0644)

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultTimezone != "America/New_York" {
		t.Errorf("expected timezone override, got '%s'", cfg.DefaultTimezone)
	}
	if cfg.PreferredMode != "file" {
		t.Errorf("expected default preferred_mode preserved, got '%s'", cfg.PreferredMode)
	}
}

func TestInvalidTOMLSyntax(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	_ = os.WriteFile(path, []byte(`not valid toml [[[`), 0644)

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestUnknownKeyWarning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	_ = os.WriteFile(path, []byte(`unknown_setting = "value"`), 0644)

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PreferredMode != "file" {
		t.Errorf("expected defaults preserved with unknown key")
	}
}
