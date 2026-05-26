package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/semos1204/komlist/internal/clock"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/storage"
	"github.com/semos1204/komlist/internal/task"
)

func newSvc() (*service.TaskService, *clock.Fake) {
	fake := clock.NewFake(time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC))
	return service.New(storage.NewMemory(), fake), fake
}

func TestAdd(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()

	got, err := svc.Add(ctx, "  buy bread  ")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if got.Title != "buy bread" {
		t.Errorf("title not trimmed: %q", got.Title)
	}
	if got.Status != task.StatusTodo {
		t.Errorf("status %q, want %q", got.Status, task.StatusTodo)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("timestamps not set")
	}
	if !got.CreatedAt.Equal(got.UpdatedAt) {
		t.Error("CreatedAt != UpdatedAt on new task")
	}
}

func TestAdd_EmptyTitle(t *testing.T) {
	svc, _ := newSvc()
	if _, err := svc.Add(context.Background(), "   "); !errors.Is(err, service.ErrEmptyTitle) {
		t.Errorf("got %v, want ErrEmptyTitle", err)
	}
}

func TestList_FilterByStatus(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	if _, err := svc.Add(ctx, "a"); err != nil {
		t.Fatalf("add a: %v", err)
	}
	b, err := svc.Add(ctx, "b")
	if err != nil {
		t.Fatalf("add b: %v", err)
	}
	if _, err := svc.ChangeStatus(ctx, b.ID, task.StatusDone); err != nil {
		t.Fatalf("change: %v", err)
	}

	done := task.StatusDone
	got, err := svc.List(ctx, service.ListFilter{Status: &done})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].ID != b.ID {
		t.Errorf("unexpected filtered list: %+v", got)
	}

	all, err := svc.List(ctx, service.ListFilter{})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 tasks total, got %d", len(all))
	}
}

func TestChangeStatus(t *testing.T) {
	svc, clk := newSvc()
	ctx := context.Background()
	created, err := svc.Add(ctx, "x")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	clk.Advance(time.Hour)

	updated, err := svc.ChangeStatus(ctx, created.ID, task.StatusInProgress)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if updated.Status != task.StatusInProgress {
		t.Errorf("status %q, want %q", updated.Status, task.StatusInProgress)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Errorf("expected UpdatedAt > CreatedAt, got %v vs %v", updated.UpdatedAt, updated.CreatedAt)
	}
}

func TestChangeStatus_InvalidStatus(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	created, err := svc.Add(ctx, "x")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if _, err := svc.ChangeStatus(ctx, created.ID, task.Status("bogus")); !errors.Is(err, service.ErrInvalidStatus) {
		t.Errorf("got %v, want ErrInvalidStatus", err)
	}
}

func TestChangeStatus_NotFound(t *testing.T) {
	svc, _ := newSvc()
	if _, err := svc.ChangeStatus(context.Background(), 999, task.StatusDone); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestRename(t *testing.T) {
	svc, clk := newSvc()
	ctx := context.Background()
	created, err := svc.Add(ctx, "old")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	clk.Advance(time.Hour)

	renamed, err := svc.Rename(ctx, created.ID, "  new title  ")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if renamed.Title != "new title" {
		t.Errorf("title = %q, want %q", renamed.Title, "new title")
	}
	if !renamed.UpdatedAt.After(renamed.CreatedAt) {
		t.Error("expected UpdatedAt to advance")
	}
}

func TestRename_EmptyTitle(t *testing.T) {
	svc, _ := newSvc()
	created, _ := svc.Add(context.Background(), "x")
	if _, err := svc.Rename(context.Background(), created.ID, "   "); !errors.Is(err, service.ErrEmptyTitle) {
		t.Errorf("got %v, want ErrEmptyTitle", err)
	}
}

func TestSetTags_DedupeAndTrim(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	created, _ := svc.Add(ctx, "x")
	got, err := svc.SetTags(ctx, created.ID, []string{" work ", "work", "", "urgent"})
	if err != nil {
		t.Fatalf("setTags: %v", err)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "work" || got.Tags[1] != "urgent" {
		t.Errorf("tags = %v, want [work urgent]", got.Tags)
	}

	cleared, err := svc.SetTags(ctx, created.ID, nil)
	if err != nil {
		t.Fatalf("clear: %v", err)
	}
	if cleared.Tags != nil {
		t.Errorf("expected nil tags after clear, got %v", cleared.Tags)
	}
}

func TestSetPriority(t *testing.T) {
	svc, _ := newSvc()
	created, _ := svc.Add(context.Background(), "x")
	got, err := svc.SetPriority(context.Background(), created.ID, task.PriorityHigh)
	if err != nil {
		t.Fatalf("setPriority: %v", err)
	}
	if got.Priority != task.PriorityHigh {
		t.Errorf("priority = %q, want %q", got.Priority, task.PriorityHigh)
	}
}

func TestSetPriority_Invalid(t *testing.T) {
	svc, _ := newSvc()
	created, _ := svc.Add(context.Background(), "x")
	if _, err := svc.SetPriority(context.Background(), created.ID, task.Priority("bogus")); !errors.Is(err, service.ErrInvalidPriority) {
		t.Errorf("got %v, want ErrInvalidPriority", err)
	}
}

func TestSetDueAt(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	created, _ := svc.Add(ctx, "x")
	due := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	got, err := svc.SetDueAt(ctx, created.ID, &due)
	if err != nil {
		t.Fatalf("setDue: %v", err)
	}
	if got.DueAt == nil || !got.DueAt.Equal(due) {
		t.Errorf("dueAt = %v, want %v", got.DueAt, due)
	}

	cleared, err := svc.SetDueAt(ctx, created.ID, nil)
	if err != nil {
		t.Fatalf("clear due: %v", err)
	}
	if cleared.DueAt != nil {
		t.Errorf("expected nil after clear, got %v", cleared.DueAt)
	}
}

func TestList_FilterByTag(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	a, _ := svc.Add(ctx, "a")
	b, _ := svc.Add(ctx, "b")
	if _, err := svc.SetTags(ctx, a.ID, []string{"work"}); err != nil {
		t.Fatalf("setTags a: %v", err)
	}
	if _, err := svc.SetTags(ctx, b.ID, []string{"personal"}); err != nil {
		t.Fatalf("setTags b: %v", err)
	}
	got, err := svc.List(ctx, service.ListFilter{Tag: "work"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].ID != a.ID {
		t.Errorf("filter by tag wrong: %+v", got)
	}
}

func TestList_SortByPriority(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	low, _ := svc.Add(ctx, "low")
	high, _ := svc.Add(ctx, "high")
	med, _ := svc.Add(ctx, "med")
	_, _ = svc.SetPriority(ctx, low.ID, task.PriorityLow)
	_, _ = svc.SetPriority(ctx, high.ID, task.PriorityHigh)
	_, _ = svc.SetPriority(ctx, med.ID, task.PriorityMedium)

	got, err := svc.List(ctx, service.ListFilter{Sort: service.SortByPriority})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
	if got[0].Priority != task.PriorityHigh ||
		got[1].Priority != task.PriorityMedium ||
		got[2].Priority != task.PriorityLow {
		t.Errorf("unexpected order: %v %v %v", got[0].Priority, got[1].Priority, got[2].Priority)
	}
}

func TestList_SortByDue_NilsLast(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	noDue, _ := svc.Add(ctx, "no-due")
	late, _ := svc.Add(ctx, "late")
	early, _ := svc.Add(ctx, "early")
	lateAt := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	earlyAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, _ = svc.SetDueAt(ctx, late.ID, &lateAt)
	_, _ = svc.SetDueAt(ctx, early.ID, &earlyAt)

	got, err := svc.List(ctx, service.ListFilter{Sort: service.SortByDueAt})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if got[0].ID != early.ID || got[1].ID != late.ID || got[2].ID != noDue.ID {
		t.Errorf("unexpected order: %v", []int{got[0].ID, got[1].ID, got[2].ID})
	}
}

func TestParseSortBy(t *testing.T) {
	cases := map[string]service.SortBy{
		"":         service.SortByID,
		"id":       service.SortByID,
		"due":      service.SortByDueAt,
		"priority": service.SortByPriority,
		"prio":     service.SortByPriority,
		"status":   service.SortByStatus,
	}
	for in, want := range cases {
		got, err := service.ParseSortBy(in)
		if err != nil {
			t.Errorf("ParseSortBy(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseSortBy(%q) = %v, want %v", in, got, want)
		}
	}
	if _, err := service.ParseSortBy("nope"); !errors.Is(err, service.ErrInvalidSort) {
		t.Errorf("got %v, want ErrInvalidSort", err)
	}
}

func TestDelete(t *testing.T) {
	svc, _ := newSvc()
	ctx := context.Background()
	created, err := svc.Add(ctx, "x")
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := svc.Delete(ctx, created.ID); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("second delete got %v, want ErrNotFound", err)
	}
}
