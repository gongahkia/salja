package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type TickTickParser struct{}

func NewTickTickParser() *TickTickParser {
	return &TickTickParser{}
}

func (p *TickTickParser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TickTick CSV: %w", err)
	}
	defer f.Close()

	return p.Parse(f, filePath)
}

func (p *TickTickParser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	tr, err := transcodeReader(r)
	if err != nil {
		return nil, fmt.Errorf("charset detection failed: %w", err)
	}
	csvReader := csv.NewReader(tr)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "ticktick"}, nil
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "ticktick",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	for i := 1; i < len(records); i++ {
		row := records[i]
		item := p.parseRow(row, colMap)
		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}

func (p *TickTickParser) parseRow(row []string, colMap map[string]int) model.CalendarItem {
	item := model.CalendarItem{
		ItemType: model.ItemTypeTask,
		Status:   model.StatusPending,
	}

	if idx, ok := colMap["title"]; ok && idx < len(row) {
		item.Title = row[idx]
	}

	if idx, ok := colMap["content"]; ok && idx < len(row) {
		item.Description = row[idx]
	}

	if idx, ok := colMap["tags"]; ok && idx < len(row) && row[idx] != "" {
		item.Tags = strings.Split(row[idx], ",")
		for i := range item.Tags {
			item.Tags[i] = strings.TrimSpace(item.Tags[i])
		}
	}

	if idx, ok := colMap["start_date"]; ok && idx < len(row) && row[idx] != "" {
		if t, err := parseTickTickDateField(row[idx]); err == nil {
			item.StartTime = &t
		}
	}

	if idx, ok := colMap["due_date"]; ok && idx < len(row) && row[idx] != "" {
		if t, err := parseTickTickDateField(row[idx]); err == nil {
			item.DueDate = &t
		}
	}

	if idx, ok := colMap["priority"]; ok && idx < len(row) {
		item.Priority = mapTickTickPriority(row[idx])
	}

	if idx, ok := colMap["status"]; ok && idx < len(row) {
		switch row[idx] {
		case "0":
			item.Status = model.StatusPending
		case "2":
			item.Status = model.StatusCompleted
		}
	}

	if idx, ok := colMap["completed_time"]; ok && idx < len(row) && row[idx] != "" {
		if t, err := time.Parse(time.RFC3339, row[idx]); err == nil {
			item.CompletionDate = &t
			item.Status = model.StatusCompleted
		}
	}

	if idx, ok := colMap["is_all_day"]; ok && idx < len(row) {
		item.IsAllDay = row[idx] == "true" || row[idx] == "1"
	}

	if idx, ok := colMap["timezone"]; ok && idx < len(row) {
		item.Timezone = row[idx]
	}

	if idx, ok := colMap["is_checklist"]; ok && idx < len(row) && (row[idx] == "true" || row[idx] == "1") {
		if contentIdx, ok := colMap["content"]; ok && contentIdx < len(row) {
			item.Subtasks = parseTickTickChecklist(row[contentIdx])
		}
	}

	if idx, ok := colMap["repeat"]; ok && idx < len(row) && row[idx] != "" {
		rec := parseTickTickRepeat(row[idx])
		if rec != nil {
			item.Recurrence = rec
		}
	}

	return item
}

func mapTickTickPriority(val string) model.Priority {
	switch val {
	case "0", "":
		return model.PriorityNone
	case "1":
		return model.PriorityLow
	case "3":
		return model.PriorityMedium
	case "5":
		return model.PriorityHigh
	default:
		return model.PriorityNone
	}
}

func parseTickTickChecklist(content string) []model.Subtask {
	var subtasks []model.Subtask
	lines := strings.Split(content, "\n")
	sortOrder := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		status := model.StatusPending
		title := line
		
		if strings.HasPrefix(line, "- [x] ") {
			status = model.StatusCompleted
			title = strings.TrimPrefix(line, "- [x] ")
		} else if strings.HasPrefix(line, "- [ ] ") {
			status = model.StatusPending
			title = strings.TrimPrefix(line, "- [ ] ")
		} else if strings.HasPrefix(line, "- ") {
			title = strings.TrimPrefix(line, "- ")
		}
		
		subtasks = append(subtasks, model.Subtask{
			Title:     title,
			Status:    status,
			SortOrder: sortOrder,
		})
		sortOrder++
	}
	
	return subtasks
}

func parseTickTickRepeat(repeat string) *model.Recurrence {
	repeat = strings.ToUpper(strings.TrimSpace(repeat))
	if repeat == "" || repeat == "NONE" {
		return nil
	}

	rec := &model.Recurrence{Interval: 1}

	if strings.Contains(repeat, "DAILY") {
		rec.Freq = model.FreqDaily
	} else if strings.Contains(repeat, "WEEKLY") {
		rec.Freq = model.FreqWeekly
	} else if strings.Contains(repeat, "MONTHLY") {
		rec.Freq = model.FreqMonthly
	} else if strings.Contains(repeat, "YEARLY") {
		rec.Freq = model.FreqYearly
	} else {
		return nil
	}

	return rec
}

func parseTickTickDateField(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
		"01/02/2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}
