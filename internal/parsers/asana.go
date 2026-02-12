package parsers

import (
	"encoding/csv"
	
	"io"
	"os"
	"strings"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type AsanaParser struct{}

func NewAsanaParser() *AsanaParser {
	return &AsanaParser{}
}

func (p *AsanaParser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(f, filePath)
}

func (p *AsanaParser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	csvReader := csv.NewReader(r)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "asana"}, nil
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "asana",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	for i := 1; i < len(records); i++ {
		item := parseAsanaRow(records[i], colMap)
		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}

func parseAsanaRow(row []string, colMap map[string]int) model.CalendarItem {
	item := model.CalendarItem{
		ItemType: model.ItemTypeTask,
		Status:   model.StatusPending,
	}

	if idx, ok := colMap["Name"]; ok && idx < len(row) {
		item.Title = row[idx]
	}

	if idx, ok := colMap["Description"]; ok && idx < len(row) {
		item.Description = row[idx]
	}

	if idx, ok := colMap["Due Date"]; ok && idx < len(row) && row[idx] != "" {
		if t, err := time.Parse("2006-01-02", row[idx]); err == nil {
			item.DueDate = &t
		}
	}

	if idx, ok := colMap["Tags"]; ok && idx < len(row) && row[idx] != "" {
		item.Tags = strings.Split(row[idx], ",")
		for i := range item.Tags {
			item.Tags[i] = strings.TrimSpace(item.Tags[i])
		}
	}

	if idx, ok := colMap["Completed"]; ok && idx < len(row) {
		if strings.ToLower(row[idx]) == "true" || row[idx] == "1" {
			item.Status = model.StatusCompleted
		}
	}

	return item
}
