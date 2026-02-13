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

type NotionParser struct{}

func NewNotionParser() *NotionParser {
	return &NotionParser{}
}

func (p *NotionParser) ParseFile(ctx context.Context, filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Notion CSV %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()
	return p.Parse(ctx, f, filePath)
}

func (p *NotionParser) Parse(ctx context.Context, r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	tr, err := transcodeReader(r)
	if err != nil {
		return nil, fmt.Errorf("charset detection failed: %w", err)
	}
	csvReader := csv.NewReader(tr)
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err == io.EOF {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "notion"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV %s: %w", sourcePath, err)
	}

	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	// Notion uses flexible column names; at least one title candidate must exist
	titleCandidates := []string{"Title", "Name", "Task"}
	hasTitle := false
	for _, c := range titleCandidates {
		if _, ok := colMap[c]; ok {
			hasTitle = true
			break
		}
	}
	if !hasTitle {
		return nil, fmt.Errorf("notion CSV %s missing a title column (expected one of: Title, Name, Task)", sourcePath)
	}

	ec := salerr.NewErrorCollector()

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "notion",
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
		item, err := parseNotionRow(row, colMap, ec, sourcePath, lineNum)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", sourcePath, lineNum, err)
		}
		collection.Items = append(collection.Items, item)
	}

	if len(ec.Warnings) > 0 {
		for _, w := range ec.Warnings {
			fmt.Fprintf(os.Stderr, "notion parser: %s\n", w)
		}
	}

	return collection, nil
}

func parseNotionRow(row []string, colMap map[string]int, ec *salerr.ErrorCollector, sourcePath string, lineNum int) (model.CalendarItem, error) {
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
			} else if t, err := parseAmbiguousDate(row[idx]); err == nil {
				item.DueDate = &t
				break
			} else {
				ec.AddWarning((&salerr.ParseError{
					File:    sourcePath,
					Line:    lineNum,
					Message: fmt.Sprintf("malformed date value %q in field %s", row[idx], candidate),
				}).Error())
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

	return item, nil
}
