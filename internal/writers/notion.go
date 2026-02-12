package writers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gongahkia/calendar-converter/internal/model"
)

type NotionWriter struct{}

func NewNotionWriter() *NotionWriter {
	return &NotionWriter{}
}

func (w *NotionWriter) WriteFile(collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return w.Write(collection, f)
}

func (w *NotionWriter) Write(collection *model.CalendarCollection, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{"Title", "Date", "Status", "Tags", "Priority"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, item := range collection.Items {
		row := make([]string, 5)
		row[0] = item.Title

		if item.DueDate != nil {
			row[1] = item.DueDate.Format("2006-01-02")
		}

		switch item.Status {
		case model.StatusCompleted:
			row[2] = "Done"
		case model.StatusInProgress:
			row[2] = "In Progress"
		case model.StatusCancelled:
			row[2] = "Cancelled"
		default:
			row[2] = "Not Started"
		}

		if len(item.Tags) > 0 {
			row[3] = strings.Join(item.Tags, ", ")
		}

		switch item.Priority {
		case model.PriorityHighest, model.PriorityHigh:
			row[4] = "High"
		case model.PriorityMedium:
			row[4] = "Medium"
		case model.PriorityLow, model.PriorityLowest:
			row[4] = "Low"
		default:
			row[4] = ""
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}
