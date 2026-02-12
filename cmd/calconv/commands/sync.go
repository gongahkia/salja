package commands

import (
"fmt"

"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
var autoResolve string

cmd := &cobra.Command{
Use:   "sync <source-service> <target-service>",
Short: "Sync between calendar services via API",
Long:  "Pull from source service, diff with target, push changes with conflict resolution.\nRequires API connectors to be configured (see `calconv auth` and `calconv config`).",
Args:  cobra.ExactArgs(2),
RunE: func(cmd *cobra.Command, args []string) error {
source := args[0]
target := args[1]
return fmt.Errorf("sync from %s to %s is not yet implemented — API connectors required.\nConfigure with: calconv auth login %s", source, target, source)
},
}

cmd.Flags().StringVar(&autoResolve, "auto-resolve", "", "Auto-resolve strategy: prefer-source, prefer-target, skip, fail")
return cmd
}
