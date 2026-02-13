package conflict

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gongahkia/salja/internal/config"
	salerr "github.com/gongahkia/salja/internal/errors"
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
		return nil, &salerr.ConflictError{SourceItem: source.Title, TargetItem: target.Title, Message: "conflict detected"}
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

	input, err := r.reader.ReadString('\n')
	if err != nil {
		// EOF or read error: fall back to preferring source
		r.log(source.Title, target.Title, "prefer-source-eof")
		return source, nil
	}
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

	if source.Status != target.Status {
		fmt.Printf("  Status: [s]%s / [t]%s? ", source.Status, target.Status)
		input, _ := r.reader.ReadString('\n')
		if strings.TrimSpace(input) == "t" {
			merged.Status = target.Status
			fields = append(fields, "status:target")
		}
	}

	if len(source.Tags) != len(target.Tags) || !tagsEqual(source.Tags, target.Tags) {
		fmt.Printf("  Tags: [s]%v / [t]%v? ", source.Tags, target.Tags)
		input, _ := r.reader.ReadString('\n')
		if strings.TrimSpace(input) == "t" {
			merged.Tags = target.Tags
			fields = append(fields, "tags:target")
		}
	}

	if (source.Recurrence == nil) != (target.Recurrence == nil) {
		fmt.Printf("  Recurrence: [s]source / [t]target? ")
		input, _ := r.reader.ReadString('\n')
		if strings.TrimSpace(input) == "t" {
			merged.Recurrence = target.Recurrence
			fields = append(fields, "recurrence:target")
		}
	}

	if len(source.Reminders) != len(target.Reminders) {
		fmt.Printf("  Reminders: [s]%d / [t]%d? ", len(source.Reminders), len(target.Reminders))
		input, _ := r.reader.ReadString('\n')
		if strings.TrimSpace(input) == "t" {
			merged.Reminders = target.Reminders
			fields = append(fields, "reminders:target")
		}
	}

	r.log(source.Title, target.Title, "merged")
	if len(r.resolutions) > 0 {
		r.resolutions[len(r.resolutions)-1].Fields = fields
	}
	return &merged, nil
}

func tagsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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

	logDir := config.ConfigDir()
	os.MkdirAll(logDir, 0755)

	logPath := filepath.Join(logDir, "conflict-log.json")

	var existing []Resolution
	if data, err := os.ReadFile(logPath); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("failed to parse existing conflict log %s: %w", logPath, err)
		}
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
