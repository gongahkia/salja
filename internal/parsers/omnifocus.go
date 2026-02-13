package parsers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type OmniFocusParser struct{}

func NewOmniFocusParser() *OmniFocusParser {
	return &OmniFocusParser{}
}

func (p *OmniFocusParser) ParseFile(ctx context.Context, filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(ctx, f, filePath)
}

func (p *OmniFocusParser) Parse(ctx context.Context, r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	tr, err := transcodeReader(r)
	if err != nil {
		return nil, fmt.Errorf("charset detection failed: %w", err)
	}

	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "omnifocus",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	scanner := bufio.NewScanner(tr)
	var currentItem *model.CalendarItem
	_ = 0 // indent tracking reserved for nested support

	for scanner.Scan() {
		line := scanner.Text()
		
		currentIndent := 0
		for strings.HasPrefix(line, "\t") {
			currentIndent++
			line = strings.TrimPrefix(line, "\t")
		}

		if strings.HasPrefix(line, "- ") {
			if currentItem != nil {
				collection.Items = append(collection.Items, *currentItem)
			}

			line = strings.TrimPrefix(line, "- ")
			item := model.CalendarItem{
				ItemType: model.ItemTypeTask,
				Status:   model.StatusPending,
			}

			parts := strings.SplitN(line, " @", 2)
			item.Title = strings.TrimSpace(parts[0])

			if len(parts) > 1 {
				parseOmniFocusTags(&item, "@"+parts[1])
			}

			currentItem = &item
			_ = currentIndent
		} else if currentItem != nil && strings.TrimSpace(line) != "" {
			currentItem.Description += strings.TrimSpace(line) + "\n"
		}
	}

	if currentItem != nil {
		collection.Items = append(collection.Items, *currentItem)
	}

	return collection, nil
}

func parseOmniFocusTags(item *model.CalendarItem, tagStr string) {
	tagRe := regexp.MustCompile(`@(\w+)(?:\(([^)]+)\))?`)
	matches := tagRe.FindAllStringSubmatch(tagStr, -1)

	for _, match := range matches {
		tagName := match[1]
		tagValue := ""
		if len(match) > 2 {
			tagValue = match[2]
		}

		switch strings.ToLower(tagName) {
		case "due":
			if t, err := parseOmniFocusDate(tagValue); err == nil {
				item.DueDate = &t
			}
		case "defer":
			if t, err := parseOmniFocusDate(tagValue); err == nil {
				item.StartTime = &t
			}
		case "done":
			item.Status = model.StatusCompleted
		case "flagged":
			item.Priority = model.PriorityHigh
		case "priority":
			switch strings.ToLower(tagValue) {
			case "high", "1":
				item.Priority = model.PriorityHigh
			case "medium", "2":
				item.Priority = model.PriorityMedium
			case "low", "3":
				item.Priority = model.PriorityLow
			}
		case "tags":
			item.Tags = append(item.Tags, strings.Split(tagValue, ",")...)
		case "note", "notes":
			item.Description += tagValue + "\n"
		default:
			item.Tags = append(item.Tags, tagName)
		}
	}
}

func parseOmniFocusDate(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"Jan 2, 2006",
		"January 2, 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}
