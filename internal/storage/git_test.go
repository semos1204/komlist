package storage_test

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semos1204/komlist/internal/storage"
	"github.com/semos1204/komlist/internal/task"
)

func TestGitRepository_CommitsOnCreate(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "test")

	if !storage.IsGitDir(dir) {
		t.Fatal("IsGitDir should be true after git init")
	}

	json, err := storage.NewJSON(filepath.Join(dir, "tasks.json"))
	if err != nil {
		t.Fatalf("NewJSON: %v", err)
	}
	t.Cleanup(func() { _ = json.Close() })

	repo := storage.NewGit(json, dir)
	if _, err := repo.Create(context.Background(), task.Task{Title: "versioned", Status: task.StatusTodo}); err != nil {
		t.Fatalf("create: %v", err)
	}

	out := gitRun(t, dir, "log", "--oneline")
	if strings.TrimSpace(out) == "" {
		t.Error("expected a commit after Create, git log is empty")
	}
	if !strings.Contains(out, "create #1") {
		t.Errorf("commit message missing, log = %q", out)
	}
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}
