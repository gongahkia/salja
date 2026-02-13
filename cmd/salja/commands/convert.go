package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gongahkia/salja/internal/config"
	"github.com/gongahkia/salja/internal/conflict"
	salerr "github.com/gongahkia/salja/internal/errors"
	"github.com/gongahkia/salja/internal/fidelity"
	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/parsers"
	"github.com/gongahkia/salja/internal/registry"
	_ "github.com/gongahkia/salja/internal/registry" // ensure format registration
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func NewConvertCmd() *cobra.Command {
	var fromFormat, toFormat string
	var dryRun, quiet, strict, jsonOutput, merge bool
	var outputFormat, fidelityMode, locale string
	var appleCalendar, appleList string

	cmd := &cobra.Command{
		Use:   "convert <input-file> <output-file>",
		Short: "Convert between calendar/task formats",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]
			outputFile := args[1]

			if fromFormat == "" {
				fromFormat = DetectFormat(inputFile)
				if fromFormat == "unknown" && !quiet && !jsonOutput {
					fromFormat = interactiveFormatPicker("source")
				} else if !quiet && !jsonOutput {
					fmt.Fprintf(os.Stderr, "Detected source format: %s\n", fromFormat)
				}
			}

			if toFormat == "" {
				toFormat = DetectFormat(outputFile)
				if toFormat == "unknown" && !quiet && !jsonOutput {
					toFormat = interactiveFormatPicker("target")
				} else if !quiet && !jsonOutput {
					fmt.Fprintf(os.Stderr, "Detected target format: %s\n", toFormat)
				}
			}

			// Load config early so streaming threshold is available
			cfg, _ := config.Load()

			if locale != "" {
				parsers.SetLocale(locale)
			}

			// Validate apple-specific flags
			if (fromFormat == "apple-calendar" || toFormat == "apple-calendar") && appleCalendar == "" {
				return fmt.Errorf("--calendar flag is required when using apple-calendar format")
			}
			if (fromFormat == "apple-reminders" || toFormat == "apple-reminders") && appleList == "" {
				return fmt.Errorf("--list flag is required when using apple-reminders format")
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			handler := salerr.NewSignalHandler(cancel)
			handler.Start()

			// File size pre-check: warn if file exceeds streaming threshold
			if inputFile != "-" {
				thresholdMB := 10
				if cfg != nil && cfg.StreamingThresholdMB > 0 {
					thresholdMB = cfg.StreamingThresholdMB
				}
				if info, statErr := os.Stat(inputFile); statErr == nil {
					fileSizeMB := info.Size() / (1024 * 1024)
					if fileSizeMB >= int64(thresholdMB) && !quiet {
						fmt.Fprintf(os.Stderr, "Warning: input file is %dMB (threshold: %dMB). Large files may use significant memory.\n", fileSizeMB, thresholdMB)
						fmt.Fprint(os.Stderr, "Continue? [Y/n] ")
						var confirm string
						fmt.Fscanln(os.Stdin, &confirm)
						if confirm == "n" || confirm == "N" {
							return fmt.Errorf("aborted by user")
						}
					}
				}
			}

			collection, err := ReadInput(ctx, inputFile, fromFormat, cfg)
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			// Validate each item, collecting successes and failures
			partial := salerr.NewPartialResult[model.CalendarItem]()
			for i, item := range collection.Items {
				if i%100 == 0 {
					if err := ctx.Err(); err != nil {
						return fmt.Errorf("aborted: %w", err)
					}
				}
				if err := item.Validate(); err != nil {
					partial.AddError(i, item.UID, err.Error(), err)
				} else {
					partial.Add(item)
				}
			}
			collection.Items = partial.Items

			if !quiet && !jsonOutput {
				if partial.HasErrors() {
					fmt.Fprintf(os.Stderr, "%s\n", partial.Summary())
				} else {
					fmt.Fprintf(os.Stderr, "Loaded %d items from %s\n", len(collection.Items), inputFile)
				}
			}

			if dryRun {
				for _, item := range collection.Items {
					fmt.Printf("  - %s (%s)\n", item.Title, item.ItemType)
				}
				return nil
			}

			// Validate that the target format supports the source item types
			caps, hasCaps := registry.GetCapabilities(toFormat)
			if hasCaps {
				hasEvents, hasTasks := false, false
				for _, item := range collection.Items {
					switch item.ItemType {
					case model.ItemTypeEvent:
						hasEvents = true
					case model.ItemTypeTask:
						hasTasks = true
					}
				}
				if hasEvents && !caps.SupportsEvents {
					return fmt.Errorf("target format %q does not support events, but source contains %s; use a format that supports events or filter items first", toFormat, "events")
				}
				if hasTasks && !caps.SupportsTasks {
					return fmt.Errorf("target format %q does not support tasks, but source contains %s; use a format that supports tasks or filter items first", toFormat, "tasks")
				}
			}

			// Determine data loss mode: --fidelity flag > --strict flag > config > default
			dataLossMode := "warn"
			if cfg != nil {
				dataLossMode = cfg.DataLossMode
			}
			if fidelityMode != "" {
				dataLossMode = fidelityMode
			}
			if strict {
				dataLossMode = "error"
			}

			var bar *progressbar.ProgressBar
			if !quiet && !jsonOutput && len(collection.Items) > 10 {
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

			// Enforce DataLossMode
			if len(warnings) > 0 {
				switch dataLossMode {
				case "error":
					if !jsonOutput {
						for _, w := range warnings {
							fmt.Fprintf(os.Stderr, "ERROR: %s\n", w)
						}
					}
					return fmt.Errorf("aborting due to %d data loss error(s); set data_loss_mode = \"warn\" or \"silent\" to continue", len(warnings))
				case "warn":
					if !jsonOutput {
						for _, w := range warnings {
							fmt.Fprintf(os.Stderr, "WARNING: %s\n", w)
						}
					}
				case "silent":
					// suppress all fidelity output
				}
			}

			// Merge with existing output file when --merge is set
			if merge {
				if _, statErr := os.Stat(outputFile); statErr == nil {
					existing, readErr := ReadInput(ctx, outputFile, toFormat, cfg)
					if readErr != nil {
						return fmt.Errorf("--merge: failed to read existing output file: %w", readErr)
					}

					strategy := conflict.StrategyAsk
					if cfg != nil && cfg.ConflictStrategy != "" {
						strategy = conflict.Strategy(cfg.ConflictStrategy)
					}

					detector := conflict.NewDetector()
					if cfg != nil {
						detector.LevenshteinThreshold = cfg.ConflictThresholds.LevenshteinThreshold
						detector.MinTitleLength = cfg.ConflictThresholds.MinTitleLength
						detector.DateProximityHours = cfg.ConflictThresholds.DateProximityHours
					}
					matches := detector.FindDuplicates(collection, existing)

					if len(matches) > 0 {
						if !quiet && !jsonOutput {
							fmt.Fprintf(os.Stderr, "Found %d duplicate(s) in existing output file\n", len(matches))
						}
						resolver := conflict.NewResolver(strategy)
						resolved := make(map[int]bool)
						for _, m := range matches {
							result, err := resolver.Resolve(&collection.Items[m.SourceIndex], &existing.Items[m.TargetIndex])
							if err != nil {
								return fmt.Errorf("--merge: conflict resolution failed: %w", err)
							}
							if result != nil {
								collection.Items[m.SourceIndex] = *result
							}
							resolved[m.TargetIndex] = true
						}
						// Append non-duplicate items from existing file
						for j, item := range existing.Items {
							if !resolved[j] {
								collection.Items = append(collection.Items, item)
							}
						}
						_ = resolver.WriteLog()
					} else {
						// No duplicates: append all existing items
						collection.Items = append(collection.Items, existing.Items...)
					}
				}
			}

			if err := WriteOutput(ctx, collection, outputFile, toFormat); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}

			if bar != nil {
				bar.Add(len(collection.Items) - len(collection.Items)/2)
				bar.Finish()
			}

			// Compute summary counts
			eventCount, taskCount, warnCount := 0, 0, len(warnings)
			for _, item := range collection.Items {
				switch item.ItemType {
				case model.ItemTypeEvent:
					eventCount++
				default:
					taskCount++
				}
			}

			// Output structured JSON report or human-readable text
			if jsonOutput || outputFormat == "json" {
				warningStrs := make([]string, len(warnings))
				for i, w := range warnings {
					warningStrs[i] = w.String()
				}
				var errStrs []string
				for _, e := range partial.Errors {
					errStrs = append(errStrs, e.Error())
				}
				report := convertReport{
					Converted:   len(collection.Items),
					Events:      eventCount,
					Tasks:       taskCount,
					Warnings:    warnCount,
					ParseErrors: len(partial.Errors),
					Source:      fromFormat,
					Target:      toFormat,
					Details:     warningStrs,
					Errors:      errStrs,
				}
				data, _ := json.MarshalIndent(report, "", "  ")
				fmt.Println(string(data))
			} else if !quiet {
				if partial.HasErrors() {
					fmt.Fprintf(os.Stderr, "Converted %d events, %d tasks (%d warnings, %d parse errors)\n", eventCount, taskCount, warnCount, len(partial.Errors))
				} else {
					fmt.Fprintf(os.Stderr, "Converted %d events, %d tasks (%d warnings)\n", eventCount, taskCount, warnCount)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromFormat, "from", "", "Source format override")
	cmd.Flags().StringVar(&toFormat, "to", "", "Target format override")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	cmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress non-error output")
	cmd.Flags().StringVar(&outputFormat, "output-format", "text", "Output format: text or json")
	cmd.Flags().StringVar(&fidelityMode, "fidelity", "", "Data loss mode override: warn|error|silent")
	cmd.Flags().BoolVar(&strict, "strict", false, "Treat any warning as a fatal error")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output structured JSON conversion report")
	cmd.Flags().BoolVar(&merge, "merge", false, "Detect duplicates and resolve conflicts when output file exists")
	cmd.Flags().StringVar(&locale, "locale", "", "Locale for ambiguous date parsing (e.g. en-gb, de, ja)")
	cmd.Flags().StringVar(&appleCalendar, "calendar", "", "Apple Calendar name (required for apple-calendar format)")
	cmd.Flags().StringVar(&appleList, "list", "", "Apple Reminders list name (required for apple-reminders format)")

	return cmd
}

type convertReport struct {
	Converted   int      `json:"converted"`
	Events      int      `json:"events"`
	Tasks       int      `json:"tasks"`
	Warnings    int      `json:"warnings"`
	ParseErrors int      `json:"parse_errors,omitempty"`
	Source      string   `json:"source"`
	Target      string   `json:"target"`
	Details     []string `json:"details,omitempty"`
	Errors      []string `json:"errors,omitempty"`
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

func ReadInput(ctx context.Context, filePath, format string, cfg *config.Config) (*model.CalendarCollection, error) {
	var r io.Reader
	if filePath == "-" {
		r = os.Stdin
	}

	// Check file size against streaming threshold
	thresholdMB := 10
	if cfg != nil && cfg.StreamingThresholdMB > 0 {
		thresholdMB = cfg.StreamingThresholdMB
	}

	if filePath != "-" {
		info, err := os.Stat(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat input file: %w", err)
		}
		fileSizeMB := info.Size() / (1024 * 1024)
		if fileSizeMB >= int64(thresholdMB) {
			fmt.Fprintf(os.Stderr, "File exceeds %dMB threshold, using streaming parser\n", thresholdMB)
			return readInputStreaming(ctx, filePath, format)
		}
	}

	p, err := registry.GetParser(format)
	if err != nil {
		return nil, err
	}
	if r != nil {
		return p.Parse(ctx, r, "stdin")
	}
	return p.ParseFile(ctx, filePath)
}

func readInputStreaming(ctx context.Context, filePath, format string) (*model.CalendarCollection, error) {
	// For CSV formats, use the StreamingCSVParser for constant-memory processing
	csvFormats := map[string]bool{"ticktick": true, "todoist": true, "gcal": true, "outlook": true, "notion": true, "asana": true}
	if csvFormats[format] {
		f, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		sp, err := salerr.NewStreamingCSVParser(f)
		if err != nil {
			return nil, fmt.Errorf("streaming CSV init failed: %w", err)
		}
		defer sp.Close()

		p, pErr := registry.GetParser(format)
		if pErr != nil {
			return nil, pErr
		}
		// Fall back to standard parser with the file reader
		return p.Parse(ctx, f, filePath)
	}

	p, err := registry.GetParser(format)
	if err != nil {
		return nil, err
	}
	return p.ParseFile(ctx, filePath)
}

func WriteOutput(ctx context.Context, collection *model.CalendarCollection, filePath, format string) error {
	var w io.Writer
	if filePath == "-" {
		w = os.Stdout
	}

	wr, err := registry.GetWriter(format)
	if err != nil {
		return err
	}
	if w != nil {
		return wr.Write(ctx, collection, w)
	}
	return wr.WriteFile(ctx, collection, filePath)
}

func interactiveFormatPicker(role string) string {
	all := registry.AllFormats()
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintf(os.Stderr, "Could not detect %s format. Select one:\n", role)
	for i, name := range names {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, name)
	}
	fmt.Fprint(os.Stderr, "> ")
	var choice int
	if _, err := fmt.Fscan(os.Stdin, &choice); err != nil || choice < 1 || choice > len(names) {
		return names[0]
	}
	return names[choice-1]
}
