package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type conflictStrategy int
var strategyLabels = []string{"prefer-source", "prefer-target", "skip", "merge-manual"}

// ConflictModel presents interactive conflict resolution UI.
type ConflictModel struct {
	keys     KeyMap
	items    []conflictPair
	cursor   int
	strategy conflictStrategy
}

type conflictPair struct {
	sourceTitle string
	targetTitle string
	resolved    bool
	chosen      string
}

// NewConflictModel creates a conflict resolution view.
func NewConflictModel(pairs []conflictPair) ConflictModel {
	return ConflictModel{
		keys:  DefaultKeyMap(),
		items: pairs,
	}
}

func (c ConflictModel) Init() tea.Cmd { return nil }

func (c ConflictModel) Update(msg tea.Msg) (ConflictModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, c.keys.Up):
			if c.cursor > 0 {
				c.cursor--
			}
		case key.Matches(msg, c.keys.Down):
			if c.cursor < len(c.items)-1 {
				c.cursor++
			}
		case key.Matches(msg, c.keys.Left):
			if c.strategy > 0 {
				c.strategy--
			}
		case key.Matches(msg, c.keys.Right):
			if c.strategy < conflictStrategy(len(strategyLabels)-1) {
				c.strategy++
			}
		case key.Matches(msg, c.keys.Enter):
			if c.cursor < len(c.items) {
				c.items[c.cursor].resolved = true
				c.items[c.cursor].chosen = strategyLabels[c.strategy]
			}
		}
	}
	return c, nil
}

func (c ConflictModel) View() string {
	header := SubtitleStyle.Render("Conflict Resolution")

	stratRow := "  Strategy: "
	for i, label := range strategyLabels {
		if conflictStrategy(i) == c.strategy {
			stratRow += SelectedStyle.Render("[" + label + "]")
		} else {
			stratRow += MutedStyle.Render(" " + label + " ")
		}
		stratRow += "  "
	}

	var rows string
	for i, pair := range c.items {
		cursor := "  "
		if i == c.cursor {
			cursor = "▸ "
		}
		status := "○"
		if pair.resolved {
			status = "●"
		}
		rows += fmt.Sprintf("%s%s %s  ←→  %s", cursor, status, pair.sourceTitle, pair.targetTitle)
		if pair.resolved {
			rows += MutedStyle.Render("  [" + pair.chosen + "]")
		}
		rows += "\n"
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, stratRow, "", rows)
}
