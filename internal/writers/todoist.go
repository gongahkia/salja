package writers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/gongahkia/salja/internal/model"
)

type TodoistWriter struct{}

func NewTodoistWriter() *TodoistWriter {
	return &TodoistWriter{}
}

func (w *TodoistWriter) WriteFile(collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create Todoist CSV: %w", err)
	}
	defer f.Close()

	return w.Write(collection, f)
}

func (w *TodoistWriter) Write(collection *model.CalendarCollection, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{
		"TYPE", "CONTENT", "DESCRIPTION", "PRIORITY", "INDENT",
		"AUTHOR", "RESPONSIBLE", "DATE", "DATE_LANG", "TIMEZONE",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, item := range collection.Items {
		rows := w.itemToRows(&item, 0)
		for _, row := range rows {
			if err := csvWriter.Write(row); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *TodoistWriter) itemToRows(item *model.CalendarItem, indent int) [][]string {
	var rows [][]string

	mainRow := make([]string, 10)
	mainRow[0] = "task"
	mainRow[1] = item.Title
	mainRow[2] = item.Description
	mainRow[3] = exportTodoistPriority(item.Priority)
	mainRow[4] = fmt.Sprintf("%d", indent)
	mainRow[5] = ""
	mainRow[6] = ""

	if item.DueDate != nil {
		mainRow[7] = item.DueDate.Format("2006-01-02")
	}

	mainRow[8] = "en"

	if item.Timezone != "" {
		mainRow[9] = item.Timezone
	}

	rows = append(rows, mainRow)

	for _, subtask := range item.Subtasks {
		subtaskRow := make([]string, 10)
		subtaskRow[0] = "task"
		subtaskRow[1] = subtask.Title
		subtaskRow[2] = ""
		subtaskRow[3] = "1"
		subtaskRow[4] = fmt.Sprintf("%d", indent+1)
		subtaskRow[5] = ""
		subtaskRow[6] = ""
		subtaskRow[7] = ""
		subtaskRow[8] = "en"
		subtaskRow[9] = ""
		rows = append(rows, subtaskRow)
	}

	return rows
}

func exportTodoistPriority(p model.Priority) string {
	switch p {
	case model.PriorityHighest:
		return "4"
	case model.PriorityHigh:
		return "3"
	case model.PriorityMedium:
		return "2"
	case model.PriorityLow:
		return "1"
	default:
		return "1"
	}
}
