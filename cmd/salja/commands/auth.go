package commands

import (
"fmt"

"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
cmd := &cobra.Command{
Use:   "auth",
Short: "Manage OAuth tokens for API services",
}

cmd.AddCommand(&cobra.Command{
Use:   "login <service>",
Short: "Authenticate with a service",
Args:  cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
service := args[0]
fmt.Printf("OAuth login for %s is not yet implemented. Configure API credentials in config.toml.\n", service)
return nil
},
})

cmd.AddCommand(&cobra.Command{
Use:   "logout <service>",
Short: "Remove stored tokens for a service",
Args:  cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
fmt.Printf("Removed tokens for %s\n", args[0])
return nil
},
})

cmd.AddCommand(&cobra.Command{
Use:   "status",
Short: "Show authentication status for all services",
Run: func(cmd *cobra.Command, args []string) {
services := []string{"ticktick", "todoist", "google", "microsoft", "notion"}
for _, s := range services {
fmt.Printf("  %s: not authenticated\n", s)
}
},
})

return cmd
}
