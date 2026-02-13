package parsers

import (
	"context"
	"strings"
	"testing"

	"github.com/gongahkia/salja/internal/model"
)

func TestTrelloCardWithChecklist(t *testing.T) {
	json := `{
		"name": "My Board",
		"cards": [{
			"name": "Task 1",
			"desc": "Description",
			"due": "",
			"closed": false,
			"labels": [],
			"checklists": [{
				"name": "Checklist",
				"checkItems": [
					{"name": "Step 1", "state": "complete"},
					{"name": "Step 2", "state": "incomplete"}
				]
			}]
		}],
		"lists": []
	}`

	p := NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(json), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(col.Items))
	}
	item := col.Items[0]
	if item.Title != "Task 1" {
		t.Errorf("title = %q", item.Title)
	}
	if item.Description != "Description" {
		t.Errorf("description = %q", item.Description)
	}
	if len(item.Subtasks) != 2 {
		t.Fatalf("expected 2 subtasks, got %d", len(item.Subtasks))
	}
	if item.Subtasks[0].Title != "Step 1" || item.Subtasks[0].Status != model.StatusCompleted {
		t.Errorf("subtask 0: title=%q status=%q", item.Subtasks[0].Title, item.Subtasks[0].Status)
	}
	if item.Subtasks[1].Title != "Step 2" || item.Subtasks[1].Status != model.StatusPending {
		t.Errorf("subtask 1: title=%q status=%q", item.Subtasks[1].Title, item.Subtasks[1].Status)
	}
}

func TestTrelloCardWithLabels(t *testing.T) {
	json := `{
		"name": "Board",
		"cards": [{
			"name": "Labeled Task",
			"desc": "",
			"due": "",
			"closed": false,
			"labels": [
				{"name": "urgent", "color": "red"},
				{"name": "feature", "color": "green"}
			],
			"checklists": []
		}],
		"lists": []
	}`

	p := NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(json), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if len(item.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(item.Tags))
	}
	if item.Tags[0] != "urgent" || item.Tags[1] != "feature" {
		t.Errorf("tags = %v", item.Tags)
	}
}

func TestTrelloCardWithDueDate(t *testing.T) {
	json := `{
		"name": "Board",
		"cards": [{
			"name": "Due Task",
			"desc": "",
			"due": "2024-06-15T14:00:00Z",
			"closed": false,
			"labels": [],
			"checklists": []
		}],
		"lists": []
	}`

	p := NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(json), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.DueDate == nil {
		t.Fatal("due_date is nil")
	}
	if item.DueDate.Format("2006-01-02") != "2024-06-15" {
		t.Errorf("due_date = %v", item.DueDate)
	}
}

func TestTrelloClosedCard(t *testing.T) {
	json := `{
		"name": "Board",
		"cards": [{
			"name": "Closed Task",
			"desc": "",
			"due": "",
			"closed": true,
			"labels": [],
			"checklists": []
		}],
		"lists": []
	}`

	p := NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(json), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	if item.Status != model.StatusCompleted {
		t.Errorf("status = %q, want %q", item.Status, model.StatusCompleted)
	}
}

func TestTrelloEmptyBoard(t *testing.T) {
	json := `{"name": "Empty Board", "cards": [], "lists": []}`

	p := NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(json), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(col.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(col.Items))
	}
}

func TestTrelloSourceApp(t *testing.T) {
	json := `{"name": "Board", "cards": [], "lists": []}`
	p := NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(json), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if col.SourceApp != "trello" {
		t.Errorf("source_app = %q, want %q", col.SourceApp, "trello")
	}
}

func TestTrelloLabelWithEmptyName(t *testing.T) {
	json := `{
		"name": "Board",
		"cards": [{
			"name": "Task",
			"desc": "",
			"due": "",
			"closed": false,
			"labels": [
				{"name": "", "color": "red"},
				{"name": "valid", "color": "blue"}
			],
			"checklists": []
		}],
		"lists": []
	}`

	p := NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(json), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	item := col.Items[0]
	// Empty-name labels are filtered out
	if len(item.Tags) != 1 {
		t.Fatalf("expected 1 tag (empty name filtered), got %d: %v", len(item.Tags), item.Tags)
	}
	if item.Tags[0] != "valid" {
		t.Errorf("tag = %q", item.Tags[0])
	}
}
