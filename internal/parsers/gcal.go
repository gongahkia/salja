package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gongahkia/calendar-converter/internal/model"
)

type GoogleCalendarParser struct{}

func NewGoogleCalendarParser() *GoogleCalendarParser {
	return &GoogleCalendarParser{}
}

func (p *GoogleCalendarParser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(f, filePath)
}

func (p *GoogleCalendarParser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	csvReader := csv.NewReader(r)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "gcal"}, nil
	}

	header := records[0]
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "gcal",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	for i := 1; i < len(records); i++ {
		item := parseGCalRow(records[i], colMap)
		collection.Items = append(collection.Items, item)
	}

	return collection, nil
}

func parseGCalRow(row []string, colMap map[string]int) model.CalendarItem {
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

	allDayIdx, hasAllDay := colMap["All Day Event"]
	isAllDay := hasAllDay && allDayIdx < len(row) && strings.ToLower(row[allDayIdx]) == "true"
	item.IsAllDay = isAllDay

	if startDateIdx, ok := colMap["Start Date"]; ok && startDateIdx < len(row) && row[startDateIdx] != "" {
		var start time.Time
		var err error
		
		if isAllDay {
			start, err = time.Parse("01/02/2006", row[startDateIdx])
		} else if startTimeIdx, ok := colMap["Start Time"]; ok && startTimeIdx < len(row) && row[startTimeIdx] != "" {
			dateTime := row[startDateIdx] + " " + row[startTimeIdx]
			start, err = time.Parse("01/02/2006 3:04 PM", dateTime)
			if err != nil {
				start, err = time.Parse("01/02/2006 15:04", dateTime)
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
			end, err = time.Parse("01/02/2006", row[endDateIdx])
		} else if endTimeIdx, ok := colMap["End Time"]; ok && endTimeIdx < len(row) && row[endTimeIdx] != "" {
			dateTime := row[endDateIdx] + " " + row[endTimeIdx]
			end, err = time.Parse("01/02/2006 3:04 PM", dateTime)
			if err != nil {
				end, err = time.Parse("01/02/2006 15:04", dateTime)
			}
		}
		
		if err == nil {
			item.EndTime = &end
		}
	}

	return item
}
