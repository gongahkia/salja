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
		// event → task-only format
		if item.ItemType == model.ItemTypeEvent && !caps.SupportsEvents && caps.SupportsTasks {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "ItemType",
				Reason:    fmt.Sprintf("event will be converted to task (target '%s' does not support events)", targetFormat),
			})
		}

		// task → event-only format
		if item.ItemType == model.ItemTypeTask && !caps.SupportsTasks && caps.SupportsEvents {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "ItemType",
				Reason:    fmt.Sprintf("task will be converted to event (target '%s' does not support tasks)", targetFormat),
			})
		}

		// subtask flattening
		if len(item.Subtasks) > 0 && !caps.SupportsSubtasks {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "Subtasks",
				Reason:    fmt.Sprintf("target format '%s' does not support subtasks; %d subtask(s) will be flattened or lost", targetFormat, len(item.Subtasks)),
			})
			for _, sub := range item.Subtasks {
				if sub.Priority > 0 {
					warnings = append(warnings, DataLossWarning{
						ItemTitle: item.Title,
						Field:     "SubtaskPriority",
						Reason:    fmt.Sprintf("subtask '%s' has priority %d which will be dropped during conversion", sub.Title, sub.Priority),
					})
					break
				}
			}
		}

		// recurrence dropping
		if item.Recurrence != nil && !caps.SupportsRecurrence {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "Recurrence",
				Reason:    fmt.Sprintf("target format '%s' does not support recurrence rules; recurrence will be dropped", targetFormat),
			})
		}

		// reminder loss (registry-driven, not hardcoded)
		if len(item.Reminders) > 0 && !caps.SupportsReminders {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "Reminders",
				Reason:    fmt.Sprintf("%d reminder(s) will be lost converting to '%s' (no reminder support)", len(item.Reminders), targetFormat),
			})
		}

		// todoist priority collision (format-specific)
		if targetFormat == "todoist" && item.Priority == model.PriorityLowest {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "Priority",
				Reason:    "PriorityLowest (1) maps to Todoist priority '1' (same as PriorityNone); priority distinction will be lost",
			})
		}

		// timezone loss
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
