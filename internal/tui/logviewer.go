package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gongahkia/salja/internal/logging"
)

// LogViewerModel displays the tail of the log file.
type LogViewerModel struct {
	lines   []logLine
	scroll  int
	follow  bool
	logPath string
	height  int
}

type logLine struct {
	raw     string
	ts      string
	level   string
	cat     string
	msg     string
	colored string
}

type logLoadedMsg struct {
	lines []logLine
}

// NewLogViewerModel creates a log viewer.
func NewLogViewerModel() LogViewerModel {
	return LogViewerModel{follow: true, logPath: logging.Default().Path()}
}

func (l LogViewerModel) Init() tea.Cmd {
	path := l.logPath
	return func() tea.Msg {
		return logLoadedMsg{lines: readLogTail(path, 200)}
	}
}

func readLogTail(path string, maxLines int) []logLine {
	f, err := os.Open(path)
	if err != nil {
		return []logLine{{raw: fmt.Sprintf("(cannot open log: %v)", err)}}
	}
	defer f.Close()
	var all []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		all = append(all, scanner.Text())
	}
	if len(all) > maxLines {
		all = all[len(all)-maxLines:]
	}
	lines := make([]logLine, 0, len(all))
	for _, raw := range all {
		lines = append(lines, parseLine(raw))
	}
	return lines
}

func parseLine(raw string) logLine {
	ll := logLine{raw: raw}
	var obj map[string]string
	if err := json.Unmarshal([]byte(raw), &obj); err == nil {
		ll.ts = obj["ts"]
		ll.level = obj["level"]
		ll.cat = obj["cat"]
		ll.msg = obj["msg"]
	}
	// pre-render colored version
	var style lipgloss.Style
	switch ll.level {
	case "error":
		style = ErrorStyle
	case "warn":
		style = WarningStyle
	default:
		style = lipgloss.NewStyle()
	}
	if ll.ts != "" {
		ts := ll.ts
		if len(ts) > 19 {
			ts = ts[:19]
		}
		ll.colored = fmt.Sprintf("%s %s %s: %s",
			MutedStyle.Render(ts),
			style.Render(ll.level),
			MutedStyle.Render(ll.cat),
			ll.msg)
	} else {
		ll.colored = raw
	}
	return ll
}

func (l LogViewerModel) Update(msg tea.Msg) (LogViewerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case logLoadedMsg:
		l.lines = msg.lines
		if l.follow && len(l.lines) > l.visibleCount() {
			l.scroll = len(l.lines) - l.visibleCount()
		}
		return l, nil
	case tea.WindowSizeMsg:
		l.height = msg.Height
		return l, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			max := len(l.lines) - l.visibleCount()
			if max < 0 {
				max = 0
			}
			if l.scroll < max {
				l.scroll++
				l.follow = false
			}
		case "k", "up":
			if l.scroll > 0 {
				l.scroll--
				l.follow = false
			}
		case "f":
			l.follow = !l.follow
			if l.follow {
				max := len(l.lines) - l.visibleCount()
				if max > 0 {
					l.scroll = max
				}
			}
		case "r":
			path := l.logPath
			return l, func() tea.Msg {
				return logLoadedMsg{lines: readLogTail(path, 200)}
			}
		}
	}
	return l, nil
}

func (l LogViewerModel) visibleCount() int {
	if l.height > 5 {
		return l.height - 5
	}
	return 15
}

func (l LogViewerModel) View() string {
	header := SubtitleStyle.Render("Log Viewer")
	if len(l.lines) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, MutedStyle.Render("  (empty log)"))
	}
	visible := l.visibleCount()
	start := l.scroll
	end := start + visible
	if end > len(l.lines) {
		end = len(l.lines)
	}
	if start < 0 {
		start = 0
	}
	var rows []string
	for _, line := range l.lines[start:end] {
		rows = append(rows, "  "+line.colored)
	}
	body := strings.Join(rows, "\n")
	followIndicator := ""
	if l.follow {
		followIndicator = SuccessStyle.Render(" [follow]")
	}
	help := HelpStyle.Render(fmt.Sprintf("↑↓ scroll · f toggle follow · r refresh%s", followIndicator))
	path := MutedStyle.Render("  " + l.logPath)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, "", help, path)
}
