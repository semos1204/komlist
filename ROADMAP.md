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
- ✅ **Notes / annotations** — `Task.Notes`, `kl note <id> <text>` (append)
  and `--clear`, surfaced by `kl show <id>`.
- ✅ **Urgency score** — Taskwarrior-inspired heuristic (priority + due
  proximity + status + age) in `internal/service/urgency.go`, exposed via
  `kl list --sort urgency` and used to order the board.
- ✅ **Recurrence** — `task.Recurrence` (daily/weekly/monthly), `kl recur`,
  completing a recurring task spawns the next occurrence with a shifted due.
- ✅ **Git-backed history** — `storage.GitRepository` decorator commits every
  mutation when `~/.komlist` is a git work tree (inspired by dstask). Pure
  win of the `Repository` port: service/CLI untouched.
- ✅ **Board view** — `kl board`, taskbook-style grouped/colored output via
  `charmbracelet/lipgloss`, ordered by urgency, with a completion footer.

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
- **Task dependencies** — Taskwarrior-style `depends`, surfaced in urgency
  (a blocked-by-incomplete task sinks) and the board.
- **Interval recurrence** — accept `2w` / `3d` in addition to the keyword
  cadences.
- **Interactive TUI** — a Bubble Tea front-end over the board view.
