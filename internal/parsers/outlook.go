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

type OutlookParser struct{}

func NewOutlookParser() *OutlookParser {
	return &OutlookParser{}
}

func (p *OutlookParser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Outlook CSV %s: %w", filePath, err)
	}
	defer f.Close()
	return p.Parse(f, filePath)
}

func (p *OutlookParser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	tr, err := transcodeReader(r)
	if err != nil {
		return nil, fmt.Errorf("charset detection failed: %w", err)
	}
	csvReader := csv.NewReader(tr)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV %s: %w", sourcePath, err)
	}

	if len(records) == 0 {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "outlook"}, nil
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	requiredCols := []string{"Subject"}
	if missing := findMissingColumns(colMap, requiredCols); len(missing) > 0 {
		return nil, fmt.Errorf("Outlook CSV %s missing required columns: %s", sourcePath, strings.Join(missing, ", "))
	}

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "outlook",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	for i := 1; i < len(records); i++ {
		item, err := parseOutlookRow(records[i], colMap)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", sourcePath, i+1, err)
		}
		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}

func parseOutlookRow(row []string, colMap map[string]int) (model.CalendarItem, error) {
	item := model.CalendarItem{
		ItemType: model.ItemTypeEvent,
		Status:   model.StatusPending,
	}

	if idx, ok := colMap["Subject"]; ok && idx < len(row) {
		item.Title = row[idx]
	}

	if idx, ok := colMap["Description"]; ok && idx < len(row) {
		item.Description = row[idx]
	}

	if idx, ok := colMap["Location"]; ok && idx < len(row) {
		item.Location = row[idx]
	}

	if idx, ok := colMap["Categories"]; ok && idx < len(row) && row[idx] != "" {
		item.Tags = strings.Split(row[idx], ";")
		for i := range item.Tags {
			item.Tags[i] = strings.TrimSpace(item.Tags[i])
		}
	}

	allDayIdx, hasAllDay := colMap["All day event"]
	isAllDay := hasAllDay && allDayIdx < len(row) && strings.ToLower(row[allDayIdx]) == "true"
	item.IsAllDay = isAllDay

	if startDateIdx, ok := colMap["Start Date"]; ok && startDateIdx < len(row) && row[startDateIdx] != "" {
		var start time.Time
		var err error
		
		if isAllDay {
			start, err = time.Parse("1/2/2006", row[startDateIdx])
		} else if startTimeIdx, ok := colMap["Start Time"]; ok && startTimeIdx < len(row) && row[startTimeIdx] != "" {
			dateTime := row[startDateIdx] + " " + row[startTimeIdx]
			start, err = time.Parse("1/2/2006 3:04:05 PM", dateTime)
			if err != nil {
				start, err = time.Parse("1/2/2006 15:04:05", dateTime)
			}
		}
		
		if err == nil {
			item.StartTime = &start
		}
	}

	if endDateIdx, ok := colMap["End Date"]; ok && endDateIdx < len(row) && row[endDateIdx] != "" {
		var end time.Time
		var err error
		
		if isAllDay {
			end, err = time.Parse("1/2/2006", row[endDateIdx])
		} else if endTimeIdx, ok := colMap["End Time"]; ok && endTimeIdx < len(row) && row[endTimeIdx] != "" {
			dateTime := row[endDateIdx] + " " + row[endTimeIdx]
			end, err = time.Parse("1/2/2006 3:04:05 PM", dateTime)
			if err != nil {
				end, err = time.Parse("1/2/2006 15:04:05", dateTime)
			}
		}
		
		if err == nil {
			item.EndTime = &end
		}
	}

	if idx, ok := colMap["Priority"]; ok && idx < len(row) {
		switch strings.ToLower(row[idx]) {
		case "high":
			item.Priority = model.PriorityHigh
		case "normal":
			item.Priority = model.PriorityMedium
		case "low":
			item.Priority = model.PriorityLow
		}
	}

	return item, nil
}
