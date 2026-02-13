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

type diffStep int

const (
	diffPickA diffStep = iota
	diffPickB
	diffRunning
	diffDone
)

// DiffModel compares two files side-by-side.
type DiffModel struct {
	step       diffStep
	filePicker FilePickerModel
	pathA      string
	pathB      string
	result     string
	err        error
	scrollY    int
}

// NewDiffModel creates a new diff view.
func NewDiffModel() DiffModel {
	return DiffModel{
		step:       diffPickA,
		filePicker: NewFilePickerModel(),
	}
}

func (d DiffModel) Init() tea.Cmd {
	return d.filePicker.Init()
}

type diffDoneMsg struct {
	result string
	err    error
}

func (d DiffModel) Update(msg tea.Msg) (DiffModel, tea.Cmd) {
	switch msg := msg.(type) {
	case FilePickerMsg:
		switch d.step {
		case diffPickA:
			d.pathA = msg.Path
			d.step = diffPickB
			d.filePicker = NewFilePickerModel()
			return d, d.filePicker.Init()
		case diffPickB:
			d.pathB = msg.Path
			d.step = diffRunning
			return d, d.runDiff()
		}
	case diffDoneMsg:
		d.step = diffDone
		d.result = msg.result
		d.err = msg.err
		return d, nil
	case tea.KeyMsg:
		if d.step == diffDone {
			switch msg.String() {
			case "j", "down":
				d.scrollY++
			case "k", "up":
				if d.scrollY > 0 {
					d.scrollY--
				}
			}
			return d, nil
		}
	}

	if d.step == diffPickA || d.step == diffPickB {
		var cmd tea.Cmd
		d.filePicker, cmd = d.filePicker.Update(msg)
		return d, cmd
	}
	return d, nil
}

func (d DiffModel) runDiff() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		allFmts := registry.AllFormats()
		var colA, colB *struct{ events, tasks int }

		for _, entry := range allFmts {
			if entry.NewParser == nil {
				continue
			}
			parser := entry.NewParser()
			a, errA := parser.ParseFile(ctx, d.pathA)
			b, errB := parser.ParseFile(ctx, d.pathB)
			if errA == nil && errB == nil {
				ae, at, be, bt := 0, 0, 0, 0
				for _, item := range a.Items {
					if item.ItemType == model.ItemTypeEvent {
						ae++
					} else {
						at++
					}
				}
				for _, item := range b.Items {
					if item.ItemType == model.ItemTypeEvent {
						be++
					} else {
						bt++
					}
				}
				colA = &struct{ events, tasks int }{ae, at}
				colB = &struct{ events, tasks int }{be, bt}
				break
			}
		}

		if colA == nil || colB == nil {
			return diffDoneMsg{err: fmt.Errorf("could not parse both files")}
		}

		var b strings.Builder
		addStyle := lipgloss.NewStyle().Foreground(ColorSuccess)
		delStyle := lipgloss.NewStyle().Foreground(ColorError)

		fmt.Fprintf(&b, "  File A: %s\n", d.pathA)
		fmt.Fprintf(&b, "  File B: %s\n\n", d.pathB)

		eDiff := colB.events - colA.events
		tDiff := colB.tasks - colA.tasks

		fmt.Fprintf(&b, "  Events: %d → %d", colA.events, colB.events)
		if eDiff > 0 {
			b.WriteString("  " + addStyle.Render(fmt.Sprintf("+%d", eDiff)))
		} else if eDiff < 0 {
			b.WriteString("  " + delStyle.Render(fmt.Sprintf("%d", eDiff)))
		}
		b.WriteString("\n")

		fmt.Fprintf(&b, "  Tasks:  %d → %d", colA.tasks, colB.tasks)
		if tDiff > 0 {
			b.WriteString("  " + addStyle.Render(fmt.Sprintf("+%d", tDiff)))
		} else if tDiff < 0 {
			b.WriteString("  " + delStyle.Render(fmt.Sprintf("%d", tDiff)))
		}
		b.WriteString("\n")

		return diffDoneMsg{result: b.String()}
	}
}

func (d DiffModel) View() string {
	header := SubtitleStyle.Render("Diff")
	switch d.step {
	case diffPickA:
		return lipgloss.JoinVertical(lipgloss.Left, header, SubtitleStyle.Render("Select file A:"), d.filePicker.View())
	case diffPickB:
		return lipgloss.JoinVertical(lipgloss.Left, header, SubtitleStyle.Render("Select file B:"), d.filePicker.View())
	case diffRunning:
		return lipgloss.JoinVertical(lipgloss.Left, header, MutedStyle.Render("  Comparing..."))
	case diffDone:
		if d.err != nil {
			return lipgloss.JoinVertical(lipgloss.Left, header, ErrorStyle.Render("  Error: "+d.err.Error()))
		}
		lines := strings.Split(d.result, "\n")
		end := d.scrollY + 20
		if end > len(lines) {
			end = len(lines)
		}
		start := d.scrollY
		if start >= len(lines) {
			start = len(lines) - 1
		}
		visible := strings.Join(lines[start:end], "\n")
		return lipgloss.JoinVertical(lipgloss.Left, header, visible)
	}
	return header
}
