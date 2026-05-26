// Command kl is the komlist command-line task manager. See the project
// README.md for usage and CONTRIBUTING.md for the architecture.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/semos1204/komlist/internal/cli"
	"github.com/semos1204/komlist/internal/clock"
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
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	path, err := storage.DefaultPath()
	if err != nil {
		return err
	}
	jsonRepo, err := storage.NewJSON(path)
	if err != nil {
		return err
	}
	defer func() { _ = jsonRepo.Close() }()

	// When ~/.komlist is a git work tree, version every mutation.
	var repo storage.Repository = jsonRepo
	if dir := filepath.Dir(path); storage.IsGitDir(dir) {
		repo = storage.NewGit(jsonRepo, dir)
	}

	svc := service.New(repo, clock.System{})
	root := cli.NewRootCommand(svc)
	root.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, date)
	return root.ExecuteContext(context.Background())
}
