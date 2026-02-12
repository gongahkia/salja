package writers

import (
	"encoding/csv"
	
	"io"
	"os"
	"strings"

	"github.com/gongahkia/calendar-converter/internal/model"
)

type OutlookWriter struct{}

func NewOutlookWriter() *OutlookWriter {
	return &OutlookWriter{}
}

func (w *OutlookWriter) WriteFile(collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return w.Write(collection, f)
}

func (w *OutlookWriter) Write(collection *model.CalendarCollection, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	header := []string{
		"Subject", "Start Date", "Start Time", "End Date", "End Time",
		"All day event", "Reminder on/off", "Reminder Date", "Reminder Time",
		"Categories", "Description", "Location", "Priority",
	}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	for _, item := range collection.Items {
		row := make([]string, 13)
		row[0] = item.Title

		if item.StartTime != nil {
			if item.IsAllDay {
				row[1] = item.StartTime.Format("1/2/2006")
				row[2] = ""
			} else {
				row[1] = item.StartTime.Format("1/2/2006")
				row[2] = item.StartTime.Format("3:04:05 PM")
			}
		}

		if item.EndTime != nil {
			if item.IsAllDay {
				row[3] = item.EndTime.Format("1/2/2006")
				row[4] = ""
			} else {
				row[3] = item.EndTime.Format("1/2/2006")
				row[4] = item.EndTime.Format("3:04:05 PM")
			}
		}

		if item.IsAllDay {
			row[5] = "True"
		} else {
			row[5] = "False"
		}

		row[6] = "False"
		row[7] = ""
		row[8] = ""

		if len(item.Tags) > 0 {
			row[9] = strings.Join(item.Tags, "; ")
		}

		row[10] = item.Description
		row[11] = item.Location

		switch item.Priority {
		case model.PriorityHigh, model.PriorityHighest:
			row[12] = "High"
		case model.PriorityMedium:
			row[12] = "Normal"
		case model.PriorityLow, model.PriorityLowest:
			row[12] = "Low"
		default:
			row[12] = "Normal"
		}

		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	return nil
}
