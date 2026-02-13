package model

import (
	"fmt"
	"time"

	salerr "github.com/gongahkia/salja/internal/errors"
)

type ItemType string

const (
	ItemTypeEvent   ItemType = "event"
	ItemTypeTask    ItemType = "task"
	ItemTypeJournal ItemType = "journal"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type Priority int

const (
	PriorityNone   Priority = 0
	PriorityLowest Priority = 1
	PriorityLow    Priority = 2
	PriorityMedium Priority = 3
	PriorityHigh   Priority = 4
	PriorityHighest Priority = 5
)

type FreqType string

const (
	FreqDaily   FreqType = "DAILY"
	FreqWeekly  FreqType = "WEEKLY"
	FreqMonthly FreqType = "MONTHLY"
	FreqYearly  FreqType = "YEARLY"
)

type Weekday string

const (
	WeekdayMO Weekday = "MO"
	WeekdayTU Weekday = "TU"
	WeekdayWE Weekday = "WE"
	WeekdayTH Weekday = "TH"
	WeekdayFR Weekday = "FR"
	WeekdaySA Weekday = "SA"
	WeekdaySU Weekday = "SU"
)

type Recurrence struct {
	Freq       FreqType
	Interval   int
	Count      *int
	Until      *time.Time
	ByDay      []Weekday
	ByMonth    []int
	ByMonthDay []int
	BySetPos   []int
	ExDates    []time.Time
	RDates     []time.Time
}

type Reminder struct {
	Offset       *time.Duration
	AbsoluteTime *time.Time
}

type Subtask struct {
	Title     string
	Status    Status
	Priority  Priority
	SortOrder int
}

type CalendarItem struct {
	UID            string
	Title          string
	Description    string
	StartTime      *time.Time
	EndTime        *time.Time
	DueDate        *time.Time
	Priority       Priority
	Status         Status
	Location       string
	Tags           []string
	Recurrence     *Recurrence
	Reminders      []Reminder
	Subtasks       []Subtask
	CompletionDate *time.Time
	Timezone       string
	IsAllDay       bool
	ItemType       ItemType
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
}

func (item *CalendarItem) Validate() error {
	if item.Title == "" {
		return &salerr.ValidationError{Field: "title", Message: "is required"}
	}
	if item.Priority < 0 || item.Priority > 5 {
		return &salerr.ValidationError{Field: "priority", Message: fmt.Sprintf("must be 0-5, got %d", item.Priority)}
	}
	if item.StartTime != nil && item.EndTime != nil && item.StartTime.After(*item.EndTime) {
		return &salerr.ValidationError{Field: "start_time", Message: fmt.Sprintf("start time %v is after end time %v", *item.StartTime, *item.EndTime)}
	}
	if item.ItemType != "" && item.ItemType != ItemTypeEvent && item.ItemType != ItemTypeTask && item.ItemType != ItemTypeJournal {
		return &salerr.ValidationError{Field: "item_type", Message: fmt.Sprintf("invalid value: %s", item.ItemType)}
	}
	return nil
}

func (item *CalendarItem) IsCompleted() bool {
	return item.Status == StatusCompleted
}

func (item *CalendarItem) HasRecurrence() bool {
	return item.Recurrence != nil
}

func (item *CalendarItem) HasSubtasks() bool {
	return len(item.Subtasks) > 0
}

func (item *CalendarItem) Duration() time.Duration {
	if item.StartTime == nil || item.EndTime == nil {
		return 0
	}
	return item.EndTime.Sub(*item.StartTime)
}

type CalendarCollection struct {
	Items            []CalendarItem
	SourceApp        string
	ExportDate       time.Time
	OriginalFilePath string
}
