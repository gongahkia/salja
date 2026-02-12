package apple

import (
"fmt"
"strings"
"time"

"github.com/gongahkia/salja/internal/model"
)

type RemindersWriter struct{}

func NewRemindersWriter() *RemindersWriter {
return &RemindersWriter{}
}

func (w *RemindersWriter) Write(items []model.CalendarItem, listName string) error {
for _, item := range items {
if err := w.createReminder(item, listName); err != nil {
return fmt.Errorf("failed to create reminder '%s': %w", item.Title, err)
}
}
return nil
}

func (w *RemindersWriter) createReminder(item model.CalendarItem, listName string) error {
var scriptParts []string
scriptParts = append(scriptParts, `tell application "Reminders"`)
scriptParts = append(scriptParts, fmt.Sprintf(`  tell list "%s"`, listName))

props := []string{fmt.Sprintf(`name:"%s"`, escapeAS(item.Title))}
if item.Description != "" {
props = append(props, fmt.Sprintf(`body:"%s"`, escapeAS(item.Description)))
}
if item.DueDate != nil {
props = append(props, fmt.Sprintf(`due date:date "%s"`, item.DueDate.Format("January 2, 2006 3:04:05 PM")))
}
if item.Priority > 0 {
asPriority := 0
switch item.Priority {
case model.PriorityHighest, model.PriorityHigh:
asPriority = 1
case model.PriorityMedium:
asPriority = 5
case model.PriorityLow, model.PriorityLowest:
asPriority = 9
}
props = append(props, fmt.Sprintf(`priority:%d`, asPriority))
}
if item.Status == model.StatusCompleted {
props = append(props, `completed:true`)
}

scriptParts = append(scriptParts, fmt.Sprintf(`    make new reminder with properties {%s}`, strings.Join(props, ", ")))
scriptParts = append(scriptParts, `  end tell`)
scriptParts = append(scriptParts, `end tell`)

_, err := RunAppleScript(strings.Join(scriptParts, "\n"))
return err
}

type RemindersReader struct{}

func NewRemindersReader() *RemindersReader {
return &RemindersReader{}
}

func (r *RemindersReader) Read() (*model.CalendarCollection, error) {
script := `tell application "Reminders"
set output to ""
repeat with l in lists
set listName to name of l
repeat with rem in reminders of l
set remName to name of rem
set remBody to ""
try
set remBody to body of rem
end try
set remDue to ""
try
set remDue to (due date of rem) as string
end try
set remDone to completed of rem
set output to output & listName & "|||" & remName & "|||" & remBody & "|||" & remDue & "|||" & remDone & linefeed
end repeat
end repeat
return output
end tell`

output, err := RunAppleScript(script)
if err != nil {
return nil, err
}

collection := &model.CalendarCollection{
Items:     []model.CalendarItem{},
SourceApp: "apple-reminders",
ExportDate: time.Now(),
}

lines := strings.Split(output, "\n")
for _, line := range lines {
line = strings.TrimSpace(line)
if line == "" {
continue
}
parts := strings.Split(line, "|||")
if len(parts) < 5 {
continue
}

item := model.CalendarItem{
ItemType:    model.ItemTypeTask,
Status:      model.StatusPending,
Title:       parts[1],
Description: parts[2],
}

if parts[4] == "true" {
item.Status = model.StatusCompleted
}

collection.Items = append(collection.Items, item)
}

return collection, nil
}

func escapeAS(s string) string {
s = strings.ReplaceAll(s, `\`, `\\`)
s = strings.ReplaceAll(s, `"`, `\"`)
return s
}
