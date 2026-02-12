package registry

import (
	"github.com/gongahkia/salja/internal/ics"
	"github.com/gongahkia/salja/internal/parsers"
	"github.com/gongahkia/salja/internal/writers"
)

func init() {
	Register(&FormatEntry{
		Name:       "ics",
		Extensions: []string{".ics"},
		NewParser:  func() Parser { return ics.NewParser() },
		NewWriter:  func() Writer { return ics.NewWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     true,
			SupportsTasks:      true,
			SupportsRecurrence: true,
			SupportsSubtasks:   false,
		},
	})

	Register(&FormatEntry{
		Name:         "ticktick",
		Extensions:   []string{".csv"},
		FilenameHint: []string{"ticktick"},
		NewParser:    func() Parser { return parsers.NewTickTickParser() },
		NewWriter:    func() Writer { return writers.NewTickTickWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     false,
			SupportsTasks:      true,
			SupportsRecurrence: true,
			SupportsSubtasks:   true,
		},
	})

	Register(&FormatEntry{
		Name:         "todoist",
		Extensions:   []string{".csv"},
		FilenameHint: []string{"todoist"},
		NewParser:    func() Parser { return parsers.NewTodoistParser() },
		NewWriter:    func() Writer { return writers.NewTodoistWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     false,
			SupportsTasks:      true,
			SupportsRecurrence: false,
			SupportsSubtasks:   true,
		},
	})

	Register(&FormatEntry{
		Name:         "gcal",
		Extensions:   []string{".csv"},
		FilenameHint: []string{"google", "gcal"},
		NewParser:    func() Parser { return parsers.NewGoogleCalendarParser() },
		NewWriter:    func() Writer { return writers.NewGoogleCalendarWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     true,
			SupportsTasks:      false,
			SupportsRecurrence: false,
			SupportsSubtasks:   false,
		},
	})

	Register(&FormatEntry{
		Name:         "outlook",
		Extensions:   []string{".csv"},
		FilenameHint: []string{"outlook"},
		NewParser:    func() Parser { return parsers.NewOutlookParser() },
		NewWriter:    func() Writer { return writers.NewOutlookWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     true,
			SupportsTasks:      false,
			SupportsRecurrence: false,
			SupportsSubtasks:   false,
		},
	})

	Register(&FormatEntry{
		Name:         "notion",
		Extensions:   []string{".csv"},
		FilenameHint: []string{"notion"},
		NewParser:    func() Parser { return parsers.NewNotionParser() },
		NewWriter:    func() Writer { return writers.NewNotionWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     false,
			SupportsTasks:      true,
			SupportsRecurrence: false,
			SupportsSubtasks:   false,
		},
	})

	Register(&FormatEntry{
		Name:       "trello",
		Extensions: []string{".json"},
		NewParser:  func() Parser { return parsers.NewTrelloParser() },
		NewWriter:  func() Writer { return writers.NewTrelloWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     false,
			SupportsTasks:      true,
			SupportsRecurrence: false,
			SupportsSubtasks:   true,
		},
	})

	Register(&FormatEntry{
		Name:         "asana",
		Extensions:   []string{".csv"},
		FilenameHint: []string{"asana"},
		NewParser:    func() Parser { return parsers.NewAsanaParser() },
		NewWriter:    func() Writer { return writers.NewAsanaWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     false,
			SupportsTasks:      true,
			SupportsRecurrence: false,
			SupportsSubtasks:   false,
		},
	})

	Register(&FormatEntry{
		Name:       "omnifocus",
		Extensions: []string{".taskpaper"},
		NewParser:  func() Parser { return parsers.NewOmniFocusParser() },
		NewWriter:  func() Writer { return writers.NewOmniFocusWriter() },
		Capabilities: FormatCapabilities{
			SupportsEvents:     false,
			SupportsTasks:      true,
			SupportsRecurrence: false,
			SupportsSubtasks:   true,
		},
	})
}
