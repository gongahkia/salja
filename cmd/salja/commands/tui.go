package commands

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/gongahkia/salja/internal/tui"
)

// NewTUICmd creates the 'tui' subcommand.
func NewTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive TUI",
		Long:  "Launch the full-screen interactive terminal user interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunTUI()
		},
	}
}

// RunTUI starts the bubbletea program with alt-screen.
func RunTUI() error {
	p := tea.NewProgram(tui.NewApp(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
