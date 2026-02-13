package fidelity

import (
	"fmt"

	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/registry"
)

type DataLossWarning struct {
	ItemTitle string
	Field     string
	Reason    string
}

func (w DataLossWarning) String() string {
	return fmt.Sprintf("[%s] %s: %s", w.ItemTitle, w.Field, w.Reason)
}

// Check compares source items against target format capabilities and returns warnings.
func Check(collection *model.CalendarCollection, targetFormat string) []DataLossWarning {
	caps, ok := registry.GetCapabilities(targetFormat)
	if !ok {
		return nil
	}

	var warnings []DataLossWarning

	for _, item := range collection.Items {
		// Subtask flattening warning
		if len(item.Subtasks) > 0 && !caps.SupportsSubtasks {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "Subtasks",
				Reason:    fmt.Sprintf("target format '%s' does not support subtasks; %d subtask(s) will be flattened or lost", targetFormat, len(item.Subtasks)),
			})
		}

		// Subtask priority loss warning
		if len(item.Subtasks) > 0 {
			for _, sub := range item.Subtasks {
				if sub.Priority > 0 && !caps.SupportsSubtasks {
					warnings = append(warnings, DataLossWarning{
						ItemTitle: item.Title,
						Field:     "SubtaskPriority",
						Reason:    fmt.Sprintf("subtask '%s' has priority %d which will be dropped during conversion", sub.Title, sub.Priority),
					})
					break
				}
			}
		}

		// Recurrence rule dropping warning
		if item.Recurrence != nil && !caps.SupportsRecurrence {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "Recurrence",
				Reason:    fmt.Sprintf("target format '%s' does not support recurrence rules; recurrence will be dropped", targetFormat),
			})
		}

		// Reminder loss warning for CSV-based formats
		if len(item.Reminders) > 0 {
			csvFormats := map[string]bool{"todoist": true, "ticktick": true, "gcal": true, "outlook": true, "asana": true, "notion": true, "trello": true}
			if csvFormats[targetFormat] {
				warnings = append(warnings, DataLossWarning{
					ItemTitle: item.Title,
					Field:     "Reminders",
					Reason:    fmt.Sprintf("%d reminder(s) will be lost converting to '%s' (no reminder support)", len(item.Reminders), targetFormat),
				})
			}
		}

		// Priority mapping collision detection for todoist
		if targetFormat == "todoist" && (item.Priority == model.PriorityNone || item.Priority == model.PriorityLowest) {
			if item.Priority == model.PriorityLowest {
				warnings = append(warnings, DataLossWarning{
					ItemTitle: item.Title,
					Field:     "Priority",
					Reason:    "PriorityLowest (1) maps to Todoist priority '1' (same as PriorityNone); priority distinction will be lost",
				})
			}
		}

		// Timezone loss warning
		if item.Timezone != "" && item.Timezone != "UTC" {
			if !caps.SupportsEvents && !caps.SupportsRecurrence {
				warnings = append(warnings, DataLossWarning{
					ItemTitle: item.Title,
					Field:     "Timezone",
					Reason:    fmt.Sprintf("timezone '%s' may be lost when converting to '%s'", item.Timezone, targetFormat),
				})
			}
		}
	}

	return warnings
}
