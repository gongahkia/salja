package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gongahkia/salja/internal/conflict"
	"github.com/gongahkia/salja/internal/logging"
	"github.com/gongahkia/salja/internal/registry"
)

type diffStep int

const (
	diffPickA diffStep = iota
	diffPickB
	diffRunning
	diffDone
)

// DiffModel compares two files with item-level matching.
type DiffModel struct {
	step       diffStep
	filePicker FilePickerModel
	pathA      string
	pathB      string
	result     *diffResult
	err        error
	scrollY    int
}

type diffResult struct {
	countA   int
	countB   int
	matched  int
	added    []string
	removed  []string
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
	result *diffResult
	err    error
}

func (d DiffModel) Update(msg tea.Msg) (DiffModel, tea.Cmd) {
	switch msg := msg.(type) {
	case FilePickerMsg:
		switch d.step {
		case diffPickA:
			d.pathA = msg.Path
			logging.Default().Info("interaction", fmt.Sprintf("diff: selected file A %s", msg.Path))
			d.step = diffPickB
			d.filePicker = NewFilePickerModel()
			return d, d.filePicker.Init()
		case diffPickB:
			d.pathB = msg.Path
			logging.Default().Info("interaction", fmt.Sprintf("diff: selected file B %s", msg.Path))
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
	pathA := d.pathA
	pathB := d.pathB
	return func() tea.Msg {
		ctx := context.Background()
		allFmts := registry.AvailableFormats()
		for _, entry := range allFmts {
			if entry.NewParser == nil {
				continue
			}
			parser := entry.NewParser()
			colA, errA := parser.ParseFile(ctx, pathA)
			colB, errB := parser.ParseFile(ctx, pathB)
			if errA != nil || errB != nil {
				continue
			}
			detector := conflict.NewDetector()
			matches := detector.FindDuplicates(colA, colB)
			matchedA := make(map[int]bool)
			matchedB := make(map[int]bool)
			for _, m := range matches {
				matchedA[m.SourceIndex] = true
				matchedB[m.TargetIndex] = true
			}
			var added, removed []string
			for j, item := range colB.Items {
				if !matchedB[j] {
					added = append(added, item.Title)
				}
			}
			for i, item := range colA.Items {
				if !matchedA[i] {
					removed = append(removed, item.Title)
				}
			}
			logging.Default().Info("interaction", fmt.Sprintf("diff: %d matched, %d added, %d removed", len(matches), len(added), len(removed)))
			return diffDoneMsg{result: &diffResult{
				countA:  len(colA.Items),
				countB:  len(colB.Items),
				matched: len(matches),
				added:   added,
				removed: removed,
			}}
		}
		return diffDoneMsg{err: fmt.Errorf("could not parse both files")}
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
		return lipgloss.JoinVertical(lipgloss.Left, header, d.renderResult())
	}
	return header
}

func (d DiffModel) renderResult() string {
	r := d.result
	var lines []string
	lines = append(lines, fmt.Sprintf("  File A: %s (%d items)", d.pathA, r.countA))
	lines = append(lines, fmt.Sprintf("  File B: %s (%d items)", d.pathB, r.countB))
	lines = append(lines, fmt.Sprintf("  Matching: %d", r.matched))
	lines = append(lines, "")
	if len(r.removed) > 0 {
		lines = append(lines, ErrorStyle.Render(fmt.Sprintf("  Removed (%d):", len(r.removed))))
		for _, t := range r.removed {
			lines = append(lines, ErrorStyle.Render("    - "+t))
		}
	}
	if len(r.added) > 0 {
		lines = append(lines, SuccessStyle.Render(fmt.Sprintf("  Added (%d):", len(r.added))))
		for _, t := range r.added {
			lines = append(lines, SuccessStyle.Render("    + "+t))
		}
	}
	if len(r.added) == 0 && len(r.removed) == 0 {
		lines = append(lines, SuccessStyle.Render("  Files are identical"))
	}
	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("  ↑↓ scroll"))

	// apply scroll
	end := d.scrollY + 20
	if end > len(lines) {
		end = len(lines)
	}
	start := d.scrollY
	if start >= len(lines) {
		start = len(lines) - 1
	}
	if start < 0 {
		start = 0
	}
	return strings.Join(lines[start:end], "\n")
}
