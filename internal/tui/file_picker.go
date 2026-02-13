package tui

import (
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

// FilePickerMsg is sent when a file is selected.
type FilePickerMsg struct {
	Path string
}

// FilePickerModel wraps the bubbles filepicker for selecting files.
type FilePickerModel struct {
	picker   filepicker.Model
	selected string
}

// NewFilePickerModel creates a file picker starting from the current directory.
func NewFilePickerModel() FilePickerModel {
	fp := filepicker.New()
	fp.CurrentDirectory = "."
	return FilePickerModel{picker: fp}
}

func (f FilePickerModel) Init() tea.Cmd {
	return f.picker.Init()
}

func (f FilePickerModel) Update(msg tea.Msg) (FilePickerModel, tea.Cmd) {
	var cmd tea.Cmd
	f.picker, cmd = f.picker.Update(msg)

	if didSelect, path := f.picker.DidSelectFile(msg); didSelect {
		f.selected = path
		return f, func() tea.Msg { return FilePickerMsg{Path: path} }
	}

	return f, cmd
}

func (f FilePickerModel) View() string {
	return SubtitleStyle.Render("Select a file:") + "\n" + f.picker.View()
}
