package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/gongahkia/salja/cmd/salja/commands"
	"github.com/gongahkia/salja/internal/config"
	salerr "github.com/gongahkia/salja/internal/errors"
	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "salja",
		Short:   "Universal calendar and task converter",
		Long:    "Convert between calendar and task management formats with conflict resolution",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildDate),
	}

	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("config", "", "Config file path")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cfgPath, _ := cmd.Flags().GetString("config"); cfgPath != "" {
			config.SetOverridePath(cfgPath)
		}
		return nil
	}

	rootCmd.AddCommand(commands.NewConvertCmd())
	rootCmd.AddCommand(commands.NewListFormatsCmd())
	rootCmd.AddCommand(commands.NewValidateCmd())
	rootCmd.AddCommand(commands.NewDiffCmd())
	rootCmd.AddCommand(commands.NewConfigCmd())
	rootCmd.AddCommand(commands.NewAuthCmd())
	rootCmd.AddCommand(commands.NewSyncCmd())
	rootCmd.AddCommand(commands.NewCompletionCmd())
	rootCmd.AddCommand(commands.NewTUICmd())
	rootCmd.AddCommand(newVersionCmd())

	// Launch TUI when invoked with no subcommand
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return commands.RunTUI()
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var parseErr *salerr.ParseError
		var validErr *salerr.ValidationError
		var conflictErr *salerr.ConflictError
		switch {
		case errors.As(err, &parseErr):
			os.Exit(1)
		case errors.As(err, &validErr):
			os.Exit(2)
		case errors.As(err, &conflictErr):
			os.Exit(3)
		default:
			os.Exit(1)
		}
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, build info, and platform details",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("salja %s\n", version)
			fmt.Printf("  commit:    %s\n", commit)
			fmt.Printf("  built:     %s\n", buildDate)
			fmt.Printf("  go:        %s\n", runtime.Version())
			fmt.Printf("  os/arch:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
