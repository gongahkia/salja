package writers

import (
	"encoding/csv"
	
	"io"
	"os"
	"strings"

	"github.com/gongahkia/salja/internal/model"
)

type AsanaWriter struct{}

func NewAsanaWriter() *AsanaWriter {
	return &AsanaWriter{}
}

func (w *AsanaWriter) WriteFile(collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return w.Write(collection, f)
}

func (w *AsanaWriter) Write(collection *model.CalendarCollection, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{"Name", "Section", "Due Date", "Assignee", "Description", "Tags", "Completed"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, item := range collection.Items {
		row := make([]string, 7)
		row[0] = item.Title
		row[1] = ""

		if item.DueDate != nil {
			row[2] = item.DueDate.Format("2006-01-02")
		}

		row[3] = ""
		desc := flattenSubtasksToDescription(item.Description, item.Subtasks)
		row[4] = recurrenceToDescription(desc, item.Recurrence)

		if len(item.Tags) > 0 {
			row[5] = strings.Join(item.Tags, ", ")
		}

		if item.Status == model.StatusCompleted {
			row[6] = "TRUE"
		} else {
			row[6] = "FALSE"
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}
