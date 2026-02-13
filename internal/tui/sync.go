package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SyncModel displays cloud sync dashboard.
type SyncModel struct {
	services []syncService
	cursor   int
	keys     KeyMap
	message  string
}

type syncService struct {
	name          string
	authenticated bool
}

var defaultServices = []syncService{
	{"google", false},
	{"microsoft", false},
	{"todoist", false},
	{"ticktick", false},
	{"notion", false},
}

// NewSyncModel creates a sync dashboard.
func NewSyncModel() SyncModel {
	services := make([]syncService, len(defaultServices))
	copy(services, defaultServices)
	return SyncModel{services: services, keys: DefaultKeyMap()}
}

func (s SyncModel) Init() tea.Cmd { return nil }

func (s SyncModel) Update(msg tea.Msg) (SyncModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if s.cursor < len(s.services)-1 {
				s.cursor++
			}
		case "k", "up":
			if s.cursor > 0 {
				s.cursor--
			}
		case "p":
			s.message = fmt.Sprintf("Push for %s not yet wired up", s.services[s.cursor].name)
		case "l":
			s.message = fmt.Sprintf("Pull for %s not yet wired up", s.services[s.cursor].name)
		}
	}
	return s, nil
}

func (s SyncModel) View() string {
	header := SubtitleStyle.Render("Cloud Sync")

	var rows string
	for i, svc := range s.services {
		cursor := "  "
		if i == s.cursor {
			cursor = "▸ "
		}
		status := ErrorStyle.Render("✗")
		if svc.authenticated {
			status = SuccessStyle.Render("✓")
		}
		rows += fmt.Sprintf("%s%s %s\n", cursor, status, svc.name)
	}

	help := HelpStyle.Render("p push · l pull · ↑↓ navigate")

	var msg string
	if s.message != "" {
		msg = "\n" + WarningStyle.Render("  "+s.message)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, "", rows, msg, help)
}
