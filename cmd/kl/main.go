// Command kl is the komlist command-line task manager. See the project
// README.md for usage and CONTRIBUTING.md for the architecture.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/semos1204/komlist/internal/cli"
	"github.com/semos1204/komlist/internal/clock"
	"github.com/semos1204/komlist/internal/i18n"
	"github.com/semos1204/komlist/internal/service"
	"github.com/semos1204/komlist/internal/storage"
)

// Build-time identifiers populated by GoReleaser via -ldflags -X.
// Defaults are used for `go run` / `go build` outside of a release.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	i18n.Configure(os.Getenv("KOMLIST_LANG"))
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, i18n.T(i18n.KeyErrPrefix), localizeError(err))
		os.Exit(1)
	}
}

// localizeError maps known sentinel errors to their localized message,
// falling back to the raw error text (e.g. for wrapped validation errors
// that embed a dynamic list of valid values).
func localizeError(err error) string {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return i18n.T(i18n.KeyErrNotFound)
	case errors.Is(err, service.ErrEmptyTitle):
		return i18n.T(i18n.KeyErrEmptyTitle)
	case errors.Is(err, service.ErrEmptyTag):
		return i18n.T(i18n.KeyErrEmptyTag)
	case errors.Is(err, service.ErrEmptyNote):
		return i18n.T(i18n.KeyErrEmptyNote)
	case errors.Is(err, service.ErrSelfDependency):
		return i18n.T(i18n.KeyErrSelfDependency)
	case errors.Is(err, service.ErrDependencyCycle):
		return i18n.T(i18n.KeyErrDependencyCycle)
	default:
		return err.Error()
	}
}

func run() error {
	path, err := storage.DefaultPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)

	repo, closer, err := openRepository(dir, path)
	if err != nil {
		return err
	}
	defer func() { _ = closer.Close() }()

	svc := service.New(repo, clock.System{})
	root := cli.NewRootCommand(svc)
	root.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	return root.ExecuteContext(context.Background())
}

// openRepository selects the storage backend. KOMLIST_BACKEND=sqlite uses a
// SQLite database; otherwise the JSON file is used, wrapped in the git
// decorator when ~/.komlist is a git work tree.
func openRepository(dir, jsonPath string) (storage.Repository, io.Closer, error) {
	if os.Getenv("KOMLIST_BACKEND") == "sqlite" {
		sq, err := storage.NewSQLite(filepath.Join(dir, "tasks.db"))
		if err != nil {
			return nil, nil, err
		}
		return sq, sq, nil
	}
	jsonRepo, err := storage.NewJSON(jsonPath)
	if err != nil {
		return nil, nil, err
	}
	var repo storage.Repository = jsonRepo
	if storage.IsGitDir(dir) {
		repo = storage.NewGit(jsonRepo, dir)
	}
	return repo, jsonRepo, nil
}
