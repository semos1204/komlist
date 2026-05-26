package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semos1204/komlist/internal/storage"
	"github.com/semos1204/komlist/internal/task"
)

func TestJSONRepository_Contract(t *testing.T) {
	runRepositoryContract(t, func(t *testing.T) storage.Repository {
		dir := t.TempDir()
		r, err := storage.NewJSON(filepath.Join(dir, "tasks.json"))
		if err != nil {
			t.Fatalf("NewJSON: %v", err)
		}
		t.Cleanup(func() { _ = r.Close() })
		return r
	})
}

func TestJSONRepository_ReloadPersists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	ctx := context.Background()

	r1, err := storage.NewJSON(path)
	if err != nil {
		t.Fatalf("NewJSON: %v", err)
	}
	created, err := r1.Create(ctx, task.Task{Title: "persisted", Status: task.StatusTodo})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := r1.Close(); err != nil {
		t.Fatalf("close r1: %v", err)
	}

	r2, err := storage.NewJSON(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = r2.Close() })

	got, err := r2.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("get after reopen: %v", err)
	}
	if got.Title != "persisted" {
		t.Errorf("got title %q, want %q", got.Title, "persisted")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp") {
			t.Errorf("found leftover tmp file: %s", e.Name())
		}
	}
}

func TestJSONRepository_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "tasks.json")
	r, err := storage.NewJSON(path)
	if err != nil {
		t.Fatalf("NewJSON: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Errorf("parent dir not created: %v", err)
	}
}

func TestJSONRepository_CloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	r, err := storage.NewJSON(filepath.Join(dir, "tasks.json"))
	if err != nil {
		t.Fatalf("NewJSON: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}
