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
- ✅ **Interval recurrence** — `kl recur 1 2w` (also `3d`, `1mo`) alongside the
  keyword cadences.
- ✅ **SQLite backend** — `KOMLIST_BACKEND=sqlite` stores in `~/.komlist/tasks.db`
  via `modernc.org/sqlite` (pure Go). Third `Repository` implementation.
- ✅ **Task dependencies** — `kl block`/`kl unblock`, cycle + self-dependency
  rejection, blocked tasks sink in urgency and show `🔒` on the board.
- ✅ **i18n (en/fr)** — runtime output, headers and errors localized via
  `internal/i18n` + `KOMLIST_LANG`.
- ✅ **Interactive TUI** — `kl ui`, a Bubble Tea front-end sharing the board's
  rendering (`internal/render`).

## Possible next steps

- **More languages** — the `internal/i18n` catalog is ready to accept more
  than fr/en.
- **Localized Cobra help** — override Cobra templates so `Usage:`/`Flags:`
  and command descriptions also translate.
- **Richer TUI** — grouping by tag, in-app editing/adding, tag/status filters.
- **Cron-like recurrence** — "1st of each month" rules beyond fixed intervals.
