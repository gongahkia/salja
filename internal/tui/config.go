package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gongahkia/salja/internal/config"
)

// ConfigModel displays the current TOML configuration read-only.
type ConfigModel struct {
	content string
	path    string
	err     error
}

// NewConfigModel loads and displays the current config.
func NewConfigModel() ConfigModel {
	cfg, err := config.Load()
	if err != nil {
		return ConfigModel{err: err}
	}
	path := config.ConfigPath()

	var b strings.Builder
	fmt.Fprintf(&b, "  Config path: %s\n\n", path)
	fmt.Fprintf(&b, "  [general]\n")
	fmt.Fprintf(&b, "    preferred_mode         = %q\n", cfg.PreferredMode)
	fmt.Fprintf(&b, "    default_timezone        = %q\n", cfg.DefaultTimezone)
	fmt.Fprintf(&b, "    conflict_strategy       = %q\n", cfg.ConflictStrategy)
	fmt.Fprintf(&b, "    data_loss_mode          = %q\n", cfg.DataLossMode)
	fmt.Fprintf(&b, "    streaming_threshold_mb  = %d\n", cfg.StreamingThresholdMB)
	fmt.Fprintf(&b, "    api_timeout_seconds     = %d\n", cfg.APITimeoutSeconds)

	return ConfigModel{content: b.String(), path: path}
}

func (c ConfigModel) Init() tea.Cmd { return nil }

func (c ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "e" && c.path != "" {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			cmd := exec.Command(editor, c.path)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return c, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return nil
			})
		}
	}
	return c, nil
}

func (c ConfigModel) View() string {
	header := SubtitleStyle.Render("Configuration")

	if c.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, ErrorStyle.Render("  "+c.err.Error()))
	}

	help := HelpStyle.Render("e open in $EDITOR")
	return lipgloss.JoinVertical(lipgloss.Left, header, "", c.content, help)
}
