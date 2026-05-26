package task

import "time"

// Task is the unit of work tracked by komlist.
//
// Optional fields (Priority, Tags, DueAt) use omitempty so unset values do
// not pollute the on-disk JSON of tasks that don't use them.
type Task struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Status    Status     `json:"status"`
	Priority  Priority   `json:"priority,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	DueAt     *time.Time `json:"due_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
