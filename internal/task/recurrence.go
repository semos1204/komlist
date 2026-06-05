package task

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Recurrence describes how a task repeats once completed. It is either a
// keyword (daily/weekly/monthly) or an interval of the form <n><unit> where
// unit is d (days), w (weeks) or mo (months) — e.g. "2w", "3d", "1mo".
type Recurrence string

// Keyword recurrences. RecurNone means the task does not repeat.
//
// RecurWeekdays and RecurWeekends are calendar-aware: their Next jumps to
// the next Monday-to-Friday (resp. Saturday-or-Sunday) date instead of
// adding a fixed interval.
const (
	RecurNone     Recurrence = ""
	RecurDaily    Recurrence = "daily"
	RecurWeekly   Recurrence = "weekly"
	RecurMonthly  Recurrence = "monthly"
	RecurWeekdays Recurrence = "weekdays"
	RecurWeekends Recurrence = "weekends"
)

// AllRecurrences returns the selectable keyword cadences (excluding None).
// Interval forms like "2w" are also accepted by ParseRecurrence.
func AllRecurrences() []Recurrence {
	return []Recurrence{RecurDaily, RecurWeekly, RecurMonthly, RecurWeekdays, RecurWeekends}
}

var intervalRe = regexp.MustCompile(`^([1-9][0-9]*)(d|w|mo)$`)

// parse decomposes a recurrence into a positive count and a unit ("d", "w"
// or "mo"). Keywords map to a count of 1. ok is false for RecurNone or any
// invalid value.
func (r Recurrence) parse() (n int, unit string, ok bool) {
	switch r {
	case RecurNone:
		return 0, "", false
	case RecurDaily:
		return 1, "d", true
	case RecurWeekly:
		return 1, "w", true
	case RecurMonthly:
		return 1, "mo", true
	}
	m := intervalRe.FindStringSubmatch(string(r))
	if m == nil {
		return 0, "", false
	}
	count, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, "", false
	}
	return count, m[2], true
}

// Valid reports whether r is RecurNone, a known keyword, or a valid interval.
func (r Recurrence) Valid() bool {
	switch r {
	case RecurNone, RecurWeekdays, RecurWeekends:
		return true
	}
	_, _, ok := r.parse()
	return ok
}

// ParseRecurrence parses s into a Recurrence. "none" and "" map to RecurNone.
func ParseRecurrence(s string) (Recurrence, error) {
	if s == "none" || s == "" {
		return RecurNone, nil
	}
	r := Recurrence(s)
	if !r.Valid() {
		return "", fmt.Errorf("invalid recurrence %q (valid: none, %v, or an interval like 2w/3d/1mo)", s, AllRecurrences())
	}
	return r, nil
}

// Next returns the instant one cadence after from. RecurNone returns from
// unchanged. RecurWeekdays/Weekends jump to the next matching weekday rather
// than adding a fixed interval.
func (r Recurrence) Next(from time.Time) time.Time {
	switch r {
	case RecurWeekdays:
		return nextMatchingDay(from, isWeekday)
	case RecurWeekends:
		return nextMatchingDay(from, isWeekend)
	}
	n, unit, ok := r.parse()
	if !ok {
		return from
	}
	switch unit {
	case "d":
		return from.AddDate(0, 0, n)
	case "w":
		return from.AddDate(0, 0, 7*n)
	case "mo":
		return from.AddDate(0, n, 0)
	default:
		return from
	}
}

// nextMatchingDay walks forward from `from` until it lands on a day the
// predicate accepts. It never returns `from` itself, so consecutive calls
// always advance.
func nextMatchingDay(from time.Time, match func(time.Weekday) bool) time.Time {
	d := from.AddDate(0, 0, 1)
	for !match(d.Weekday()) {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

func isWeekday(d time.Weekday) bool { return d != time.Saturday && d != time.Sunday }
func isWeekend(d time.Weekday) bool { return d == time.Saturday || d == time.Sunday }
