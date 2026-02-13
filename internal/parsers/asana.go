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

type AsanaParser struct{}

func NewAsanaParser() *AsanaParser {
	return &AsanaParser{}
}

func (p *AsanaParser) ParseFile(ctx context.Context, filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Asana CSV %s: %w", filePath, err)
	}
	defer func() { _ = f.Close() }()
	return p.Parse(ctx, f, filePath)
}

func (p *AsanaParser) Parse(ctx context.Context, r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	tr, err := transcodeReader(r)
	if err != nil {
		return nil, fmt.Errorf("charset detection failed: %w", err)
	}
	csvReader := csv.NewReader(tr)
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true

	header, err := csvReader.Read()
	if err == io.EOF {
		return &model.CalendarCollection{Items: []model.CalendarItem{}, SourceApp: "asana"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV %s: %w", sourcePath, err)
	}

	colMap := make(map[string]int)
	for i, col := range header {
		colMap[col] = i
	}

	requiredCols := []string{"Name"}
	if missing := findMissingColumns(colMap, requiredCols); len(missing) > 0 {
		return nil, fmt.Errorf("Asana CSV %s missing required columns: %s", sourcePath, strings.Join(missing, ", "))
	}

	ec := salerr.NewErrorCollector()

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "asana",
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
		item, err := parseAsanaRow(row, colMap, ec, sourcePath, lineNum)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", sourcePath, lineNum, err)
		}
		collection.Items = append(collection.Items, item)
	}

	if len(ec.Warnings) > 0 {
		for _, w := range ec.Warnings {
			fmt.Fprintf(os.Stderr, "asana parser: %s\n", w)
		}
	}

	return collection, nil
}

func parseAsanaRow(row []string, colMap map[string]int, ec *salerr.ErrorCollector, sourcePath string, lineNum int) (model.CalendarItem, error) {
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
		} else if t, err := parseAmbiguousDate(row[idx]); err == nil {
			item.DueDate = &t
		} else {
			ec.AddWarning((&salerr.ParseError{
				File:    sourcePath,
				Line:    lineNum,
				Message: fmt.Sprintf("malformed date value %q in field %s", row[idx], "Due Date"),
			}).Error())
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

	return item, nil
}
