package writers

import (
	"bytes"
	"context"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/parsers"
)

func TestTickTickWriterRoundtrip(t *testing.T) {
	input := `title,tags,content,is_checklist,start_date,due_date,priority,status,timezone,is_all_day,repeat,completed_time
Buy groceries,"food, errands",Milk and eggs,false,2024-01-15T09:00:00Z,2024-01-16T09:00:00Z,5,0,America/New_York,false,DAILY,`

	p := parsers.NewTickTickParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.csv")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var buf bytes.Buffer
	w := NewTickTickWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Buy groceries") {
		t.Error("output should contain title")
	}
	if !strings.Contains(output, "food, errands") {
		t.Error("output should contain tags")
	}
	if !strings.Contains(output, "DAILY") {
		t.Error("output should contain repeat rule")
	}
}

func TestTickTickWriterSubtaskSerialization(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:    "Parent",
				ItemType: model.ItemTypeTask,
				Subtasks: []model.Subtask{
					{Title: "Sub 1", Status: model.StatusCompleted},
					{Title: "Sub 2", Status: model.StatusPending},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewTickTickWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "- [x] Sub 1") {
		t.Error("output should contain completed subtask")
	}
	if !strings.Contains(output, "- [ ] Sub 2") {
		t.Error("output should contain pending subtask")
	}

	// Verify is_checklist is "true"
	reader := csv.NewReader(strings.NewReader(output))
	records, _ := reader.ReadAll()
	// Find is_checklist column index
	header := records[0]
	checklistIdx := -1
	for i, h := range header {
		if h == "is_checklist" {
			checklistIdx = i
			break
		}
	}
	if checklistIdx == -1 {
		t.Fatal("is_checklist column not found")
	}
	if records[1][checklistIdx] != "true" {
		t.Errorf("is_checklist = %q, want %q", records[1][checklistIdx], "true")
	}
}

func TestTickTickWriterPriorityMapping(t *testing.T) {
	tests := []struct {
		priority model.Priority
		want     string
	}{
		{model.PriorityNone, "0"},
		{model.PriorityLowest, "0"},
		{model.PriorityLow, "1"},
		{model.PriorityMedium, "3"},
		{model.PriorityHigh, "5"},
		{model.PriorityHighest, "5"},
	}
	for _, tc := range tests {
		got := exportTickTickPriority(tc.priority)
		if got != tc.want {
			t.Errorf("exportTickTickPriority(%d) = %q, want %q", tc.priority, got, tc.want)
		}
	}
}

func TestTickTickWriterCompletedStatus(t *testing.T) {
	now := time.Now()
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:          "Done Task",
				Status:         model.StatusCompleted,
				CompletionDate: &now,
			},
		},
	}

	var buf bytes.Buffer
	w := NewTickTickWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()
	header := records[0]
	statusIdx := -1
	for i, h := range header {
		if h == "status" {
			statusIdx = i
			break
		}
	}
	if statusIdx >= 0 && records[1][statusIdx] != "2" {
		t.Errorf("status = %q, want %q", records[1][statusIdx], "2")
	}
}

func TestTickTickWriterAllDay(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{Title: "All Day", IsAllDay: true},
			{Title: "Not All Day", IsAllDay: false},
		},
	}

	var buf bytes.Buffer
	w := NewTickTickWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()
	header := records[0]
	allDayIdx := -1
	for i, h := range header {
		if h == "is_all_day" {
			allDayIdx = i
			break
		}
	}
	if allDayIdx < 0 {
		t.Fatal("is_all_day column not found")
	}
	if records[1][allDayIdx] != "true" {
		t.Errorf("all day item: is_all_day = %q", records[1][allDayIdx])
	}
	if records[2][allDayIdx] != "false" {
		t.Errorf("non-all day item: is_all_day = %q", records[2][allDayIdx])
	}
}
