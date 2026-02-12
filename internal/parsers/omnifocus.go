package parsers

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gongahkia/calendar-converter/internal/model"
)

type OmniFocusParser struct{}

func NewOmniFocusParser() *OmniFocusParser {
	return &OmniFocusParser{}
}

func (p *OmniFocusParser) ParseFile(filePath string) (*model.CalendarCollection, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(f, filePath)
}

func (p *OmniFocusParser) Parse(r io.Reader, sourcePath string) (*model.CalendarCollection, error) {
	collection := &model.CalendarCollection{
		Items:            []model.CalendarItem{},
		SourceApp:        "omnifocus",
		ExportDate:       time.Now(),
		OriginalFilePath: sourcePath,
	}

	scanner := bufio.NewScanner(r)
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
		case "tags":
			item.Tags = append(item.Tags, strings.Split(tagValue, ",")...)
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
