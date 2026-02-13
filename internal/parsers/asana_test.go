package parsers

import (
	"context"
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestAsanaAllFields(t *testing.T) {
	csv := `Name,Description,Due Date,Tags,Completed
Buy groceries,Milk and bread,2024-01-15,"food, errands",false`

	p := NewAsanaParser()
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
	if item.Description != "Milk and bread" {
		t.Errorf("description = %q", item.Description)
	}
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.DueDate.Format("2006-01-02") != "2024-01-15" {
		t.Errorf("due_date = %v", item.DueDate)
	}
	if len(item.Tags) != 2 || item.Tags[0] != "food" || item.Tags[1] != "errands" {
		t.Errorf("tags = %v", item.Tags)
	}
	if item.Status != model.StatusPending {
		t.Errorf("status = %q, want %q", item.Status, model.StatusPending)
	}
	if item.ItemType != model.ItemTypeTask {
		t.Errorf("item_type = %q", item.ItemType)
	}
}

func TestAsanaCompletedTask(t *testing.T) {
	csv := `Name,Completed
Done Task,true`

	p := NewAsanaParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Status != model.StatusCompleted {
		t.Errorf("status = %q, want %q", item.Status, model.StatusCompleted)
	}
}

func TestAsanaCompletedTaskNumeric(t *testing.T) {
	csv := `Name,Completed
Done Task,1`

	p := NewAsanaParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.Items[0].Status != model.StatusCompleted {
		t.Errorf("status = %q, want %q", col.Items[0].Status, model.StatusCompleted)
	}
}

func TestAsanaMissingNameError(t *testing.T) {
	csv := `Description,Due Date
Some description,2024-01-01`

	p := NewAsanaParser()
	_, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err == nil {
		t.Fatal("expected error for missing Name column")
	}
	if !strings.Contains(err.Error(), "Name") {
		t.Errorf("error should mention Name: %v", err)
	}
}

func TestAsanaMinimalTask(t *testing.T) {
	csv := `Name
Simple Task`

	p := NewAsanaParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Title != "Simple Task" {
		t.Errorf("title = %q", item.Title)
	}
	if item.DueDate != nil {
		t.Error("due_date should be nil")
	}
	if len(item.Tags) != 0 {
		t.Errorf("tags should be empty, got %v", item.Tags)
	}
}

func TestAsanaSourceApp(t *testing.T) {
	csv := "Name\nTest\n"
	p := NewAsanaParser()
	col, err := p.Parse(context.Background(), strings.NewReader(csv), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "asana" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "asana")
	}
}

func TestAsanaEmptyInput(t *testing.T) {
	p := NewAsanaParser()
	col, err := p.Parse(context.Background(), strings.NewReader(""), "test.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(col.Items))
	}
}
