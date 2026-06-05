package tui

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/semos1204/komlist/internal/clock"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/storage"
	"github.com/semos1204/komlist/internal/task"
)

func testModel(t *testing.T) model {
	t.Helper()
	svc := service.New(storage.NewMemory(), clock.NewFake(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	for _, title := range []string{"a", "b", "c"} {
		if _, err := svc.Add(ctx, title); err != nil {
			t.Fatalf("add: %v", err)
		}
	}
	return newModel(svc)
}

func testModelTagged(t *testing.T) model {
	t.Helper()
	svc := service.New(storage.NewMemory(), clock.NewFake(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	ctx := context.Background()
	type seed struct {
		title string
		tags  []string
	}
	for _, s := range []seed{
		{"a-untagged", nil},
		{"b-travail", []string{"travail"}},
		{"c-perso", []string{"perso"}},
		{"d-travail-2", []string{"travail"}},
	} {
		tk, err := svc.Add(ctx, s.title)
		if err != nil {
			t.Fatalf("add: %v", err)
		}
		if s.tags != nil {
			if _, err := svc.SetTags(ctx, tk.ID, s.tags); err != nil {
				t.Fatalf("setTags: %v", err)
			}
		}
	}
	return newModel(svc)
}

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func step(m model, msg tea.Msg) model {
	out, _ := m.Update(msg)
	return out.(model)
}

// ---- navigation & basic actions (kept from v0.3) ----

func TestTUI_Navigation(t *testing.T) {
	m := testModel(t)
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}
	m = step(m, key("j"))
	if m.cursor != 1 {
		t.Errorf("after j cursor = %d", m.cursor)
	}
	m = step(m, key("k"))
	if m.cursor != 0 {
		t.Errorf("after k cursor = %d", m.cursor)
	}
}

func TestTUI_CycleStatus(t *testing.T) {
	m := testModel(t)
	firstID := m.tasks[0].ID
	m = step(m, key(" "))
	for _, tk := range m.tasks {
		if tk.ID == firstID && tk.Status != task.StatusInProgress {
			t.Errorf("status = %q, want in-progress", tk.Status)
		}
	}
}

func TestTUI_Quit(t *testing.T) {
	m := testModel(t)
	out, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Error("q should return a quit command")
	}
	if !out.(model).quitting {
		t.Error("q should set quitting")
	}
}

// ---- new TUI v2 flows ----

func TestTUI_Add(t *testing.T) {
	m := testModel(t)
	initial := len(m.tasks)
	m = step(m, key("a"))
	if m.mode != modeAdd {
		t.Fatalf("mode = %v, want modeAdd", m.mode)
	}
	m.input.SetValue("brand new task")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeNormal {
		t.Fatalf("mode = %v, want modeNormal", m.mode)
	}
	if len(m.tasks) != initial+1 {
		t.Errorf("count = %d, want %d", len(m.tasks), initial+1)
	}
	found := false
	for _, tk := range m.tasks {
		if tk.Title == "brand new task" {
			found = true
			if m.tasks[m.cursor].ID != tk.ID {
				t.Error("cursor should be on the new task")
			}
		}
	}
	if !found {
		t.Error("new task not in list")
	}
}

func TestTUI_AddEmptyIgnored(t *testing.T) {
	m := testModel(t)
	initial := len(m.tasks)
	m = step(m, key("a"))
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeNormal {
		t.Fatalf("mode = %v, want modeNormal", m.mode)
	}
	if len(m.tasks) != initial {
		t.Error("empty add should not create a task")
	}
}

func TestTUI_Edit(t *testing.T) {
	m := testModel(t)
	targetID := m.tasks[m.cursor].ID
	m = step(m, key("e"))
	if m.mode != modeEdit {
		t.Fatalf("mode = %v, want modeEdit", m.mode)
	}
	if m.input.Value() != m.tasks[0].Title {
		t.Errorf("input not prefilled, got %q", m.input.Value())
	}
	m.input.SetValue("renamed")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != modeNormal {
		t.Fatal("should return to normal")
	}
	for _, tk := range m.tasks {
		if tk.ID == targetID && tk.Title != "renamed" {
			t.Errorf("title = %q, want renamed", tk.Title)
		}
	}
}

func TestTUI_EditCancelWithEsc(t *testing.T) {
	m := testModel(t)
	original := m.tasks[m.cursor].Title
	m = step(m, key("e"))
	m.input.SetValue("should-be-ignored")
	m = step(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeNormal {
		t.Fatal("esc should exit input")
	}
	if m.tasks[0].Title != original {
		t.Errorf("title changed despite esc: %q", m.tasks[0].Title)
	}
}

func TestTUI_DeleteConfirm(t *testing.T) {
	m := testModel(t)
	targetID := m.tasks[m.cursor].ID
	m = step(m, key("x"))
	if m.mode != modeConfirmDelete {
		t.Fatalf("mode = %v, want modeConfirmDelete", m.mode)
	}
	m = step(m, key("y"))
	if m.mode != modeNormal {
		t.Fatal("should return to normal")
	}
	for _, tk := range m.tasks {
		if tk.ID == targetID {
			t.Error("task should have been deleted")
		}
	}
}

func TestTUI_DeleteCancel(t *testing.T) {
	m := testModel(t)
	targetID := m.tasks[m.cursor].ID
	m = step(m, key("x"))
	m = step(m, key("n"))
	if m.mode != modeNormal {
		t.Fatal("should return to normal")
	}
	found := false
	for _, tk := range m.tasks {
		if tk.ID == targetID {
			found = true
		}
	}
	if !found {
		t.Error("task should still exist after n")
	}
}

func TestTUI_StatusFilterCycle(t *testing.T) {
	m := testModel(t)
	if m.statusFilter != nil {
		t.Fatal("initial filter should be nil")
	}
	m = step(m, key("s"))
	if m.statusFilter == nil || *m.statusFilter != task.StatusTodo {
		t.Errorf("after 1×s, filter = %v, want todo", m.statusFilter)
	}
	m = step(m, key("s"))
	if *m.statusFilter != task.StatusInProgress {
		t.Errorf("after 2×s, filter = %v", *m.statusFilter)
	}
	// cycle 3 more times: blocked, done, nil
	m = step(m, key("s"))
	m = step(m, key("s"))
	m = step(m, key("s"))
	if m.statusFilter != nil {
		t.Errorf("after 5×s, filter = %v, want nil", m.statusFilter)
	}
}

func TestTUI_TagFilter(t *testing.T) {
	m := testModelTagged(t)
	initial := len(m.tasks)
	m = step(m, key("f"))
	if m.mode != modeFilterTag {
		t.Fatalf("mode = %v, want modeFilterTag", m.mode)
	}
	m.input.SetValue("travail")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.tagFilter != "travail" {
		t.Errorf("tagFilter = %q", m.tagFilter)
	}
	if len(m.tasks) >= initial {
		t.Errorf("expected filtered list, got %d (initial %d)", len(m.tasks), initial)
	}
	for _, tk := range m.tasks {
		found := false
		for _, tag := range tk.Tags {
			if tag == "travail" {
				found = true
			}
		}
		if !found {
			t.Errorf("task %q has no travail tag", tk.Title)
		}
	}
}

func TestTUI_ClearFilters(t *testing.T) {
	m := testModelTagged(t)
	m = step(m, key("s")) // status filter on
	m.tagFilter = "travail"
	m.reload()
	m = step(m, key("c"))
	if m.statusFilter != nil || m.tagFilter != "" {
		t.Errorf("after c, filters = (%v, %q)", m.statusFilter, m.tagFilter)
	}
}

func TestTUI_CyclePriority(t *testing.T) {
	m := testModel(t)
	id := m.tasks[m.cursor].ID
	if m.tasks[m.cursor].Priority != "" {
		t.Fatalf("initial priority should be unset, got %q", m.tasks[m.cursor].Priority)
	}
	// "" -> low
	m = step(m, key("p"))
	if got := taskByID(m.tasks, id).Priority; got != task.PriorityLow {
		t.Errorf("after 1×p, priority = %q, want low", got)
	}
	// low -> medium -> high -> "" back
	m = step(m, key("p"))
	m = step(m, key("p"))
	if got := taskByID(m.tasks, id).Priority; got != task.PriorityHigh {
		t.Errorf("after 3×p, priority = %q, want high", got)
	}
	m = step(m, key("p"))
	if got := taskByID(m.tasks, id).Priority; got != "" {
		t.Errorf("after 4×p, priority should clear, got %q", got)
	}
}

func TestTUI_CycleRecurrence(t *testing.T) {
	m := testModel(t)
	id := m.tasks[m.cursor].ID
	m = step(m, key("R"))
	if got := taskByID(m.tasks, id).Recur; got != task.RecurDaily {
		t.Errorf("after 1×R, recur = %q, want daily", got)
	}
	m = step(m, key("R"))
	m = step(m, key("R"))
	if got := taskByID(m.tasks, id).Recur; got != task.RecurMonthly {
		t.Errorf("after 3×R, recur = %q, want monthly", got)
	}
	m = step(m, key("R"))
	if got := taskByID(m.tasks, id).Recur; got != task.RecurNone {
		t.Errorf("after 4×R, recur should clear, got %q", got)
	}
}

func TestTUI_EditTags(t *testing.T) {
	m := testModel(t)
	id := m.tasks[m.cursor].ID
	m = step(m, key("t"))
	if m.mode != modeEditTags {
		t.Fatalf("mode = %v, want modeEditTags", m.mode)
	}
	m.input.SetValue("work,urgent")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	got := taskByID(m.tasks, id).Tags
	if len(got) != 2 || got[0] != "work" || got[1] != "urgent" {
		t.Errorf("tags = %v, want [work urgent]", got)
	}
}

func TestTUI_EditTagsClear(t *testing.T) {
	m := testModel(t)
	id := m.tasks[m.cursor].ID
	if _, err := m.svc.SetTags(m.ctx, id, []string{"existing"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m.reload()
	m = step(m, key("t"))
	m.input.SetValue("")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := taskByID(m.tasks, id).Tags; len(got) != 0 {
		t.Errorf("tags should clear, got %v", got)
	}
}

func TestTUI_EditDue(t *testing.T) {
	m := testModel(t)
	id := m.tasks[m.cursor].ID
	m = step(m, key("u"))
	if m.mode != modeEditDue {
		t.Fatalf("mode = %v, want modeEditDue", m.mode)
	}
	m.input.SetValue("2026-06-30")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	got := taskByID(m.tasks, id).DueAt
	if got == nil || got.Format(time.DateOnly) != "2026-06-30" {
		t.Errorf("due = %v, want 2026-06-30", got)
	}
}

func TestTUI_EditDueClear(t *testing.T) {
	m := testModel(t)
	id := m.tasks[m.cursor].ID
	due := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if _, err := m.svc.SetDueAt(m.ctx, id, &due); err != nil {
		t.Fatalf("seed: %v", err)
	}
	m.reload()
	m = step(m, key("u"))
	m.input.SetValue("")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := taskByID(m.tasks, id).DueAt; got != nil {
		t.Errorf("due should clear, got %v", got)
	}
}

func TestTUI_EditDueInvalid(t *testing.T) {
	m := testModel(t)
	m = step(m, key("u"))
	m.input.SetValue("not-a-date")
	m = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.err == nil {
		t.Error("expected error for invalid date")
	}
}

// taskByID is a small lookup used by the field-edit tests; returns the
// zero value when the ID is not in the list (the caller's assertion then
// surfaces the mistake clearly).
func taskByID(tasks []task.Task, id int) task.Task {
	for _, t := range tasks {
		if t.ID == id {
			return t
		}
	}
	return task.Task{}
}

func TestTUI_GroupedToggle(t *testing.T) {
	m := testModelTagged(t)
	if m.grouped {
		t.Fatal("should start ungrouped")
	}
	m = step(m, key("g"))
	if !m.grouped {
		t.Error("g should toggle on")
	}
	// After grouping, first-tag buckets are alphabetical with untagged last.
	// Expected order: perso, travail, travail, (untagged).
	if len(m.tasks) > 0 && (len(m.tasks[0].Tags) == 0 || m.tasks[0].Tags[0] != "perso") {
		t.Errorf("first task should be in perso group, got tags=%v", m.tasks[0].Tags)
	}
	if last := m.tasks[len(m.tasks)-1]; len(last.Tags) != 0 {
		t.Errorf("last task should be untagged, got tags=%v", last.Tags)
	}
	m = step(m, key("g"))
	if m.grouped {
		t.Error("g should toggle off")
	}
}
