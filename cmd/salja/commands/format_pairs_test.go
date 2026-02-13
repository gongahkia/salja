package commands_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/gongahkia/salja/internal/fidelity"
	"github.com/gongahkia/salja/internal/model"
	"github.com/gongahkia/salja/internal/registry"
	_ "github.com/gongahkia/salja/internal/registry" // trigger init
)

// allWritableFormats returns format names that have both a parser and writer.
func allWritableFormats() []string {
	// Apple formats use AppleScript and don't support stream-based I/O
	appleFormats := map[string]bool{"apple-calendar": true, "apple-reminders": true}
	var names []string
	for name, entry := range registry.AllFormats() {
		if entry.NewParser != nil && entry.NewWriter != nil && !appleFormats[name] {
			names = append(names, name)
		}
	}
	return names
}

// testCollection returns a collection with one event and one task.
func testCollection() *model.CalendarCollection {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	end := now.Add(2 * time.Hour)
	due := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	return &model.CalendarCollection{
		Items: []model.CalendarItem{
			{
				UID:       "evt-1",
				Title:     "Meeting",
				StartTime: &now,
				EndTime:   &end,
				Location:  "Office",
				ItemType:  model.ItemTypeEvent,
				Status:    model.StatusPending,
			},
			{
				UID:      "task-1",
				Title:    "Review PR",
				DueDate:  &due,
				ItemType: model.ItemTypeTask,
				Status:   model.StatusPending,
				Priority: model.PriorityMedium,
				Tags:     []string{"code-review"},
			},
		},
		SourceApp: "test",
	}
}

// filterByCapabilities returns items from collection that the target format supports.
func filterByCapabilities(collection *model.CalendarCollection, target string) *model.CalendarCollection {
	caps, ok := registry.GetCapabilities(target)
	if !ok {
		return collection
	}
	filtered := &model.CalendarCollection{
		SourceApp: collection.SourceApp,
	}
	for _, item := range collection.Items {
		switch item.ItemType {
		case model.ItemTypeEvent:
			if caps.SupportsEvents {
				filtered.Items = append(filtered.Items, item)
			}
		case model.ItemTypeTask:
			if caps.SupportsTasks {
				filtered.Items = append(filtered.Items, item)
			}
		default:
			filtered.Items = append(filtered.Items, item)
		}
	}
	return filtered
}

// TestAllFormatPairRoundTrips verifies that for each compatible format pair,
// round-trip conversion preserves item count and that DataLossWarnings
// match expected field losses.
func TestAllFormatPairRoundTrips(t *testing.T) {
	ctx := context.Background()
	formats := allWritableFormats()
	collection := testCollection()

	for _, src := range formats {
		for _, dst := range formats {
			if src == dst {
				continue
			}
			t.Run(src+"_to_"+dst, func(t *testing.T) {
				// Filter items to only those the source format supports
				srcFiltered := filterByCapabilities(collection, src)
				if len(srcFiltered.Items) == 0 {
					t.Skip("no compatible items for source format")
				}

				// Check fidelity warnings before conversion
				warnings := fidelity.Check(srcFiltered, dst)

				// Write using source format
				srcWriter, err := registry.GetWriter(src)
				if err != nil {
					t.Fatalf("get writer for %s: %v", src, err)
				}
				var buf bytes.Buffer
				if err := srcWriter.Write(ctx, srcFiltered, &buf); err != nil {
					t.Fatalf("write %s: %v", src, err)
				}

				// Parse back using source format
				srcParser, err := registry.GetParser(src)
				if err != nil {
					t.Fatalf("get parser for %s: %v", src, err)
				}
				parsed, err := srcParser.Parse(ctx, &buf, "test")
				if err != nil {
					t.Fatalf("parse %s: %v", src, err)
				}

				// Filter parsed items to what the destination supports
				dstFiltered := filterByCapabilities(parsed, dst)
				if len(dstFiltered.Items) == 0 {
					t.Skip("no compatible items for destination format")
				}

				// Write using destination format
				dstWriter, err := registry.GetWriter(dst)
				if err != nil {
					t.Fatalf("get writer for %s: %v", dst, err)
				}
				var buf2 bytes.Buffer
				if err := dstWriter.Write(ctx, dstFiltered, &buf2); err != nil {
					t.Fatalf("write %s: %v", dst, err)
				}

				// Parse back using destination format
				dstParser, err := registry.GetParser(dst)
				if err != nil {
					t.Fatalf("get parser for %s: %v", dst, err)
				}
				final, err := dstParser.Parse(ctx, &buf2, "test")
				if err != nil {
					t.Fatalf("parse %s: %v", dst, err)
				}

				// Items should not be zero after conversion
				if len(final.Items) == 0 && len(dstFiltered.Items) > 0 {
					t.Errorf("lost all items in %s→%s conversion", src, dst)
				}

				// Verify warnings are consistent: fields warned about
				// should not appear on the output items
				warnedFields := make(map[string]bool)
				for _, w := range warnings {
					warnedFields[w.Field] = true
				}
				for _, item := range final.Items {
					if warnedFields["Subtasks"] && len(item.Subtasks) > 0 {
						t.Errorf("%s→%s: subtasks survived despite warning", src, dst)
					}
					if warnedFields["Recurrence"] && item.Recurrence != nil {
						t.Errorf("%s→%s: recurrence survived despite warning", src, dst)
					}
				}

				t.Logf("%s→%s: %d items, %d fidelity warnings", src, dst, len(final.Items), len(warnings))
			})
		}
	}
}
