package config

import (
"fmt"
"os"
"path/filepath"

"github.com/BurntSushi/toml"
)

type Config struct {
PreferredMode    string            `toml:"preferred_mode"`
DefaultTimezone  string            `toml:"default_timezone"`
ConflictStrategy string            `toml:"conflict_strategy"`
DataLossMode     string            `toml:"data_loss_mode"`
PriorityMap      map[string]int    `toml:"priority_map"`
TagMap           map[string]string `toml:"tag_map"`
API              APIConfig         `toml:"api"`
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
PreferredMode:    "file",
DefaultTimezone:  "UTC",
ConflictStrategy: "ask",
DataLossMode:     "warn",
PriorityMap:      map[string]int{},
TagMap:           map[string]string{},
}
}

func Load() (*Config, error) {
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
return fmt.Errorf("invalid preferred_mode '%s': must be 'file' or 'api'", cfg.PreferredMode)
}

validStrategies := map[string]bool{"ask": true, "prefer-source": true, "prefer-target": true, "skip": true, "fail": true}
if !validStrategies[cfg.ConflictStrategy] {
return fmt.Errorf("invalid conflict_strategy '%s'", cfg.ConflictStrategy)
}

validDataLossModes := map[string]bool{"warn": true, "error": true, "silent": true}
if !validDataLossModes[cfg.DataLossMode] {
return fmt.Errorf("invalid data_loss_mode '%s': must be 'warn', 'error', or 'silent'", cfg.DataLossMode)
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
