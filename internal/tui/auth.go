package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AuthModel manages OAuth login/logout/status per service.
type AuthModel struct {
	services []authService
	cursor   int
	keys     KeyMap
	message  string
}

type authService struct {
	name   string
	authed bool
	expiry string
}

// NewAuthModel creates a new auth view.
func NewAuthModel() AuthModel {
	services := []authService{
		{"google", false, ""},
		{"microsoft", false, ""},
		{"todoist", false, ""},
		{"ticktick", false, ""},
		{"notion", false, ""},
	}
	return AuthModel{services: services, keys: DefaultKeyMap()}
}

func (a AuthModel) Init() tea.Cmd { return nil }

func (a AuthModel) Update(msg tea.Msg) (AuthModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if a.cursor < len(a.services)-1 {
				a.cursor++
			}
		case "k", "up":
			if a.cursor > 0 {
				a.cursor--
			}
		case "enter":
			svc := a.services[a.cursor]
			if svc.authed {
				a.message = fmt.Sprintf("Logout for %s — not yet wired", svc.name)
			} else {
				a.message = fmt.Sprintf("Login for %s — not yet wired", svc.name)
			}
		}
	}
	return a, nil
}

func (a AuthModel) View() string {
	header := SubtitleStyle.Render("Authentication")

	var rows string
	for i, svc := range a.services {
		cursor := "  "
		if i == a.cursor {
			cursor = "▸ "
		}
		status := ErrorStyle.Render("not authenticated")
		if svc.authed {
			status = SuccessStyle.Render("authenticated")
			if svc.expiry != "" {
				status += MutedStyle.Render(" (expires: " + svc.expiry + ")")
			}
		}
		rows += fmt.Sprintf("%s%-12s %s\n", cursor, svc.name, status)
	}

	help := HelpStyle.Render("enter login/logout · ↑↓ navigate")
	var msg string
	if a.message != "" {
		msg = "\n" + WarningStyle.Render("  "+a.message)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", rows, msg, help)
}
