package apple

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

// CalendarParserAdapter wraps CalendarReader to implement the registry.Parser interface.
type CalendarParserAdapter struct {
	calendarName string
}

func NewCalendarParserAdapter() *CalendarParserAdapter {
	return &CalendarParserAdapter{}
}

func (a *CalendarParserAdapter) SetCalendarName(name string) {
	a.calendarName = name
}

func (a *CalendarParserAdapter) Parse(ctx context.Context, r io.Reader, source string) (*model.CalendarCollection, error) {
	return nil, fmt.Errorf("apple-calendar does not support reading from a stream; use ParseFile or sync pull")
}

func (a *CalendarParserAdapter) ParseFile(ctx context.Context, filePath string) (*model.CalendarCollection, error) {
	reader := NewCalendarReader()
	now := time.Now()
	return reader.Read(a.calendarName, now.AddDate(0, -1, 0), now.AddDate(0, 3, 0))
}

// CalendarWriterAdapter wraps CalendarWriter to implement the registry.Writer interface.
type CalendarWriterAdapter struct {
	calendarName string
}

func NewCalendarWriterAdapter() *CalendarWriterAdapter {
	return &CalendarWriterAdapter{calendarName: "Calendar"}
}

func (a *CalendarWriterAdapter) SetCalendarName(name string) {
	a.calendarName = name
}

func (a *CalendarWriterAdapter) Write(ctx context.Context, collection *model.CalendarCollection, w io.Writer) error {
	return fmt.Errorf("apple-calendar does not support writing to a stream; use WriteFile")
}

func (a *CalendarWriterAdapter) WriteFile(ctx context.Context, collection *model.CalendarCollection, filePath string) error {
	writer := NewCalendarWriter()
	return writer.Write(collection.Items, a.calendarName)
}

// RemindersParserAdapter wraps RemindersReader to implement the registry.Parser interface.
type RemindersParserAdapter struct{}

func NewRemindersParserAdapter() *RemindersParserAdapter {
	return &RemindersParserAdapter{}
}

func (a *RemindersParserAdapter) Parse(ctx context.Context, r io.Reader, source string) (*model.CalendarCollection, error) {
	return nil, fmt.Errorf("apple-reminders does not support reading from a stream; use ParseFile or sync pull")
}

func (a *RemindersParserAdapter) ParseFile(ctx context.Context, filePath string) (*model.CalendarCollection, error) {
	reader := NewRemindersReader()
	return reader.Read()
}

// RemindersWriterAdapter wraps RemindersWriter to implement the registry.Writer interface.
type RemindersWriterAdapter struct {
	listName string
}

func NewRemindersWriterAdapter() *RemindersWriterAdapter {
	return &RemindersWriterAdapter{listName: "Reminders"}
}

func (a *RemindersWriterAdapter) SetListName(name string) {
	a.listName = name
}

func (a *RemindersWriterAdapter) Write(ctx context.Context, collection *model.CalendarCollection, w io.Writer) error {
	return fmt.Errorf("apple-reminders does not support writing to a stream; use WriteFile")
}

func (a *RemindersWriterAdapter) WriteFile(ctx context.Context, collection *model.CalendarCollection, filePath string) error {
	writer := NewRemindersWriter()
	return writer.Write(collection.Items, a.listName)
}
