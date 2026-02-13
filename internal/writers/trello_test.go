package writers

import (
	"context"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/parsers"
)

func TestTrelloWriterRoundtrip(t *testing.T) {
	input := `{
		"name": "Board",
		"cards": [{
			"name": "Task 1",
			"desc": "Description",
			"due": "2024-06-15T14:00:00Z",
			"closed": false,
			"labels": [{"name": "urgent", "color": "red"}],
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

	p := parsers.NewTrelloParser()
	col, err := p.Parse(context.Background(), strings.NewReader(input), "test.json")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	var buf bytes.Buffer
	w := NewTrelloWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Task 1") {
		t.Error("output should contain card name")
	}
	if !strings.Contains(output, "urgent") {
		t.Error("output should contain label")
	}
	if !strings.Contains(output, "Step 1") {
		t.Error("output should contain checklist item")
	}
}

func TestTrelloWriterJSONValidity(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{Title: "Task 1", Description: "Desc"},
			{Title: "Task 2"},
		},
	}

	var buf bytes.Buffer
	w := NewTrelloWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["name"] != "Exported Board" {
		t.Errorf("board name = %v", result["name"])
	}
	cards, ok := result["cards"].([]interface{})
	if !ok {
		t.Fatal("cards is not an array")
	}
	if len(cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(cards))
	}
}

func TestTrelloWriterChecklistFromSubtasks(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title: "Parent",
				Subtasks: []model.Subtask{
					{Title: "Sub 1", Status: model.StatusCompleted},
					{Title: "Sub 2", Status: model.StatusPending},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewTrelloWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var result trelloExport
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	card := result.Cards[0]
	if len(card.Checklists) != 1 {
		t.Fatalf("expected 1 checklist, got %d", len(card.Checklists))
	}
	cl := card.Checklists[0]
	if cl.Name != "Checklist" {
		t.Errorf("checklist name = %q, want %q", cl.Name, "Checklist")
	}
	if len(cl.CheckItems) != 2 {
		t.Fatalf("expected 2 check items, got %d", len(cl.CheckItems))
	}
	if cl.CheckItems[0].Name != "Sub 1" || cl.CheckItems[0].State != "complete" {
		t.Errorf("check item 0: %+v", cl.CheckItems[0])
	}
	if cl.CheckItems[1].Name != "Sub 2" || cl.CheckItems[1].State != "incomplete" {
		t.Errorf("check item 1: %+v", cl.CheckItems[1])
	}
}

func TestTrelloWriterConfigurableChecklistName(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title: "Task",
				Subtasks: []model.Subtask{
					{Title: "Item", Status: model.StatusPending},
				},
			},
		},
	}

	var buf bytes.Buffer
	w := NewTrelloWriter()
	w.ChecklistName = "My Custom Checklist"
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var result trelloExport
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result.Cards[0].Checklists[0].Name != "My Custom Checklist" {
		t.Errorf("checklist name = %q, want %q", result.Cards[0].Checklists[0].Name, "My Custom Checklist")
	}
}

func TestTrelloWriterLabelsFromTags(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title: "Tagged Task",
				Tags:  []string{"urgent", "feature"},
			},
		},
	}

	var buf bytes.Buffer
	w := NewTrelloWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var result trelloExport
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	card := result.Cards[0]
	if len(card.Labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(card.Labels))
	}
	if card.Labels[0].Name != "urgent" || card.Labels[1].Name != "feature" {
		t.Errorf("labels = %+v", card.Labels)
	}
}

func TestTrelloWriterDueDateFormat(t *testing.T) {
	dueDate := time.Date(2024, 6, 15, 14, 0, 0, 0, time.UTC)
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{Title: "Task", DueDate: &dueDate},
		},
	}

	var buf bytes.Buffer
	w := NewTrelloWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var result trelloExport
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if result.Cards[0].Due != "2024-06-15T14:00:00.000Z" {
		t.Errorf("due = %q", result.Cards[0].Due)
	}
}

func TestTrelloWriterClosedStatus(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{Title: "Done", Status: model.StatusCompleted},
			{Title: "Open", Status: model.StatusPending},
		},
	}

	var buf bytes.Buffer
	w := NewTrelloWriter()
	if err := w.Write(context.Background(), col, &buf); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var result trelloExport
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if !result.Cards[0].Closed {
		t.Error("completed task should be closed")
	}
	if result.Cards[1].Closed {
		t.Error("pending task should not be closed")
	}
}
