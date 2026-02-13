package conflict

import (
	"errors"
	"testing"

	salerr "github.com/gongahkia/salja/internal/errors"
	"github.com/gongahkia/salja/internal/model"
)

func makeSourceItem() *model.CalendarItem {
	return &model.CalendarItem{
		UID:      "src-001",
		Title:    "Source Meeting",
		Priority: model.PriorityHigh,
		Status:   model.StatusPending,
		ItemType: model.ItemTypeEvent,
		Tags:     []string{"work"},
	}
}

func makeTargetItem() *model.CalendarItem {
	return &model.CalendarItem{
		UID:      "tgt-001",
		Title:    "Target Meeting",
		Priority: model.PriorityLow,
		Status:   model.StatusCompleted,
		ItemType: model.ItemTypeTask,
		Tags:     []string{"personal"},
	}
}

func TestResolvePreferSource(t *testing.T) {
	r := NewResolver(StrategyPreferSource)
	source := makeSourceItem()
	target := makeTargetItem()

	result, err := r.Resolve(source, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != source {
		t.Fatal("expected source item to be returned")
	}
	if result.Title != "Source Meeting" {
		t.Fatalf("expected title 'Source Meeting', got '%s'", result.Title)
	}
}

func TestResolvePreferTarget(t *testing.T) {
	r := NewResolver(StrategyPreferTarget)
	source := makeSourceItem()
	target := makeTargetItem()

	result, err := r.Resolve(source, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != target {
		t.Fatal("expected target item to be returned")
	}
	if result.Title != "Target Meeting" {
		t.Fatalf("expected title 'Target Meeting', got '%s'", result.Title)
	}
}

func TestResolveSkip(t *testing.T) {
	r := NewResolver(StrategySkip)
	source := makeSourceItem()
	target := makeTargetItem()

	result, err := r.Resolve(source, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result for skip strategy")
	}
}

func TestResolveFail(t *testing.T) {
	r := NewResolver(StrategyFail)
	source := makeSourceItem()
	target := makeTargetItem()

	result, err := r.Resolve(source, target)
	if result != nil {
		t.Fatal("expected nil result for fail strategy")
	}
	if err == nil {
		t.Fatal("expected error for fail strategy")
	}

	var conflictErr *salerr.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("expected ConflictError, got %T", err)
	}
	if conflictErr.SourceItem != "Source Meeting" {
		t.Fatalf("expected SourceItem 'Source Meeting', got '%s'", conflictErr.SourceItem)
	}
	if conflictErr.TargetItem != "Target Meeting" {
		t.Fatalf("expected TargetItem 'Target Meeting', got '%s'", conflictErr.TargetItem)
	}
}
