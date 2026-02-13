package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gongahkia/salja/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage salja configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print config file path",
		Run: func(cmd *cobra.Command, args []string) {
			configDir := getConfigDir()
			fmt.Println(filepath.Join(configDir, "config.toml"))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := getConfigDir()
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return err
			}
			configPath := filepath.Join(configDir, "config.toml")
			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("config already exists at %s", configPath)
			}
			defaultConfig := `# salja configuration
preferred_mode = "file"
default_timezone = "UTC"
conflict_strategy = "ask"
data_loss_mode = "warn"
streaming_threshold_mb = 10
api_timeout_seconds = 30

[conflict_thresholds]
levenshtein_threshold = 3
min_title_length = 10
date_proximity_hours = 24

[priority_map]

[tag_map]

[api.ticktick]
client_id = ""
client_secret = ""

[api.todoist]
client_id = ""
client_secret = ""

[api.google]
client_id = ""
client_secret = ""

[api.microsoft]
client_id = ""
client_secret = ""

[api.notion]
token = ""
`
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				return err
			}
			fmt.Printf("Created config at %s\n", configPath)
			return nil
		},
	})

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print current configuration as TOML",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			return toml.NewEncoder(os.Stdout).Encode(cfg)
		},
	}
}

func getConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "salja")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "salja")
}
