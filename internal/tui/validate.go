package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	format    string
	itemCount int
	events    int
	tasks     int
	errors    []string
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
	return func() tea.Msg {
		ctx := context.Background()
		// Try all formats
		allFmts := registry.AllFormats()
		for id, entry := range allFmts {
			if entry.NewParser == nil {
				continue
			}
			parser := entry.NewParser()
			col, err := parser.ParseFile(ctx, v.filePath)
			if err != nil {
				continue
			}
			events, tasks := 0, 0
			for _, item := range col.Items {
				switch item.ItemType {
				case model.ItemTypeEvent:
					events++
				case model.ItemTypeTask:
					tasks++
				}
			}
			return validateDoneMsg{result: &validateResult{
				format:    id,
				itemCount: len(col.Items),
				events:    events,
				tasks:     tasks,
			}}
		}
		return validateDoneMsg{err: fmt.Errorf("no parser could read %s", v.filePath)}
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
	fmt.Fprintf(&b, "  File:    %s\n", v.filePath)
	fmt.Fprintf(&b, "  Format:  %s\n", r.format)
	fmt.Fprintf(&b, "  Items:   %d (events: %d, tasks: %d)\n", r.itemCount, r.events, r.tasks)
	if len(r.errors) > 0 {
		fmt.Fprintf(&b, "  Errors:  %d\n", len(r.errors))
		for _, e := range r.errors {
			fmt.Fprintf(&b, "    - %s\n", e)
		}
	} else {
		b.WriteString("  Status:  ✓ valid\n")
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, SuccessStyle.Render(b.String()))
}
