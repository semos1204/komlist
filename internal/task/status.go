// Package task defines the core Task entity and its lifecycle statuses.
package task

import "fmt"

// Status represents the lifecycle state of a Task.
type Status string

// Canonical Task statuses.
const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
)

// AllStatuses returns the list of valid statuses in canonical order.
func AllStatuses() []Status {
	return []Status{StatusTodo, StatusInProgress, StatusBlocked, StatusDone}
}

// Valid reports whether s is a known Status.
func (s Status) Valid() bool {
	for _, v := range AllStatuses() {
		if v == s {
			return true
		}
	}
	return false
}

// ParseStatus parses s into a Status. It returns an error listing the valid
// values when s is not recognized.
func ParseStatus(s string) (Status, error) {
	st := Status(s)
	if !st.Valid() {
		return "", fmt.Errorf("invalid status %q (valid: %v)", s, AllStatuses())
	}
	return st, nil
}
