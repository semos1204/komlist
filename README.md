# komlist

A small command-line task manager written in Go.

[![CI](https://github.com/semos1204/komlist/actions/workflows/ci.yml/badge.svg)](https://github.com/semos1204/komlist/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`komlist` tracks lightweight tasks with four statuses — `todo`,
`in-progress`, `blocked`, `done` — stored as JSON in `~/.komlist/tasks.json`.
No daemon, no database, no network.

## Install

### Homebrew (recommended, no Go required)

```bash
brew install semos1204/tap/kl
```

This pulls the latest signed binary from the GitHub release. Your users
never need a Go toolchain.

### Direct download (any platform, no Go required)

Grab the archive for your OS/arch from the
[latest release](https://github.com/semos1204/komlist/releases/latest),
extract `kl`, and put it on your `PATH`. Example for macOS arm64:

```bash
curl -L -o kl.tar.gz \
  https://github.com/semos1204/komlist/releases/latest/download/kl_<version>_darwin_arm64.tar.gz
tar -xzf kl.tar.gz
sudo mv kl /usr/local/bin/
```

### Go install (for developers)

```bash
go install github.com/semos1204/komlist/cmd/kl@latest
```

### From source

```bash
git clone https://github.com/semos1204/komlist
cd komlist
make build
./bin/kl --help
```

After install, you're done. The first `kl add` auto-creates
`~/.komlist/tasks.json` — no init step, no config file.

## Usage

### Core commands

```console
$ kl add "Write the README"
Created: #1 [todo] Write the README

$ kl list
ID  STATUS  TITLE             UPDATED
1   todo    Write the README  2026-05-25 23:06

$ kl status 1 in-progress
Updated: #1 [in-progress] Write the README

$ kl edit 1 "Polish the README"
Renamed: #1 [in-progress] Polish the README

$ kl delete 1
Deleted: #1
```

### Priority, tags, due date

```console
$ kl prio 2 high
Priority: #2 [high]

$ kl tag 1 doc,onboarding
Tagged: #1 [doc,onboarding]

$ kl due 2 2026-07-01
Due: #2 2026-07-01

$ kl due 2 none
Due: #2 (cleared)
```

The `list` command shows additional columns (`PRIO`, `TAGS`, `DUE`) only
when at least one task uses the feature, so the default output stays
compact:

```console
$ kl list
ID  STATUS  PRIO  TAGS            DUE         TITLE             UPDATED
1   todo    -     doc,onboarding  -           Write the README  2026-05-25 23:06
2   todo    high  -               2026-07-01  Ship V2           2026-05-25 23:06
```

### Filtering and sorting

```console
$ kl list --status todo
$ kl list --tag doc
$ kl list --wide            # always show PRIO / TAGS / DUE columns
$ kl list --sort priority   # high → medium → low → unset
$ kl list --sort due        # earliest first, no-due last
$ kl list --sort status
$ kl list --sort urgency    # computed score: priority + due proximity + age
```

The **urgency** score is a Taskwarrior-inspired heuristic: high priority,
soon/overdue due dates and in-progress status push a task up; done and
blocked tasks sink.

### Notes

```console
$ kl note 1 "purge inactive accounts first"
Note added: #1 (1 total)

$ kl show 1            # full detail view, including notes
$ kl note 1 --clear   # drop all notes
```

### Recurrence

```console
$ kl recur 1 weekly         # daily | weekly | monthly | none
$ kl recur 1 2w             # or an interval: 3d, 2w, 1mo, …
```

When a recurring task is marked `done`, komlist spawns a fresh `todo` copy
with its due date advanced by one cadence (from the old due date, or from
now if it had none).

### Dependencies

A task can depend on others; while any dependency is not `done` it is
**blocked** — marked `🔒` on the board and sunk to the bottom of the urgency
order. Cycles and self-dependencies are rejected.

```console
$ kl block 3 4       # task #3 now depends on #4
$ kl unblock 3 4     # remove the dependency
```

### Interactive UI

```console
$ kl ui
```

A Bubble Tea terminal app over your tasks: `j`/`k` to move, `space` to cycle
the status (todo → in-progress → done), `d` to mark done, `r` to reload,
`q` to quit.

### Board view

`kl board` is a colored, taskbook-style view grouping tasks by tag, each
group ordered by urgency, with a completion footer. `kl list` stays the
plain, scriptable table.

```console
$ kl board               # all tasks, grouped by tag
$ kl board travail       # only the "travail" board
$ kl board --status todo # only pending items
```

```
 perso
  4. ☐ acheter du pain ⚑ 2026-06-02 ⟳weekly
  3. ✔ acheter du pain ⟳weekly

 travail
  1. ☐ clean bdd 30j ·high ⚑ 2026-06-30
  2. ▶ migrate API

 1 done · 1 doing · 2 todo — 25% complete
```

Colors are disabled automatically when output is piped or `NO_COLOR` is set.

### Errors

Errors are reported on stderr with a non-zero exit code:

```console
$ kl status 999 done
Error: task not found

$ kl status 2 nope
Error: invalid status "nope" (valid: [todo in-progress blocked done])

$ kl add ""
Error: title must not be empty
```

Run `kl --help` for the full command reference, or `kl --version` for the
build identifier.

### Shell completion

Cobra ships completion for the major shells. Enable it once per shell:

```bash
# bash
kl completion bash > /usr/local/etc/bash_completion.d/kl

# zsh (with completions enabled)
kl completion zsh > "${fpath[1]}/_kl"

# fish
kl completion fish > ~/.config/fish/completions/kl.fish
```

### Storage

Tasks are stored as JSON in `~/.komlist/tasks.json`. Writes are atomic
(write-to-tmp then rename), so the file is never left half-written. An
exclusive interprocess lock (sidecar `tasks.json.lock`, via `flock(2)`)
serialises concurrent `kl` invocations — running two shells against the
same store is safe.

### Git-backed history (optional)

Turn `~/.komlist` into a git repository and komlist will commit every
mutation automatically — a full, diffable history of your tasks:

```bash
git -C ~/.komlist init
kl add "now versioned"
git -C ~/.komlist log --oneline   # kl: create #1 now versioned
```

Detection is automatic (presence of `~/.komlist/.git`). Commits are
best-effort: if git is missing or fails, the task operation still
succeeds. This is a second `storage.Repository` implementation decorating
the JSON one — the service and CLI are unchanged, illustrating the
hexagonal design.

### SQLite backend (optional)

Set `KOMLIST_BACKEND=sqlite` to store tasks in `~/.komlist/tasks.db` instead
of JSON, using a pure-Go SQLite driver (no cgo):

```bash
export KOMLIST_BACKEND=sqlite
kl add "stored in sqlite"
```

This is a third `storage.Repository` implementation — again, no change to the
service or CLI.

## Language

komlist's runtime output is localized. Set `KOMLIST_LANG=fr` for French
(English is the default; locale forms like `fr_FR.UTF-8` are accepted):

```console
$ KOMLIST_LANG=fr kl add "acheter du pain"
Créée : #1 [todo] acheter du pain
```

Cobra's structural help words (`Usage:`, `Flags:`, …) remain English; only
komlist's own messages, table headers and errors are translated.

## Architecture

Lightweight hexagonal layout. A `service` package containing the use cases
depends only on the `task` domain and on two ports — `storage.Repository`
and `clock.Clock`. The CLI (`internal/cli`) wires Cobra around the service;
`main.go` is the only place where concrete implementations are instantiated.

```
internal/cli       (Cobra adapter)
       │
       ▼
internal/service   (use cases)
       │
       ├──► internal/task     (entities)
       ├──► internal/storage  (Repository port + JSON / memory)
       └──► internal/clock    (Clock port + system / fake)
```

Swapping the JSON file for SQLite, BoltDB or a remote store only requires a
new `Repository` implementation; the service and CLI stay unchanged.

## Development

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the dev workflow, project layout
and how to add a storage backend.

## Testing

`make test` runs unit tests with the race detector on every package,
including end-to-end CLI tests that drive `cli.NewRootCommand` with an
in-memory repository and a fake clock.

## Releases

Tagging a commit `vX.Y.Z` and pushing the tag triggers GoReleaser
(`.github/workflows/release.yml`) which:

- builds multi-arch binaries (linux / darwin / windows × amd64 / arm64),
- publishes them to a GitHub release with checksums and a changelog,
- updates the Homebrew formula in your tap repo so `brew install <tap>/kl`
  picks up the new version (see one-time setup below).

The build embeds `version`, `commit` and `date` so `kl --version`
identifies any binary unambiguously.

### One-time Homebrew tap setup

To enable `brew install semos1204/tap/kl`:

1. Create an empty GitHub repo **`homebrew-tap`** under your account or
   org. (The `homebrew-` prefix matters — `brew tap` requires it.)
2. Generate a [Personal Access Token](https://github.com/settings/tokens)
   with **`repo`** scope.
3. On the komlist repo, add it as a secret named
   **`HOMEBREW_TAP_GITHUB_TOKEN`** (Settings → Secrets and variables →
   Actions → New repository secret).
4. In `.goreleaser.yml`, replace `semos1204` with your handle in
   both `brews[].repository.owner` and `brews[].homepage`.

That's it. Every `git tag vX.Y.Z && git push --tags` will now bump the
formula automatically. Without the secret, releases still publish raw
binaries — the Homebrew step is skipped.

## Known limitations

- **English-only messages.** All CLI output (errors, info) is in English.
  This is a deliberate choice for open-source reach; i18n is on the
  [roadmap](ROADMAP.md) if there's demand.

## Roadmap

See [`ROADMAP.md`](ROADMAP.md) for remaining items: optional i18n.

## License

MIT — see [`LICENSE`](LICENSE).
