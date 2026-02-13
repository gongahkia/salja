package commands

import (
"encoding/json"
"fmt"

"github.com/gongahkia/salja/internal/conflict"
"github.com/spf13/cobra"
)

func NewDiffCmd() *cobra.Command {
var fromFormat, toFormat string
var outputFormat string

cmd := &cobra.Command{
Use:   "diff <file1> <file2>",
Short: "Compare two calendar/task files",
Args:  cobra.ExactArgs(2),
RunE: func(cmd *cobra.Command, args []string) error {
if fromFormat == "" {
fromFormat = DetectFormat(args[0])
}
if toFormat == "" {
toFormat = DetectFormat(args[1])
}

col1, err := ReadInput(args[0], fromFormat, nil)
if err != nil {
return fmt.Errorf("failed to read file1: %w", err)
}
col2, err := ReadInput(args[1], toFormat, nil)
if err != nil {
return fmt.Errorf("failed to read file2: %w", err)
}

detector := conflict.NewDetector()
matches := detector.FindDuplicates(col1, col2)

matched1 := make(map[int]bool)
matched2 := make(map[int]bool)
for _, m := range matches {
matched1[m.SourceIndex] = true
matched2[m.TargetIndex] = true
}

var addedTitles, removedTitles []string
for j, item := range col2.Items {
if !matched2[j] {
addedTitles = append(addedTitles, item.Title)
}
}
for i, item := range col1.Items {
if !matched1[i] {
removedTitles = append(removedTitles, item.Title)
}
}

switch outputFormat {
case "json":
result := map[string]interface{}{
"file1_count":    len(col1.Items),
"file2_count":    len(col2.Items),
"matching_count": len(matches),
"added_count":    len(addedTitles),
"removed_count":  len(removedTitles),
"added":          addedTitles,
"removed":        removedTitles,
}
data, _ := json.MarshalIndent(result, "", "  ")
fmt.Println(string(data))

case "patch":
for _, title := range removedTitles {
fmt.Printf("- %s\n", title)
}
for _, title := range addedTitles {
fmt.Printf("+ %s\n", title)
}

default: // "table"
fmt.Printf("File 1: %d items\n", len(col1.Items))
fmt.Printf("File 2: %d items\n", len(col2.Items))
fmt.Printf("Matching: %d items\n", len(matches))
fmt.Printf("Added:   %d items (in file2 only)\n", len(addedTitles))
fmt.Printf("Removed: %d items (in file1 only)\n", len(removedTitles))

if len(removedTitles) > 0 {
fmt.Println("\n--- Removed ---")
for _, title := range removedTitles {
fmt.Printf("  - %s\n", title)
}
}
if len(addedTitles) > 0 {
fmt.Println("\n--- Added ---")
for _, title := range addedTitles {
fmt.Printf("  + %s\n", title)
}
}
}

return nil
},
}

cmd.Flags().StringVar(&fromFormat, "from", "", "Format of file1")
cmd.Flags().StringVar(&toFormat, "to", "", "Format of file2")
cmd.Flags().StringVar(&outputFormat, "format", "table", "Output format: table, json, or patch")
return cmd
}
