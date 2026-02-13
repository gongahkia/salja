package parsers

import (
	"context"
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestOmniFocusBasicTask(t *testing.T) {
	input := "- Buy groceries @due(2024-01-15) @defer(2024-01-10)\n"

	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.txt")
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
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.DueDate.Format("2006-01-02") != "2024-01-15" {
		t.Errorf("due_date = %v", item.DueDate)
	}
	if item.StartTime == nil {
		t.Fatal("start_time (defer) is nil")
	}
	if item.StartTime.Format("2006-01-02") != "2024-01-10" {
		t.Errorf("start_time = %v", item.StartTime)
	}
	if item.ItemType != model.ItemTypeTask {
		t.Errorf("item_type = %q", item.ItemType)
	}
}

func TestOmniFocusDoneTag(t *testing.T) {
	input := "- Completed task @done\n"

	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Status != model.StatusCompleted {
		t.Errorf("status = %q, want %q", item.Status, model.StatusCompleted)
	}
}

func TestOmniFocusFlaggedTag(t *testing.T) {
	input := "- Important task @flagged\n"

	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Priority != model.PriorityHigh {
		t.Errorf("priority = %d, want %d (PriorityHigh)", item.Priority, model.PriorityHigh)
	}
}

func TestOmniFocusPriorityParsing(t *testing.T) {
	tests := []struct {
		input string
		want  model.Priority
	}{
		{"- Task @priority(high)\n", model.PriorityHigh},
		{"- Task @priority(medium)\n", model.PriorityMedium},
		{"- Task @priority(low)\n", model.PriorityLow},
	}
	for _, tc := range tests {
		p := NewOmniFocusParser()
		col, err := p.Parse(context.Background(), strings.NewReader(tc.input), "test.txt")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if col.Items[0].Priority != tc.want {
			t.Errorf("priority for %q = %d, want %d", tc.input, col.Items[0].Priority, tc.want)
		}
	}
}

func TestOmniFocusNotesBlock(t *testing.T) {
	input := "- Task with notes @due(2024-01-15)\n\tThis is a note line\n\tAnother note line\n"

	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if !strings.Contains(item.Description, "This is a note line") {
		t.Errorf("description should contain note line, got %q", item.Description)
	}
	if !strings.Contains(item.Description, "Another note line") {
		t.Errorf("description should contain second note line, got %q", item.Description)
	}
}

func TestOmniFocusMultipleTasks(t *testing.T) {
	input := "- Task 1 @due(2024-01-01)\n- Task 2 @flagged\n- Task 3 @done\n"

	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(col.Items))
	}
	if col.Items[0].Title != "Task 1" {
		t.Errorf("item 0 title = %q", col.Items[0].Title)
	}
	if col.Items[1].Title != "Task 2" {
		t.Errorf("item 1 title = %q", col.Items[1].Title)
	}
	if col.Items[2].Title != "Task 3" {
		t.Errorf("item 2 title = %q", col.Items[2].Title)
	}
}

func TestOmniFocusSourceApp(t *testing.T) {
	input := "- Test\n"
	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "omnifocus" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "omnifocus")
	}
}

func TestOmniFocusTagsParsing(t *testing.T) {
	input := "- Task @tags(work,personal)\n"

	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if len(item.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d: %v", len(item.Tags), item.Tags)
	}
	if item.Tags[0] != "work" || item.Tags[1] != "personal" {
		t.Errorf("tags = %v", item.Tags)
	}
}

func TestOmniFocusEmptyInput(t *testing.T) {
	p := NewOmniFocusParser()
	col, err := p.Parse(context.Background(), strings.NewReader(""), "test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(col.Items))
	}
}
