package conflict

import (
	"strings"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type Detector struct{}

func NewDetector() *Detector {
	return &Detector{}
}

type Match struct {
	SourceIndex int
	TargetIndex int
	Confidence  float64
}

func (d *Detector) FindDuplicates(source, target *model.CalendarCollection) []Match {
	var matches []Match

	// Build UID index for O(1) UID lookups
	uidIndex := make(map[string]int)
	for j, tgtItem := range target.Items {
		if tgtItem.UID != "" {
			uidIndex[tgtItem.UID] = j
		}
	}

	matched := make(map[int]bool)

	// First pass: exact UID matches
	for i, srcItem := range source.Items {
		if srcItem.UID != "" {
			if j, ok := uidIndex[srcItem.UID]; ok {
				matches = append(matches, Match{
					SourceIndex: i,
					TargetIndex: j,
					Confidence:  d.calculateConfidence(&srcItem, &target.Items[j]),
				})
				matched[j] = true
			}
		}
	}

	// Second pass: fuzzy matching for items without UID match
	for i, srcItem := range source.Items {
		if srcItem.UID != "" {
			if _, ok := uidIndex[srcItem.UID]; ok {
				continue // Already matched by UID
			}
		}
		for j, tgtItem := range target.Items {
			if matched[j] {
				continue
			}
			if d.isDuplicate(&srcItem, &tgtItem) {
				matches = append(matches, Match{
					SourceIndex: i,
					TargetIndex: j,
					Confidence:  d.calculateConfidence(&srcItem, &tgtItem),
				})
			}
		}
	}

	return matches
}

func (d *Detector) isDuplicate(a, b *model.CalendarItem) bool {
	if a.UID != "" && b.UID != "" && a.UID == b.UID {
		return true
	}

	if !d.fuzzyTitleMatch(a.Title, b.Title) {
		return false
	}

	if a.StartTime != nil && b.StartTime != nil {
		if !d.timeMatch(*a.StartTime, *b.StartTime) {
			return false
		}
		return true
	}

	if a.DueDate != nil && b.DueDate != nil {
		if !d.timeMatch(*a.DueDate, *b.DueDate) {
			return false
		}
		return true
	}

	return false
}

func (d *Detector) fuzzyTitleMatch(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	
	if a == b {
		return true
	}

	if len(a) > 10 && len(b) > 10 {
		return levenshteinDistance(a, b) < 3
	}

	return false
}

func (d *Detector) timeMatch(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func (d *Detector) calculateConfidence(a, b *model.CalendarItem) float64 {
	score := 0.0

	if a.UID != "" && b.UID != "" && a.UID == b.UID {
		score += 1.0
	}

	if strings.ToLower(a.Title) == strings.ToLower(b.Title) {
		score += 0.5
	} else if d.fuzzyTitleMatch(a.Title, b.Title) {
		score += 0.3
	}

	if a.StartTime != nil && b.StartTime != nil && d.timeMatch(*a.StartTime, *b.StartTime) {
		score += 0.3
	}

	if a.DueDate != nil && b.DueDate != nil && d.timeMatch(*a.DueDate, *b.DueDate) {
		score += 0.3
	}

	return score
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(nums ...int) int {
	m := nums[0]
	for _, n := range nums[1:] {
		if n < m {
			m = n
		}
	}
	return m
}
