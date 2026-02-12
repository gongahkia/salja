package ics

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/gongahkia/calendar-converter/internal/model"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ICS file: %w", err)
	}
	defer f.Close()

	return p.Parse(f, filePath)
}

func (p *Parser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	dec := ical.NewDecoder(r)
	
	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "ics",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	for {
		cal, err := dec.Decode()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode ICS: %w", err)
		}

		for _, comp := range cal.Children {
			switch comp.Name {
			case ical.CompEvent:
				item, err := p.parseEvent(comp)
				if err != nil {
					return nil, err
				}
				collection.Items = append(collection.Items, *item)
			case ical.CompToDo:
				item, err := p.parseTodo(comp)
				if err != nil {
					return nil, err
				}
				collection.Items = append(collection.Items, *item)
			case ical.CompJournal:
				item, err := p.parseJournal(comp)
				if err != nil {
					return nil, err
				}
				collection.Items = append(collection.Items, *item)
			}
		}
	}

	return collection, nil
}

func (p *Parser) parseEvent(comp *ical.Component) (*model.CalendarItem, error) {
	item := &model.CalendarItem{
		ItemType: model.ItemTypeEvent,
		Status:   model.StatusPending,
	}

	if uid, err := comp.Props.Text("UID"); err == nil {
		item.UID = uid
	}
	if summary, err := comp.Props.Text("SUMMARY"); err == nil {
		item.Title = summary
	}
	if desc, err := comp.Props.Text("DESCRIPTION"); err == nil {
		item.Description = desc
	}
	if loc, err := comp.Props.Text("LOCATION"); err == nil {
		item.Location = loc
	}
	
	if dtstart := comp.Props.Get("DTSTART"); dtstart != nil {
		dt, isAllDay, tz := parseDateTime(dtstart)
		item.StartTime = &dt
		item.IsAllDay = isAllDay
		if tz != "" {
			item.Timezone = tz
		}
	}
	
	if dtend := comp.Props.Get("DTEND"); dtend != nil {
		dt, _, _ := parseDateTime(dtend)
		item.EndTime = &dt
	}
	
	if rrule := comp.Props.Get("RRULE"); rrule != nil {
		rec, err := parseRRule(rrule.Value)
		if err != nil {
			return nil, err
		}
		item.Recurrence = rec
	}
	
	for _, exdate := range comp.Props.Values("EXDATE") {
		exdates := parseExDate(&exdate)
		if item.Recurrence == nil {
			item.Recurrence = &model.Recurrence{}
		}
		item.Recurrence.ExDates = append(item.Recurrence.ExDates, exdates...)
	}
	
	for _, rdate := range comp.Props.Values("RDATE") {
		rdates := parseRDate(&rdate)
		if item.Recurrence == nil {
			item.Recurrence = &model.Recurrence{}
		}
		item.Recurrence.RDates = append(item.Recurrence.RDates, rdates...)
	}
	
	if categories, err := comp.Props.Text("CATEGORIES"); err == nil {
		item.Tags = parseCategories(categories)
	}
	
	if priority := comp.Props.Get("PRIORITY"); priority != nil {
		item.Priority = parsePriority(priority.Value)
	}

	for _, child := range comp.Children {
		if child.Name == "VALARM" {
			reminder := parseAlarm(child)
			if reminder != nil {
				item.Reminders = append(item.Reminders, *reminder)
			}
		}
	}

	return item, nil
}

func (p *Parser) parseTodo(comp *ical.Component) (*model.CalendarItem, error) {
	item := &model.CalendarItem{
		ItemType: model.ItemTypeTask,
		Status:   model.StatusPending,
	}

	if uid, err := comp.Props.Text("UID"); err == nil {
		item.UID = uid
	}
	if summary, err := comp.Props.Text("SUMMARY"); err == nil {
		item.Title = summary
	}
	if desc, err := comp.Props.Text("DESCRIPTION"); err == nil {
		item.Description = desc
	}
	
	if due := comp.Props.Get("DUE"); due != nil {
		dt, _, tz := parseDateTime(due)
		item.DueDate = &dt
		if tz != "" {
			item.Timezone = tz
		}
	}
	
	if dtstart := comp.Props.Get("DTSTART"); dtstart != nil {
		dt, _, tz := parseDateTime(dtstart)
		item.StartTime = &dt
		if tz != "" {
			item.Timezone = tz
		}
	}
	
	if status, err := comp.Props.Text("STATUS"); err == nil {
		item.Status = parseStatus(status)
	}
	
	if completed := comp.Props.Get("COMPLETED"); completed != nil {
		dt, _, _ := parseDateTime(completed)
		item.CompletionDate = &dt
	}
	
	if percent, err := comp.Props.Text("PERCENT-COMPLETE"); err == nil {
		if percent == "100" && item.Status != model.StatusCompleted {
			item.Status = model.StatusCompleted
		}
	}
	
	if priority := comp.Props.Get("PRIORITY"); priority != nil {
		item.Priority = parsePriority(priority.Value)
	}
	
	if categories, err := comp.Props.Text("CATEGORIES"); err == nil {
		item.Tags = parseCategories(categories)
	}
	
	if rrule := comp.Props.Get("RRULE"); rrule != nil {
		rec, err := parseRRule(rrule.Value)
		if err != nil {
			return nil, err
		}
		item.Recurrence = rec
	}

	for _, child := range comp.Children {
		if child.Name == "VALARM" {
			reminder := parseAlarm(child)
			if reminder != nil {
				item.Reminders = append(item.Reminders, *reminder)
			}
		}
	}

	return item, nil
}

func (p *Parser) parseJournal(comp *ical.Component) (*model.CalendarItem, error) {
	item := &model.CalendarItem{
		ItemType: model.ItemTypeJournal,
		Status:   model.StatusPending,
	}

	if uid, err := comp.Props.Text("UID"); err == nil {
		item.UID = uid
	}
	if summary, err := comp.Props.Text("SUMMARY"); err == nil {
		item.Title = summary
	}
	if desc, err := comp.Props.Text("DESCRIPTION"); err == nil {
		item.Description = desc
	}
	
	if dtstart := comp.Props.Get("DTSTART"); dtstart != nil {
		dt, _, tz := parseDateTime(dtstart)
		item.StartTime = &dt
		if tz != "" {
			item.Timezone = tz
		}
	}

	return item, nil
}

func parseDateTime(prop *ical.Prop) (time.Time, bool, string) {
	isAllDay := false
	tz := ""

	if valueType, ok := prop.Params[ical.ParamValue]; ok && len(valueType) > 0 && valueType[0] == "DATE" {
		isAllDay = true
	}

	if tzid, ok := prop.Params[ical.ParamTimezoneID]; ok && len(tzid) > 0 {
		tz = tzid[0]
	}

	var t time.Time
	var err error

	if isAllDay {
		t, err = time.Parse("20060102", prop.Value)
	} else if strings.HasSuffix(prop.Value, "Z") {
		t, err = time.Parse("20060102T150405Z", prop.Value)
	} else {
		t, err = time.Parse("20060102T150405", prop.Value)
		if err == nil && tz != "" {
			loc, locErr := time.LoadLocation(tz)
			if locErr == nil {
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
			}
		}
	}

	if err != nil {
		return time.Time{}, false, ""
	}

	return t, isAllDay, tz
}

func parseRRule(value string) (*model.Recurrence, error) {
	rec := &model.Recurrence{
		Interval: 1,
	}

	parts := strings.Split(value, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		val := kv[1]

		switch key {
		case "FREQ":
			rec.Freq = model.FreqType(val)
		case "INTERVAL":
			fmt.Sscanf(val, "%d", &rec.Interval)
		case "COUNT":
			var count int
			fmt.Sscanf(val, "%d", &count)
			rec.Count = &count
		case "UNTIL":
			until, _ := time.Parse("20060102T150405Z", val)
			rec.Until = &until
		case "BYDAY":
			days := strings.Split(val, ",")
			for _, day := range days {
				rec.ByDay = append(rec.ByDay, model.Weekday(strings.TrimSpace(day)))
			}
		case "BYMONTH":
			months := strings.Split(val, ",")
			for _, month := range months {
				var m int
				fmt.Sscanf(month, "%d", &m)
				rec.ByMonth = append(rec.ByMonth, m)
			}
		case "BYMONTHDAY":
			days := strings.Split(val, ",")
			for _, day := range days {
				var d int
				fmt.Sscanf(day, "%d", &d)
				rec.ByMonthDay = append(rec.ByMonthDay, d)
			}
		case "BYSETPOS":
			positions := strings.Split(val, ",")
			for _, pos := range positions {
				var p int
				fmt.Sscanf(pos, "%d", &p)
				rec.BySetPos = append(rec.BySetPos, p)
			}
		}
	}

	return rec, nil
}

func parseExDate(prop *ical.Prop) []time.Time {
	var dates []time.Time
	values := strings.Split(prop.Value, ",")
	for _, val := range values {
		t, err := time.Parse("20060102T150405Z", val)
		if err != nil {
			t, err = time.Parse("20060102", val)
		}
		if err == nil {
			dates = append(dates, t)
		}
	}
	return dates
}

func parseRDate(prop *ical.Prop) []time.Time {
	return parseExDate(prop)
}

func parseCategories(value string) []string {
	return strings.Split(value, ",")
}

func parsePriority(value string) model.Priority {
	var p int
	fmt.Sscanf(value, "%d", &p)
	
	if p == 0 {
		return model.PriorityNone
	} else if p >= 1 && p <= 2 {
		return model.PriorityHighest
	} else if p >= 3 && p <= 4 {
		return model.PriorityHigh
	} else if p == 5 {
		return model.PriorityMedium
	} else if p >= 6 && p <= 7 {
		return model.PriorityLow
	} else if p >= 8 && p <= 9 {
		return model.PriorityLowest
	}
	return model.PriorityNone
}

func parseStatus(value string) model.Status {
	switch strings.ToUpper(value) {
	case "COMPLETED":
		return model.StatusCompleted
	case "IN-PROCESS":
		return model.StatusInProgress
	case "CANCELLED":
		return model.StatusCancelled
	default:
		return model.StatusPending
	}
}

func parseAlarm(comp *ical.Component) *model.Reminder {
	reminder := &model.Reminder{}
	
	if trigger := comp.Props.Get("TRIGGER"); trigger != nil {
		if strings.HasPrefix(trigger.Value, "-") || strings.HasPrefix(trigger.Value, "+") || strings.HasPrefix(trigger.Value, "P") {
			duration, err := parseDuration(trigger.Value)
			if err == nil {
				reminder.Offset = &duration
			}
		} else {
			t, err := time.Parse("20060102T150405Z", trigger.Value)
			if err == nil {
				reminder.AbsoluteTime = &t
			}
		}
	}
	
	if reminder.Offset == nil && reminder.AbsoluteTime == nil {
		return nil
	}
	
	return reminder
}

func parseDuration(value string) (time.Duration, error) {
	value = strings.TrimPrefix(value, "-")
	value = strings.TrimPrefix(value, "+")
	
	isNegative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(value, "P")
	
	var duration time.Duration
	
	if strings.Contains(value, "W") {
		var weeks int
		fmt.Sscanf(value, "%dW", &weeks)
		duration = time.Duration(weeks) * 7 * 24 * time.Hour
	} else {
		parts := strings.Split(value, "T")
		if len(parts) > 0 {
			datePart := parts[0]
			var days int
			fmt.Sscanf(datePart, "%dD", &days)
			duration += time.Duration(days) * 24 * time.Hour
		}
		if len(parts) > 1 {
			timePart := parts[1]
			var hours, minutes, seconds int
			fmt.Sscanf(timePart, "%dH%dM%dS", &hours, &minutes, &seconds)
			duration += time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
		}
	}
	
	if isNegative {
		duration = -duration
	}
	
	return duration, nil
}
