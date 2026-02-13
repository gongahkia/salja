package tui

import (
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/gongahkia/salja/internal/registry"
)

// FormatPickerMsg is sent when a format is selected.
type FormatPickerMsg struct {
	FormatID string
}

type formatItem struct {
	id   string
	name string
	exts string
}

func (f formatItem) Title() string       { return f.name }
func (f formatItem) Description() string { return f.exts }
func (f formatItem) FilterValue() string { return f.name + " " + f.id }

// FormatPickerModel shows a filterable list of supported formats.
type FormatPickerModel struct {
	list list.Model
	keys KeyMap
}

// NewFormatPickerModel creates a format picker populated from the registry.
func NewFormatPickerModel() FormatPickerModel {
	all := registry.AllFormats()
	items := make([]list.Item, 0, len(all))
	for id, entry := range all {
		exts := ""
		for i, e := range entry.Extensions {
			if i > 0 {
				exts += ", "
			}
			exts += e
		}
		items = append(items, formatItem{id: id, name: entry.Name, exts: exts})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].(formatItem).name < items[j].(formatItem).name
	})

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 50, 14)
	l.Title = "Select format"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return FormatPickerModel{list: l, keys: DefaultKeyMap()}
}

func (f FormatPickerModel) Init() tea.Cmd { return nil }

func (f FormatPickerModel) Update(msg tea.Msg) (FormatPickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if f.list.FilterState() == list.Filtering {
			break
		}
		if key.Matches(msg, f.keys.Enter) {
			if item, ok := f.list.SelectedItem().(formatItem); ok {
				return f, func() tea.Msg { return FormatPickerMsg{FormatID: item.id} }
			}
		}
	}
	var cmd tea.Cmd
	f.list, cmd = f.list.Update(msg)
	return f, cmd
}

func (f FormatPickerModel) View() string {
	return f.list.View()
}
