package parsers

import (
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestGCalAllDayEvent(t *testing.T) {
	csv := `Subject,Start Date,Start Time,End Date,End Time,All Day Event,Description,Location
Holiday,01/25/2024,,01/26/2024,,True,Day off,`

	p := NewGoogleCalendarParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Items))
	}
	item := col.Items[0]
	if item.Title != "Holiday" {
		t.Errorf("title = %q, want %q", item.Title, "Holiday")
	}
	if !item.IsAllDay {
		t.Error("expected all day event")
	}
	if item.StartTime == nil {
		t.Fatal("start_time is nil")
	}
	if item.StartTime.Format("01/02/2006") != "01/25/2024" {
		t.Errorf("start_date = %v", item.StartTime)
	}
	if item.EndTime == nil {
		t.Fatal("end_time is nil")
	}
	if item.EndTime.Format("01/02/2006") != "01/26/2024" {
		t.Errorf("end_date = %v", item.EndTime)
	}
	if item.Description != "Day off" {
		t.Errorf("description = %q", item.Description)
	}
	if item.ItemType != model.ItemTypeEvent {
		t.Errorf("item_type = %q, want %q", item.ItemType, model.ItemTypeEvent)
	}
}

func TestGCalTimedEvent(t *testing.T) {
	csv := `Subject,Start Date,Start Time,End Date,End Time,All Day Event,Description,Location
Meeting,01/15/2024,2:00 PM,01/15/2024,3:00 PM,False,Team sync,Room 101`

	p := NewGoogleCalendarParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Title != "Meeting" {
		t.Errorf("title = %q", item.Title)
	}
	if item.IsAllDay {
		t.Error("should not be all day")
	}
	if item.StartTime == nil {
		t.Fatal("start_time is nil")
	}
	if item.StartTime.Hour() != 14 || item.StartTime.Minute() != 0 {
		t.Errorf("start_time = %v, want 14:00", item.StartTime)
	}
	if item.EndTime == nil {
		t.Fatal("end_time is nil")
	}
	if item.EndTime.Hour() != 15 || item.EndTime.Minute() != 0 {
		t.Errorf("end_time = %v, want 15:00", item.EndTime)
	}
	if item.Location != "Room 101" {
		t.Errorf("location = %q", item.Location)
	}
}

func TestGCalMissingOptionalColumns(t *testing.T) {
	csv := `Subject
Just a title`

	p := NewGoogleCalendarParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Items))
	}
	item := col.Items[0]
	if item.Title != "Just a title" {
		t.Errorf("title = %q", item.Title)
	}
	if item.StartTime != nil {
		t.Error("start_time should be nil")
	}
	if item.EndTime != nil {
		t.Error("end_time should be nil")
	}
}

func TestGCalMissingSubjectError(t *testing.T) {
	csv := `Description,Location
Some desc,Somewhere`

	p := NewGoogleCalendarParser()
	_, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err == nil {
		t.Fatal("expected error for missing Subject column")
	}
	if !strings.Contains(err.Error(), "Subject") {
		t.Errorf("error should mention Subject: %v", err)
	}
}

func TestGCalSourceApp(t *testing.T) {
	csv := "Subject\nTest\n"
	p := NewGoogleCalendarParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "gcal" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "gcal")
	}
}

func TestGCalEmptyInput(t *testing.T) {
	p := NewGoogleCalendarParser()
	col, err := p.Parse(strings.NewReader(""), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(col.Items))
	}
}
