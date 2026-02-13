package ics

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

func TestParseSingleEvent(t *testing.T) {
	icsData := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:test-event-1
SUMMARY:Team Meeting
DESCRIPTION:Weekly team sync
DTSTART:20240115T140000Z
DTEND:20240115T150000Z
LOCATION:Conference Room A
END:VEVENT
END:VCALENDAR`

	parser := NewParser()
	collection, err := parser.Parse(context.Background(), strings.NewReader(icsData), "test.ics")

	if err != nil {
		t.Fatalf("Failed to parse ICS: %v", err)
	}

	if len(collection.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(collection.Items))
	}

	item := collection.Items[0]
	if item.Title != "Team Meeting" {
		t.Errorf("Expected title 'Team Meeting', got '%s'", item.Title)
	}
	if item.Description != "Weekly team sync" {
		t.Errorf("Expected description 'Weekly team sync', got '%s'", item.Description)
	}
	if item.Location != "Conference Room A" {
		t.Errorf("Expected location 'Conference Room A', got '%s'", item.Location)
	}
	if item.ItemType != model.ItemTypeEvent {
		t.Errorf("Expected ItemType event, got %s", item.ItemType)
	}
}

func TestParseRecurringEventWithExDate(t *testing.T) {
	icsData := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:recurring-1
SUMMARY:Daily Standup
DTSTART:20240101T090000Z
DTEND:20240101T091500Z
RRULE:FREQ=DAILY;COUNT=10
EXDATE:20240103T090000Z,20240105T090000Z
END:VEVENT
END:VCALENDAR`

	parser := NewParser()
	collection, err := parser.Parse(context.Background(), strings.NewReader(icsData), "test.ics")

	if err != nil {
		t.Fatalf("Failed to parse ICS: %v", err)
	}

	item := collection.Items[0]
	if item.Recurrence == nil {
		t.Fatal("Expected recurrence, got nil")
	}
	if item.Recurrence.Freq != model.FreqDaily {
		t.Errorf("Expected DAILY frequency, got %s", item.Recurrence.Freq)
	}
	if *item.Recurrence.Count != 10 {
		t.Errorf("Expected count 10, got %d", *item.Recurrence.Count)
	}
	if len(item.Recurrence.ExDates) != 2 {
		t.Errorf("Expected 2 ExDates, got %d", len(item.Recurrence.ExDates))
	}
}

func TestParseVTODOWithDue(t *testing.T) {
	icsData := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VTODO
UID:task-1
SUMMARY:Complete project proposal
DUE:20240120T170000Z
PRIORITY:1
STATUS:IN-PROCESS
PERCENT-COMPLETE:50
END:VTODO
END:VCALENDAR`

	parser := NewParser()
	collection, err := parser.Parse(context.Background(), strings.NewReader(icsData), "test.ics")

	if err != nil {
		t.Fatalf("Failed to parse ICS: %v", err)
	}

	item := collection.Items[0]
	if item.ItemType != model.ItemTypeTask {
		t.Errorf("Expected ItemType task, got %s", item.ItemType)
	}
	if item.Title != "Complete project proposal" {
		t.Errorf("Expected title 'Complete project proposal', got '%s'", item.Title)
	}
	if item.DueDate == nil {
		t.Fatal("Expected DueDate, got nil")
	}
	if item.Status != model.StatusInProgress {
		t.Errorf("Expected status in_progress, got %s", item.Status)
	}
	if item.Priority != model.PriorityHighest {
		t.Errorf("Expected priority highest, got %d", item.Priority)
	}
}

func TestParseAllDayEvent(t *testing.T) {
	icsData := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:allday-1
SUMMARY:Company Holiday
DTSTART;VALUE=DATE:20240704
DTEND;VALUE=DATE:20240705
END:VEVENT
END:VCALENDAR`

	parser := NewParser()
	collection, err := parser.Parse(context.Background(), strings.NewReader(icsData), "test.ics")

	if err != nil {
		t.Fatalf("Failed to parse ICS: %v", err)
	}

	item := collection.Items[0]
	if !item.IsAllDay {
		t.Error("Expected IsAllDay to be true")
	}
	if item.StartTime == nil {
		t.Fatal("Expected StartTime, got nil")
	}
}

func TestParseMultiTimezoneFile(t *testing.T) {
	icsData := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:tz-event-1
SUMMARY:US Meeting
DTSTART;TZID=America/New_York:20240115T140000
DTEND;TZID=America/New_York:20240115T150000
END:VEVENT
BEGIN:VEVENT
UID:tz-event-2
SUMMARY:Tokyo Meeting
DTSTART;TZID=Asia/Tokyo:20240116T100000
DTEND;TZID=Asia/Tokyo:20240116T110000
END:VEVENT
END:VCALENDAR`

	parser := NewParser()
	collection, err := parser.Parse(context.Background(), strings.NewReader(icsData), "test.ics")

	if err != nil {
		t.Fatalf("Failed to parse ICS: %v", err)
	}

	if len(collection.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(collection.Items))
	}

	if collection.Items[0].Timezone != "America/New_York" {
		t.Errorf("Expected timezone America/New_York, got %s", collection.Items[0].Timezone)
	}
	if collection.Items[1].Timezone != "Asia/Tokyo" {
		t.Errorf("Expected timezone Asia/Tokyo, got %s", collection.Items[1].Timezone)
	}
}

func TestParseMalformedInput(t *testing.T) {
	icsData := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
THIS IS NOT VALID
END:VEVENT`

	parser := NewParser()
	_, err := parser.Parse(context.Background(), strings.NewReader(icsData), "test.ics")

	if err == nil {
		t.Error("Expected error for malformed input, got nil")
	}
}

func TestRRuleParserDaily(t *testing.T) {
	rec, err := parseRRule("FREQ=DAILY;INTERVAL=2;COUNT=5")
	if err != nil {
		t.Fatalf("Failed to parse RRULE: %v", err)
	}

	if rec.Freq != model.FreqDaily {
		t.Errorf("Expected DAILY, got %s", rec.Freq)
	}
	if rec.Interval != 2 {
		t.Errorf("Expected interval 2, got %d", rec.Interval)
	}
	if *rec.Count != 5 {
		t.Errorf("Expected count 5, got %d", *rec.Count)
	}
}

func TestRRuleParserWeeklyWithByDay(t *testing.T) {
	rec, err := parseRRule("FREQ=WEEKLY;BYDAY=MO,WE,FR;INTERVAL=1")
	if err != nil {
		t.Fatalf("Failed to parse RRULE: %v", err)
	}

	if rec.Freq != model.FreqWeekly {
		t.Errorf("Expected WEEKLY, got %s", rec.Freq)
	}
	if len(rec.ByDay) != 3 {
		t.Errorf("Expected 3 ByDay values, got %d", len(rec.ByDay))
	}
}

func TestRRuleParserMonthlyWithByMonthDay(t *testing.T) {
	rec, err := parseRRule("FREQ=MONTHLY;BYMONTHDAY=15;COUNT=12")
	if err != nil {
		t.Fatalf("Failed to parse RRULE: %v", err)
	}

	if rec.Freq != model.FreqMonthly {
		t.Errorf("Expected MONTHLY, got %s", rec.Freq)
	}
	if len(rec.ByMonthDay) != 1 || rec.ByMonthDay[0] != 15 {
		t.Errorf("Expected ByMonthDay [15], got %v", rec.ByMonthDay)
	}
}

func TestRRuleParserYearly(t *testing.T) {
	until := "20250101T000000Z"
	rec, err := parseRRule("FREQ=YEARLY;UNTIL=" + until)
	if err != nil {
		t.Fatalf("Failed to parse RRULE: %v", err)
	}

	if rec.Freq != model.FreqYearly {
		t.Errorf("Expected YEARLY, got %s", rec.Freq)
	}
	if rec.Until == nil {
		t.Fatal("Expected Until, got nil")
	}
}

func TestRRuleParserCountLimited(t *testing.T) {
	rec, err := parseRRule("FREQ=DAILY;COUNT=30")
	if err != nil {
		t.Fatalf("Failed to parse RRULE: %v", err)
	}

	if *rec.Count != 30 {
		t.Errorf("Expected count 30, got %d", *rec.Count)
	}
}

func TestRRuleParserUntilLimited(t *testing.T) {
	rec, err := parseRRule("FREQ=WEEKLY;UNTIL=20241231T235959Z")
	if err != nil {
		t.Fatalf("Failed to parse RRULE: %v", err)
	}

	if rec.Until == nil {
		t.Fatal("Expected Until, got nil")
	}
}

func TestRRuleParserComplexBySetPos(t *testing.T) {
	rec, err := parseRRule("FREQ=MONTHLY;BYDAY=FR;BYSETPOS=-1")
	if err != nil {
		t.Fatalf("Failed to parse RRULE: %v", err)
	}

	if len(rec.BySetPos) != 1 || rec.BySetPos[0] != -1 {
		t.Errorf("Expected BySetPos [-1], got %v", rec.BySetPos)
	}
}

func TestWriterRoundTrip(t *testing.T) {
	start := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 15, 0, 0, 0, time.UTC)

	original := model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				UID:       "test-1",
				Title:     "Test Event",
				ItemType:  model.ItemTypeEvent,
				StartTime: &start,
				EndTime:   &end,
				Priority:  model.PriorityHigh,
				Tags:      []string{"work", "meeting"},
			},
		},
		SourceApp: "test",
	}

	writer := NewWriter()
	var buf strings.Builder
	err := writer.Write(context.Background(), &original, &buf)
	if err != nil {
		t.Fatalf("Failed to write ICS: %v", err)
	}

	parser := NewParser()
	parsed, err := parser.Parse(context.Background(), strings.NewReader(buf.String()), "test.ics")
	if err != nil {
		t.Fatalf("Failed to parse written ICS: %v", err)
	}

	if len(parsed.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(parsed.Items))
	}

	item := parsed.Items[0]
	if item.Title != "Test Event" {
		t.Errorf("Expected title 'Test Event', got '%s'", item.Title)
	}
	if len(item.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(item.Tags))
	}
}
