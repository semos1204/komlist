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

// ParsePriority parses s into a Priority. It returns an error listing the
// valid values when s is not recognized.
func ParsePriority(s string) (Priority, error) {
	p := Priority(s)
	if !p.Valid() {
		return "", fmt.Errorf("invalid priority %q (valid: %v)", s, AllPriorities())
	}
	return p, nil
}
