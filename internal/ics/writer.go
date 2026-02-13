package ics

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/gongahkia/salja/internal/model"
)

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) WriteFile(ctx context.Context, collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create ICS file: %w", err)
	}
	defer f.Close()

	return w.Write(ctx, collection, f)
}

func (w *Writer) Write(ctx context.Context, collection *model.CalendarCollection, writer io.Writer) error {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//gongahkia//salja//EN")

	for _, item := range collection.Items {
		var comp *ical.Component

		switch item.ItemType {
		case model.ItemTypeEvent:
			comp = w.writeEvent(&item)
		case model.ItemTypeTask:
			comp = w.writeTodo(&item)
		case model.ItemTypeJournal:
			comp = w.writeJournal(&item)
		default:
			continue
		}

		if comp != nil {
			cal.Children = append(cal.Children, comp)
		}
	}

	enc := ical.NewEncoder(writer)
	return enc.Encode(cal)
}

func (w *Writer) writeEvent(item *model.CalendarItem) *ical.Component {
	event := ical.NewComponent("VEVENT")

	event.Props.SetText("UID", item.UID)
	event.Props.SetText("SUMMARY", item.Title)
	event.Props.SetDateTime("DTSTAMP", time.Now().UTC())

	if item.Description != "" {
		event.Props.SetText("DESCRIPTION", item.Description)
	}

	if item.Location != "" {
		event.Props.SetText("LOCATION", item.Location)
	}

	if item.StartTime != nil {
		w.setDateTime(event.Props, "DTSTART", *item.StartTime, item.IsAllDay, item.Timezone)
	}

	if item.EndTime != nil {
		w.setDateTime(event.Props, "DTEND", *item.EndTime, item.IsAllDay, item.Timezone)
	}

	if item.Priority > 0 {
		event.Props.SetText("PRIORITY", fmt.Sprintf("%d", exportPriority(item.Priority)))
	}

	if len(item.Tags) > 0 {
		event.Props.SetText("CATEGORIES", strings.Join(item.Tags, ","))
	}

	if item.Recurrence != nil {
		rrule := w.formatRRule(item.Recurrence)
		event.Props.SetText("RRULE", rrule)

		if len(item.Recurrence.ExDates) > 0 {
			w.setExDates(event.Props, item.Recurrence.ExDates, item.Timezone)
		}

		if len(item.Recurrence.RDates) > 0 {
			w.setRDates(event.Props, item.Recurrence.RDates, item.Timezone)
		}
	}

	for _, reminder := range item.Reminders {
		alarm := w.writeAlarm(&reminder)
		if alarm != nil {
			event.Children = append(event.Children, alarm)
		}
	}

	return event
}

func (w *Writer) writeTodo(item *model.CalendarItem) *ical.Component {
	todo := ical.NewComponent("VTODO")

	todo.Props.SetText("UID", item.UID)
	todo.Props.SetText("SUMMARY", item.Title)
	todo.Props.SetDateTime("DTSTAMP", time.Now().UTC())

	if item.Description != "" {
		todo.Props.SetText("DESCRIPTION", item.Description)
	}

	if item.DueDate != nil {
		w.setDateTime(todo.Props, "DUE", *item.DueDate, false, item.Timezone)
	}

	if item.StartTime != nil {
		w.setDateTime(todo.Props, "DTSTART", *item.StartTime, false, item.Timezone)
	}

	if item.Status != model.StatusPending {
		todo.Props.SetText("STATUS", exportStatus(item.Status))
	}

	if item.CompletionDate != nil {
		w.setDateTime(todo.Props, "COMPLETED", *item.CompletionDate, false, item.Timezone)
		todo.Props.SetText("PERCENT-COMPLETE", "100")
	} else if item.Status == model.StatusCompleted {
		todo.Props.SetText("PERCENT-COMPLETE", "100")
	}

	if item.Priority > 0 {
		todo.Props.SetText("PRIORITY", fmt.Sprintf("%d", exportPriority(item.Priority)))
	}

	if len(item.Tags) > 0 {
		todo.Props.SetText("CATEGORIES", strings.Join(item.Tags, ","))
	}

	if item.Recurrence != nil {
		rrule := w.formatRRule(item.Recurrence)
		todo.Props.SetText("RRULE", rrule)
	}

	for _, reminder := range item.Reminders {
		alarm := w.writeAlarm(&reminder)
		if alarm != nil {
			todo.Children = append(todo.Children, alarm)
		}
	}

	return todo
}

func (w *Writer) writeJournal(item *model.CalendarItem) *ical.Component {
	journal := ical.NewComponent("VJOURNAL")

	journal.Props.SetText("UID", item.UID)
	journal.Props.SetText("SUMMARY", item.Title)

	if item.Description != "" {
		journal.Props.SetText("DESCRIPTION", item.Description)
	}

	if item.StartTime != nil {
		w.setDateTime(journal.Props, "DTSTART", *item.StartTime, false, item.Timezone)
	}

	return journal
}

func (w *Writer) writeAlarm(reminder *model.Reminder) *ical.Component {
	alarm := ical.NewComponent("VALARM")
	alarm.Props.SetText("ACTION", "DISPLAY")
	alarm.Props.SetText("DESCRIPTION", "Reminder")

	if reminder.Offset != nil {
		alarm.Props.SetText("TRIGGER", formatDuration(*reminder.Offset))
	} else if reminder.AbsoluteTime != nil {
		alarm.Props.SetDateTime("TRIGGER", *reminder.AbsoluteTime)
	} else {
		return nil
	}

	return alarm
}

func (w *Writer) setDateTime(props ical.Props, propName string, t time.Time, isAllDay bool, tz string) {
	if isAllDay {
		props.SetText(propName, t.Format("20060102"))
		props.Get(propName).Params["VALUE"] = []string{"DATE"}
	} else if tz != "" && tz != "UTC" {
		props.SetText(propName, t.Format("20060102T150405"))
		props.Get(propName).Params["TZID"] = []string{tz}
	} else {
		props.SetText(propName, t.UTC().Format("20060102T150405Z"))
	}
}

func (w *Writer) setExDates(props ical.Props, dates []time.Time, tz string) {
	var formatted []string
	for _, d := range dates {
		if tz != "" && tz != "UTC" {
			formatted = append(formatted, d.Format("20060102T150405"))
		} else {
			formatted = append(formatted, d.UTC().Format("20060102T150405Z"))
		}
	}
	props.SetText("EXDATE", strings.Join(formatted, ","))
	if tz != "" && tz != "UTC" {
		props.Get("EXDATE").Params["TZID"] = []string{tz}
	}
}

func (w *Writer) setRDates(props ical.Props, dates []time.Time, tz string) {
	var formatted []string
	for _, d := range dates {
		if tz != "" && tz != "UTC" {
			formatted = append(formatted, d.Format("20060102T150405"))
		} else {
			formatted = append(formatted, d.UTC().Format("20060102T150405Z"))
		}
	}
	props.SetText("RDATE", strings.Join(formatted, ","))
	if tz != "" && tz != "UTC" {
		props.Get("RDATE").Params["TZID"] = []string{tz}
	}
}

func (w *Writer) formatRRule(rec *model.Recurrence) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("FREQ=%s", rec.Freq))

	if rec.Interval > 1 {
		parts = append(parts, fmt.Sprintf("INTERVAL=%d", rec.Interval))
	}

	if rec.Count != nil {
		parts = append(parts, fmt.Sprintf("COUNT=%d", *rec.Count))
	}

	if rec.Until != nil {
		parts = append(parts, fmt.Sprintf("UNTIL=%s", rec.Until.UTC().Format("20060102T150405Z")))
	}

	if len(rec.ByDay) > 0 {
		var days []string
		for _, day := range rec.ByDay {
			days = append(days, string(day))
		}
		parts = append(parts, fmt.Sprintf("BYDAY=%s", strings.Join(days, ",")))
	}

	if len(rec.ByMonth) > 0 {
		var months []string
		for _, month := range rec.ByMonth {
			months = append(months, fmt.Sprintf("%d", month))
		}
		parts = append(parts, fmt.Sprintf("BYMONTH=%s", strings.Join(months, ",")))
	}

	if len(rec.ByMonthDay) > 0 {
		var days []string
		for _, day := range rec.ByMonthDay {
			days = append(days, fmt.Sprintf("%d", day))
		}
		parts = append(parts, fmt.Sprintf("BYMONTHDAY=%s", strings.Join(days, ",")))
	}

	if len(rec.BySetPos) > 0 {
		var positions []string
		for _, pos := range rec.BySetPos {
			positions = append(positions, fmt.Sprintf("%d", pos))
		}
		parts = append(parts, fmt.Sprintf("BYSETPOS=%s", strings.Join(positions, ",")))
	}

	return strings.Join(parts, ";")
}

func formatDuration(d time.Duration) string {
	isNegative := d < 0
	if isNegative {
		d = -d
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	days := hours / 24
	hours = hours % 24

	var parts []string
	if isNegative {
		parts = append(parts, "-P")
	} else {
		parts = append(parts, "P")
	}

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dD", days))
	}

	if hours > 0 || minutes > 0 || seconds > 0 {
		parts = append(parts, "T")
		if hours > 0 {
			parts = append(parts, fmt.Sprintf("%dH", hours))
		}
		if minutes > 0 {
			parts = append(parts, fmt.Sprintf("%dM", minutes))
		}
		if seconds > 0 {
			parts = append(parts, fmt.Sprintf("%dS", seconds))
		}
	}

	if len(parts) == 1 {
		return "PT0S"
	}

	return strings.Join(parts, "")
}

func exportPriority(p model.Priority) int {
	switch p {
	case model.PriorityHighest:
		return 1
	case model.PriorityHigh:
		return 3
	case model.PriorityMedium:
		return 5
	case model.PriorityLow:
		return 7
	case model.PriorityLowest:
		return 9
	default:
		return 0
	}
}

func exportStatus(s model.Status) string {
	switch s {
	case model.StatusCompleted:
		return "COMPLETED"
	case model.StatusInProgress:
		return "IN-PROCESS"
	case model.StatusCancelled:
		return "CANCELLED"
	default:
		return "NEEDS-ACTION"
	}
}
