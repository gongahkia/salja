package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gongahkia/salja/internal/api"
	"github.com/gongahkia/salja/internal/logging"
)

// AuthModel manages OAuth login/logout/status per service.
type AuthModel struct {
	services []authService
	cursor   int
	keys     KeyMap
	message  string
	loaded   bool
}

type authService struct {
	name   string
	authed bool
	expiry string
	status string // "authenticated", "expired", "not authenticated"
}

var serviceNames = []string{"google", "microsoft", "todoist", "ticktick", "notion"}

type authStatusMsg struct {
	services []authService
}

type authActionMsg struct {
	message string
}

// NewAuthModel creates a new auth view.
func NewAuthModel() AuthModel {
	services := make([]authService, len(serviceNames))
	for i, name := range serviceNames {
		services[i] = authService{name: name, status: "loading..."}
	}
	return AuthModel{services: services, keys: DefaultKeyMap()}
}

func (a AuthModel) Init() tea.Cmd {
	return loadAuthStatus
}

func loadAuthStatus() tea.Msg {
	store, err := api.DefaultTokenStore()
	if err != nil {
		return authStatusMsg{} // fallback to unknown
	}
	tokens, err := store.Load()
	if err != nil {
		return authStatusMsg{}
	}
	services := make([]authService, len(serviceNames))
	for i, name := range serviceNames {
		svc := authService{name: name, status: "not authenticated"}
		tok, ok := tokens[name]
		if ok && tok != nil {
			if tok.IsExpired() {
				svc.status = "expired"
				if tok.RefreshToken != "" {
					svc.status = "expired (has refresh token)"
				}
			} else {
				svc.authed = true
				svc.status = "authenticated"
				svc.expiry = tok.ExpiresAt.Format(time.RFC3339)
			}
		}
		services[i] = svc
	}
	return authStatusMsg{services: services}
}

func (a AuthModel) Update(msg tea.Msg) (AuthModel, tea.Cmd) {
	switch msg := msg.(type) {
	case authStatusMsg:
		if len(msg.services) > 0 {
			a.services = msg.services
		}
		a.loaded = true
		return a, nil
	case authActionMsg:
		a.message = msg.message
		return a, loadAuthStatus // refresh after action
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
				logging.Default().Info("interaction", fmt.Sprintf("auth: logout %s", svc.name))
				return a, logoutService(svc.name)
			}
			logging.Default().Info("interaction", fmt.Sprintf("auth: login %s requested", svc.name))
			a.message = fmt.Sprintf("Login for %s — use CLI: salja auth login %s", svc.name, svc.name)
		case "d":
			svc := a.services[a.cursor]
			if svc.authed {
				logging.Default().Info("interaction", fmt.Sprintf("auth: logout %s", svc.name))
				return a, logoutService(svc.name)
			}
		}
	}
	return a, nil
}

func logoutService(name string) tea.Cmd {
	return func() tea.Msg {
		store, err := api.DefaultSecureStore()
		if err != nil {
			return authActionMsg{message: fmt.Sprintf("Error: %v", err)}
		}
		if err := store.Delete(name); err != nil {
			return authActionMsg{message: fmt.Sprintf("Failed to logout %s: %v", name, err)}
		}
		logging.Default().Info("interaction", fmt.Sprintf("auth: logged out %s", name))
		return authActionMsg{message: fmt.Sprintf("Logged out of %s", name)}
	}
}

func (a AuthModel) View() string {
	header := SubtitleStyle.Render("Authentication")
	var rows string
	for i, svc := range a.services {
		cursor := "  "
		if i == a.cursor {
			cursor = "▸ "
		}
		var status string
		switch {
		case svc.authed:
			status = SuccessStyle.Render("authenticated")
			if svc.expiry != "" {
				status += MutedStyle.Render(" (expires: " + svc.expiry + ")")
			}
		case svc.status == "expired" || svc.status == "expired (has refresh token)":
			status = WarningStyle.Render(svc.status)
		default:
			status = MutedStyle.Render(svc.status)
		}
		rows += fmt.Sprintf("%s%-12s %s\n", cursor, svc.name, status)
	}
	help := HelpStyle.Render("enter login/logout · d delete token · ↑↓ navigate")
	var msg string
	if a.message != "" {
		msg = "\n" + WarningStyle.Render("  "+a.message)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, "", rows, msg, help)
}
