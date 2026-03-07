package errors

import (
	"fmt"
	"strings"
)

// PartialResult holds successfully converted items and collected errors.
type PartialResult[T any] struct {
	Items  []T
	Errors []ItemError
	Total  int
}

// ItemError records a failure for a specific item.
type ItemError struct {
	Index   int
	ItemID  string
	Message string
	Err     error
}

func (e *ItemError) Error() string {
	if e.ItemID != "" {
		return fmt.Sprintf("item %d (%s): %s", e.Index, e.ItemID, e.Message)
	}
	return fmt.Sprintf("item %d: %s", e.Index, e.Message)
}

// NewPartialResult creates a new partial result collector.
func NewPartialResult[T any]() *PartialResult[T] {
	return &PartialResult[T]{}
}

// Add records a successful item.
func (p *PartialResult[T]) Add(item T) {
	p.Items = append(p.Items, item)
	p.Total++
}

// AddError records a failed item.
func (p *PartialResult[T]) AddError(index int, itemID, message string, err error) {
	p.Errors = append(p.Errors, ItemError{
		Index:   index,
		ItemID:  itemID,
		Message: message,
		Err:     err,
	})
	p.Total++
}

// HasErrors returns true if any items failed.
func (p *PartialResult[T]) HasErrors() bool {
	return len(p.Errors) > 0
}

// SuccessCount returns number of successful items.
func (p *PartialResult[T]) SuccessCount() int {
	return len(p.Items)
}

// AllErrors returns the full untruncated error list.
func (p *PartialResult[T]) AllErrors() []ItemError {
	return p.Errors
}

// Summary returns a human-readable summary, capped at 20 errors.
func (p *PartialResult[T]) Summary() string {
	if !p.HasErrors() {
		return fmt.Sprintf("all %d items converted successfully", p.Total)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d/%d items converted (%d errors):\n", p.SuccessCount(), p.Total, len(p.Errors))
	limit := len(p.Errors)
	if limit > 20 {
		limit = 20
	}
	for _, e := range p.Errors[:limit] {
		fmt.Fprintf(&sb, "  - %s\n", e.Error())
	}
	if len(p.Errors) > 20 {
		fmt.Fprintf(&sb, "  ... and %d more errors (see log for full details)\n", len(p.Errors)-20)
	}
	return sb.String()
}
