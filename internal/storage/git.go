package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/semos1204/komlist/internal/task"
)

// GitRepository decorates another Repository, committing the underlying data
// file to a git repository after every mutation. It is a thin adapter: the
// inner repository owns persistence; this layer only versions it.
//
// Commits are best-effort. If git is missing or a commit fails, a warning is
// written to stderr and the task operation still succeeds — versioning must
// never block task management.
type GitRepository struct {
	inner Repository
	dir   string
}

// NewGit wraps inner so that mutations are committed in the git working tree
// rooted at dir.
func NewGit(inner Repository, dir string) *GitRepository {
	return &GitRepository{inner: inner, dir: dir}
}

// IsGitDir reports whether dir contains a .git entry (i.e. is a git work tree).
func IsGitDir(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

// List implements Repository.
func (g *GitRepository) List(ctx context.Context) ([]task.Task, error) {
	return g.inner.List(ctx)
}

// Get implements Repository.
func (g *GitRepository) Get(ctx context.Context, id int) (task.Task, error) {
	return g.inner.Get(ctx, id)
}

// Create implements Repository and commits the result.
func (g *GitRepository) Create(ctx context.Context, t task.Task) (task.Task, error) {
	created, err := g.inner.Create(ctx, t)
	if err != nil {
		return created, err
	}
	g.commit(fmt.Sprintf("create #%d %s", created.ID, created.Title))
	return created, nil
}

// Update implements Repository and commits the change.
func (g *GitRepository) Update(ctx context.Context, t task.Task) error {
	if err := g.inner.Update(ctx, t); err != nil {
		return err
	}
	g.commit(fmt.Sprintf("update #%d %s", t.ID, t.Title))
	return nil
}

// Delete implements Repository and commits the removal.
func (g *GitRepository) Delete(ctx context.Context, id int) error {
	if err := g.inner.Delete(ctx, id); err != nil {
		return err
	}
	g.commit(fmt.Sprintf("delete #%d", id))
	return nil
}

// Close releases the inner repository if it is a Closer.
func (g *GitRepository) Close() error {
	if c, ok := g.inner.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (g *GitRepository) commit(msg string) {
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, "kl: git not found, skipping commit")
		return
	}
	if out, err := g.run("add", "-A"); err != nil {
		fmt.Fprintf(os.Stderr, "kl: git add failed: %v: %s\n", err, out)
		return
	}
	out, err := g.run("commit", "-m", "kl: "+msg)
	if err != nil {
		// A "nothing to commit" exit is benign (e.g. no net change); stay quiet
		// only for the noisy cases. Surface anything unexpected.
		fmt.Fprintf(os.Stderr, "kl: git commit skipped: %s\n", out)
	}
}

func (g *GitRepository) run(args ...string) (string, error) {
	full := append([]string{"-C", g.dir}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	return string(out), err
}
