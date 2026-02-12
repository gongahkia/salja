package conflict

import (
"bufio"
"encoding/json"
"fmt"
"os"
"path/filepath"
"strings"
"time"

"github.com/gongahkia/salja/internal/model"
)

type Strategy string

const (
StrategyAsk          Strategy = "ask"
StrategyPreferSource Strategy = "prefer-source"
StrategyPreferTarget Strategy = "prefer-target"
StrategySkip         Strategy = "skip-conflicts"
StrategyFail         Strategy = "fail-on-conflict"
)

type Resolution struct {
SourceTitle string    `json:"source_title"`
TargetTitle string    `json:"target_title"`
Action      string    `json:"action"`
Timestamp   time.Time `json:"timestamp"`
Fields      []string  `json:"fields,omitempty"`
}

type Resolver struct {
strategy    Strategy
resolutions []Resolution
reader      *bufio.Reader
}

func NewResolver(strategy Strategy) *Resolver {
return &Resolver{
strategy: strategy,
reader:   bufio.NewReader(os.Stdin),
}
}

func (r *Resolver) Resolve(source, target *model.CalendarItem) (*model.CalendarItem, error) {
switch r.strategy {
case StrategyPreferSource:
r.log(source.Title, target.Title, "prefer-source")
return source, nil
case StrategyPreferTarget:
r.log(source.Title, target.Title, "prefer-target")
return target, nil
case StrategySkip:
r.log(source.Title, target.Title, "skip")
return nil, nil
case StrategyFail:
return nil, fmt.Errorf("conflict detected between '%s' and '%s'", source.Title, target.Title)
case StrategyAsk:
return r.interactiveResolve(source, target)
default:
return source, nil
}
}

func (r *Resolver) interactiveResolve(source, target *model.CalendarItem) (*model.CalendarItem, error) {
fmt.Println("\n=== CONFLICT DETECTED ===")
fmt.Printf("Source: %s\n", source.Title)
fmt.Printf("Target: %s\n", target.Title)
printDiff(source, target)
fmt.Println("\nOptions: [s]ource / [t]arget / [m]erge / [k]skip")
fmt.Print("> ")

input, _ := r.reader.ReadString('\n')
input = strings.TrimSpace(strings.ToLower(input))

switch input {
case "s":
r.log(source.Title, target.Title, "chose-source")
return source, nil
case "t":
r.log(source.Title, target.Title, "chose-target")
return target, nil
case "m":
return r.fieldMerge(source, target)
case "k":
r.log(source.Title, target.Title, "skipped")
return nil, nil
default:
return source, nil
}
}

func (r *Resolver) fieldMerge(source, target *model.CalendarItem) (*model.CalendarItem, error) {
merged := *source
var fields []string

fmt.Println("\nField-level merge (press enter to keep source, 't' for target):")

if source.Title != target.Title {
fmt.Printf("  Title: [s]'%s' / [t]'%s'? ", source.Title, target.Title)
input, _ := r.reader.ReadString('\n')
if strings.TrimSpace(input) == "t" {
merged.Title = target.Title
fields = append(fields, "title:target")
}
}

if source.Description != target.Description {
fmt.Printf("  Description: [s]source / [t]target? ")
input, _ := r.reader.ReadString('\n')
if strings.TrimSpace(input) == "t" {
merged.Description = target.Description
fields = append(fields, "description:target")
}
}

if source.Priority != target.Priority {
fmt.Printf("  Priority: [s]%d / [t]%d? ", source.Priority, target.Priority)
input, _ := r.reader.ReadString('\n')
if strings.TrimSpace(input) == "t" {
merged.Priority = target.Priority
fields = append(fields, "priority:target")
}
}

r.log(source.Title, target.Title, "merged")
r.resolutions[len(r.resolutions)-1].Fields = fields
return &merged, nil
}

func (r *Resolver) log(sourceTitle, targetTitle, action string) {
r.resolutions = append(r.resolutions, Resolution{
SourceTitle: sourceTitle,
TargetTitle: targetTitle,
Action:      action,
Timestamp:   time.Now(),
})
}

func (r *Resolver) WriteLog() error {
if len(r.resolutions) == 0 {
return nil
}

configDir := os.Getenv("XDG_CONFIG_HOME")
if configDir == "" {
home, _ := os.UserHomeDir()
configDir = filepath.Join(home, ".config")
}
logDir := filepath.Join(configDir, "salja")
os.MkdirAll(logDir, 0755)

logPath := filepath.Join(logDir, "conflict-log.json")

var existing []Resolution
if data, err := os.ReadFile(logPath); err == nil {
json.Unmarshal(data, &existing)
}

all := append(existing, r.resolutions...)
data, err := json.MarshalIndent(all, "", "  ")
if err != nil {
return err
}
return os.WriteFile(logPath, data, 0644)
}

func printDiff(a, b *model.CalendarItem) {
if a.Title != b.Title {
fmt.Printf("  Title:       '%s' vs '%s'\n", a.Title, b.Title)
}
if a.Description != b.Description {
fmt.Printf("  Description: differs\n")
}
if a.Priority != b.Priority {
fmt.Printf("  Priority:    %d vs %d\n", a.Priority, b.Priority)
}
if a.Status != b.Status {
fmt.Printf("  Status:      %s vs %s\n", a.Status, b.Status)
}
}

type DataLossChecker struct{}

func NewDataLossChecker() *DataLossChecker {
return &DataLossChecker{}
}

type DataLossWarning struct {
Field   string
Message string
}

func (c *DataLossChecker) Check(items []model.CalendarItem, targetFormat string) []DataLossWarning {
var warnings []DataLossWarning

noSubtasks := map[string]bool{"gcal": true, "outlook": true, "ics": true}
noRecurrence := map[string]bool{"notion": true, "trello": true, "asana": true}
noReminders := map[string]bool{"trello": true, "asana": true, "notion": true}

for _, item := range items {
if len(item.Subtasks) > 0 && noSubtasks[targetFormat] {
warnings = append(warnings, DataLossWarning{
Field:   "subtasks",
Message: fmt.Sprintf("'%s' has %d subtasks which %s doesn't support", item.Title, len(item.Subtasks), targetFormat),
})
}
if item.Recurrence != nil && noRecurrence[targetFormat] {
warnings = append(warnings, DataLossWarning{
Field:   "recurrence",
Message: fmt.Sprintf("'%s' has recurrence which %s doesn't support", item.Title, targetFormat),
})
}
if len(item.Reminders) > 0 && noReminders[targetFormat] {
warnings = append(warnings, DataLossWarning{
Field:   "reminders",
Message: fmt.Sprintf("'%s' has reminders which %s doesn't support", item.Title, targetFormat),
})
}
}

return warnings
}
