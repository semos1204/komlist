package service

import (
	"context"

	"github.com/semos1204/komlist/internal/task"
)

// Block makes task id depend on task onID. It rejects self-dependencies
// (ErrSelfDependency) and dependencies that would create a cycle
// (ErrDependencyCycle). Adding an existing dependency is a no-op.
func (s *TaskService) Block(ctx context.Context, id, onID int) (task.Task, error) {
	if id == onID {
		return task.Task{}, ErrSelfDependency
	}
	if _, err := s.repo.Get(ctx, onID); err != nil {
		return task.Task{}, err
	}
	all, err := s.repo.List(ctx)
	if err != nil {
		return task.Task{}, err
	}
	// Adding edge id -> onID forms a cycle iff onID already reaches id.
	if reaches(depGraph(all), onID, id) {
		return task.Task{}, ErrDependencyCycle
	}
	return s.mutate(ctx, id, func(t *task.Task) {
		if !containsInt(t.DependsOn, onID) {
			t.DependsOn = append(t.DependsOn, onID)
		}
	})
}

// Unblock removes onID from task id's dependencies.
func (s *TaskService) Unblock(ctx context.Context, id, onID int) (task.Task, error) {
	return s.mutate(ctx, id, func(t *task.Task) {
		t.DependsOn = removeInt(t.DependsOn, onID)
	})
}

// BlockedSet returns the set of task IDs that have at least one dependency
// which is not yet done. Dependencies referencing missing tasks are ignored.
func (s *TaskService) BlockedSet(ctx context.Context) (map[int]bool, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return blockedSet(all), nil
}

// IncompleteDeps returns the IDs of t's dependencies that are not yet done.
func (s *TaskService) IncompleteDeps(ctx context.Context, t task.Task) []int {
	if len(t.DependsOn) == 0 {
		return nil
	}
	var out []int
	for _, dep := range t.DependsOn {
		d, err := s.repo.Get(ctx, dep)
		if err != nil {
			continue
		}
		if d.Status != task.StatusDone {
			out = append(out, dep)
		}
	}
	return out
}

// blockedSet computes, over the full task list, which task IDs are blocked by
// an incomplete dependency.
func blockedSet(all []task.Task) map[int]bool {
	status := make(map[int]task.Status, len(all))
	for _, t := range all {
		status[t.ID] = t.Status
	}
	out := make(map[int]bool)
	for _, t := range all {
		for _, dep := range t.DependsOn {
			if st, ok := status[dep]; ok && st != task.StatusDone {
				out[t.ID] = true
				break
			}
		}
	}
	return out
}

func depGraph(all []task.Task) map[int][]int {
	g := make(map[int][]int, len(all))
	for _, t := range all {
		g[t.ID] = t.DependsOn
	}
	return g
}

// reaches reports whether target is reachable from start by following
// dependency edges.
func reaches(graph map[int][]int, start, target int) bool {
	seen := make(map[int]bool)
	stack := append([]int(nil), graph[start]...)
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n == target {
			return true
		}
		if seen[n] {
			continue
		}
		seen[n] = true
		stack = append(stack, graph[n]...)
	}
	return false
}

func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func removeInt(xs []int, v int) []int {
	out := make([]int, 0, len(xs))
	for _, x := range xs {
		if x != v {
			out = append(out, x)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
