package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type TrelloParser struct{}

func NewTrelloParser() *TrelloParser {
	return &TrelloParser{}
}

type TrelloBoard struct {
	Name  string        `json:"name"`
	Cards []TrelloCard  `json:"cards"`
	Lists []TrelloList  `json:"lists"`
}

type TrelloCard struct {
	Name       string             `json:"name"`
	Desc       string             `json:"desc"`
	Due        string             `json:"due"`
	Closed     bool               `json:"closed"`
	Labels     []TrelloLabel      `json:"labels"`
	Checklists []TrelloChecklist  `json:"checklists"`
}

type TrelloLabel struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type TrelloChecklist struct {
	Name       string            `json:"name"`
	CheckItems []TrelloCheckItem `json:"checkItems"`
}

type TrelloCheckItem struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

type TrelloList struct {
	Name string `json:"name"`
}

func (p *TrelloParser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(f, filePath)
}

func (p *TrelloParser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	var board TrelloBoard
	if err := json.NewDecoder(r).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode Trello JSON: %w", err)
	}

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "trello",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	for _, card := range board.Cards {
		item := model.CalendarItem{
			Title:       card.Name,
			Description: card.Desc,
			ItemType:    model.ItemTypeTask,
			Status:      model.StatusPending,
		}

		if card.Closed {
			item.Status = model.StatusCompleted
		}

		if card.Due != "" {
			if t, err := time.Parse(time.RFC3339, card.Due); err == nil {
				item.DueDate = &t
			}
		}

		for _, label := range card.Labels {
			if label.Name != "" {
				item.Tags = append(item.Tags, label.Name)
			}
		}

		for _, checklist := range card.Checklists {
			for _, checkItem := range checklist.CheckItems {
				status := model.StatusPending
				if checkItem.State == "complete" {
					status = model.StatusCompleted
				}
				item.Subtasks = append(item.Subtasks, model.Subtask{
					Title:     checkItem.Name,
					Status:    status,
					SortOrder: len(item.Subtasks),
				})
			}
		}

		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}
