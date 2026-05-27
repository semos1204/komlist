package storage_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/semos1204/komlist/internal/storage"
	"github.com/semos1204/komlist/internal/task"
)

func TestSQLiteRepository_Contract(t *testing.T) {
	runRepositoryContract(t, func(t *testing.T) storage.Repository {
		r, err := storage.NewSQLite(filepath.Join(t.TempDir(), "tasks.db"))
		if err != nil {
			t.Fatalf("NewSQLite: %v", err)
		}
		t.Cleanup(func() { _ = r.Close() })
		return r
	})
}

func TestSQLiteRepository_ReloadPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.db")
	ctx := context.Background()

	r1, err := storage.NewSQLite(path)
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	created, err := r1.Create(ctx, task.Task{
		Title:    "persisted",
		Status:   task.StatusTodo,
		Tags:     []string{"work", "urgent"},
		Notes:    []string{"a note"},
		Priority: task.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := r1.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	r2, err := storage.NewSQLite(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = r2.Close() })

	got, err := r2.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if got.Title != "persisted" || got.Priority != task.PriorityHigh {
		t.Errorf("scalar fields not persisted: %+v", got)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "work" {
		t.Errorf("tags not persisted: %v", got.Tags)
	}
	if len(got.Notes) != 1 || got.Notes[0] != "a note" {
		t.Errorf("notes not persisted: %v", got.Notes)
	}
}
