package commands

import (
"fmt"

"github.com/gongahkia/calendar-converter/internal/conflict"
"github.com/spf13/cobra"
)

func NewDiffCmd() *cobra.Command {
var fromFormat, toFormat string

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

col1, err := ReadInput(args[0], fromFormat)
if err != nil {
return fmt.Errorf("failed to read file1: %w", err)
}
col2, err := ReadInput(args[1], toFormat)
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

fmt.Printf("File 1: %d items\n", len(col1.Items))
fmt.Printf("File 2: %d items\n", len(col2.Items))
fmt.Printf("Matching: %d items\n", len(matches))

added := 0
for j := range col2.Items {
if !matched2[j] {
added++
}
}
removed := 0
for i := range col1.Items {
if !matched1[i] {
removed++
}
}

fmt.Printf("Added:   %d items (in file2 only)\n", added)
fmt.Printf("Removed: %d items (in file1 only)\n", removed)

if removed > 0 {
fmt.Println("\n--- Removed ---")
for i, item := range col1.Items {
if !matched1[i] {
fmt.Printf("  - %s\n", item.Title)
}
}
}
if added > 0 {
fmt.Println("\n--- Added ---")
for j, item := range col2.Items {
if !matched2[j] {
fmt.Printf("  + %s\n", item.Title)
}
}
}

return nil
},
}

cmd.Flags().StringVar(&fromFormat, "from", "", "Format of file1")
cmd.Flags().StringVar(&toFormat, "to", "", "Format of file2")
return cmd
}
