package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gongahkia/salja/internal/config"
	"github.com/gongahkia/salja/internal/logging"
)

type configSetting struct {
	label   string
	value   string
	options []string // if non-empty, supports cycling
}

// ConfigModel displays and edits the current config.
type ConfigModel struct {
	settings []configSetting
	cursor   int
	cfg      *config.Config
	path     string
	err      error
	message  string
}

// NewConfigModel loads and displays the current config.
func NewConfigModel() ConfigModel {
	cfg, err := config.Load()
	if err != nil {
		return ConfigModel{err: err}
	}
	path := config.ConfigPath()
	return ConfigModel{
		cfg:      cfg,
		path:     path,
		settings: buildSettings(cfg),
	}
}

func buildSettings(cfg *config.Config) []configSetting {
	return []configSetting{
		{"preferred_mode", cfg.PreferredMode, []string{"file", "api"}},
		{"default_timezone", cfg.DefaultTimezone, nil},
		{"conflict_strategy", cfg.ConflictStrategy, []string{"ask", "prefer-source", "prefer-target", "skip", "fail"}},
		{"data_loss_mode", cfg.DataLossMode, []string{"warn", "error", "silent"}},
		{"streaming_threshold_mb", fmt.Sprintf("%d", cfg.StreamingThresholdMB), nil},
		{"api_timeout_seconds", fmt.Sprintf("%d", cfg.APITimeoutSeconds), nil},
		{"google client_id", maskSecret(cfg.API.Google.ClientID), nil},
		{"microsoft client_id", maskSecret(cfg.API.Microsoft.ClientID), nil},
		{"todoist client_id", maskSecret(cfg.API.Todoist.ClientID), nil},
		{"ticktick client_id", maskSecret(cfg.API.TickTick.ClientID), nil},
		{"notion client_id", maskSecret(cfg.API.Notion.ClientID), nil},
	}
}

func maskSecret(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "***"
}

func (c ConfigModel) Init() tea.Cmd { return nil }

type configSavedMsg struct{ err error }

func (c ConfigModel) Update(msg tea.Msg) (ConfigModel, tea.Cmd) {
	switch msg := msg.(type) {
	case configSavedMsg:
		if msg.err != nil {
			c.message = fmt.Sprintf("Save error: %v", msg.err)
		} else {
			c.message = "Saved"
		}
		return c, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if c.cursor < len(c.settings)-1 {
				c.cursor++
			}
		case "k", "up":
			if c.cursor > 0 {
				c.cursor--
			}
		case "enter", "tab", " ":
			s := c.settings[c.cursor]
			if len(s.options) > 0 {
				idx := 0
				for i, o := range s.options {
					if o == s.value {
						idx = i
						break
					}
				}
				next := (idx + 1) % len(s.options)
				c.settings[c.cursor].value = s.options[next]
				c.applyToConfig()
				c.message = fmt.Sprintf("%s → %s", s.label, s.options[next])
				logging.Default().Info("interaction", fmt.Sprintf("config: %s = %s", s.label, s.options[next]))
			}
		case "s":
			cfg := c.cfg
			path := c.path
			return c, func() tea.Msg {
				return configSavedMsg{err: config.Save(cfg, path)}
			}
		}
	}
	return c, nil
}

func (c *ConfigModel) applyToConfig() {
	for _, s := range c.settings {
		switch s.label {
		case "preferred_mode":
			c.cfg.PreferredMode = s.value
		case "conflict_strategy":
			c.cfg.ConflictStrategy = s.value
		case "data_loss_mode":
			c.cfg.DataLossMode = s.value
		}
	}
}

func (c ConfigModel) View() string {
	header := SubtitleStyle.Render("Configuration")

	if c.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, ErrorStyle.Render("  "+c.err.Error()))
	}

	var lines []string
	for i, s := range c.settings {
		cursor := "  "
		style := lipgloss.NewStyle()
		if i == c.cursor {
			cursor = "▸ "
			style = SelectedStyle
		}
		val := s.value
		cyclable := ""
		if len(s.options) > 0 {
			cyclable = MutedStyle.Render(" [Enter to cycle]")
		}
		lines = append(lines, fmt.Sprintf("%s%-24s %s%s", cursor, style.Render(s.label), val, cyclable))
	}
	body := strings.Join(lines, "\n")

	var msg string
	if c.message != "" {
		msg = "\n" + WarningStyle.Render("  "+c.message)
	}

	help := HelpStyle.Render("↑↓ navigate · Enter cycle value · s save")
	path := MutedStyle.Render("  " + c.path)
	return lipgloss.JoinVertical(lipgloss.Left, header, "", body, msg, "", help, path)
}
