package commands

import (
"fmt"
"os"
"path/filepath"

"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
cmd := &cobra.Command{
Use:   "config",
Short: "Manage calconv configuration",
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
defaultConfig := `# calconv configuration
preferred_mode = "file"
default_timezone = "UTC"
conflict_strategy = "ask"

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

func getConfigDir() string {
if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
return filepath.Join(xdg, "calconv")
}
home, _ := os.UserHomeDir()
return filepath.Join(home, ".config", "calconv")
}
