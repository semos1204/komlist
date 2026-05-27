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

func key(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func step(m model, msg tea.Msg) model {
	out, _ := m.Update(msg)
	return out.(model)
}

func TestTUI_Navigation(t *testing.T) {
	m := testModel(t)
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}
	m = step(m, key("j"))
	if m.cursor != 1 {
		t.Errorf("after j cursor = %d, want 1", m.cursor)
	}
	m = step(m, key("k"))
	if m.cursor != 0 {
		t.Errorf("after k cursor = %d, want 0", m.cursor)
	}
	m = step(m, key("k")) // at top, stays
	if m.cursor != 0 {
		t.Errorf("k at top cursor = %d, want 0", m.cursor)
	}
}

func TestTUI_CycleStatus(t *testing.T) {
	m := testModel(t)
	firstID := m.tasks[0].ID
	if m.tasks[0].Status != task.StatusTodo {
		t.Fatalf("want initial todo, got %q", m.tasks[0].Status)
	}
	m = step(m, key(" "))

	var found bool
	for _, tk := range m.tasks {
		if tk.ID == firstID {
			found = true
			if tk.Status != task.StatusInProgress {
				t.Errorf("status = %q, want in-progress", tk.Status)
			}
		}
	}
	if !found {
		t.Error("task disappeared after cycle")
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
