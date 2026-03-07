package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gongahkia/salja/internal/fidelity"
	"github.com/gongahkia/salja/internal/logging"
	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/registry"
)

type convertStep int

const (
	stepInputFile convertStep = iota
	stepFromFormat
	stepOutputFile
	stepToFormat
	stepPreview
	stepRunning
	stepDone
)

// ConvertModel implements the multi-step conversion workflow.
type ConvertModel struct {
	step       convertStep
	filePicker FilePickerModel
	fmtPicker  FormatPickerModel
	inputPath  string
	outputPath string
	fromFormat string
	toFormat   string
	result     string
	err        error
	warnings   []fidelity.DataLossWarning
	parsed     *model.CalendarCollection
}

// NewConvertModel creates a new conversion workflow.
func NewConvertModel() ConvertModel {
	return ConvertModel{
		step:       stepInputFile,
		filePicker: NewFilePickerModel(),
	}
}

func (c ConvertModel) Init() tea.Cmd {
	return c.filePicker.Init()
}

type previewMsg struct {
	collection *model.CalendarCollection
	warnings   []fidelity.DataLossWarning
	err        error
}

func (c ConvertModel) Update(msg tea.Msg) (ConvertModel, tea.Cmd) {
	switch msg := msg.(type) {
	case FilePickerMsg:
		switch c.step {
		case stepInputFile:
			c.inputPath = msg.Path
			logging.Default().Info("interaction", fmt.Sprintf("convert: selected source %s", msg.Path))
			ext := strings.ToLower(filepath.Ext(msg.Path))
			matches := registry.DetectByExtension(ext)
			if len(matches) == 1 {
				c.fromFormat = matches[0]
				c.step = stepOutputFile
				c.filePicker = NewFilePickerModel()
				return c, c.filePicker.Init()
			}
			c.step = stepFromFormat
			c.fmtPicker = NewFormatPickerModel(matches...)
			return c, c.fmtPicker.Init()
		case stepOutputFile:
			c.outputPath = msg.Path
			logging.Default().Info("interaction", fmt.Sprintf("convert: selected output %s", msg.Path))
			ext := strings.ToLower(filepath.Ext(msg.Path))
			matches := registry.DetectByExtension(ext)
			if len(matches) == 1 {
				c.toFormat = matches[0]
				c.step = stepPreview
				return c, c.loadPreview()
			}
			c.step = stepToFormat
			c.fmtPicker = NewFormatPickerModel(matches...)
			return c, c.fmtPicker.Init()
		}
	case FormatPickerMsg:
		switch c.step {
		case stepFromFormat:
			c.fromFormat = msg.FormatID
			logging.Default().Info("interaction", fmt.Sprintf("convert: source format %s", msg.FormatID))
			c.step = stepOutputFile
			c.filePicker = NewFilePickerModel()
			return c, c.filePicker.Init()
		case stepToFormat:
			c.toFormat = msg.FormatID
			logging.Default().Info("interaction", fmt.Sprintf("convert: target format %s", msg.FormatID))
			c.step = stepPreview
			return c, c.loadPreview()
		}
	case previewMsg:
		if msg.err != nil {
			c.step = stepDone
			c.err = msg.err
			return c, nil
		}
		c.parsed = msg.collection
		c.warnings = msg.warnings
		c.step = stepPreview
		return c, nil
	case convertDoneMsg:
		c.step = stepDone
		c.result = msg.summary
		c.err = msg.err
		return c, nil
	case tea.KeyMsg:
		if c.step == stepPreview {
			if msg.String() == "enter" || msg.String() == "y" {
				c.step = stepRunning
				return c, c.runConvertFromParsed()
			}
			if msg.String() == "n" || msg.String() == "esc" {
				c.step = stepDone
				c.result = "Conversion cancelled"
				return c, nil
			}
		}
	}

	var cmd tea.Cmd
	switch c.step {
	case stepInputFile, stepOutputFile:
		c.filePicker, cmd = c.filePicker.Update(msg)
	case stepFromFormat, stepToFormat:
		c.fmtPicker, cmd = c.fmtPicker.Update(msg)
	}
	return c, cmd
}

type convertDoneMsg struct {
	summary string
	err     error
}

func (c ConvertModel) loadPreview() tea.Cmd {
	from := c.fromFormat
	to := c.toFormat
	input := c.inputPath
	return func() tea.Msg {
		ctx := context.Background()
		parser, err := registry.GetParser(from)
		if err != nil {
			return previewMsg{err: err}
		}
		col, err := parser.ParseFile(ctx, input)
		if err != nil {
			return previewMsg{err: fmt.Errorf("parse: %w", err)}
		}
		warnings := fidelity.Check(col, to)
		return previewMsg{collection: col, warnings: warnings}
	}
}

func (c ConvertModel) runConvertFromParsed() tea.Cmd {
	col := c.parsed
	toFmt := c.toFormat
	fromFmt := c.fromFormat
	output := c.outputPath
	return func() tea.Msg {
		ctx := context.Background()
		writer, err := registry.GetWriter(toFmt)
		if err != nil {
			return convertDoneMsg{err: err}
		}
		if err := writer.WriteFile(ctx, col, output); err != nil {
			return convertDoneMsg{err: fmt.Errorf("write: %w", err)}
		}
		count := len(col.Items)
		logging.Default().Info("interaction", fmt.Sprintf("convert: wrote %d items to %s", count, output))
		return convertDoneMsg{summary: fmt.Sprintf("Converted %d items from %s to %s\nOutput: %s", count, fromFmt, toFmt, output)}
	}
}

func (c ConvertModel) View() string {
	stepNames := []string{"input file", "source format", "output file", "target format", "preview", "converting", "done"}
	progress := fmt.Sprintf("Step %d/%d: %s", c.step+1, len(stepNames), stepNames[c.step])
	header := SubtitleStyle.Render("Convert · " + progress)

	var body string
	switch c.step {
	case stepInputFile:
		body = c.filePicker.View()
	case stepFromFormat:
		body = c.fmtPicker.View()
	case stepOutputFile:
		body = c.filePicker.View()
	case stepToFormat:
		body = c.fmtPicker.View()
	case stepPreview:
		body = c.viewPreview()
	case stepRunning:
		body = MutedStyle.Render("  Converting...")
	case stepDone:
		if c.err != nil {
			body = ErrorStyle.Render("  Error: " + c.err.Error())
		} else {
			body = SuccessStyle.Render("  " + c.result)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (c ConvertModel) viewPreview() string {
	if c.parsed == nil {
		return MutedStyle.Render("  Loading preview...")
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("  Source: %s (%s)", c.inputPath, c.fromFormat))
	lines = append(lines, fmt.Sprintf("  Target: %s (%s)", c.outputPath, c.toFormat))
	lines = append(lines, fmt.Sprintf("  Items: %d", len(c.parsed.Items)))
	lines = append(lines, "")
	if len(c.warnings) == 0 {
		lines = append(lines, SuccessStyle.Render("  No fidelity warnings"))
	} else {
		lines = append(lines, WarningStyle.Render(fmt.Sprintf("  %d fidelity warning(s):", len(c.warnings))))
		limit := len(c.warnings)
		if limit > 10 {
			limit = 10
		}
		for _, w := range c.warnings[:limit] {
			lines = append(lines, WarningStyle.Render("    - "+w.String()))
		}
		if len(c.warnings) > 10 {
			lines = append(lines, MutedStyle.Render(fmt.Sprintf("    ... and %d more", len(c.warnings)-10)))
		}
	}
	lines = append(lines, "")
	lines = append(lines, MutedStyle.Render("  Press Enter/y to proceed, n/Esc to cancel"))
	return strings.Join(lines, "\n")
}
