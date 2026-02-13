package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	salerr "github.com/gongahkia/salja/internal/errors"
)

type Config struct {
	PreferredMode        string             `toml:"preferred_mode"`
	DefaultTimezone      string             `toml:"default_timezone"`
	ConflictStrategy     string             `toml:"conflict_strategy"`
	DataLossMode         string             `toml:"data_loss_mode"`
	StreamingThresholdMB int                `toml:"streaming_threshold_mb"`
	APITimeoutSeconds    int                `toml:"api_timeout_seconds"`
	PriorityMap          map[string]int     `toml:"priority_map"`
	TagMap               map[string]string  `toml:"tag_map"`
	ConflictThresholds   ConflictThresholds `toml:"conflict_thresholds"`
	API                  APIConfig          `toml:"api"`
}

type ConflictThresholds struct {
	LevenshteinThreshold int `toml:"levenshtein_threshold"`
	MinTitleLength       int `toml:"min_title_length"`
	DateProximityHours   int `toml:"date_proximity_hours"`
}

type APIConfig struct {
	TickTick  ServiceAuth `toml:"ticktick"`
	Todoist   ServiceAuth `toml:"todoist"`
	Google    ServiceAuth `toml:"google"`
	Microsoft ServiceAuth `toml:"microsoft"`
	Notion    ServiceAuth `toml:"notion"`
}

type ServiceAuth struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	RedirectURI  string `toml:"redirect_uri"`
	Token        string `toml:"token"`
}

func DefaultConfig() *Config {
	return &Config{
		PreferredMode:        "file",
		DefaultTimezone:      "UTC",
		ConflictStrategy:     "ask",
		DataLossMode:         "warn",
		StreamingThresholdMB: 10,
		APITimeoutSeconds:    30,
		PriorityMap:          map[string]int{},
		TagMap:               map[string]string{},
		ConflictThresholds: ConflictThresholds{
			LevenshteinThreshold: 3,
			MinTitleLength:       10,
			DateProximityHours:   24,
		},
	}
}

var overridePath string

// SetOverridePath sets a custom config file path, overriding the default.
func SetOverridePath(path string) {
	overridePath = path
}

func Load() (*Config, error) {
	if overridePath != "" {
		return LoadFrom(overridePath)
	}
	configPath := ConfigPath()
	return LoadFrom(configPath)
}

func LoadFrom(path string) (*Config, error) {
	cfg := DefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	md, err := toml.DecodeFile(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config at %s: %w", path, err)
	}

	undecoded := md.Undecoded()
	if len(undecoded) > 0 {
		for _, key := range undecoded {
			fmt.Fprintf(os.Stderr, "Warning: unknown config key '%s'\n", key)
		}
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config) error {
	validModes := map[string]bool{"file": true, "api": true}
	if !validModes[cfg.PreferredMode] {
		return &salerr.ValidationError{Field: "preferred_mode", Message: "must be 'file' or 'api', got '" + cfg.PreferredMode + "'"}
	}

	validStrategies := map[string]bool{"ask": true, "prefer-source": true, "prefer-target": true, "skip": true, "fail": true}
	if !validStrategies[cfg.ConflictStrategy] {
		return &salerr.ValidationError{Field: "conflict_strategy", Message: "invalid value '" + cfg.ConflictStrategy + "'"}
	}

	validDataLossModes := map[string]bool{"warn": true, "error": true, "silent": true}
	if !validDataLossModes[cfg.DataLossMode] {
		return &salerr.ValidationError{Field: "data_loss_mode", Message: "must be 'warn', 'error', or 'silent', got '" + cfg.DataLossMode + "'"}
	}

	if cfg.DefaultTimezone != "" {
		if _, err := time.LoadLocation(cfg.DefaultTimezone); err != nil {
			return &salerr.ValidationError{Field: "default_timezone", Message: "invalid value '" + cfg.DefaultTimezone + "'", Err: err}
		}
	}

	services := map[string]ServiceAuth{
		"ticktick": cfg.API.TickTick, "todoist": cfg.API.Todoist,
		"google": cfg.API.Google, "microsoft": cfg.API.Microsoft, "notion": cfg.API.Notion,
	}
	for name, svc := range services {
		if svc.ClientID != "" && svc.ClientSecret == "" {
			fmt.Fprintf(os.Stderr, "Warning: api.%s has client_id set but client_secret is empty\n", name)
		}
	}

	return nil
}

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "salja")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "salja")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

func CacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "salja")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "salja")
}

func DataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "salja")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "salja")
}
