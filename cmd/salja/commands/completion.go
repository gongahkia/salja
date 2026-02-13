package commands

import (
"os"

"github.com/spf13/cobra"
)

func NewCompletionCmd() *cobra.Command {
cmd := &cobra.Command{
Use:   "completion [bash|zsh|fish|powershell]",
Short: "Generate shell completion scripts",
Long: `Generate shell completion scripts for salja.

To load completions:

Bash:
  $ source <(salja completion bash)
  # Or add to ~/.bashrc:
  $ salja completion bash > /etc/bash_completion.d/salja

Zsh:
  $ salja completion zsh > "${fpath[1]}/_salja"

Fish:
  $ salja completion fish | source
  $ salja completion fish > ~/.config/fish/completions/salja.fish

PowerShell:
  PS> salja completion powershell | Out-String | Invoke-Expression
`,
ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
RunE: func(cmd *cobra.Command, args []string) error {
switch args[0] {
case "bash":
return cmd.Root().GenBashCompletion(os.Stdout)
case "zsh":
return cmd.Root().GenZshCompletion(os.Stdout)
case "fish":
return cmd.Root().GenFishCompletion(os.Stdout, true)
case "powershell":
return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
}
return nil
},
}
return cmd
}
