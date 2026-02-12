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

func TestTodoistWriterRoundtrip(t *testing.T) {
	input := `TYPE,CONTENT,DESCRIPTION,PRIORITY,INDENT,DATE,TIMEZONE
task,Buy milk,From the store,4,0,2024-01-15,America/New_York`

	p := parsers.NewTodoistParser()
	col, err := p.Parse(strings.NewReader(input), "test.csv")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var buf bytes.Buffer
	w := NewTodoistWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Buy milk") {
		t.Error("output should contain title")
	}
	if !strings.Contains(output, "From the store") {
		t.Error("output should contain description")
	}
	if !strings.Contains(output, "2024-01-15") {
		t.Error("output should contain date")
	}
}

func TestTodoistWriterPriorityInversion(t *testing.T) {
	tests := []struct {
		priority model.Priority
		want     string
	}{
		{model.PriorityHighest, "4"},
		{model.PriorityHigh, "3"},
		{model.PriorityMedium, "2"},
		{model.PriorityLow, "1"},
		{model.PriorityNone, "1"},
	}
	for _, tc := range tests {
		got := exportTodoistPriority(tc.priority)
		if got != tc.want {
			t.Errorf("exportTodoistPriority(%d) = %q, want %q", tc.priority, got, tc.want)
		}
	}
}

func TestTodoistWriterSubtasksAsIndent(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:    "Parent",
				Priority: model.PriorityHigh,
				Subtasks: []model.Subtask{
					{Title: "Child 1", Status: model.StatusPending},
					{Title: "Child 2", Status: model.StatusCompleted},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewTodoistWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()

	// Header + parent + 2 subtasks = 4 rows
	if len(records) != 4 {
		t.Fatalf("expected 4 rows (header+parent+2 subtasks), got %d", len(records))
	}

	// All rows should be TYPE=task
	for i := 1; i < len(records); i++ {
		if records[i][0] != "task" {
			t.Errorf("row %d TYPE = %q, want %q", i, records[i][0], "task")
		}
	}

	// Parent indent = 0, subtask indent = 1
	if records[1][4] != "0" {
		t.Errorf("parent INDENT = %q, want %q", records[1][4], "0")
	}
	if records[2][4] != "1" {
		t.Errorf("subtask 1 INDENT = %q, want %q", records[2][4], "1")
	}
	if records[3][4] != "1" {
		t.Errorf("subtask 2 INDENT = %q, want %q", records[3][4], "1")
	}
}

func TestTodoistWriterDateFormat(t *testing.T) {
	dueDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{Title: "Task", DueDate: &dueDate},
		},
	}

	var buf bytes.Buffer
	w := NewTodoistWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	if !strings.Contains(buf.String(), "2024-06-15") {
		t.Error("output should contain date in YYYY-MM-DD format")
	}
}

func TestTodoistWriterTimezone(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{Title: "Task", Timezone: "Europe/London"},
		},
	}

	var buf bytes.Buffer
	w := NewTodoistWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	if !strings.Contains(buf.String(), "Europe/London") {
		t.Error("output should contain timezone")
	}
}

func TestTodoistWriterHeader(t *testing.T) {
	col := &model.CalendarCollection{Items: []model.CalendarItem{}}

	var buf bytes.Buffer
	w := NewTodoistWriter()
	if err := w.Write(col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	reader := csv.NewReader(strings.NewReader(buf.String()))
	records, _ := reader.ReadAll()
	if len(records) < 1 {
		t.Fatal("expected at least header row")
	}
	expected := []string{"TYPE", "CONTENT", "DESCRIPTION", "PRIORITY", "INDENT", "AUTHOR", "RESPONSIBLE", "DATE", "DATE_LANG", "TIMEZONE"}
	for i, h := range expected {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}
}
