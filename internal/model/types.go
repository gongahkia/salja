package model

import "time"

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
}

type CalendarCollection struct {
	Items            []CalendarItem
	SourceApp        string
	ExportDate       time.Time
	OriginalFilePath string
}
