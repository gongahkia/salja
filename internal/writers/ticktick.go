package writers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type TickTickWriter struct{}

func NewTickTickWriter() *TickTickWriter {
	return &TickTickWriter{}
}

func (w *TickTickWriter) WriteFile(collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create TickTick CSV: %w", err)
	}
	defer f.Close()

	return w.Write(collection, f)
}

func (w *TickTickWriter) Write(collection *model.CalendarCollection, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{
		"folder", "list", "title", "tags", "content", "is_checklist",
		"start_date", "due_date", "reminder", "repeat", "priority",
		"status", "created_time", "completed_time", "timezone", "is_all_day",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, item := range collection.Items {
		row := w.itemToRow(&item)
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (w *TickTickWriter) itemToRow(item *model.CalendarItem) []string {
	row := make([]string, 16)

	row[0] = ""
	row[1] = ""
	row[2] = item.Title

	if len(item.Tags) > 0 {
		row[3] = strings.Join(item.Tags, ", ")
	}

	content := item.Description
	if len(item.Subtasks) > 0 {
		var checklistItems []string
		for _, subtask := range item.Subtasks {
			checkbox := "[ ]"
			if subtask.Status == model.StatusCompleted {
				checkbox = "[x]"
			}
			checklistItems = append(checklistItems, fmt.Sprintf("- %s %s", checkbox, subtask.Title))
		}
		content = strings.Join(checklistItems, "\n")
		row[5] = "true"
	} else {
		row[5] = "false"
	}
	row[4] = content

	if item.StartTime != nil {
		row[6] = item.StartTime.Format(time.RFC3339)
	}

	if item.DueDate != nil {
		row[7] = item.DueDate.Format(time.RFC3339)
	}

	row[8] = ""

	if item.Recurrence != nil {
		row[9] = exportTickTickRepeat(item.Recurrence)
	}

	row[10] = exportTickTickPriority(item.Priority)

	if item.Status == model.StatusCompleted {
		row[11] = "2"
	} else {
		row[11] = "0"
	}

	row[12] = ""

	if item.CompletionDate != nil {
		row[13] = item.CompletionDate.Format(time.RFC3339)
	}

	if item.Timezone != "" {
		row[14] = item.Timezone
	}

	if item.IsAllDay {
		row[15] = "true"
	} else {
		row[15] = "false"
	}

	return row
}

func exportTickTickPriority(p model.Priority) string {
	switch p {
	case model.PriorityNone, model.PriorityLowest:
		return "0"
	case model.PriorityLow:
		return "1"
	case model.PriorityMedium:
		return "3"
	case model.PriorityHigh, model.PriorityHighest:
		return "5"
	default:
		return "0"
	}
}

func exportTickTickRepeat(rec *model.Recurrence) string {
	switch rec.Freq {
	case model.FreqDaily:
		return "DAILY"
	case model.FreqWeekly:
		return "WEEKLY"
	case model.FreqMonthly:
		return "MONTHLY"
	case model.FreqYearly:
		return "YEARLY"
	default:
		return ""
	}
}
