package ics

import (
	"fmt"
	"os"
	"time"

	"github.com/gongahkia/salja/internal/model"
)

type Flattener struct{}

func NewFlattener() *Flattener {
	return &Flattener{}
}

func (f *Flattener) FlattenRecurrence(item *model.CalendarItem, maxOccurrences int, endDate time.Time) []model.CalendarItem {
	if item.Recurrence == nil {
		return []model.CalendarItem{*item}
	}

	fmt.Fprintf(os.Stderr, "Warning: Flattening recurring item '%s' - target app lacks RRULE support\n", item.Title)

	instances := []model.CalendarItem{}
	rec := item.Recurrence

	current := item.StartTime
	if current == nil && item.DueDate != nil {
		current = item.DueDate
	}
	if current == nil {
		return []model.CalendarItem{*item}
	}

	count := 0
	maxCount := maxOccurrences
	if rec.Count != nil && *rec.Count < maxCount {
		maxCount = *rec.Count
	}

	until := endDate
	if rec.Until != nil && rec.Until.Before(until) {
		until = *rec.Until
	}

	for count < maxCount {
		instance := *item
		// Copy recurrence to avoid mutating the original
		instance.Recurrence = nil

		if isExDate(*current, rec.ExDates) {
			current = f.advance(current, rec)
			continue
		}

		if item.StartTime != nil {
			newStart := *current
			instance.StartTime = &newStart
		}
		if item.DueDate != nil {
			newDue := *current
			instance.DueDate = &newDue
		}
		if item.EndTime != nil && item.StartTime != nil {
			duration := item.EndTime.Sub(*item.StartTime)
			newEnd := current.Add(duration)
			instance.EndTime = &newEnd
		}

		instance.UID = fmt.Sprintf("%s-occurrence-%d", item.UID, count)
		instances = append(instances, instance)
		count++

		current = f.advance(current, rec)
		if current.After(until) {
			break
		}
	}

	// Add RDATE instances, deduplicating against RRULE-generated dates
	seen := make(map[string]bool)
	for _, inst := range instances {
		if inst.StartTime != nil {
			seen[inst.StartTime.Format("2006-01-02")] = true
		}
	}
	for _, rdate := range rec.RDates {
		dateKey := rdate.Format("2006-01-02")
		if !isExDate(rdate, rec.ExDates) && rdate.Before(until) && !seen[dateKey] {
			instance := *item
			instance.Recurrence = nil
			rdateCopy := rdate
			if item.StartTime != nil {
				instance.StartTime = &rdateCopy
			}
			if item.DueDate != nil {
				instance.DueDate = &rdateCopy
			}
			instance.UID = fmt.Sprintf("%s-rdate-%d", item.UID, rdate.Unix())
			instances = append(instances, instance)
			seen[dateKey] = true
		}
	}

	return instances
}

func (f *Flattener) advance(current *time.Time, rec *model.Recurrence) *time.Time {
	if current == nil {
		return nil
	}

	next := *current
	interval := rec.Interval
	if interval == 0 {
		interval = 1
	}

	switch rec.Freq {
	case model.FreqDaily:
		next = next.AddDate(0, 0, interval)
	case model.FreqWeekly:
		if len(rec.ByDay) > 0 {
			// Advance day-by-day until we hit the next BYDAY match within the interval
			for i := 0; i < 7*interval; i++ {
				next = next.AddDate(0, 0, 1)
				if matchesWeekday(next, rec.ByDay) {
					return &next
				}
			}
			// Fallback if no match
			next = current.AddDate(0, 0, 7*interval)
		} else {
			next = next.AddDate(0, 0, 7*interval)
		}
	case model.FreqMonthly:
		if len(rec.ByMonthDay) > 0 {
			// Advance to next month, then to the specified day
			next = next.AddDate(0, interval, 0)
			for _, day := range rec.ByMonthDay {
				candidate := time.Date(next.Year(), next.Month(), day, next.Hour(), next.Minute(), next.Second(), next.Nanosecond(), next.Location())
				if candidate.After(*current) {
					return &candidate
				}
			}
		} else {
			next = next.AddDate(0, interval, 0)
		}
	case model.FreqYearly:
		if len(rec.ByMonth) > 0 {
			next = next.AddDate(interval, 0, 0)
			for _, month := range rec.ByMonth {
				candidate := time.Date(next.Year(), time.Month(month), next.Day(), next.Hour(), next.Minute(), next.Second(), next.Nanosecond(), next.Location())
				if candidate.After(*current) {
					return &candidate
				}
			}
		} else {
			next = next.AddDate(interval, 0, 0)
		}
	}

	return &next
}

func matchesWeekday(t time.Time, byDay []model.Weekday) bool {
	dayMap := map[time.Weekday]model.Weekday{
		time.Monday:    model.WeekdayMO,
		time.Tuesday:   model.WeekdayTU,
		time.Wednesday: model.WeekdayWE,
		time.Thursday:  model.WeekdayTH,
		time.Friday:    model.WeekdayFR,
		time.Saturday:  model.WeekdaySA,
		time.Sunday:    model.WeekdaySU,
	}
	wd, ok := dayMap[t.Weekday()]
	if !ok {
		return false
	}
	for _, d := range byDay {
		if d == wd {
			return true
		}
	}
	return false
}

func isExDate(t time.Time, exdates []time.Time) bool {
	for _, ex := range exdates {
		if t.Year() == ex.Year() && t.Month() == ex.Month() && t.Day() == ex.Day() {
			return true
		}
	}
	return false
}
