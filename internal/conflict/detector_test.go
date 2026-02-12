package conflict

import (
"testing"
"time"

"github.com/gongahkia/calendar-converter/internal/model"
)

func TestExactUIDMatch(t *testing.T) {
d := NewDetector()
start := time.Now()
src := &model.CalendarCollection{Items: []model.CalendarItem{{UID: "abc-123", Title: "Test", StartTime: &start}}}
tgt := &model.CalendarCollection{Items: []model.CalendarItem{{UID: "abc-123", Title: "Test", StartTime: &start}}}

matches := d.FindDuplicates(src, tgt)
if len(matches) != 1 {
t.Fatalf("expected 1 match, got %d", len(matches))
}
}

func TestFuzzyTitleDateMatch(t *testing.T) {
d := NewDetector()
start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
src := &model.CalendarCollection{Items: []model.CalendarItem{{Title: "Weekly Team Meeting", StartTime: &start}}}
tgt := &model.CalendarCollection{Items: []model.CalendarItem{{Title: "Weekly Team Meeting", StartTime: &start}}}

matches := d.FindDuplicates(src, tgt)
if len(matches) != 1 {
t.Fatalf("expected 1 match, got %d", len(matches))
}
}

func TestNoMatch(t *testing.T) {
d := NewDetector()
s1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
s2 := time.Date(2024, 6, 20, 14, 0, 0, 0, time.UTC)
src := &model.CalendarCollection{Items: []model.CalendarItem{{Title: "Meeting A", StartTime: &s1}}}
tgt := &model.CalendarCollection{Items: []model.CalendarItem{{Title: "Different Event", StartTime: &s2}}}

matches := d.FindDuplicates(src, tgt)
if len(matches) != 0 {
t.Fatalf("expected 0 matches, got %d", len(matches))
}
}

func TestDataLossSubtaskFlattening(t *testing.T) {
checker := NewDataLossChecker()
items := []model.CalendarItem{
{Title: "Task", Subtasks: []model.Subtask{{Title: "Sub1"}, {Title: "Sub2"}}},
}
warnings := checker.Check(items, "gcal")
if len(warnings) == 0 {
t.Fatal("expected data loss warning for subtasks -> gcal")
}
if warnings[0].Field != "subtasks" {
t.Errorf("expected subtasks warning, got %s", warnings[0].Field)
}
}

func TestDataLossRecurrenceDrop(t *testing.T) {
checker := NewDataLossChecker()
items := []model.CalendarItem{
{Title: "Recurring", Recurrence: &model.Recurrence{Freq: model.FreqDaily}},
}
warnings := checker.Check(items, "trello")
if len(warnings) == 0 {
t.Fatal("expected data loss warning for recurrence -> trello")
}
}
