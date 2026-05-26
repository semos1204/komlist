package service

import (
	"testing"
	"time"

	"github.com/semos1204/komlist/internal/task"
)

func TestUrgency_Ordering(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	mk := func(mut func(*task.Task)) task.Task {
		tk := task.Task{Status: task.StatusTodo, CreatedAt: now}
		mut(&tk)
		return tk
	}

	overdueHigh := mk(func(t *task.Task) {
		t.Priority = task.PriorityHigh
		due := now.AddDate(0, 0, -1)
		t.DueAt = &due
	})
	plainTodo := mk(func(t *task.Task) {})
	doneTask := mk(func(t *task.Task) { t.Status = task.StatusDone; t.Priority = task.PriorityHigh })
	blockedTask := mk(func(t *task.Task) { t.Status = task.StatusBlocked })

	if urgency(overdueHigh, now) <= urgency(plainTodo, now) {
		t.Error("overdue+high should outrank a plain todo")
	}
	if urgency(doneTask, now) != 0 {
		t.Errorf("done task urgency = %v, want 0", urgency(doneTask, now))
	}
	if urgency(blockedTask, now) >= urgency(plainTodo, now) {
		t.Error("blocked task should sink below a plain todo")
	}
}

func TestParseSortBy_Urgency(t *testing.T) {
	got, err := ParseSortBy("urgency")
	if err != nil {
		t.Fatalf("ParseSortBy: %v", err)
	}
	if got != SortByUrgency {
		t.Errorf("got %v, want SortByUrgency", got)
	}
}
