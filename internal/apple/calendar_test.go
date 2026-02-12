//go:build darwin && integration

package apple

import (
"testing"
"time"

"github.com/gongahkia/salja/internal/model"
)

func TestCalendarReadWriteCycle(t *testing.T) {
calName := "salja-test-calendar"

// Create test calendar
createScript := `tell application "Calendar"
make new calendar with properties {name:"` + calName + `"}
end tell`
_, err := RunAppleScript(createScript)
if err != nil {
t.Skipf("Cannot access Calendar.app: %v", err)
}

// Cleanup at end
defer func() {
deleteScript := `tell application "Calendar"
try
delete calendar "` + calName + `"
end try
end tell`
RunAppleScript(deleteScript)
}()

now := time.Now().Truncate(time.Minute)
start := now
end := now.Add(time.Hour)

writer := NewCalendarWriter()
items := []model.CalendarItem{
{
Title:       "Integration Test Event",
Description: "A test event",
ItemType:    model.ItemTypeEvent,
StartTime:   &start,
EndTime:     &end,
Location:    "Test Location",
},
}

if err := writer.Write(items, calName); err != nil {
t.Fatalf("Write failed: %v", err)
}

reader := NewCalendarReader()
rangeStart := now.Add(-time.Hour)
rangeEnd := now.Add(2 * time.Hour)
collection, err := reader.Read(calName, rangeStart, rangeEnd)
if err != nil {
t.Fatalf("Read failed: %v", err)
}

found := false
for _, item := range collection.Items {
if item.Title == "Integration Test Event" {
found = true
if item.Description != "A test event" {
t.Errorf("description mismatch: got %q", item.Description)
}
if item.Location != "Test Location" {
t.Errorf("location mismatch: got %q", item.Location)
}
break
}
}
if !found {
t.Error("Written event not found in read results")
}
}
