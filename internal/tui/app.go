package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View identifies the active TUI view.
type View int

const (
	ViewHome View = iota
	ViewConvert
	ViewValidate
	ViewDiff
	ViewSync
	ViewAuth
	ViewConfig
	ViewHelp
)

// App is the root bubbletea model that routes to child views.
type App struct {
	keys       KeyMap
	activeView View
	width      int
	height     int
	home       HomeModel
	convert    ConvertModel
	validate   ValidateModel
	diff       DiffModel
	sync       SyncModel
	auth       AuthModel
	config     ConfigModel
	help       HelpModel
	showHelp   bool
	err        error
}

// NewApp creates a new root TUI application model.
func NewApp() App {
	return App{
		keys:       DefaultKeyMap(),
		activeView: ViewHome,
		home:       NewHomeModel(),
	}
}

func (a App) Init() tea.Cmd {
	return a.home.Init()
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a.propagateSize(msg)

	case tea.KeyMsg:
		if key.Matches(msg, a.keys.Quit) && a.activeView == ViewHome && !a.showHelp {
			return a, tea.Quit
		}
		if key.Matches(msg, a.keys.Help) {
			a.showHelp = !a.showHelp
			return a, nil
		}
		if key.Matches(msg, a.keys.Back) {
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
			if a.activeView != ViewHome {
				a.activeView = ViewHome
				return a, nil
			}
		}

	case NavigateMsg:
		a.activeView = msg.Target
		return a.initView(msg.Target)

	case ErrorMsg:
		a.err = msg.Err
		return a, nil
	}

	return a.updateChild(msg)
}

func (a App) View() string {
	if a.showHelp {
		return a.help.View()
	}

	var content string
	switch a.activeView {
	case ViewHome:
		content = a.home.View()
	case ViewConvert:
		content = a.convert.View()
	case ViewValidate:
		content = a.validate.View()
	case ViewDiff:
		content = a.diff.View()
	case ViewSync:
		content = a.sync.View()
	case ViewAuth:
		content = a.auth.View()
	case ViewConfig:
		content = a.config.View()
	default:
		content = a.home.View()
	}

	header := TitleStyle.Render("salja")
	status := StatusBarStyle.Render("? help · esc back · q quit")

	return lipgloss.JoinVertical(lipgloss.Left, header, content, status)
}

func (a App) updateChild(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch a.activeView {
	case ViewHome:
		a.home, cmd = a.home.Update(msg)
	case ViewConvert:
		a.convert, cmd = a.convert.Update(msg)
	case ViewValidate:
		a.validate, cmd = a.validate.Update(msg)
	case ViewDiff:
		a.diff, cmd = a.diff.Update(msg)
	case ViewSync:
		a.sync, cmd = a.sync.Update(msg)
	case ViewAuth:
		a.auth, cmd = a.auth.Update(msg)
	case ViewConfig:
		a.config, cmd = a.config.Update(msg)
	}
	return a, cmd
}

func (a App) propagateSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	// Reserve space for header and status bar
	childMsg := tea.WindowSizeMsg{Width: msg.Width, Height: msg.Height - 3}
	var cmd tea.Cmd
	switch a.activeView {
	case ViewHome:
		a.home, cmd = a.home.Update(childMsg)
	default:
		return a.updateChild(childMsg)
	}
	return a, cmd
}

func (a App) initView(v View) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch v {
	case ViewConvert:
		a.convert = NewConvertModel()
		cmd = a.convert.Init()
	case ViewValidate:
		a.validate = NewValidateModel()
		cmd = a.validate.Init()
	case ViewDiff:
		a.diff = NewDiffModel()
		cmd = a.diff.Init()
	case ViewSync:
		a.sync = NewSyncModel()
		cmd = a.sync.Init()
	case ViewAuth:
		a.auth = NewAuthModel()
		cmd = a.auth.Init()
	case ViewConfig:
		a.config = NewConfigModel()
		cmd = a.config.Init()
	}
	return a, cmd
}

// NavigateMsg tells the app to switch to a different view.
type NavigateMsg struct {
	Target View
}

// ErrorMsg carries an error to display.
type ErrorMsg struct {
	Err error
}
