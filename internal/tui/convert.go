package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/gongahkia/salja/internal/registry"
)

type convertStep int

const (
	stepInputFile convertStep = iota
	stepFromFormat
	stepOutputFile
	stepToFormat
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

func (c ConvertModel) Update(msg tea.Msg) (ConvertModel, tea.Cmd) {
	switch msg := msg.(type) {
	case FilePickerMsg:
		switch c.step {
		case stepInputFile:
			c.inputPath = msg.Path
			c.step = stepFromFormat
			c.fmtPicker = NewFormatPickerModel()
			return c, c.fmtPicker.Init()
		case stepOutputFile:
			c.outputPath = msg.Path
			c.step = stepToFormat
			c.fmtPicker = NewFormatPickerModel()
			return c, c.fmtPicker.Init()
		}
	case FormatPickerMsg:
		switch c.step {
		case stepFromFormat:
			c.fromFormat = msg.FormatID
			c.step = stepOutputFile
			c.filePicker = NewFilePickerModel()
			return c, c.filePicker.Init()
		case stepToFormat:
			c.toFormat = msg.FormatID
			c.step = stepRunning
			return c, c.runConvert()
		}
	case convertDoneMsg:
		c.step = stepDone
		c.result = msg.summary
		c.err = msg.err
		return c, nil
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

func (c ConvertModel) runConvert() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		parser, err := registry.GetParser(c.fromFormat)
		if err != nil {
			return convertDoneMsg{err: err}
		}
		writer, err := registry.GetWriter(c.toFormat)
		if err != nil {
			return convertDoneMsg{err: err}
		}
		col, err := parser.ParseFile(ctx, c.inputPath)
		if err != nil {
			return convertDoneMsg{err: fmt.Errorf("parse: %w", err)}
		}
		if err := writer.WriteFile(ctx, col, c.outputPath); err != nil {
			return convertDoneMsg{err: fmt.Errorf("write: %w", err)}
		}
		count := len(col.Items)
		return convertDoneMsg{summary: fmt.Sprintf("Converted %d items from %s to %s", count, c.fromFormat, c.toFormat)}
	}
}

func (c ConvertModel) View() string {
	steps := []string{"input file", "source format", "output file", "target format", "converting", "done"}
	progress := fmt.Sprintf("Step %d/%d: %s", c.step+1, len(steps), steps[c.step])
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
