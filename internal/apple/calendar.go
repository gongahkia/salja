package apple

import (
"fmt"
"strings"
"time"

"github.com/gongahkia/salja/internal/model"
)

type CalendarWriter struct{}

func NewCalendarWriter() *CalendarWriter {
return &CalendarWriter{}
}

func (w *CalendarWriter) Write(items []model.CalendarItem, calendarName string) error {
for _, item := range items {
if err := w.createEvent(item, calendarName); err != nil {
return fmt.Errorf("failed to create event '%s': %w", item.Title, err)
}
}
return nil
}

func (w *CalendarWriter) createEvent(item model.CalendarItem, calendarName string) error {
if item.StartTime == nil {
return fmt.Errorf("event '%s' has no start time", item.Title)
}

startStr := item.StartTime.Format("January 2, 2006 3:04:05 PM")
endStr := startStr
if item.EndTime != nil {
endStr = item.EndTime.Format("January 2, 2006 3:04:05 PM")
}

var scriptParts []string
scriptParts = append(scriptParts, `tell application "Calendar"`)
scriptParts = append(scriptParts, fmt.Sprintf(`  tell calendar "%s"`, calendarName))

props := []string{
fmt.Sprintf(`summary:"%s"`, escapeAS(item.Title)),
fmt.Sprintf(`start date:date "%s"`, startStr),
fmt.Sprintf(`end date:date "%s"`, endStr),
}

if item.Description != "" {
props = append(props, fmt.Sprintf(`description:"%s"`, escapeAS(item.Description)))
}
if item.Location != "" {
props = append(props, fmt.Sprintf(`location:"%s"`, escapeAS(item.Location)))
}
if item.IsAllDay {
props = append(props, `allday event:true`)
}

scriptParts = append(scriptParts, fmt.Sprintf(`    make new event with properties {%s}`, strings.Join(props, ", ")))
scriptParts = append(scriptParts, `  end tell`)
scriptParts = append(scriptParts, `end tell`)

_, err := scriptRunnerFn(strings.Join(scriptParts, "\n"))
return err
}

type CalendarReader struct{}

func NewCalendarReader() *CalendarReader {
return &CalendarReader{}
}

func (r *CalendarReader) Read(calendarName string, startDate, endDate time.Time) (*model.CalendarCollection, error) {
script := fmt.Sprintf(`tell application "Calendar"
set output to ""
set startD to date "%s"
set endD to date "%s"
tell calendar "%s"
set evts to every event whose start date >= startD and start date <= endD
repeat with evt in evts
set evtSummary to summary of evt
set evtStart to (start date of evt) as string
set evtEnd to (end date of evt) as string
set evtDesc to ""
try
set evtDesc to description of evt
end try
set evtLoc to ""
try
set evtLoc to location of evt
end try
set output to output & evtSummary & "|||" & evtStart & "|||" & evtEnd & "|||" & evtDesc & "|||" & evtLoc & linefeed
end repeat
end tell
return output
end tell`,
startDate.Format("January 2, 2006"),
endDate.Format("January 2, 2006"),
calendarName,
)

output, err := scriptRunnerFn(script)
if err != nil {
return nil, err
}

collection := &model.CalendarCollection{
Items:      []model.CalendarItem{},
SourceApp:  "apple-calendar",
ExportDate: time.Now(),
}

lines := strings.Split(output, "\n")
for _, line := range lines {
line = strings.TrimSpace(line)
if line == "" {
continue
}
parts := strings.Split(line, "|||")
if len(parts) < 5 {
continue
}

item := model.CalendarItem{
ItemType:    model.ItemTypeEvent,
Status:      model.StatusPending,
Title:       parts[0],
Description: parts[3],
Location:    parts[4],
}

collection.Items = append(collection.Items, item)
}

return collection, nil
}
