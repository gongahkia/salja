package fidelity

import (
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/model"
	_ "github.com/gongahkia/salja/internal/registry"
)

func TestCheckSubtasksToGcal(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:    "Task with subtasks",
				Subtasks: []model.Subtask{{Title: "Sub1"}},
			},
		},
	}
	warnings := Check(col, "gcal")
	found := false
	for _, w := range warnings {
		if w.Field == "Subtasks" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a Subtasks warning when converting to gcal")
	}
}

func TestCheckRecurrenceToNotion(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:      "Recurring task",
				Recurrence: &model.Recurrence{Freq: model.FreqDaily},
			},
		},
	}
	warnings := Check(col, "notion")
	found := false
	for _, w := range warnings {
		if w.Field == "Recurrence" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a Recurrence warning when converting to notion")
	}
}

func TestCheckRemindersToTrello(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title:     "Task with reminders",
				Reminders: []model.Reminder{{Offset: durationPtr(15 * time.Minute)}},
			},
		},
	}
	warnings := Check(col, "trello")
	found := false
	for _, w := range warnings {
		if w.Field == "Reminders" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a Reminders warning when converting to trello")
	}
}

func TestCheckNoWarnings(t *testing.T) {
	col := &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				Title: "Simple task",
			},
		},
	}
	warnings := Check(col, "ics")
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for simple task to ics, got %d: %v", len(warnings), warnings)
	}
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
