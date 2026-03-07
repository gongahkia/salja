package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gongahkia/salja/internal/logging"
)

type ErrorEntry struct {
	Time    time.Time
	Level   string
	Message string
}

type ErrorPanel struct {
	errors  []ErrorEntry
	visible bool
	scroll  int
}

func NewErrorPanel() ErrorPanel {
	return ErrorPanel{}
}

func (e *ErrorPanel) Push(level, msg string) {
	e.errors = append(e.errors, ErrorEntry{
		Time: time.Now(), Level: level, Message: msg,
	})
	e.visible = true
	if len(e.errors) > 100 {
		e.errors = e.errors[len(e.errors)-100:]
	}
	logging.Default().Log(level, "error", msg)
}

func (e *ErrorPanel) Dismiss() {
	e.visible = false
}

func (e ErrorPanel) Update(msg tea.Msg) (ErrorPanel, tea.Cmd) {
	if !e.visible {
		return e, nil
	}
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(kmsg, key.NewBinding(key.WithKeys("esc"))) {
			e.visible = false
		}
		if key.Matches(kmsg, key.NewBinding(key.WithKeys("up", "k"))) && e.scroll > 0 {
			e.scroll--
		}
		maxScroll := len(e.errors) - 10
		if maxScroll < 0 {
			maxScroll = 0
		}
		if key.Matches(kmsg, key.NewBinding(key.WithKeys("down", "j"))) && e.scroll < maxScroll {
			e.scroll++
		}
	}
	return e, nil
}

func (e ErrorPanel) View() string {
	if !e.visible || len(e.errors) == 0 {
		return ""
	}
	header := ErrorStyle.Render("--- Errors (esc to dismiss) ---")
	maxVisible := 10
	start := e.scroll
	end := start + maxVisible
	if end > len(e.errors) {
		end = len(e.errors)
	}
	var lines string
	for _, entry := range e.errors[start:end] {
		ts := entry.Time.Format("15:04:05")
		style := ErrorStyle
		if entry.Level == "warn" {
			style = WarningStyle
		}
		lines += style.Render(fmt.Sprintf("[%s] %s: %s", ts, entry.Level, entry.Message)) + "\n"
	}
	if len(e.errors) > maxVisible {
		lines += MutedStyle.Render(fmt.Sprintf("  showing %d-%d of %d", start+1, end, len(e.errors)))
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, lines)
}
