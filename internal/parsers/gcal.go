package parsers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	salerr "github.com/gongahkia/salja/internal/errors"
	"github.com/gongahkia/salja/internal/model"
)

type GoogleCalendarParser struct{}

func NewGoogleCalendarParser() *GoogleCalendarParser {
	return &GoogleCalendarParser{}
}

func (p *GoogleCalendarParser) ParseFile(ctx context.Context, filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Google Calendar CSV %s: %w", filePath, err)
	}
	defer f.Close()
	return p.Parse(ctx, f, filePath)
}

func (p *GoogleCalendarParser) Parse(ctx context.Context, r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	tr, err := transcodeReader(r)
	if err != nil {
		return nil, fmt.Errorf("charset detection failed: %w", err)
	}
	csvReader := csv.NewReader(tr)
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err == io.EOF {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "gcal"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV %s: %w", sourcePath, err)
	}

	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	requiredCols := []string{"Subject"}
	if missing := findMissingColumns(colMap, requiredCols); len(missing) > 0 {
		return nil, fmt.Errorf("Google Calendar CSV %s missing required columns: %s", sourcePath, strings.Join(missing, ", "))
	}

	ec := salerr.NewErrorCollector()

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "gcal",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	lineNum := 1
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV %s: %w", sourcePath, err)
		}
		lineNum++
		item, err := parseGCalRow(row, colMap, ec, sourcePath, lineNum)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", sourcePath, lineNum, err)
		}
		collection.Items = append(collection.Items, item)
	}

	if len(ec.Warnings) > 0 {
		for _, w := range ec.Warnings {
			fmt.Fprintf(os.Stderr, "gcal parser: %s\n", w)
		}
	}

	return collection, nil
}

func parseGCalRow(row []string, colMap map[string]int, ec *salerr.ErrorCollector, sourcePath string, lineNum int) (model.CalendarItem, error) {
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
		} else if err != nil {
			ec.AddWarning((&salerr.ParseError{
				File:    sourcePath,
				Line:    lineNum,
				Message: fmt.Sprintf("malformed date value %q in field %s", row[startDateIdx], "Start Date"),
			}).Error())
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
		} else if err != nil {
			ec.AddWarning((&salerr.ParseError{
				File:    sourcePath,
				Line:    lineNum,
				Message: fmt.Sprintf("malformed date value %q in field %s", row[endDateIdx], "End Date"),
			}).Error())
		}
	}

	return item, nil
}
