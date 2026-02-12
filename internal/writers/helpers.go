package writers

import (
	"fmt"
	"strings"

	"github.com/gongahkia/salja/internal/model"
)

// flattenSubtasksToDescription appends subtasks as a markdown checklist to the description.
func flattenSubtasksToDescription(desc string, subtasks []model.Subtask) string {
	if len(subtasks) == 0 {
		return desc
	}
	var lines []string
	for _, st := range subtasks {
		checkbox := "[ ]"
		if st.Status == model.StatusCompleted {
			checkbox = "[x]"
		}
		lines = append(lines, fmt.Sprintf("- %s %s", checkbox, st.Title))
	}
	checklist := strings.Join(lines, "\n")
	if desc != "" {
		return desc + "\n\n" + checklist
	}
	return checklist
}

// recurrenceToDescription appends a human-readable recurrence summary to the description.
func recurrenceToDescription(desc string, rec *model.Recurrence) string {
	if rec == nil {
		return desc
	}
	summary := formatRecurrenceSummary(rec)
	if summary == "" {
		return desc
	}
	annotation := fmt.Sprintf("[Recurrence: %s]", summary)
	if desc != "" {
		return desc + "\n\n" + annotation
	}
	return annotation
}

func formatRecurrenceSummary(rec *model.Recurrence) string {
	if rec.Freq == "" {
		return ""
	}
	var parts []string
	freq := strings.ToLower(string(rec.Freq))
	if rec.Interval > 1 {
		parts = append(parts, fmt.Sprintf("Every %d %s", rec.Interval, freq))
	} else {
		parts = append(parts, string(rec.Freq))
	}
	if len(rec.ByDay) > 0 {
		var days []string
		for _, d := range rec.ByDay {
			days = append(days, string(d))
		}
		parts = append(parts, fmt.Sprintf("on %s", strings.Join(days, ", ")))
	}
	if rec.Count != nil {
		parts = append(parts, fmt.Sprintf("for %d occurrences", *rec.Count))
	}
	if rec.Until != nil {
		parts = append(parts, fmt.Sprintf("until %s", rec.Until.Format("2006-01-02")))
	}
	return strings.Join(parts, " ")
}
