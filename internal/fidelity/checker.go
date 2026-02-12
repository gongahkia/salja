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

		// Recurrence rule dropping warning
		if item.Recurrence != nil && !caps.SupportsRecurrence {
			warnings = append(warnings, DataLossWarning{
				ItemTitle: item.Title,
				Field:     "Recurrence",
				Reason:    fmt.Sprintf("target format '%s' does not support recurrence rules; recurrence will be dropped", targetFormat),
			})
		}

		// Timezone loss warning
		if item.Timezone != "" && item.Timezone != "UTC" {
			// Formats that only support events without timezone fields
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
