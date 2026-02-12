package writers

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gongahkia/salja/internal/model"
)

type OmniFocusWriter struct{}

func NewOmniFocusWriter() *OmniFocusWriter {
	return &OmniFocusWriter{}
}

func (w *OmniFocusWriter) WriteFile(collection *model.CalendarCollection, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return w.Write(collection, f)
}

func (w *OmniFocusWriter) Write(collection *model.CalendarCollection, writer io.Writer) error {
	for _, item := range collection.Items {
		line := "- " + item.Title

		if item.DueDate != nil {
			line += fmt.Sprintf(" @due(%s)", item.DueDate.Format("2006-01-02"))
		}

		if item.StartTime != nil {
			line += fmt.Sprintf(" @defer(%s)", item.StartTime.Format("2006-01-02"))
		}

		if len(item.Tags) > 0 {
			line += " @tags(" + strings.Join(item.Tags, ",") + ")"
		}

		if item.Status == model.StatusCompleted {
			line += " @done"
		}

		if _, err := fmt.Fprintln(writer, line); err != nil {
			return err
		}

		if item.Description != "" {
			descLines := strings.Split(strings.TrimSpace(item.Description), "\n")
			for _, descLine := range descLines {
				if _, err := fmt.Fprintln(writer, "\t"+descLine); err != nil {
					return err
				}
			}
		}

		for _, subtask := range item.Subtasks {
			subLine := "\t- " + subtask.Title
			if subtask.Status == model.StatusCompleted {
				subLine += " @done"
			}
			if _, err := fmt.Fprintln(writer, subLine); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintln(writer); err != nil {
			return err
		}
	}

	return nil
}
