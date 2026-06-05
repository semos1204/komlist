package task

import "fmt"

// Priority represents the importance of a Task.
type Priority string

// Canonical priorities, ordered from least to most important.
const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// AllPriorities returns the list of valid priorities in canonical order.
func AllPriorities() []Priority {
	return []Priority{PriorityLow, PriorityMedium, PriorityHigh}
}

// Valid reports whether p is a known Priority.
func (p Priority) Valid() bool {
	for _, v := range AllPriorities() {
		if v == p {
			return true
		}
	}
	return false
}

// Rank returns an ordering number for sorting: high > medium > low. An
// unset priority sorts last.
func (p Priority) Rank() int {
	switch p {
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

// ParsePriority parses s into a Priority. "none" (and the empty string) map
// to an unset Priority; any other unknown value returns an error listing the
// valid keywords.
func ParsePriority(s string) (Priority, error) {
	if s == "none" || s == "" {
		return "", nil
	}
	p := Priority(s)
	if !p.Valid() {
		return "", fmt.Errorf("invalid priority %q (valid: none, %v)", s, AllPriorities())
	}
	return p, nil
}
