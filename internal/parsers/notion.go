package parsers

import (
	"encoding/csv"
	
	"io"
	"os"
	"strings"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type NotionParser struct{}

func NewNotionParser() *NotionParser {
	return &NotionParser{}
}

func (p *NotionParser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(f, filePath)
}

func (p *NotionParser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	csvReader := csv.NewReader(r)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "notion"}, nil
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "notion",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	for i := 1; i < len(records); i++ {
		item := parseNotionRow(records[i], colMap)
		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}

func parseNotionRow(row []string, colMap map[string]int) model.CalendarItem {
	item := model.CalendarItem{
		ItemType: model.ItemTypeTask,
		Status:   model.StatusPending,
	}

	titleCandidates := []string{"Title", "Name", "Task"}
	for _, candidate := range titleCandidates {
		if idx, ok := colMap[candidate]; ok && idx < len(row) {
			item.Title = row[idx]
			break
		}
	}

	dateCandidates := []string{"Date", "Due Date", "Due"}
	for _, candidate := range dateCandidates {
		if idx, ok := colMap[candidate]; ok && idx < len(row) && row[idx] != "" {
			if t, err := time.Parse("2006-01-02", row[idx]); err == nil {
				item.DueDate = &t
				break
			} else if t, err := time.Parse("January 2, 2006", row[idx]); err == nil {
				item.DueDate = &t
				break
			}
		}
	}

	statusCandidates := []string{"Status"}
	for _, candidate := range statusCandidates {
		if idx, ok := colMap[candidate]; ok && idx < len(row) {
			switch strings.ToLower(row[idx]) {
			case "done", "completed":
				item.Status = model.StatusCompleted
			case "in progress", "doing":
				item.Status = model.StatusInProgress
			case "cancelled":
				item.Status = model.StatusCancelled
			default:
				item.Status = model.StatusPending
			}
			break
		}
	}

	tagsCandidates := []string{"Tags", "Labels"}
	for _, candidate := range tagsCandidates {
		if idx, ok := colMap[candidate]; ok && idx < len(row) && row[idx] != "" {
			item.Tags = strings.Split(row[idx], ",")
			for i := range item.Tags {
				item.Tags[i] = strings.TrimSpace(item.Tags[i])
			}
			break
		}
	}

	priorityCandidates := []string{"Priority"}
	for _, candidate := range priorityCandidates {
		if idx, ok := colMap[candidate]; ok && idx < len(row) {
			switch strings.ToLower(row[idx]) {
			case "high", "urgent":
				item.Priority = model.PriorityHigh
			case "medium", "normal":
				item.Priority = model.PriorityMedium
			case "low":
				item.Priority = model.PriorityLow
			}
			break
		}
	}

	return item
}
