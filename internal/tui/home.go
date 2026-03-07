package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gongahkia/salja/internal/appdetect"
	"github.com/gongahkia/salja/internal/platform"
)

type suggestedAction struct {
	label string
	key   string
	view  View
}

// HomeModel is the main menu view with platform info and detected apps.
type HomeModel struct {
	keys       KeyMap
	platInfo   platform.PlatformInfo
	apps       []appdetect.DetectedApp
	appsLoaded bool
	actions    []suggestedAction
	cursor     int
	width      int
	height     int
}

// NewHomeModel creates a new home menu.
func NewHomeModel() HomeModel {
	return HomeModel{
		keys:     DefaultKeyMap(),
		platInfo: platform.Summary(),
		actions:  defaultActions(),
	}
}

func defaultActions() []suggestedAction {
	return []suggestedAction{
		{"Convert a file", "c", ViewConvert},
		{"Validate a file", "v", ViewValidate},
		{"Diff two files", "d", ViewDiff},
		{"Cloud sync", "s", ViewSync},
		{"Manage auth", "a", ViewAuth},
		{"Settings", "o", ViewConfig},
	}
}

type appsDetectedMsg struct {
	apps []appdetect.DetectedApp
}

func detectAppsCmd() tea.Msg {
	return appsDetectedMsg{apps: appdetect.DetectAll()}
}

func (h HomeModel) Init() tea.Cmd {
	return detectAppsCmd
}

func (h HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
		return h, nil
	case appsDetectedMsg:
		h.apps = msg.apps
		h.appsLoaded = true
		return h, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, h.keys.Enter):
			if h.cursor < len(h.actions) {
				target := h.actions[h.cursor].view
				return h, func() tea.Msg { return NavigateMsg{Target: target} }
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if h.cursor < len(h.actions)-1 {
				h.cursor++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if h.cursor > 0 {
				h.cursor--
			}
		default:
			for i, a := range h.actions {
				if key.Matches(msg, key.NewBinding(key.WithKeys(a.key))) {
					h.cursor = i
					return h, func() tea.Msg { return NavigateMsg{Target: a.view} }
				}
			}
		}
	}
	return h, nil
}

func (h HomeModel) View() string {
	var sections []string

	// header with platform badge
	title := TitleStyle.Render("salja")
	badge := MutedStyle.Render(fmt.Sprintf("%s %s", h.platInfo.OS, h.platInfo.Arch))
	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Center, title, " ", badge))
	sections = append(sections, MutedStyle.Render("  Universal Calendar & Task Converter"))
	sections = append(sections, "")

	// detected apps
	if !h.appsLoaded {
		sections = append(sections, MutedStyle.Render("  Detecting installed apps..."))
	} else if len(h.apps) > 0 {
		sections = append(sections, SubtitleStyle.Render("Detected Apps"))
		for _, app := range h.apps {
			indicator := lipgloss.NewStyle().Foreground(ColorSuccess).Render("●")
			if !app.Installed {
				indicator = lipgloss.NewStyle().Foreground(ColorMuted).Render("○")
			}
			line := fmt.Sprintf("  %s %s", indicator, app.Name)
			if len(app.DataPaths) > 0 {
				line += MutedStyle.Render(fmt.Sprintf(" (%d data paths)", len(app.DataPaths)))
			}
			sections = append(sections, line)
		}
	}
	sections = append(sections, "")

	// suggested actions
	sections = append(sections, SubtitleStyle.Render("Actions"))
	for i, a := range h.actions {
		prefix := "  "
		style := lipgloss.NewStyle()
		if i == h.cursor {
			prefix = "▸ "
			style = SelectedStyle
		}
		keyHint := MutedStyle.Render(fmt.Sprintf("[%s]", a.key))
		sections = append(sections, fmt.Sprintf("%s%s %s", prefix, style.Render(a.label), keyHint))
	}

	return strings.Join(sections, "\n")
}
