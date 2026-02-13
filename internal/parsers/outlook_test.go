package parsers

import (
	"context"
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestOutlook12HourTime(t *testing.T) {
	csv := `Subject,Start Date,Start Time,End Date,End Time,All day event,Description,Location,Categories,Priority
Meeting,1/15/2024,3:04:05 PM,1/15/2024,4:04:05 PM,False,Sync call,Office,,Normal`

	p := NewOutlookParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Title != "Meeting" {
		t.Errorf("title = %q", item.Title)
	}
	if item.StartTime == nil {
		t.Fatal("start_time is nil")
	}
	if item.StartTime.Hour() != 15 || item.StartTime.Minute() != 4 {
		t.Errorf("start_time = %v, want 15:04", item.StartTime)
	}
	if item.EndTime == nil {
		t.Fatal("end_time is nil")
	}
	if item.EndTime.Hour() != 16 || item.EndTime.Minute() != 4 {
		t.Errorf("end_time = %v, want 16:04", item.EndTime)
	}
	if item.Location != "Office" {
		t.Errorf("location = %q", item.Location)
	}
}

func TestOutlook24HourTime(t *testing.T) {
	csv := `Subject,Start Date,Start Time,End Date,End Time,All day event
Standup,1/10/2024,15:04:05,1/10/2024,15:34:05,False`

	p := NewOutlookParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.StartTime == nil {
		t.Fatal("start_time is nil")
	}
	if item.StartTime.Hour() != 15 || item.StartTime.Minute() != 4 {
		t.Errorf("start_time = %v, want 15:04", item.StartTime)
	}
}

func TestOutlookCategories(t *testing.T) {
	csv := `Subject,Categories
Team Event,"work; meeting; important"`

	p := NewOutlookParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if len(item.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d: %v", len(item.Tags), item.Tags)
	}
	if item.Tags[0] != "work" || item.Tags[1] != "meeting" || item.Tags[2] != "important" {
		t.Errorf("tags = %v", item.Tags)
	}
}

func TestOutlookPriorityMapping(t *testing.T) {
	csv := `Subject,Priority
High Task,High
Normal Task,Normal
Low Task,Low`

	p := NewOutlookParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Items[0].Priority != model.PriorityHigh {
		t.Errorf("high priority = %d, want %d", col.Items[0].Priority, model.PriorityHigh)
	}
	if col.Items[1].Priority != model.PriorityMedium {
		t.Errorf("normal priority = %d, want %d", col.Items[1].Priority, model.PriorityMedium)
	}
	if col.Items[2].Priority != model.PriorityLow {
		t.Errorf("low priority = %d, want %d", col.Items[2].Priority, model.PriorityLow)
	}
}

func TestOutlookAllDayEvent(t *testing.T) {
	csv := `Subject,Start Date,All day event
Holiday,1/25/2024,True`

	p := NewOutlookParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if !item.IsAllDay {
		t.Error("expected all day event")
	}
	if item.StartTime == nil {
		t.Fatal("start_time is nil for all-day event")
	}
}

func TestOutlookMissingSubjectError(t *testing.T) {
	csv := `Description,Location
Something,Somewhere`

	p := NewOutlookParser()
	_, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err == nil {
		t.Fatal("expected error for missing Subject column")
	}
	if !strings.Contains(err.Error(), "Subject") {
		t.Errorf("error should mention Subject: %v", err)
	}
}

func TestOutlookSourceApp(t *testing.T) {
	csv := "Subject\nTest\n"
	p := NewOutlookParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "outlook" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "outlook")
	}
}
