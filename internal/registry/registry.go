package registry

import (
	"context"
	"fmt"
	"io"

	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/platform"
)

type Parser interface {
	Parse(context.Context, io.Reader, string) (*model.CalendarCollection, error)
	ParseFile(context.Context, string) (*model.CalendarCollection, error)
}

type Writer interface {
	Write(context.Context, *model.CalendarCollection, io.Writer) error
	WriteFile(context.Context, *model.CalendarCollection, string) error
}

type ParserFactory func() Parser

type WriterFactory func() Writer

type FormatCapabilities struct {
	SupportsEvents     bool
	SupportsTasks      bool
	SupportsRecurrence bool
	SupportsSubtasks   bool
	SupportsReminders  bool
}

type FormatEntry struct {
	Name         string
	Extensions   []string
	FilenameHint []string // substrings in filename used for CSV disambiguation
	Platform     string   // "" = all platforms, "darwin" = macOS only
	NewParser    ParserFactory
	NewWriter    WriterFactory
	Capabilities FormatCapabilities
}

var formats = map[string]*FormatEntry{}

func Register(entry *FormatEntry) {
	formats[entry.Name] = entry
}

func GetParser(name string) (Parser, error) {
	entry, ok := formats[name]
	if !ok || entry.NewParser == nil {
		return nil, fmt.Errorf("unsupported input format: %s", name)
	}
	if entry.Platform != "" && entry.Platform != platform.DetectOS() {
		return nil, fmt.Errorf("format %q requires %s (current: %s)", name, entry.Platform, platform.DetectOS())
	}
	return entry.NewParser(), nil
}

func GetWriter(name string) (Writer, error) {
	entry, ok := formats[name]
	if !ok || entry.NewWriter == nil {
		return nil, fmt.Errorf("unsupported output format: %s", name)
	}
	if entry.Platform != "" && entry.Platform != platform.DetectOS() {
		return nil, fmt.Errorf("format %q requires %s (current: %s)", name, entry.Platform, platform.DetectOS())
	}
	return entry.NewWriter(), nil
}

func DetectByExtension(ext string) []string {
	var matches []string
	for name, entry := range formats {
		for _, e := range entry.Extensions {
			if e == ext {
				matches = append(matches, name)
			}
		}
	}
	return matches
}

func DetectByFilenameHint(basename string) string {
	for name, entry := range formats {
		for _, hint := range entry.FilenameHint {
			if containsLower(basename, hint) {
				return name
			}
		}
	}
	return ""
}

func AllFormats() map[string]*FormatEntry {
	return formats
}

// AvailableFormats returns only formats supported on the current platform.
func AvailableFormats() map[string]*FormatEntry {
	os := platform.DetectOS()
	result := make(map[string]*FormatEntry, len(formats))
	for name, entry := range formats {
		if entry.Platform == "" || entry.Platform == os {
			result[name] = entry
		}
	}
	return result
}

func GetCapabilities(name string) (FormatCapabilities, bool) {
	entry, ok := formats[name]
	if !ok {
		return FormatCapabilities{}, false
	}
	return entry.Capabilities, true
}

func containsLower(s, substr string) bool {
	// caller is expected to pass already-lowercased values
	return len(s) >= len(substr) && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
