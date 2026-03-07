package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gongahkia/salja/internal/logging"
	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/registry"
)

// ValidateModel validates a file and shows results.
type ValidateModel struct {
	filePicker FilePickerModel
	filePath   string
	result     *validateResult
	err        error
	done       bool
}

type validateResult struct {
	format     string
	itemCount  int
	events     int
	tasks      int
	journals   int
	recurrence int
	subtasks   int
	errors     []string
}

// NewValidateModel creates a new validate view.
func NewValidateModel() ValidateModel {
	return ValidateModel{filePicker: NewFilePickerModel()}
}

func (v ValidateModel) Init() tea.Cmd {
	return v.filePicker.Init()
}

type validateDoneMsg struct {
	result *validateResult
	err    error
}

func (v ValidateModel) Update(msg tea.Msg) (ValidateModel, tea.Cmd) {
	switch msg := msg.(type) {
	case FilePickerMsg:
		v.filePath = msg.Path
		logging.Default().Info("interaction", fmt.Sprintf("validate: selected %s", msg.Path))
		return v, v.runValidate()
	case validateDoneMsg:
		v.done = true
		v.result = msg.result
		v.err = msg.err
		return v, nil
	}

	if !v.done {
		var cmd tea.Cmd
		v.filePicker, cmd = v.filePicker.Update(msg)
		return v, cmd
	}
	return v, nil
}

func (v ValidateModel) runValidate() tea.Cmd {
	filePath := v.filePath
	return func() tea.Msg {
		ctx := context.Background()
		allFmts := registry.AvailableFormats()
		for id, entry := range allFmts {
			if entry.NewParser == nil {
				continue
			}
			parser := entry.NewParser()
			col, err := parser.ParseFile(ctx, filePath)
			if err != nil {
				continue
			}
			events, tasks, journals, recurrence, subtasks := 0, 0, 0, 0, 0
			for _, item := range col.Items {
				switch item.ItemType {
				case model.ItemTypeEvent:
					events++
				case model.ItemTypeTask:
					tasks++
				case model.ItemTypeJournal:
					journals++
				}
				if item.Recurrence != nil {
					recurrence++
				}
				subtasks += len(item.Subtasks)
			}
			logging.Default().Info("interaction", fmt.Sprintf("validate: %s detected as %s, %d items", filePath, id, len(col.Items)))
			return validateDoneMsg{result: &validateResult{
				format:     id,
				itemCount:  len(col.Items),
				events:     events,
				tasks:      tasks,
				journals:   journals,
				recurrence: recurrence,
				subtasks:   subtasks,
			}}
		}
		return validateDoneMsg{err: fmt.Errorf("no parser could read %s", filePath)}
	}
}

func (v ValidateModel) View() string {
	header := SubtitleStyle.Render("Validate")

	if !v.done && v.filePath == "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, v.filePicker.View())
	}

	if v.err != nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, ErrorStyle.Render("  Error: "+v.err.Error()))
	}

	if v.result == nil {
		return lipgloss.JoinVertical(lipgloss.Left, header, MutedStyle.Render("  Validating..."))
	}

	r := v.result
	var b strings.Builder
	fmt.Fprintf(&b, "  File:       %s\n", v.filePath)
	fmt.Fprintf(&b, "  Format:     %s\n", r.format)
	fmt.Fprintf(&b, "  Items:      %d\n", r.itemCount)
	fmt.Fprintf(&b, "  Events:     %d\n", r.events)
	fmt.Fprintf(&b, "  Tasks:      %d\n", r.tasks)
	if r.journals > 0 {
		fmt.Fprintf(&b, "  Journals:   %d\n", r.journals)
	}
	fmt.Fprintf(&b, "  Recurrence: %d\n", r.recurrence)
	fmt.Fprintf(&b, "  Subtasks:   %d\n", r.subtasks)
	if len(r.errors) > 0 {
		fmt.Fprintf(&b, "\n  Errors: %d\n", len(r.errors))
		for _, e := range r.errors {
			b.WriteString(ErrorStyle.Render("    - "+e) + "\n")
		}
	} else {
		b.WriteString("\n  " + SuccessStyle.Render("✓ Valid") + "\n")
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, b.String())
}
