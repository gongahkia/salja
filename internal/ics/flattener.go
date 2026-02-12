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

	for _, rdate := range rec.RDates {
		if !isExDate(rdate, rec.ExDates) && rdate.Before(until) {
			instance := *item
			instance.Recurrence = nil
			if item.StartTime != nil {
				instance.StartTime = &rdate
			}
			if item.DueDate != nil {
				instance.DueDate = &rdate
			}
			instance.UID = fmt.Sprintf("%s-rdate-%d", item.UID, rdate.Unix())
			instances = append(instances, instance)
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
		next = next.AddDate(0, 0, 7*interval)
	case model.FreqMonthly:
		next = next.AddDate(0, interval, 0)
	case model.FreqYearly:
		next = next.AddDate(interval, 0, 0)
	}

	return &next
}

func isExDate(t time.Time, exdates []time.Time) bool {
	for _, ex := range exdates {
		if t.Year() == ex.Year() && t.Month() == ex.Month() && t.Day() == ex.Day() {
			return true
		}
	}
	return false
}
