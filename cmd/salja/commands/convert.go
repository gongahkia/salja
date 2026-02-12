package commands

import (
"fmt"
"io"
"os"
"path/filepath"
"strings"

"github.com/gongahkia/salja/internal/config"
"github.com/gongahkia/salja/internal/fidelity"
"github.com/gongahkia/salja/internal/model"
"github.com/gongahkia/salja/internal/registry"
_ "github.com/gongahkia/salja/internal/registry" // ensure format registration
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

// Run data fidelity check with progress updates
cfg, _ := config.Load()
dataLossMode := "warn"
if cfg != nil {
dataLossMode = cfg.DataLossMode
}

var bar *progressbar.ProgressBar
if !quiet && len(collection.Items) > 10 {
bar = progressbar.NewOptions(len(collection.Items),
progressbar.OptionSetDescription("Converting"),
progressbar.OptionSetWriter(os.Stderr),
progressbar.OptionShowCount(),
progressbar.OptionClearOnFinish(),
)
}

warnings := fidelity.Check(collection, toFormat)
if bar != nil {
bar.Add(len(collection.Items) / 2)
}

if len(warnings) > 0 {
switch dataLossMode {
case "error":
for _, w := range warnings {
fmt.Fprintf(os.Stderr, "ERROR: %s\n", w)
}
return fmt.Errorf("aborting due to %d data loss error(s); set data_loss_mode = \"warn\" or \"silent\" to continue", len(warnings))
case "warn":
for _, w := range warnings {
fmt.Fprintf(os.Stderr, "WARNING: %s\n", w)
}
}
}

if err := WriteOutput(collection, outputFile, toFormat); err != nil {
return fmt.Errorf("failed to write output: %w", err)
}

if bar != nil {
bar.Add(len(collection.Items) - len(collection.Items)/2)
bar.Finish()
}

// Print item-count summary
eventCount, taskCount, warnCount := 0, 0, len(warnings)
for _, item := range collection.Items {
switch item.ItemType {
case model.ItemTypeEvent:
eventCount++
default:
taskCount++
}
}

if outputFormat == "json" {
fmt.Printf("{\"converted\": %d, \"events\": %d, \"tasks\": %d, \"warnings\": %d, \"source\": \"%s\", \"target\": \"%s\"}\n",
len(collection.Items), eventCount, taskCount, warnCount, fromFormat, toFormat)
} else if !quiet {
fmt.Fprintf(os.Stderr, "Converted %d events, %d tasks (%d warnings)\n", eventCount, taskCount, warnCount)
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
base := strings.ToLower(filepath.Base(filePath))

// Try filename hint first (for CSV disambiguation)
if hint := registry.DetectByFilenameHint(base); hint != "" {
return hint
}

// Fall back to extension matching
matches := registry.DetectByExtension(ext)
if len(matches) == 1 {
return matches[0]
}
if len(matches) > 1 {
// Multiple formats share this extension (e.g. .csv); can't disambiguate
return strings.TrimPrefix(ext, ".")
}
return "unknown"
}

func ReadInput(filePath, format string) (*model.CalendarCollection, error) {
var r io.Reader
if filePath == "-" {
r = os.Stdin
}

p, err := registry.GetParser(format)
if err != nil {
return nil, err
}
if r != nil {
return p.Parse(r, "stdin")
}
return p.ParseFile(filePath)
}

func WriteOutput(collection *model.CalendarCollection, filePath, format string) error {
var w io.Writer
if filePath == "-" {
w = os.Stdout
}

wr, err := registry.GetWriter(format)
if err != nil {
return err
}
if w != nil {
return wr.Write(collection, w)
}
return wr.WriteFile(collection, filePath)
}
