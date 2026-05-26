package task

import (
	"fmt"
	"time"
)

// Recurrence describes how a task repeats once completed.
type Recurrence string

// Supported recurrence cadences. RecurNone means the task does not repeat.
const (
	RecurNone    Recurrence = ""
	RecurDaily   Recurrence = "daily"
	RecurWeekly  Recurrence = "weekly"
	RecurMonthly Recurrence = "monthly"
)

// AllRecurrences returns the selectable recurrence cadences (excluding None).
func AllRecurrences() []Recurrence {
	return []Recurrence{RecurDaily, RecurWeekly, RecurMonthly}
}

// Valid reports whether r is a known cadence. RecurNone is valid (no repeat).
func (r Recurrence) Valid() bool {
	switch r {
	case RecurNone, RecurDaily, RecurWeekly, RecurMonthly:
		return true
	default:
		return false
	}
}

// ParseRecurrence parses s into a Recurrence. "none" (and the empty string)
// map to RecurNone. Unknown values return an error listing valid cadences.
func ParseRecurrence(s string) (Recurrence, error) {
	if s == "none" || s == "" {
		return RecurNone, nil
	}
	r := Recurrence(s)
	if !r.Valid() {
		return "", fmt.Errorf("invalid recurrence %q (valid: none, %v)", s, AllRecurrences())
	}
	return r, nil
}

// Next returns the instant one cadence after from. RecurNone returns from
// unchanged.
func (r Recurrence) Next(from time.Time) time.Time {
	switch r {
	case RecurDaily:
		return from.AddDate(0, 0, 1)
	case RecurWeekly:
		return from.AddDate(0, 0, 7)
	case RecurMonthly:
		return from.AddDate(0, 1, 0)
	default:
		return from
	}
}
