package parsers

import (
	"context"
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestNotionAllColumns(t *testing.T) {
	csv := `Title,Date,Status,Tags,Priority
Buy groceries,2024-01-15,Done,"food, errands",High`

	p := NewNotionParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Items))
	}
	item := col.Items[0]
	if item.Title != "Buy groceries" {
		t.Errorf("title = %q", item.Title)
	}
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.DueDate.Format("2006-01-02") != "2024-01-15" {
		t.Errorf("due_date = %v", item.DueDate)
	}
	if item.Status != model.StatusCompleted {
		t.Errorf("status = %q, want %q", item.Status, model.StatusCompleted)
	}
	if len(item.Tags) != 2 || item.Tags[0] != "food" || item.Tags[1] != "errands" {
		t.Errorf("tags = %v", item.Tags)
	}
	if item.Priority != model.PriorityHigh {
		t.Errorf("priority = %d, want %d", item.Priority, model.PriorityHigh)
	}
}

func TestNotionNameColumnVariant(t *testing.T) {
	csv := `Name,Status
My Task,In Progress`

	p := NewNotionParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Title != "My Task" {
		t.Errorf("title = %q, want %q", item.Title, "My Task")
	}
	if item.Status != model.StatusInProgress {
		t.Errorf("status = %q, want %q", item.Status, model.StatusInProgress)
	}
}

func TestNotionTaskColumnVariant(t *testing.T) {
	csv := `Task,Priority
Another Task,Low`

	p := NewNotionParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Title != "Another Task" {
		t.Errorf("title = %q", item.Title)
	}
	if item.Priority != model.PriorityLow {
		t.Errorf("priority = %d, want %d", item.Priority, model.PriorityLow)
	}
}

func TestNotionMissingTitleError(t *testing.T) {
	csv := `Status,Date
Done,2024-01-01`

	p := NewNotionParser()
	_, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err == nil {
		t.Fatal("expected error for missing title column")
	}
	if !strings.Contains(err.Error(), "title column") {
		t.Errorf("error should mention title column: %v", err)
	}
}

func TestNotionMissingColumnsGraceful(t *testing.T) {
	csv := `Title
Just a title`

	p := NewNotionParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Title != "Just a title" {
		t.Errorf("title = %q", item.Title)
	}
	if item.DueDate != nil {
		t.Error("due_date should be nil")
	}
	if item.Status != model.StatusPending {
		t.Errorf("status = %q, want %q", item.Status, model.StatusPending)
	}
	if len(item.Tags) != 0 {
		t.Errorf("tags should be empty, got %v", item.Tags)
	}
}

func TestNotionStatusMapping(t *testing.T) {
	csv := `Title,Status
Task1,Done
Task2,Completed
Task3,In Progress
Task4,Doing
Task5,Cancelled
Task6,Not Started`

	p := NewNotionParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []model.Status{
		model.StatusCompleted,
		model.StatusCompleted,
		model.StatusInProgress,
		model.StatusInProgress,
		model.StatusCancelled,
		model.StatusPending,
	}
	for i, want := range expected {
		if col.Items[i].Status != want {
			t.Errorf("item %d status = %q, want %q", i, col.Items[i].Status, want)
		}
	}
}

func TestNotionSourceApp(t *testing.T) {
	csv := "Title\nTest\n"
	p := NewNotionParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "notion" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "notion")
	}
}

func TestNotionAlternateDateFormat(t *testing.T) {
	// "January 15, 2024" contains a comma so it must be quoted in CSV
	csv := "Title,Date\nTask,\"January 15, 2024\"\n"

	p := NewNotionParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.DueDate.Month() != 1 || item.DueDate.Day() != 15 {
		t.Errorf("due_date = %v", item.DueDate)
	}
}
