package tui

import (
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	title string
	desc  string
	view  View
}

func (m menuItem) Title() string       { return m.title }
func (m menuItem) Description() string { return m.desc }
func (m menuItem) FilterValue() string { return m.title }

var menuItems = []list.Item{
	menuItem{"convert", "Convert between calendar/task formats", ViewConvert},
	menuItem{"validate", "Validate a calendar/task file", ViewValidate},
	menuItem{"diff", "Compare two calendar/task files", ViewDiff},
	menuItem{"sync", "Cloud sync (push/pull)", ViewSync},
	menuItem{"auth", "Manage service authentication", ViewAuth},
	menuItem{"config", "View configuration", ViewConfig},
}

// HomeModel is the main menu view.
type HomeModel struct {
	list list.Model
	keys KeyMap
}

// NewHomeModel creates a new home menu.
func NewHomeModel() HomeModel {
	items := make([]list.Item, len(menuItems))
	copy(items, menuItems)
	sort.Slice(items, func(i, j int) bool {
		return items[i].(menuItem).title < items[j].(menuItem).title
	})

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(ColorPrimary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(ColorSecondary)

	l := list.New(items, delegate, 60, 14)
	l.Title = "salja"
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).MarginLeft(1)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	return HomeModel{list: l, keys: DefaultKeyMap()}
}

func (h HomeModel) Init() tea.Cmd {
	return nil
}

func (h HomeModel) Update(msg tea.Msg) (HomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.list.SetSize(msg.Width, msg.Height)
		return h, nil
	case tea.KeyMsg:
		if h.list.FilterState() == list.Filtering {
			break
		}
		if key.Matches(msg, h.keys.Enter) {
			if item, ok := h.list.SelectedItem().(menuItem); ok {
				return h, func() tea.Msg { return NavigateMsg{Target: item.view} }
			}
		}
	}
	var cmd tea.Cmd
	h.list, cmd = h.list.Update(msg)
	return h, cmd
}

func (h HomeModel) View() string {
	return h.list.View()
}
