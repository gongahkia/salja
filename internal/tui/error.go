package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorDisplay renders salja errors with styled formatting.
type ErrorDisplay struct {
	title   string
	message string
	code    int
}

// NewErrorDisplay creates an error display component.
func NewErrorDisplay(title, message string, code int) ErrorDisplay {
	return ErrorDisplay{title: title, message: message, code: code}
}

func (e ErrorDisplay) Init() tea.Cmd { return nil }

func (e ErrorDisplay) Update(msg tea.Msg) (ErrorDisplay, tea.Cmd) {
	return e, nil
}

func (e ErrorDisplay) View() string {
	header := ErrorStyle.Render("✗ " + e.title)
	body := fmt.Sprintf("  %s\n  Exit code: %d", e.message, e.code)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body)
}
