package commands

import (
"fmt"
"io"
"os"
"path/filepath"
"strings"

"github.com/gongahkia/calendar-converter/internal/ics"
"github.com/gongahkia/calendar-converter/internal/model"
"github.com/gongahkia/calendar-converter/internal/parsers"
"github.com/gongahkia/calendar-converter/internal/writers"
"github.com/schollz/progressbar/v3"
"github.com/spf13/cobra"
)

func NewConvertCmd() *cobra.Command {
var fromFormat, toFormat string
var dryRun, quiet bool
var outputFormat string

cmd := &cobra.Command{
Use:   "convert <input-file> <output-file>",
Short: "Convert between calendar/task formats",
Args:  cobra.ExactArgs(2),
RunE: func(cmd *cobra.Command, args []string) error {
inputFile := args[0]
outputFile := args[1]

if fromFormat == "" {
fromFormat = DetectFormat(inputFile)
if !quiet {
fmt.Fprintf(os.Stderr, "Detected source format: %s\n", fromFormat)
}
}

if toFormat == "" {
toFormat = DetectFormat(outputFile)
if !quiet {
fmt.Fprintf(os.Stderr, "Detected target format: %s\n", toFormat)
}
}

collection, err := ReadInput(inputFile, fromFormat)
if err != nil {
return fmt.Errorf("failed to read input: %w", err)
}

if !quiet {
fmt.Fprintf(os.Stderr, "Loaded %d items from %s\n", len(collection.Items), inputFile)
}

if dryRun {
for _, item := range collection.Items {
fmt.Printf("  - %s (%s)\n", item.Title, item.ItemType)
}
return nil
}

var bar *progressbar.ProgressBar
if !quiet && len(collection.Items) > 10 {
bar = progressbar.NewOptions(len(collection.Items),
progressbar.OptionSetDescription("Converting"),
progressbar.OptionSetWriter(os.Stderr),
progressbar.OptionShowCount(),
progressbar.OptionClearOnFinish(),
)
for range collection.Items {
bar.Add(1)
}
bar.Finish()
}

if err := WriteOutput(collection, outputFile, toFormat); err != nil {
return fmt.Errorf("failed to write output: %w", err)
}

if outputFormat == "json" {
fmt.Printf("{\"converted\": %d, \"source\": \"%s\", \"target\": \"%s\"}\n", len(collection.Items), fromFormat, toFormat)
} else if !quiet {
fmt.Fprintf(os.Stderr, "Successfully converted %d items to %s\n", len(collection.Items), outputFile)
}
return nil
},
}

cmd.Flags().StringVar(&fromFormat, "from", "", "Source format override")
cmd.Flags().StringVar(&toFormat, "to", "", "Target format override")
cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress non-error output")
cmd.Flags().StringVar(&outputFormat, "output-format", "text", "Output format: text or json")

return cmd
}

func DetectFormat(filePath string) string {
if filePath == "-" {
return "ics"
}
ext := strings.ToLower(filepath.Ext(filePath))
switch ext {
case ".ics":
return "ics"
case ".csv":
base := strings.ToLower(filepath.Base(filePath))
if strings.Contains(base, "ticktick") {
return "ticktick"
} else if strings.Contains(base, "todoist") {
return "todoist"
} else if strings.Contains(base, "google") || strings.Contains(base, "gcal") {
return "gcal"
} else if strings.Contains(base, "outlook") {
return "outlook"
} else if strings.Contains(base, "notion") {
return "notion"
} else if strings.Contains(base, "asana") {
return "asana"
}
return "csv"
case ".json":
return "trello"
case ".taskpaper":
return "omnifocus"
default:
return "unknown"
}
}

func ReadInput(filePath, format string) (*model.CalendarCollection, error) {
var r io.Reader
if filePath == "-" {
r = os.Stdin
}

switch format {
case "ics":
p := ics.NewParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "ticktick":
p := parsers.NewTickTickParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "todoist":
p := parsers.NewTodoistParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "gcal":
p := parsers.NewGoogleCalendarParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "outlook":
p := parsers.NewOutlookParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "notion":
p := parsers.NewNotionParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "trello":
p := parsers.NewTrelloParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "asana":
p := parsers.NewAsanaParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
case "omnifocus":
p := parsers.NewOmniFocusParser()
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
default:
return nil, fmt.Errorf("unsupported input format: %s", format)
}
}

func WriteOutput(collection *model.CalendarCollection, filePath, format string) error {
var w io.Writer
if filePath == "-" {
w = os.Stdout
}

switch format {
case "ics":
wr := ics.NewWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "ticktick":
wr := writers.NewTickTickWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "todoist":
wr := writers.NewTodoistWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "gcal":
wr := writers.NewGoogleCalendarWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "outlook":
wr := writers.NewOutlookWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "notion":
wr := writers.NewNotionWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "trello":
wr := writers.NewTrelloWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "asana":
wr := writers.NewAsanaWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
case "omnifocus":
wr := writers.NewOmniFocusWriter()
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
default:
return fmt.Errorf("unsupported output format: %s", format)
}
}
