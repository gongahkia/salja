package parsers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	salerr "github.com/gongahkia/salja/internal/errors"
	"github.com/gongahkia/salja/internal/model"
)

type TodoistParser struct{}

func NewTodoistParser() *TodoistParser {
	return &TodoistParser{}
}

func (p *TodoistParser) ParseFile(ctx context.Context, filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Todoist CSV: %w", err)
	}
	defer f.Close()

	return p.Parse(ctx, f, filePath)
}

func (p *TodoistParser) Parse(ctx context.Context, r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	tr, err := transcodeReader(r)
	if err != nil {
		return nil, fmt.Errorf("charset detection failed: %w", err)
	}
	csvReader := csv.NewReader(tr)
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err == io.EOF {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "todoist"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	ec := salerr.NewErrorCollector()

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "todoist",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	var taskStack []struct {
		item   *model.CalendarItem
		indent int
	}

	lineNum := 1
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV: %w", err)
		}
		lineNum++

		typeIdx, hasType := colMap["TYPE"]
		if hasType && typeIdx < len(row) && row[typeIdx] != "task" {
			ec.AddWarning(fmt.Sprintf("%s line %d: skipping non-task row (TYPE=%q)", sourcePath, lineNum, row[typeIdx]))
			continue
		}

		item := p.parseRow(row, colMap, ec, sourcePath, lineNum)

		indentIdx, hasIndent := colMap["INDENT"]
		currentIndent := 0
		if hasIndent && indentIdx < len(row) {
			currentIndent, _ = strconv.Atoi(row[indentIdx])
		}

		for len(taskStack) > 0 && taskStack[len(taskStack)-1].indent >= currentIndent {
			taskStack = taskStack[:len(taskStack)-1]
		}

		if len(taskStack) > 0 && currentIndent > taskStack[len(taskStack)-1].indent {
			parent := taskStack[len(taskStack)-1].item
			parent.Subtasks = append(parent.Subtasks, model.Subtask{
				Title:     item.Title,
				Status:    item.Status,
				Priority:  item.Priority,
				SortOrder: len(parent.Subtasks),
			})
		} else {
			collection.Items = append(collection.Items, item)
		}

		taskStack = append(taskStack, struct {
			item   *model.CalendarItem
			indent int
		}{item: &item, indent: currentIndent})
	}

	if len(ec.Warnings) > 0 {
		for _, w := range ec.Warnings {
			fmt.Fprintf(os.Stderr, "todoist parser: %s\n", w)
		}
	}

	return collection, nil
}

func (p *TodoistParser) parseRow(row []string, colMap map[string]int, ec *salerr.ErrorCollector, sourcePath string, lineNum int) model.CalendarItem {
	item := model.CalendarItem{
		ItemType: model.ItemTypeTask,
		Status:   model.StatusPending,
	}

	if idx, ok := colMap["CONTENT"]; ok && idx < len(row) {
		item.Title = row[idx]
	}

	if idx, ok := colMap["DESCRIPTION"]; ok && idx < len(row) {
		item.Description = row[idx]
	}

	if idx, ok := colMap["PRIORITY"]; ok && idx < len(row) {
		item.Priority = mapTodoistPriority(row[idx])
	}

	if idx, ok := colMap["DATE"]; ok && idx < len(row) && row[idx] != "" {
		if t, err := time.Parse("2006-01-02", row[idx]); err == nil {
			item.DueDate = &t
		} else if t, err := time.Parse(time.RFC3339, row[idx]); err == nil {
			item.DueDate = &t
		} else if t, err := parseAmbiguousDate(row[idx]); err == nil {
			item.DueDate = &t
		} else {
			ec.AddWarning((&salerr.ParseError{
				File:    sourcePath,
				Line:    lineNum,
				Message: fmt.Sprintf("malformed date value %q in field %s", row[idx], "DATE"),
			}).Error())
		}
	}

	if idx, ok := colMap["TIMEZONE"]; ok && idx < len(row) {
		item.Timezone = row[idx]
	}

	return item
}

func mapTodoistPriority(val string) model.Priority {
	priority, err := strconv.Atoi(val)
	if err != nil {
		return model.PriorityNone
	}

	switch priority {
	case 4:
		return model.PriorityHighest
	case 3:
		return model.PriorityHigh
	case 2:
		return model.PriorityMedium
	case 1:
		return model.PriorityLow
	default:
		return model.PriorityNone
	}
}
