//go:build darwin && integration

package apple

import (
"testing"

"github.com/gongahkia/calendar-converter/internal/model"
)

func TestRemindersReadWriteCycle(t *testing.T) {
listName := "calconv-test-list"

// Create test list
createScript := `tell application "Reminders"
make new list with properties {name:"` + listName + `"}
end tell`
_, err := RunAppleScript(createScript)
if err != nil {
t.Skipf("Cannot access Reminders.app: %v", err)
}

// Cleanup at end
defer func() {
deleteScript := `tell application "Reminders"
try
delete list "` + listName + `"
end try
end tell`
RunAppleScript(deleteScript)
}()

writer := NewRemindersWriter()
items := []model.CalendarItem{
{
Title:       "Integration Test Task",
Description: "A test reminder",
ItemType:    model.ItemTypeTask,
Priority:    model.PriorityHigh,
Status:      model.StatusPending,
},
}

if err := writer.Write(items, listName); err != nil {
t.Fatalf("Write failed: %v", err)
}

reader := NewRemindersReader()
collection, err := reader.Read()
if err != nil {
t.Fatalf("Read failed: %v", err)
}

found := false
for _, item := range collection.Items {
if item.Title == "Integration Test Task" {
found = true
if item.Description != "A test reminder" {
t.Errorf("description mismatch: got %q", item.Description)
}
break
}
}
if !found {
t.Error("Written reminder not found in read results")
}
}
