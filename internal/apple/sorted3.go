package apple

import (
"fmt"
"os"

"github.com/gongahkia/salja/internal/model"
)

type Sorted3Importer struct {
calendarName string
reminderList string
}

func NewSorted3Importer(calendarName, reminderList string) *Sorted3Importer {
return &Sorted3Importer{calendarName: calendarName, reminderList: reminderList}
}

func (s *Sorted3Importer) Import(collection *model.CalendarCollection) error {
calWriter := NewCalendarWriter()
remWriter := NewRemindersWriter()

var events, tasks []model.CalendarItem
for _, item := range collection.Items {
if item.ItemType == model.ItemTypeEvent {
events = append(events, item)
} else {
tasks = append(tasks, item)
}
}

if len(events) > 0 {
if err := calWriter.Write(events, s.calendarName); err != nil {
return fmt.Errorf("failed to write events to Apple Calendar: %w", err)
}
fmt.Fprintf(os.Stderr, "Wrote %d events to Apple Calendar '%s'\n", len(events), s.calendarName)
}

if len(tasks) > 0 {
if err := remWriter.Write(tasks, s.reminderList); err != nil {
return fmt.Errorf("failed to write tasks to Apple Reminders: %w", err)
}
fmt.Fprintf(os.Stderr, "Wrote %d tasks to Apple Reminders '%s'\n", len(tasks), s.reminderList)
}

fmt.Fprintln(os.Stderr, "\nTo sync with Sorted 3:")
fmt.Fprintln(os.Stderr, "  1. Open Sorted 3")
fmt.Fprintln(os.Stderr, "  2. Pull down to refresh")
fmt.Fprintf(os.Stderr, "  3. Ensure calendar '%s' and reminder list '%s' are enabled in Sorted 3 settings\n", s.calendarName, s.reminderList)

return nil
}

type Sorted3Exporter struct {
calendarName string
reminderList string
}

func NewSorted3Exporter(calendarName, reminderList string) *Sorted3Exporter {
return &Sorted3Exporter{calendarName: calendarName, reminderList: reminderList}
}

func (s *Sorted3Exporter) Export() (*model.CalendarCollection, error) {
collection := &model.CalendarCollection{
Items:     []model.CalendarItem{},
SourceApp: "sorted3",
}

remReader := NewRemindersReader()
reminders, err := remReader.Read()
if err != nil {
return nil, fmt.Errorf("failed to read reminders: %w", err)
}
collection.Items = append(collection.Items, reminders.Items...)

return collection, nil
}
