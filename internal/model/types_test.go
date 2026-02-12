package model

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		item    CalendarItem
		wantErr bool
	}{
		{"valid", CalendarItem{Title: "Test"}, false},
		{"empty title", CalendarItem{}, true},
		{"priority too high", CalendarItem{Title: "T", Priority: 6}, true},
		{"priority negative", CalendarItem{Title: "T", Priority: -1}, true},
		{"start after end", CalendarItem{
			Title:     "T",
			StartTime: timePtr(time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)),
			EndTime:   timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		}, true},
		{"valid times", CalendarItem{
			Title:     "T",
			StartTime: timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			EndTime:   timePtr(time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)),
		}, false},
		{"invalid type", CalendarItem{Title: "T", ItemType: "bad"}, true},
		{"valid type", CalendarItem{Title: "T", ItemType: ItemTypeEvent}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestHelpers(t *testing.T) {
	item := CalendarItem{Title: "Test", Status: StatusCompleted}
	if !item.IsCompleted() {
		t.Error("expected IsCompleted true")
	}
	item.Status = StatusPending
	if item.IsCompleted() {
		t.Error("expected IsCompleted false")
	}

	if item.HasRecurrence() {
		t.Error("expected HasRecurrence false")
	}
	item.Recurrence = &Recurrence{Freq: FreqDaily}
	if !item.HasRecurrence() {
		t.Error("expected HasRecurrence true")
	}

	if item.HasSubtasks() {
		t.Error("expected HasSubtasks false")
	}
	item.Subtasks = []Subtask{{Title: "sub"}}
	if !item.HasSubtasks() {
		t.Error("expected HasSubtasks true")
	}
}

func TestDuration(t *testing.T) {
	item := CalendarItem{Title: "T"}
	if item.Duration() != 0 {
		t.Error("expected 0 duration with nil times")
	}
	start := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 1, 11, 30, 0, 0, time.UTC)
	item.StartTime = &start
	item.EndTime = &end
	if item.Duration() != 90*time.Minute {
		t.Errorf("expected 90m, got %v", item.Duration())
	}
}

func timePtr(t time.Time) *time.Time { return &t }
