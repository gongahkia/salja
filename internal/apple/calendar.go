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

// asDateVar builds AppleScript lines that construct a date variable from Go time.Time.
// This is locale-independent — it sets year/month/day/hours/minutes/seconds explicitly.
func asDateVar(varName string, t time.Time) string {
	return fmt.Sprintf(`set %s to current date
set year of %s to %d
set month of %s to %d
set day of %s to %d
set hours of %s to %d
set minutes of %s to %d
set seconds of %s to %d`,
		varName,
		varName, t.Year(),
		varName, int(t.Month()),
		varName, t.Day(),
		varName, t.Hour(),
		varName, t.Minute(),
		varName, t.Second())
}

func (w *CalendarWriter) createEvent(item model.CalendarItem, calendarName string) error {
	if item.StartTime == nil {
		return fmt.Errorf("event '%s' has no start time", item.Title)
	}

	endTime := *item.StartTime
	if item.EndTime != nil {
		endTime = *item.EndTime
	}

	var scriptParts []string
	scriptParts = append(scriptParts, asDateVar("startD", *item.StartTime))
	scriptParts = append(scriptParts, asDateVar("endD", endTime))
	scriptParts = append(scriptParts, `tell application "Calendar"`)
	scriptParts = append(scriptParts, fmt.Sprintf(`  tell calendar "%s"`, escapeAS(calendarName)))

	props := []string{
		fmt.Sprintf(`summary:"%s"`, escapeAS(item.Title)),
		`start date:startD`,
		`end date:endD`,
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

// isoDateScript generates AppleScript that formats a date object as ISO 8601.
const isoDateScript = `on isoDate(d)
set y to year of d as string
set m to (month of d as integer) as string
if length of m < 2 then set m to "0" & m
set dd to day of d as string
if length of dd < 2 then set dd to "0" & dd
set h to hours of d as string
if length of h < 2 then set h to "0" & h
set mi to minutes of d as string
if length of mi < 2 then set mi to "0" & mi
set s to seconds of d as string
if length of s < 2 then set s to "0" & s
return y & "-" & m & "-" & dd & "T" & h & ":" & mi & ":" & s
end isoDate`

func (r *CalendarReader) Read(calendarName string, startDate, endDate time.Time) (*model.CalendarCollection, error) {
	script := fmt.Sprintf(`%s

%s
%s
tell application "Calendar"
set output to ""
tell calendar "%s"
set evts to every event whose start date >= startD and start date <= endD
repeat with evt in evts
set evtSummary to summary of evt
set evtStart to my isoDate(start date of evt)
set evtEnd to my isoDate(end date of evt)
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
		isoDateScript,
		asDateVar("startD", startDate),
		asDateVar("endD", endDate),
		escapeAS(calendarName),
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

		if t, err := time.Parse("2006-01-02T15:04:05", parts[1]); err == nil {
			item.StartTime = &t
		}
		if t, err := time.Parse("2006-01-02T15:04:05", parts[2]); err == nil {
			item.EndTime = &t
		}

		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}
