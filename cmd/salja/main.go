package main

import (
	"fmt"
	"os"

	"github.com/gongahkia/salja/cmd/salja/commands"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "salja",
		Short: "Universal calendar and task converter",
		Long:  "Convert between calendar and task management formats with conflict resolution",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
	}

	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("config", "", "Config file path")

	rootCmd.AddCommand(commands.NewConvertCmd())
	rootCmd.AddCommand(commands.NewListFormatsCmd())
	rootCmd.AddCommand(commands.NewValidateCmd())
	rootCmd.AddCommand(commands.NewDiffCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewAuthCmd())
	rootCmd.AddCommand(commands.NewSyncCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
