package writers

import (
	"encoding/csv"
	
	"io"
	"os"

	"github.com/gongahkia/salja/internal/model"
)

type GoogleCalendarWriter struct{}

func NewGoogleCalendarWriter() *GoogleCalendarWriter {
	return &GoogleCalendarWriter{}
}

func (w *GoogleCalendarWriter) WriteFile(collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return w.Write(collection, f)
}

func (w *GoogleCalendarWriter) Write(collection *model.CalendarCollection, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{
		"Subject", "Start Date", "Start Time", "End Date", "End Time",
		"All Day Event", "Description", "Location", "Private",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, item := range collection.Items {
		row := make([]string, 9)
		row[0] = item.Title

		if item.StartTime != nil {
			if item.IsAllDay {
				row[1] = item.StartTime.Format("01/02/2006")
				row[2] = ""
			} else {
				row[1] = item.StartTime.Format("01/02/2006")
				row[2] = item.StartTime.Format("3:04 PM")
			}
		}

		if item.EndTime != nil {
			if item.IsAllDay {
				row[3] = item.EndTime.Format("01/02/2006")
				row[4] = ""
			} else {
				row[3] = item.EndTime.Format("01/02/2006")
				row[4] = item.EndTime.Format("3:04 PM")
			}
		}

		if item.IsAllDay {
			row[5] = "True"
		} else {
			row[5] = "False"
		}

		row[6] = flattenSubtasksToDescription(item.Description, item.Subtasks)
		row[7] = item.Location
		row[8] = "False"

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}
