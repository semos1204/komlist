# Roadmap

## Shipped

- ✅ **CLI end-to-end tests** — `internal/cli/cli_test.go` covers every
  sub-command with shared in-memory repo + fake clock.
- ✅ **Interprocess lock** — `flock(2)` via `github.com/gofrs/flock` on a
  sidecar `tasks.json.lock`, acquired in `storage.NewJSON`, released by
  `Close()` from `main.go`.
- ✅ **GoReleaser** — `.goreleaser.yml` + `.github/workflows/release.yml`
  build linux/darwin/windows × amd64/arm64 on `v*` tags. `kl --version`
  reflects the embedded `version`/`commit`/`date`.
- ✅ **Homebrew tap publishing** — `brews:` block in `.goreleaser.yml`
  pushes the formula to `semos1204/homebrew-tap` on every
  release (skipped gracefully if the `HOMEBREW_TAP_GITHUB_TOKEN` secret
  is absent). See the README *Releases* section for the one-time tap
  setup.
- ✅ **Edit a task's title** — `TaskService.Rename` + `kl edit <id> <title>`.
- ✅ **Tags** — `Task.Tags []string` with dedupe/trim, `kl tag <id> a,b`,
  `kl list --tag a` filter.
- ✅ **Due dates** — `Task.DueAt *time.Time`, `kl due <id> YYYY-MM-DD`
  (or `none` to clear), `kl list --sort due` (nil last).
- ✅ **Priorities** — `task.Priority` (low/medium/high), `kl prio <id> high`,
  `kl list --sort priority` (high first).
- ✅ **Adaptive list output** — optional columns appear only when at least
  one task uses them, so the default `kl list` stays compact.

## Open

- **i18n of CLI messages** — currently English-only by convention. If user
  demand justifies it, wire `golang.org/x/text/message` behind a
  `KOMLIST_LANG` env var (or a `--lang` global flag). Until then, error /
  info strings live inline.

## Possible next steps

- **More storage backends** — drop a new `Repository` implementation
  alongside `JSONRepository` (SQLite via `modernc.org/sqlite` for a pure-Go
  build, or BoltDB for embedded). The service and CLI stay unchanged; pick
  the backend in `cmd/kl/main.go` via a flag or env var.
- **Recurrence / snoozing** — derived from `DueAt`. Out of scope today.
