package writers

import (
	"context"
	"encoding/json"
	
	"io"
	"os"

	"github.com/gongahkia/salja/internal/model"
)

type TrelloWriter struct {
	ChecklistName string
}

func NewTrelloWriter() *TrelloWriter {
	return &TrelloWriter{ChecklistName: "Checklist"}
}

type trelloExport struct {
	Name  string             `json:"name"`
	Cards []trelloCardExport `json:"cards"`
}

type trelloCardExport struct {
	Name       string                  `json:"name"`
	Desc       string                  `json:"desc"`
	Due        string                  `json:"due,omitempty"`
	Closed     bool                    `json:"closed"`
	Labels     []trelloLabelExport     `json:"labels"`
	Checklists []trelloChecklistExport `json:"checklists,omitempty"`
}

type trelloLabelExport struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type trelloChecklistExport struct {
	Name       string                  `json:"name"`
	CheckItems []trelloCheckItemExport `json:"checkItems"`
}

type trelloCheckItemExport struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

func (w *TrelloWriter) WriteFile(ctx context.Context, collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return w.Write(ctx, collection, f)
}

func (w *TrelloWriter) Write(ctx context.Context, collection *model.CalendarCollection, writer io.Writer) error {
	board := trelloExport{
		Name:  "Exported Board",
		Cards: []trelloCardExport{},
	}

	for _, item := range collection.Items {
		card := trelloCardExport{
			Name:   item.Title,
			Desc:   item.Description,
			Closed: item.Status == model.StatusCompleted,
			Labels: []trelloLabelExport{},
		}

		if item.DueDate != nil {
			card.Due = item.DueDate.Format("2006-01-02T15:04:05.000Z")
		}

		for _, tag := range item.Tags {
			card.Labels = append(card.Labels, trelloLabelExport{
				Name:  tag,
				Color: "blue",
			})
		}

		if len(item.Subtasks) > 0 {
			checklist := trelloChecklistExport{
				Name:       w.ChecklistName,
				CheckItems: []trelloCheckItemExport{},
			}
			for _, subtask := range item.Subtasks {
				state := "incomplete"
				if subtask.Status == model.StatusCompleted {
					state = "complete"
				}
				checklist.CheckItems = append(checklist.CheckItems, trelloCheckItemExport{
					Name:  subtask.Title,
					State: state,
				})
			}
			card.Checklists = append(card.Checklists, checklist)
		}

		board.Cards = append(board.Cards, card)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(board)
}
