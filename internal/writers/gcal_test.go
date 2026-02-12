package writers

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/parsers"
)

func TestGCalWriterRoundtrip(t *testing.T) {
	input := `Subject,Start Date,Start Time,End Date,End Time,All Day Event,Description,Location
Meeting,01/15/2024,2:00 PM,01/15/2024,3:00 PM,False,Team sync,Room 101`

	p := parsers.NewGoogleCalendarParser()
	col, err := p.Parse(strings.NewReader(input), "test.csv")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var buf bytes.Buffer
	w := NewGoogleCalendarWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Meeting") {
		t.Error("output should contain title")
	}
	if !strings.Contains(output, "Room 101") {
		t.Error("output should contain location")
	}
	if !strings.Contains(output, "Team sync") {
		t.Error("output should contain description")
	}
}

func TestGCalWriterAllDayDateFormat(t *testing.T) {
	start := time.Date(2024, 1, 25, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 26, 0, 0, 0, 0, time.UTC)
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:     "Holiday",
				StartTime: &start,
				EndTime:   &end,
				IsAllDay:  true,
			},
		},
	}

	var buf bytes.Buffer
	w := NewGoogleCalendarWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()
	if len(records) < 2 {
		t.Fatal("expected header + data row")
	}

	row := records[1]
	// Start Date = 01/25/2024
	if row[1] != "01/25/2024" {
		t.Errorf("start date = %q, want %q", row[1], "01/25/2024")
	}
	// Start Time should be empty for all-day
	if row[2] != "" {
		t.Errorf("start time should be empty for all-day, got %q", row[2])
	}
	// All Day Event = True
	if row[5] != "True" {
		t.Errorf("all day event = %q, want %q", row[5], "True")
	}
}

func TestGCalWriterSubtaskFlattening(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:       "Parent",
				Description: "Main desc",
				Subtasks: []model.Subtask{
					{Title: "Sub 1", Status: model.StatusCompleted},
					{Title: "Sub 2", Status: model.StatusPending},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewGoogleCalendarWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()

	// Description should be in column 6
	desc := records[1][6]
	if !strings.Contains(desc, "Main desc") {
		t.Error("description should contain original text")
	}
	if !strings.Contains(desc, "- [x] Sub 1") {
		t.Error("description should contain completed subtask")
	}
	if !strings.Contains(desc, "- [ ] Sub 2") {
		t.Error("description should contain pending subtask")
	}
}

func TestGCalWriterTimedEvent(t *testing.T) {
	start := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 15, 30, 0, 0, time.UTC)
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:     "Meeting",
				StartTime: &start,
				EndTime:   &end,
				IsAllDay:  false,
			},
		},
	}

	var buf bytes.Buffer
	w := NewGoogleCalendarWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()
	row := records[1]
	if row[2] == "" {
		t.Error("start time should not be empty for timed event")
	}
	if row[5] != "False" {
		t.Errorf("all day event = %q, want %q", row[5], "False")
	}
}

func TestGCalWriterHeader(t *testing.T) {
	col := &model.CalendarCollection{Items: []model.CalendarItem{}}

	var buf bytes.Buffer
	w := NewGoogleCalendarWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()
	expected := []string{"Subject", "Start Date", "Start Time", "End Date", "End Time", "All Day Event", "Description", "Location", "Private"}
	for i, h := range expected {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}
}
