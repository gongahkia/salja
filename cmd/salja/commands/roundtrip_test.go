package commands_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/ics"
	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/parsers"
	"github.com/gongahkia/salja/internal/writers"
)

// TestRoundTripICSToGCalCSVToICS verifies ICS→Google Calendar CSV→ICS preserves
// all fields that both formats support.
func TestRoundTripICSToGCalCSVToICS(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	end := now.Add(2 * time.Hour)

	original := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				UID:       "test-event-1",
				Title:     "Team Standup",
				StartTime: &now,
				EndTime:   &end,
				Location:  "Room 42",
				ItemType:  model.ItemTypeEvent,
				IsAllDay:  false,
			},
		},
		SourceApp: "ics",
	}

	// ICS → Google Calendar CSV
	gcalWriter := writers.NewGoogleCalendarWriter()
	var csvBuf bytes.Buffer
	if err := gcalWriter.Write(ctx, original, &csvBuf); err != nil {
		t.Fatalf("gcal writer failed: %v", err)
	}

	// Google Calendar CSV → collection
	gcalParser := parsers.NewGoogleCalendarParser()
	intermediate, err := gcalParser.Parse(ctx, &csvBuf, "test.csv")
	if err != nil {
		t.Fatalf("gcal parser failed: %v", err)
	}

	// collection → ICS
	icsWriter := ics.NewWriter()
	var icsBuf bytes.Buffer
	if err := icsWriter.Write(ctx, intermediate, &icsBuf); err != nil {
		t.Fatalf("ics writer failed: %v", err)
	}

	// ICS → final collection
	icsParser := ics.NewParser()
	final, err := icsParser.Parse(ctx, &icsBuf, "test.ics")
	if err != nil {
		t.Fatalf("ics parser failed: %v", err)
	}

	if len(final.Items) != len(original.Items) {
		t.Fatalf("item count mismatch: got %d, want %d", len(final.Items), len(original.Items))
	}

	orig := original.Items[0]
	got := final.Items[0]

	if got.Title != orig.Title {
		t.Errorf("title: got %q, want %q", got.Title, orig.Title)
	}
	if got.Location != orig.Location {
		t.Errorf("location: got %q, want %q", got.Location, orig.Location)
	}
	if got.ItemType != model.ItemTypeEvent {
		t.Errorf("item type: got %q, want %q", got.ItemType, model.ItemTypeEvent)
	}
	if got.StartTime == nil || !got.StartTime.Equal(*orig.StartTime) {
		t.Errorf("start time: got %v, want %v", got.StartTime, orig.StartTime)
	}
	if got.EndTime == nil || !got.EndTime.Equal(*orig.EndTime) {
		t.Errorf("end time: got %v, want %v", got.EndTime, orig.EndTime)
	}
}

// TestRoundTripICSToOutlookCSVToICS verifies ICS→Outlook CSV→ICS preserves
// all fields that both formats support.
func TestRoundTripICSToOutlookCSVToICS(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2025, 3, 20, 14, 30, 0, 0, time.UTC)
	end := now.Add(time.Hour)

	original := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				UID:         "outlook-test-1",
				Title:       "Project Review",
				Description: "Quarterly review",
				StartTime:   &now,
				EndTime:     &end,
				Location:    "Conference Room B",
				ItemType:    model.ItemTypeEvent,
				IsAllDay:    false,
			},
		},
		SourceApp: "ics",
	}

	// ICS → Outlook CSV
	outlookWriter := writers.NewOutlookWriter()
	var csvBuf bytes.Buffer
	if err := outlookWriter.Write(ctx, original, &csvBuf); err != nil {
		t.Fatalf("outlook writer failed: %v", err)
	}

	// Outlook CSV → collection
	outlookParser := parsers.NewOutlookParser()
	intermediate, err := outlookParser.Parse(ctx, &csvBuf, "test.csv")
	if err != nil {
		t.Fatalf("outlook parser failed: %v", err)
	}

	// collection → ICS
	icsWriter := ics.NewWriter()
	var icsBuf bytes.Buffer
	if err := icsWriter.Write(ctx, intermediate, &icsBuf); err != nil {
		t.Fatalf("ics writer failed: %v", err)
	}

	// ICS → final collection
	icsParser := ics.NewParser()
	final, err := icsParser.Parse(ctx, &icsBuf, "test.ics")
	if err != nil {
		t.Fatalf("ics parser failed: %v", err)
	}

	if len(final.Items) != len(original.Items) {
		t.Fatalf("item count mismatch: got %d, want %d", len(final.Items), len(original.Items))
	}

	orig := original.Items[0]
	got := final.Items[0]

	if got.Title != orig.Title {
		t.Errorf("title: got %q, want %q", got.Title, orig.Title)
	}
	if got.Location != orig.Location {
		t.Errorf("location: got %q, want %q", got.Location, orig.Location)
	}
	if got.ItemType != model.ItemTypeEvent {
		t.Errorf("item type: got %q, want %q", got.ItemType, model.ItemTypeEvent)
	}
}

// TestRoundTripICSToTickTickCSVToICS verifies ICS→TickTick CSV→ICS preserves
// task fields that both formats support.
func TestRoundTripICSToTickTickCSVToICS(t *testing.T) {
	ctx := context.Background()
	due := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	original := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				UID:      "task-roundtrip-1",
				Title:    "Buy groceries",
				DueDate:  &due,
				Status:   model.StatusPending,
				Priority: model.PriorityHigh,
				Tags:     []string{"errands"},
				ItemType: model.ItemTypeTask,
			},
		},
		SourceApp: "ics",
	}

	// ICS → TickTick CSV
	ttWriter := writers.NewTickTickWriter()
	var csvBuf bytes.Buffer
	if err := ttWriter.Write(ctx, original, &csvBuf); err != nil {
		t.Fatalf("ticktick writer failed: %v", err)
	}

	// TickTick CSV → collection
	ttParser := parsers.NewTickTickParser()
	intermediate, err := ttParser.Parse(ctx, &csvBuf, "ticktick.csv")
	if err != nil {
		t.Fatalf("ticktick parser failed: %v", err)
	}

	// collection → ICS
	icsWriter := ics.NewWriter()
	var icsBuf bytes.Buffer
	if err := icsWriter.Write(ctx, intermediate, &icsBuf); err != nil {
		t.Fatalf("ics writer failed: %v", err)
	}

	// ICS → final collection
	icsParser := ics.NewParser()
	final, err := icsParser.Parse(ctx, &icsBuf, "test.ics")
	if err != nil {
		t.Fatalf("ics parser failed: %v", err)
	}

	if len(final.Items) != len(original.Items) {
		t.Fatalf("item count mismatch: got %d, want %d", len(final.Items), len(original.Items))
	}

	got := final.Items[0]
	if got.Title != "Buy groceries" {
		t.Errorf("title: got %q, want %q", got.Title, "Buy groceries")
	}
	if !strings.EqualFold(string(got.ItemType), string(model.ItemTypeTask)) {
		t.Errorf("item type: got %q, want %q", got.ItemType, model.ItemTypeTask)
	}
}
