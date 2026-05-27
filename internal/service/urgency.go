package service

import (
	"time"

	"github.com/semos1204/komlist/internal/task"
)

// urgency computes a heuristic urgency score for a task, used by
// SortByUrgency. Higher means more pressing. The weights are inspired by
// Taskwarrior and are intentionally simple to reason about; tune freely.
//
// Contributions:
//   - done tasks score 0 (they sink to the bottom)
//   - priority: high +6, medium +3.9, low +1.8, unset 0
//   - status: in-progress +4, blocked -5 (not actionable, sinks)
//   - blocked by an incomplete dependency: -8 (sinks below plain todos)
//   - due date: overdue +12, today +9, <=1d +8, <=3d +6, <=7d +4, <=14d +2, else +0.5
//   - age: +0.02 per day since creation, capped at +2
func urgency(t task.Task, now time.Time, blocked bool) float64 {
	if t.Status == task.StatusDone {
		return 0
	}

	var score float64

	switch t.Priority {
	case task.PriorityHigh:
		score += 6
	case task.PriorityMedium:
		score += 3.9
	case task.PriorityLow:
		score += 1.8
	}

	switch t.Status {
	case task.StatusInProgress:
		score += 4
	case task.StatusBlocked:
		score -= 5
	}

	if blocked {
		score -= 8
	}

	if t.DueAt != nil {
		score += dueUrgency(t.DueAt.Sub(now))
	}

	ageDays := now.Sub(t.CreatedAt).Hours() / 24
	if ageDays > 0 {
		bonus := ageDays * 0.02
		if bonus > 2 {
			bonus = 2
		}
		score += bonus
	}

	return score
}

func dueUrgency(remaining time.Duration) float64 {
	days := remaining.Hours() / 24
	switch {
	case days < 0:
		return 12
	case days < 1:
		return 9
	case days <= 1:
		return 8
	case days <= 3:
		return 6
	case days <= 7:
		return 4
	case days <= 14:
		return 2
	default:
		return 0.5
	}
}
