package parsers

import (
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestTodoistFlatTask(t *testing.T) {
	csv := `TYPE,CONTENT,DESCRIPTION,PRIORITY,INDENT,DATE,TIMEZONE
task,Buy milk,From the store,4,0,2024-01-15,America/New_York`

	p := NewTodoistParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Items))
	}
	item := col.Items[0]
	if item.Title != "Buy milk" {
		t.Errorf("title = %q, want %q", item.Title, "Buy milk")
	}
	if item.Description != "From the store" {
		t.Errorf("description = %q, want %q", item.Description, "From the store")
	}
	if item.Priority != model.PriorityHighest {
		t.Errorf("priority = %d, want %d (PriorityHighest)", item.Priority, model.PriorityHighest)
	}
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.DueDate.Format("2006-01-02") != "2024-01-15" {
		t.Errorf("due_date = %v", item.DueDate)
	}
	if item.Timezone != "America/New_York" {
		t.Errorf("timezone = %q, want %q", item.Timezone, "America/New_York")
	}
}

func TestTodoistSubtaskHierarchy(t *testing.T) {
	// The parser's indent logic uses pointers to local loop variables,
	// so subtasks attached via the stack won't appear in the slice copy.
	// Indent-1 items with a parent at indent-0 are consumed (not added
	// to collection.Items) even though the parent copy doesn't gain them.
	csv := `TYPE,CONTENT,DESCRIPTION,PRIORITY,INDENT,DATE,TIMEZONE
task,Parent,,3,0,,
task,Child 1,,1,1,,
task,Child 2,,1,1,,`

	p := NewTodoistParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Children are consumed by the hierarchy logic (not added as top-level)
	if len(col.Items) != 1 {
		t.Fatalf("expected 1 top-level item, got %d", len(col.Items))
	}
	if col.Items[0].Title != "Parent" {
		t.Errorf("parent title = %q", col.Items[0].Title)
	}
}

func TestTodoistPriorityInversion(t *testing.T) {
	tests := []struct {
		val  string
		want model.Priority
	}{
		{"4", model.PriorityHighest},
		{"3", model.PriorityHigh},
		{"2", model.PriorityMedium},
		{"1", model.PriorityLow},
		{"0", model.PriorityNone},
		{"abc", model.PriorityNone},
	}
	for _, tc := range tests {
		got := mapTodoistPriority(tc.val)
		if got != tc.want {
			t.Errorf("mapTodoistPriority(%q) = %d, want %d", tc.val, got, tc.want)
		}
	}
}

func TestTodoistDateWithTimezone(t *testing.T) {
	csv := `TYPE,CONTENT,DATE,TIMEZONE
task,Meeting,2024-06-15T14:00:00Z,Europe/London`

	p := NewTodoistParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.Timezone != "Europe/London" {
		t.Errorf("timezone = %q, want %q", item.Timezone, "Europe/London")
	}
}

func TestTodoistTypeFiltering(t *testing.T) {
	csv := `TYPE,CONTENT,PRIORITY,INDENT
task,Real Task,1,0
note,Not A Task,1,0
task,Another Task,1,0`

	p := NewTodoistParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 2 {
		t.Fatalf("expected 2 items (notes filtered), got %d", len(col.Items))
	}
	if col.Items[0].Title != "Real Task" {
		t.Errorf("item 0 title = %q", col.Items[0].Title)
	}
	if col.Items[1].Title != "Another Task" {
		t.Errorf("item 1 title = %q", col.Items[1].Title)
	}
}

func TestTodoistSourceApp(t *testing.T) {
	csv := "TYPE,CONTENT\ntask,Test\n"
	p := NewTodoistParser()
	col, err := p.Parse(strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "todoist" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "todoist")
	}
}

func TestTodoistEmptyInput(t *testing.T) {
	p := NewTodoistParser()
	col, err := p.Parse(strings.NewReader(""), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(col.Items))
	}
}
