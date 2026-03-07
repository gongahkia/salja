package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppInit(t *testing.T) {
	app := NewApp()
	if app.activeView != ViewHome {
		t.Errorf("expected initial view ViewHome, got %d", app.activeView)
	}
}

func TestAppQuitFromHome(t *testing.T) {
	app := NewApp()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	model, cmd := app.Update(msg)
	_ = model
	if cmd == nil {
		t.Error("expected quit command, got nil")
	}
}

func TestAppHelpToggle(t *testing.T) {
	app := NewApp()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}

	model, _ := app.Update(msg)
	a := model.(App)
	if !a.showHelp {
		t.Error("expected showHelp=true after pressing ?")
	}

	model, _ = a.Update(msg)
	a = model.(App)
	if a.showHelp {
		t.Error("expected showHelp=false after pressing ? again")
	}
}

func TestAppBackFromChild(t *testing.T) {
	app := NewApp()
	app.activeView = ViewConvert

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	model, _ := app.Update(msg)
	a := model.(App)
	if a.activeView != ViewHome {
		t.Errorf("expected ViewHome after esc, got %d", a.activeView)
	}
}

func TestAppNavigateMsg(t *testing.T) {
	app := NewApp()
	msg := NavigateMsg{Target: ViewValidate}
	model, _ := app.Update(msg)
	a := model.(App)
	if a.activeView != ViewValidate {
		t.Errorf("expected ViewValidate, got %d", a.activeView)
	}
}

func TestAppViewRenders(t *testing.T) {
	app := NewApp()
	view := app.View()
	if !strings.Contains(view, "salja") {
		t.Error("expected View to contain 'salja' title")
	}
}

func TestHomeModelView(t *testing.T) {
	h := NewHomeModel()
	view := h.View()
	if !strings.Contains(view, "salja") {
		t.Error("expected home view to contain 'salja'")
	}
}

func TestHelpModelView(t *testing.T) {
	h := NewHelpModel()
	view := h.View()
	if !strings.Contains(view, "Keybindings") {
		t.Error("expected help view to contain 'Keybindings'")
	}
	if !strings.Contains(view, "Quit") || !strings.Contains(view, "Toggle help") {
		t.Error("expected help view to list quit and help bindings")
	}
}

func TestErrorPanelView(t *testing.T) {
	e := NewErrorPanel()
	e.Push("error", "unexpected token")
	view := e.View()
	if !strings.Contains(view, "unexpected token") {
		t.Error("expected error panel to contain error message")
	}
	if !e.visible {
		t.Error("expected error panel to be visible after push")
	}
	e.Dismiss()
	if e.visible {
		t.Error("expected error panel to be hidden after dismiss")
	}
}

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()
	if len(km.Quit.Keys()) == 0 {
		t.Error("expected Quit binding to have keys")
	}
	if len(km.Help.Keys()) == 0 {
		t.Error("expected Help binding to have keys")
	}
}

func TestConvertModelInit(t *testing.T) {
	c := NewConvertModel()
	if c.step != stepInputFile {
		t.Errorf("expected initial step stepInputFile, got %d", c.step)
	}
}

func TestSyncModelNavigation(t *testing.T) {
	s := NewSyncModel()
	if s.cursor != 0 {
		t.Error("expected initial cursor at 0")
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	s, _ = s.Update(msg)
	if s.cursor != 1 {
		t.Errorf("expected cursor=1 after j, got %d", s.cursor)
	}

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	s, _ = s.Update(msg)
	if s.cursor != 0 {
		t.Errorf("expected cursor=0 after k, got %d", s.cursor)
	}
}

func TestAuthModelNavigation(t *testing.T) {
	a := NewAuthModel()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	a, _ = a.Update(msg)
	if a.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", a.cursor)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	app := NewApp()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	model, _ := app.Update(msg)
	a := model.(App)
	if a.width != 120 || a.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", a.width, a.height)
	}
}
