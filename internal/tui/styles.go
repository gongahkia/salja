package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	ColorPrimary   = lipgloss.Color("#7C3AED") // violet
	ColorSecondary = lipgloss.Color("#06B6D4") // cyan
	ColorSuccess   = lipgloss.Color("#22C55E") // green
	ColorWarning   = lipgloss.Color("#EAB308") // yellow
	ColorError     = lipgloss.Color("#EF4444") // red
	ColorMuted     = lipgloss.Color("#6B7280") // gray
	ColorText      = lipgloss.Color("#F9FAFB") // near-white
	ColorBg        = lipgloss.Color("#1F2937") // dark gray
)

// Shared styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			PaddingLeft(1).
			PaddingRight(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			PaddingLeft(1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingLeft(1).
			PaddingTop(1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true).
			PaddingLeft(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true).
			PaddingLeft(1)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			PaddingLeft(1)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingLeft(1).
			PaddingTop(1)
)
