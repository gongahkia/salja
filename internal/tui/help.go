package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpModel displays keybinding reference as a modal overlay.
type HelpModel struct {
	keys KeyMap
}

// NewHelpModel creates a help overlay.
func NewHelpModel() HelpModel {
	return HelpModel{keys: DefaultKeyMap()}
}

func (h HelpModel) Init() tea.Cmd { return nil }

func (h HelpModel) Update(_ tea.Msg) (HelpModel, tea.Cmd) {
	return h, nil
}

func (h HelpModel) View() string {
	bindings := []struct{ key, desc string }{
		{"q / Ctrl+C", "Quit"},
		{"?", "Toggle help"},
		{"Esc", "Back / close"},
		{"Tab", "Cycle focus"},
		{"Enter", "Select / confirm"},
		{"↑ / k", "Move up"},
		{"↓ / j", "Move down"},
		{"← / h", "Move left"},
		{"→ / l", "Move right"},
		{"e", "Open $EDITOR (config view)"},
		{"p", "Push (sync view)"},
		{"l", "Pull (sync view)"},
	}

	var b strings.Builder
	for _, bind := range bindings {
		b.WriteString(fmt.Sprintf("  %-14s %s\n", bind.key, bind.desc))
	}

	content := BorderStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("Keybindings"),
			"",
			b.String(),
		),
	)

	return lipgloss.Place(80, 24, lipgloss.Center, lipgloss.Center, content)
}
