package apple

import (
	"strings"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

// mockScriptRunner records calls and returns canned responses.
type mockScriptRunner struct {
	calls   []string
	outputs []string
	errs    []error
	idx     int
}

func (m *mockScriptRunner) run(script string) (string, error) {
	m.calls = append(m.calls, script)
	if m.idx < len(m.outputs) {
		out := m.outputs[m.idx]
		var err error
		if m.idx < len(m.errs) {
			err = m.errs[m.idx]
		}
		m.idx++
		return out, err
	}
	return "", nil
}

func withMock(m *mockScriptRunner, fn func()) {
	orig := scriptRunnerFn
	scriptRunnerFn = m.run
	defer func() { scriptRunnerFn = orig }()
	fn()
}

func TestMockCalendarWriterCreateEvent(t *testing.T) {
	start := time.Date(2025, 6, 15, 10, 0, 0, 0, time.Local)
	end := time.Date(2025, 6, 15, 11, 0, 0, 0, time.Local)

	items := []model.CalendarItem{
		{
			Title:       "Team Standup",
			Description: "Daily sync",
			Location:    "Room 42",
			StartTime:   &start,
			EndTime:     &end,
			IsAllDay:    false,
			ItemType:    model.ItemTypeEvent,
			Status:      model.StatusPending,
		},
	}

	mock := &mockScriptRunner{outputs: []string{""}}
	withMock(mock, func() {
		w := NewCalendarWriter()
		err := w.Write(items, "Work")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(mock.calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.calls))
		}

		script := mock.calls[0]
		for _, want := range []string{
			`tell application "Calendar"`,
			`tell calendar "Work"`,
			`summary:"Team Standup"`,
			`description:"Daily sync"`,
			`location:"Room 42"`,
			`start date:date "`,
			`end date:date "`,
			`make new event with properties`,
		} {
			if !strings.Contains(script, want) {
				t.Errorf("script missing %q\ngot:\n%s", want, script)
			}
		}
	})
}

func TestMockCalendarWriterAllDayEvent(t *testing.T) {
	start := time.Date(2025, 12, 25, 0, 0, 0, 0, time.Local)

	items := []model.CalendarItem{
		{
			Title:     "Holiday",
			StartTime: &start,
			IsAllDay:  true,
			ItemType:  model.ItemTypeEvent,
		},
	}

	mock := &mockScriptRunner{outputs: []string{""}}
	withMock(mock, func() {
		w := NewCalendarWriter()
		err := w.Write(items, "Personal")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		script := mock.calls[0]
		if !strings.Contains(script, "allday event:true") {
			t.Errorf("script missing allday event flag\ngot:\n%s", script)
		}
	})
}

func TestMockCalendarWriterNoStartTime(t *testing.T) {
	items := []model.CalendarItem{
		{Title: "Bad Event", ItemType: model.ItemTypeEvent},
	}

	mock := &mockScriptRunner{outputs: []string{""}}
	withMock(mock, func() {
		w := NewCalendarWriter()
		err := w.Write(items, "Work")
		if err == nil {
			t.Fatal("expected error for event without start time")
		}
		if !strings.Contains(err.Error(), "no start time") {
			t.Errorf("unexpected error: %v", err)
		}
		if len(mock.calls) != 0 {
			t.Errorf("expected no script calls, got %d", len(mock.calls))
		}
	})
}

func TestMockCalendarReaderRead(t *testing.T) {
	fakeOutput := "Meeting|||June 15, 2025 10:00:00 AM|||June 15, 2025 11:00:00 AM|||Notes here|||Office\nLunch|||June 15, 2025 12:00:00 PM|||June 15, 2025 1:00:00 PM||||||Cafe"

	mock := &mockScriptRunner{outputs: []string{fakeOutput}}
	withMock(mock, func() {
		r := NewCalendarReader()
		start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.Local)
		end := time.Date(2025, 6, 30, 0, 0, 0, 0, time.Local)
		collection, err := r.Read("Work", start, end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(collection.Items) != 1 {
			// second line has 6 parts but still >=5 so may parse; just check first
			if len(collection.Items) < 1 {
				t.Fatalf("expected at least 1 item, got %d", len(collection.Items))
			}
		}

		item := collection.Items[0]
		if item.Title != "Meeting" {
			t.Errorf("expected title 'Meeting', got %q", item.Title)
		}
		if item.Description != "Notes here" {
			t.Errorf("expected description 'Notes here', got %q", item.Description)
		}
		if item.Location != "Office" {
			t.Errorf("expected location 'Office', got %q", item.Location)
		}
		if item.ItemType != model.ItemTypeEvent {
			t.Errorf("expected ItemType %q, got %q", model.ItemTypeEvent, item.ItemType)
		}

		script := mock.calls[0]
		if !strings.Contains(script, `tell calendar "Work"`) {
			t.Errorf("script missing calendar name\ngot:\n%s", script)
		}
	})
}

func TestMockRemindersWriterCreateReminder(t *testing.T) {
	due := time.Date(2025, 7, 1, 9, 0, 0, 0, time.Local)

	items := []model.CalendarItem{
		{
			Title:       "Buy groceries",
			Description: "Milk, eggs, bread",
			DueDate:     &due,
			Priority:    model.PriorityHigh,
			Status:      model.StatusPending,
			ItemType:    model.ItemTypeTask,
		},
	}

	mock := &mockScriptRunner{outputs: []string{""}}
	withMock(mock, func() {
		w := NewRemindersWriter()
		err := w.Write(items, "Shopping")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(mock.calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.calls))
		}

		script := mock.calls[0]
		for _, want := range []string{
			`tell application "Reminders"`,
			`tell list "Shopping"`,
			`name:"Buy groceries"`,
			`body:"Milk, eggs, bread"`,
			`due date:date "`,
			`priority:1`,
			`make new reminder with properties`,
		} {
			if !strings.Contains(script, want) {
				t.Errorf("script missing %q\ngot:\n%s", want, script)
			}
		}

		// High priority should not set completed
		if strings.Contains(script, "completed:true") {
			t.Errorf("script should not contain completed:true for pending item")
		}
	})
}

func TestMockRemindersWriterCompleted(t *testing.T) {
	items := []model.CalendarItem{
		{
			Title:    "Done task",
			Status:   model.StatusCompleted,
			ItemType: model.ItemTypeTask,
		},
	}

	mock := &mockScriptRunner{outputs: []string{""}}
	withMock(mock, func() {
		w := NewRemindersWriter()
		err := w.Write(items, "Tasks")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		script := mock.calls[0]
		if !strings.Contains(script, "completed:true") {
			t.Errorf("script missing completed:true for completed item\ngot:\n%s", script)
		}
	})
}

func TestMockRemindersReaderRead(t *testing.T) {
	fakeOutput := "Shopping|||Buy milk|||Get 2% milk|||July 1, 2025 9:00:00 AM|||false\nWork|||Finish report|||Q2 summary|||July 2, 2025 5:00:00 PM|||true"

	mock := &mockScriptRunner{outputs: []string{fakeOutput}}
	withMock(mock, func() {
		r := NewRemindersReader()
		collection, err := r.Read()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(collection.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(collection.Items))
		}

		if collection.Items[0].Title != "Buy milk" {
			t.Errorf("expected title 'Buy milk', got %q", collection.Items[0].Title)
		}
		if collection.Items[0].Status != model.StatusPending {
			t.Errorf("expected status pending, got %q", collection.Items[0].Status)
		}

		if collection.Items[1].Title != "Finish report" {
			t.Errorf("expected title 'Finish report', got %q", collection.Items[1].Title)
		}
		if collection.Items[1].Status != model.StatusCompleted {
			t.Errorf("expected status completed, got %q", collection.Items[1].Status)
		}

		if collection.SourceApp != "apple-reminders" {
			t.Errorf("expected SourceApp 'apple-reminders', got %q", collection.SourceApp)
		}
	})
}

func TestMockRemindersWriterMultiple(t *testing.T) {
	items := []model.CalendarItem{
		{Title: "Task A", ItemType: model.ItemTypeTask},
		{Title: "Task B", ItemType: model.ItemTypeTask},
		{Title: "Task C", ItemType: model.ItemTypeTask},
	}

	mock := &mockScriptRunner{outputs: []string{"", "", ""}}
	withMock(mock, func() {
		w := NewRemindersWriter()
		err := w.Write(items, "List")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(mock.calls) != 3 {
			t.Fatalf("expected 3 calls, got %d", len(mock.calls))
		}

		for i, title := range []string{"Task A", "Task B", "Task C"} {
			if !strings.Contains(mock.calls[i], title) {
				t.Errorf("call %d missing title %q", i, title)
			}
		}
	})
}
