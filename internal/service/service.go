// Package service contains the komlist use cases. It depends only on the
// task domain and on two ports (storage.Repository, clock.Clock) — never on
// a concrete persistence or I/O mechanism.
package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/semos1204/komlist/internal/clock"
	"github.com/semos1204/komlist/internal/storage"
	"github.com/semos1204/komlist/internal/task"
)

// TaskService orchestrates the task use cases.
type TaskService struct {
	repo  storage.Repository
	clock clock.Clock
}

// New returns a TaskService wired with the given repository and clock.
func New(repo storage.Repository, clk clock.Clock) *TaskService {
	return &TaskService{repo: repo, clock: clk}
}

// SortBy selects an ordering for List results.
type SortBy int

// Sort keys understood by List.
const (
	SortByID SortBy = iota
	SortByDueAt
	SortByPriority
	SortByStatus
)

// ParseSortBy parses a user-supplied sort key. An empty string maps to
// SortByID.
func ParseSortBy(s string) (SortBy, error) {
	switch s {
	case "", "id":
		return SortByID, nil
	case "due", "due-date":
		return SortByDueAt, nil
	case "priority", "prio":
		return SortByPriority, nil
	case "status":
		return SortByStatus, nil
	default:
		return 0, fmt.Errorf("%w: %q (valid: id, due, priority, status)", ErrInvalidSort, s)
	}
}

// ListFilter narrows and orders the result of List. The zero value returns
// every task, sorted by ID ascending.
type ListFilter struct {
	Status *task.Status
	Tag    string
	Sort   SortBy
}

// Add creates a new Task with status "todo" and timestamps set from the
// clock. The title is trimmed; a blank title returns ErrEmptyTitle.
func (s *TaskService) Add(ctx context.Context, title string) (task.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return task.Task{}, ErrEmptyTitle
	}
	now := s.clock.Now()
	t := task.Task{
		Title:     title,
		Status:    task.StatusTodo,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repo.Create(ctx, t)
}

// List returns tasks matching the filter, sorted according to f.Sort.
func (s *TaskService) List(ctx context.Context, f ListFilter) ([]task.Task, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]task.Task, 0, len(all))
	for _, t := range all {
		if f.Status != nil && t.Status != *f.Status {
			continue
		}
		if f.Tag != "" && !hasTag(t.Tags, f.Tag) {
			continue
		}
		out = append(out, t)
	}
	sortTasks(out, f.Sort)
	return out, nil
}

// ChangeStatus updates the task's status and refreshes UpdatedAt. Returns
// ErrInvalidStatus if st is not a known status, storage.ErrNotFound if no
// task has the given id.
func (s *TaskService) ChangeStatus(ctx context.Context, id int, st task.Status) (task.Task, error) {
	if !st.Valid() {
		return task.Task{}, fmt.Errorf("%w: %q (valid: %v)", ErrInvalidStatus, st, task.AllStatuses())
	}
	return s.mutate(ctx, id, func(t *task.Task) { t.Status = st })
}

// Rename changes a task's title and refreshes UpdatedAt.
func (s *TaskService) Rename(ctx context.Context, id int, title string) (task.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return task.Task{}, ErrEmptyTitle
	}
	return s.mutate(ctx, id, func(t *task.Task) { t.Title = title })
}

// SetTags replaces a task's tags with the given slice. Tags are trimmed,
// de-duplicated, and a nil/empty slice clears tags entirely.
func (s *TaskService) SetTags(ctx context.Context, id int, tags []string) (task.Task, error) {
	cleaned := normalizeTags(tags)
	return s.mutate(ctx, id, func(t *task.Task) { t.Tags = cleaned })
}

// SetPriority changes a task's priority.
func (s *TaskService) SetPriority(ctx context.Context, id int, p task.Priority) (task.Task, error) {
	if !p.Valid() {
		return task.Task{}, fmt.Errorf("%w: %q (valid: %v)", ErrInvalidPriority, p, task.AllPriorities())
	}
	return s.mutate(ctx, id, func(t *task.Task) { t.Priority = p })
}

// SetDueAt sets or clears a task's due date. Pass nil to clear.
func (s *TaskService) SetDueAt(ctx context.Context, id int, due *time.Time) (task.Task, error) {
	return s.mutate(ctx, id, func(t *task.Task) { t.DueAt = due })
}

// Delete removes the task with the given id.
func (s *TaskService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// TagCount associates a tag with the number of tasks carrying it.
type TagCount struct {
	Tag   string
	Count int
}

// Tags returns every distinct tag currently in use across all tasks, with
// the number of tasks each tag is attached to, sorted alphabetically.
func (s *TaskService) Tags(ctx context.Context) ([]TagCount, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, t := range all {
		for _, tag := range t.Tags {
			counts[tag]++
		}
	}
	out := make([]TagCount, 0, len(counts))
	for tag, n := range counts {
		out = append(out, TagCount{Tag: tag, Count: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Tag < out[j].Tag })
	return out, nil
}

// DeleteTag removes the given tag from every task that carries it. Returns
// the number of tasks affected. Removing a tag that no task carries is a
// no-op and returns 0 without an error.
func (s *TaskService) DeleteTag(ctx context.Context, tag string) (int, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return 0, err
	}
	affected := 0
	for _, t := range all {
		if !hasTag(t.Tags, tag) {
			continue
		}
		t.Tags = withoutTag(t.Tags, tag)
		t.UpdatedAt = s.clock.Now()
		if err := s.repo.Update(ctx, t); err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

// RenameTag renames the tag `from` to `to` across every task. The new name
// is trimmed; an empty target returns ErrEmptyTag. Tasks that already carry
// `to` see the duplicate de-duplicated. Returns the number of tasks
// affected; renaming to the same name is a no-op.
func (s *TaskService) RenameTag(ctx context.Context, from, to string) (int, error) {
	to = strings.TrimSpace(to)
	if to == "" {
		return 0, ErrEmptyTag
	}
	if from == to {
		return 0, nil
	}
	all, err := s.repo.List(ctx)
	if err != nil {
		return 0, err
	}
	affected := 0
	for _, t := range all {
		if !hasTag(t.Tags, from) {
			continue
		}
		renamed := make([]string, 0, len(t.Tags))
		for _, tag := range t.Tags {
			if tag == from {
				renamed = append(renamed, to)
			} else {
				renamed = append(renamed, tag)
			}
		}
		t.Tags = normalizeTags(renamed)
		t.UpdatedAt = s.clock.Now()
		if err := s.repo.Update(ctx, t); err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func withoutTag(tags []string, drop string) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if t != drop {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// mutate loads a task, applies fn, refreshes UpdatedAt and persists.
func (s *TaskService) mutate(ctx context.Context, id int, fn func(*task.Task)) (task.Task, error) {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return task.Task{}, err
	}
	fn(&t)
	t.UpdatedAt = s.clock.Now()
	if err := s.repo.Update(ctx, t); err != nil {
		return task.Task{}, err
	}
	return t, nil
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		t := strings.TrimSpace(raw)
		if t == "" {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func hasTag(tags []string, target string) bool {
	for _, t := range tags {
		if t == target {
			return true
		}
	}
	return false
}

func sortTasks(tasks []task.Task, by SortBy) {
	switch by {
	case SortByDueAt:
		sort.SliceStable(tasks, func(i, j int) bool {
			return lessDue(tasks[i].DueAt, tasks[j].DueAt)
		})
	case SortByPriority:
		sort.SliceStable(tasks, func(i, j int) bool {
			return tasks[i].Priority.Rank() > tasks[j].Priority.Rank()
		})
	case SortByStatus:
		sort.SliceStable(tasks, func(i, j int) bool {
			return statusRank(tasks[i].Status) < statusRank(tasks[j].Status)
		})
	default:
		sort.SliceStable(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	}
}

// lessDue orders by due date ascending; nil dues sort last.
func lessDue(a, b *time.Time) bool {
	switch {
	case a == nil && b == nil:
		return false
	case a == nil:
		return false
	case b == nil:
		return true
	default:
		return a.Before(*b)
	}
}

func statusRank(s task.Status) int {
	switch s {
	case task.StatusTodo:
		return 0
	case task.StatusInProgress:
		return 1
	case task.StatusBlocked:
		return 2
	case task.StatusDone:
		return 3
	default:
		return 99
	}
}
