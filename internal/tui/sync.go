package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gongahkia/salja/internal/api"
	"github.com/gongahkia/salja/internal/logging"
	"github.com/gongahkia/salja/internal/registry"
)

type syncStep int

const (
	syncStepMenu syncStep = iota
	syncStepPickFile
	syncStepPickFormat
	syncStepRunning
	syncStepDone
)

// SyncModel displays cloud sync dashboard.
type SyncModel struct {
	services   []syncService
	cursor     int
	keys       KeyMap
	message    string
	step       syncStep
	mode       string // "push" or "pull"
	filePicker FilePickerModel
	fmtPicker  FormatPickerModel
	filePath   string
	format     string
	result     string
	err        error
}

type syncService struct {
	name          string
	authenticated bool
}

type syncStatusMsg struct {
	services []syncService
}

type syncDoneMsg struct {
	summary string
	err     error
}

// NewSyncModel creates a sync dashboard.
func NewSyncModel() SyncModel {
	services := make([]syncService, len(serviceNames))
	for i, name := range serviceNames {
		services[i] = syncService{name: name}
	}
	return SyncModel{services: services, keys: DefaultKeyMap(), step: syncStepMenu}
}

func (s SyncModel) Init() tea.Cmd {
	return loadSyncStatus
}

func loadSyncStatus() tea.Msg {
	store, err := api.DefaultTokenStore()
	if err != nil {
		return syncStatusMsg{}
	}
	tokens, err := store.Load()
	if err != nil {
		return syncStatusMsg{}
	}
	services := make([]syncService, len(serviceNames))
	for i, name := range serviceNames {
		svc := syncService{name: name}
		tok, ok := tokens[name]
		if ok && tok != nil && !tok.IsExpired() {
			svc.authenticated = true
		}
		services[i] = svc
	}
	return syncStatusMsg{services: services}
}

func (s SyncModel) Update(msg tea.Msg) (SyncModel, tea.Cmd) {
	switch msg := msg.(type) {
	case syncStatusMsg:
		if len(msg.services) > 0 {
			s.services = msg.services
		}
		return s, nil
	case syncDoneMsg:
		s.step = syncStepDone
		s.result = msg.summary
		s.err = msg.err
		return s, nil
	case FilePickerMsg:
		if s.step == syncStepPickFile {
			s.filePath = msg.Path
			s.step = syncStepPickFormat
			s.fmtPicker = NewFormatPickerModel()
			return s, s.fmtPicker.Init()
		}
	case FormatPickerMsg:
		if s.step == syncStepPickFormat {
			s.format = msg.FormatID
			s.step = syncStepRunning
			logging.Default().Info("interaction", fmt.Sprintf("sync: %s %s as %s to %s", s.mode, s.filePath, s.format, s.services[s.cursor].name))
			return s, s.runSync()
		}
	case tea.KeyMsg:
		switch s.step {
		case syncStepMenu:
			switch msg.String() {
			case "j", "down":
				if s.cursor < len(s.services)-1 {
					s.cursor++
				}
			case "k", "up":
				if s.cursor > 0 {
					s.cursor--
				}
			case "p":
				svc := s.services[s.cursor]
				if !svc.authenticated {
					s.message = fmt.Sprintf("%s not authenticated — run: salja auth login %s", svc.name, svc.name)
					return s, nil
				}
				s.mode = "push"
				s.step = syncStepPickFile
				s.filePicker = NewFilePickerModel()
				return s, s.filePicker.Init()
			case "l":
				svc := s.services[s.cursor]
				if !svc.authenticated {
					s.message = fmt.Sprintf("%s not authenticated — run: salja auth login %s", svc.name, svc.name)
					return s, nil
				}
				s.mode = "pull"
				s.step = syncStepRunning
				logging.Default().Info("interaction", fmt.Sprintf("sync: pull from %s", svc.name))
				s.message = fmt.Sprintf("Pull from %s — use CLI: salja sync pull --from %s", svc.name, svc.name)
				s.step = syncStepDone
				s.result = s.message
				return s, nil
			}
		case syncStepPickFile:
			s.filePicker, _ = s.filePicker.Update(msg)
		case syncStepPickFormat:
			s.fmtPicker, _ = s.fmtPicker.Update(msg)
		}
	}

	var cmd tea.Cmd
	switch s.step {
	case syncStepPickFile:
		s.filePicker, cmd = s.filePicker.Update(msg)
	case syncStepPickFormat:
		s.fmtPicker, cmd = s.fmtPicker.Update(msg)
	}
	return s, cmd
}

func (s SyncModel) runSync() tea.Cmd {
	svcName := s.services[s.cursor].name
	filePath := s.filePath
	format := s.format
	return func() tea.Msg {
		ctx := context.Background()
		parser, err := registry.GetParser(format)
		if err != nil {
			return syncDoneMsg{err: err}
		}
		col, err := parser.ParseFile(ctx, filePath)
		if err != nil {
			return syncDoneMsg{err: fmt.Errorf("parse: %w", err)}
		}
		count := len(col.Items)
		logging.Default().Info("interaction", fmt.Sprintf("sync: parsed %d items for push to %s", count, svcName))
		return syncDoneMsg{summary: fmt.Sprintf("Parsed %d items for push to %s\nUse CLI for actual push: salja sync push %s --to %s", count, svcName, filePath, svcName)}
	}
}

func (s SyncModel) View() string {
	header := SubtitleStyle.Render("Cloud Sync")

	switch s.step {
	case syncStepPickFile:
		return lipgloss.JoinVertical(lipgloss.Left, header, SubtitleStyle.Render("Select file to push"), s.filePicker.View())
	case syncStepPickFormat:
		return lipgloss.JoinVertical(lipgloss.Left, header, SubtitleStyle.Render("Select file format"), s.fmtPicker.View())
	case syncStepRunning:
		return lipgloss.JoinVertical(lipgloss.Left, header, MutedStyle.Render("  Syncing..."))
	case syncStepDone:
		var body string
		if s.err != nil {
			body = ErrorStyle.Render("  Error: " + s.err.Error())
		} else {
			body = SuccessStyle.Render("  " + s.result)
		}
		return lipgloss.JoinVertical(lipgloss.Left, header, body)
	}

	// menu view
	var rows string
	for i, svc := range s.services {
		cursor := "  "
		if i == s.cursor {
			cursor = "▸ "
		}
		status := MutedStyle.Render("✗ not authenticated")
		if svc.authenticated {
			status = SuccessStyle.Render("✓ authenticated")
		}
		rows += fmt.Sprintf("%s%-12s %s\n", cursor, svc.name, status)
	}
	help := HelpStyle.Render("p push · l pull · ↑↓ navigate")
	var msg string
	if s.message != "" {
		msg = "\n" + WarningStyle.Render("  "+s.message)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, "", rows, msg, help)
}
