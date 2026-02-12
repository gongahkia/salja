package parsers

import (
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestTickTickBasicTask(t *testing.T) {
	csv := `title,content,tags,start_date,due_date,priority,status,timezone,is_all_day,is_checklist,repeat,completed_time
Buy groceries,Milk and eggs,"food, errands",2024-01-15T09:00:00Z,2024-01-16T09:00:00Z,5,0,America/New_York,false,false,,`

	p := NewTickTickParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Items))
	}
	item := col.Items[0]
	if item.Title != "Buy groceries" {
		t.Errorf("title = %q, want %q", item.Title, "Buy groceries")
	}
	if item.Description != "Milk and eggs" {
		t.Errorf("description = %q, want %q", item.Description, "Milk and eggs")
	}
	if len(item.Tags) != 2 || item.Tags[0] != "food" || item.Tags[1] != "errands" {
		t.Errorf("tags = %v, want [food errands]", item.Tags)
	}
	if item.StartTime == nil {
		t.Fatal("start_time is nil")
	}
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.Priority != model.PriorityHigh {
		t.Errorf("priority = %d, want %d (PriorityHigh)", item.Priority, model.PriorityHigh)
	}
	if item.Status != model.StatusPending {
		t.Errorf("status = %q, want %q", item.Status, model.StatusPending)
	}
	if item.Timezone != "America/New_York" {
		t.Errorf("timezone = %q, want %q", item.Timezone, "America/New_York")
	}
	if item.IsAllDay {
		t.Error("is_all_day should be false")
	}
	if item.ItemType != model.ItemTypeTask {
		t.Errorf("item_type = %q, want %q", item.ItemType, model.ItemTypeTask)
	}
}

func TestTickTickChecklist(t *testing.T) {
	csv := "title,content,is_checklist\nParent Task,\"- [x] Sub 1\n- [ ] Sub 2\n- Sub 3\",true\n"

	p := NewTickTickParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if len(item.Subtasks) != 3 {
		t.Fatalf("expected 3 subtasks, got %d", len(item.Subtasks))
	}
	if item.Subtasks[0].Title != "Sub 1" || item.Subtasks[0].Status != model.StatusCompleted {
		t.Errorf("subtask 0: title=%q status=%q", item.Subtasks[0].Title, item.Subtasks[0].Status)
	}
	if item.Subtasks[1].Title != "Sub 2" || item.Subtasks[1].Status != model.StatusPending {
		t.Errorf("subtask 1: title=%q status=%q", item.Subtasks[1].Title, item.Subtasks[1].Status)
	}
	if item.Subtasks[2].Title != "Sub 3" {
		t.Errorf("subtask 2: title=%q", item.Subtasks[2].Title)
	}
	if item.Subtasks[0].SortOrder != 0 || item.Subtasks[1].SortOrder != 1 || item.Subtasks[2].SortOrder != 2 {
		t.Error("sort order mismatch")
	}
}

func TestTickTickPriorityMapping(t *testing.T) {
	tests := []struct {
		val  string
		want model.Priority
	}{
		{"0", model.PriorityNone},
		{"1", model.PriorityLow},
		{"3", model.PriorityMedium},
		{"5", model.PriorityHigh},
		{"", model.PriorityNone},
		{"99", model.PriorityNone},
	}
	for _, tc := range tests {
		got := mapTickTickPriority(tc.val)
		if got != tc.want {
			t.Errorf("mapTickTickPriority(%q) = %d, want %d", tc.val, got, tc.want)
		}
	}
}

func TestTickTickRepeatRules(t *testing.T) {
	tests := []struct {
		input string
		want  model.FreqType
	}{
		{"DAILY", model.FreqDaily},
		{"WEEKLY", model.FreqWeekly},
		{"MONTHLY", model.FreqMonthly},
		{"YEARLY", model.FreqYearly},
	}
	for _, tc := range tests {
		csv := "title,repeat\nTask," + tc.input + "\n"
		p := NewTickTickParser()
		col, err := p.Parse(strings.NewReader(csv), "test.csv")
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", tc.input, err)
		}
		item := col.Items[0]
		if item.Recurrence == nil {
			t.Fatalf("recurrence is nil for %s", tc.input)
		}
		if item.Recurrence.Freq != tc.want {
			t.Errorf("freq = %q, want %q for input %q", item.Recurrence.Freq, tc.want, tc.input)
		}
	}
}

func TestTickTickRepeatNone(t *testing.T) {
	rec := parseTickTickRepeat("NONE")
	if rec != nil {
		t.Error("expected nil for NONE repeat")
	}
	rec = parseTickTickRepeat("")
	if rec != nil {
		t.Error("expected nil for empty repeat")
	}
}

func TestTickTickEmptyFields(t *testing.T) {
	csv := "title,content,tags,start_date,due_date,priority,status\nMinimal,,,,,,\n"

	p := NewTickTickParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Title != "Minimal" {
		t.Errorf("title = %q", item.Title)
	}
	if item.Description != "" {
		t.Errorf("description should be empty, got %q", item.Description)
	}
	if len(item.Tags) != 0 {
		t.Errorf("tags should be empty, got %v", item.Tags)
	}
	if item.StartTime != nil {
		t.Error("start_time should be nil")
	}
	if item.DueDate != nil {
		t.Error("due_date should be nil")
	}
}

func TestTickTickDateFormats(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-01-15T09:00:00Z", false},
		{"2024-01-15T09:00:00", false},
		{"2024-01-15", false},
		{"01/15/2024", false},
		{"not-a-date", true},
	}
	for _, tc := range tests {
		_, err := parseTickTickDateField(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("parseTickTickDateField(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
		}
	}
}

func TestTickTickCompletedTask(t *testing.T) {
	csv := "title,status,completed_time\nDone Task,2,2024-06-01T12:00:00Z\n"

	p := NewTickTickParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Status != model.StatusCompleted {
		t.Errorf("status = %q, want %q", item.Status, model.StatusCompleted)
	}
	if item.CompletionDate == nil {
		t.Fatal("completion_date is nil")
	}
}

func TestTickTickSourceApp(t *testing.T) {
	csv := "title\nTest\n"
	p := NewTickTickParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "ticktick" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "ticktick")
	}
}

func TestTickTickIsAllDay(t *testing.T) {
	csv := "title,is_all_day\nAll Day,true\nNot All Day,false\n"
	p := NewTickTickParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !col.Items[0].IsAllDay {
		t.Error("first item should be all day")
	}
	if col.Items[1].IsAllDay {
		t.Error("second item should not be all day")
	}
}
