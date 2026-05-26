package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/semos1204/komlist/internal/storage"
	"github.com/semos1204/komlist/internal/task"
)

func TestMemoryRepository_Contract(t *testing.T) {
	runRepositoryContract(t, func(_ *testing.T) storage.Repository { return storage.NewMemory() })
}

// runRepositoryContract is the shared behavioural contract every
// storage.Repository implementation must satisfy. New backends should call
// this helper from their own _test.go. The factory receives the subtest's
// *testing.T so it can scope per-test temp dirs and cleanups.
func runRepositoryContract(t *testing.T, factory func(t *testing.T) storage.Repository) {
	t.Helper()
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		r := factory(t)
		got, err := r.List(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty list, got %d", len(got))
		}
	})

	t.Run("create assigns id and persists", func(t *testing.T) {
		r := factory(t)
		created, err := r.Create(ctx, task.Task{Title: "first", Status: task.StatusTodo})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.ID == 0 {
			t.Fatal("expected non-zero id after create")
		}
		got, err := r.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Title != "first" {
			t.Errorf("got title %q, want %q", got.Title, "first")
		}
	})

	t.Run("get unknown returns ErrNotFound", func(t *testing.T) {
		r := factory(t)
		if _, err := r.Get(ctx, 999); !errors.Is(err, storage.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("update unknown returns ErrNotFound", func(t *testing.T) {
		r := factory(t)
		err := r.Update(ctx, task.Task{ID: 999, Title: "x"})
		if !errors.Is(err, storage.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("delete unknown returns ErrNotFound", func(t *testing.T) {
		r := factory(t)
		if err := r.Delete(ctx, 999); !errors.Is(err, storage.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("list is sorted by id", func(t *testing.T) {
		r := factory(t)
		for _, title := range []string{"a", "b", "c"} {
			if _, err := r.Create(ctx, task.Task{Title: title, Status: task.StatusTodo}); err != nil {
				t.Fatalf("create %q: %v", title, err)
			}
		}
		got, err := r.List(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for i := 1; i < len(got); i++ {
			if got[i-1].ID >= got[i].ID {
				t.Errorf("not sorted at %d: %v", i, got)
			}
		}
	})

	t.Run("delete removes", func(t *testing.T) {
		r := factory(t)
		created, _ := r.Create(ctx, task.Task{Title: "to delete", Status: task.StatusTodo})
		if err := r.Delete(ctx, created.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
		if _, err := r.Get(ctx, created.ID); !errors.Is(err, storage.ErrNotFound) {
			t.Errorf("after delete got %v, want ErrNotFound", err)
		}
	})
}
