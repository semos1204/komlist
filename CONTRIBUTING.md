# Contributing to komlist

Thanks for considering a contribution!

## Local development

```bash
git clone https://github.com/semos1204/komlist
cd komlist
make build      # produces ./bin/kl
make test       # go test -race ./...
make lint       # golangci-lint run (install: brew install golangci-lint)
```

## Project layout

```
cmd/kl/main.go    — binary entrypoint (installed as `kl`)
internal/task     — domain (Task, Status)
internal/clock    — Clock port + system & fake implementations
internal/storage  — Repository port + JSON & in-memory implementations
internal/service  — use cases (TaskService)
internal/cli      — Cobra adapters wiring the service
```

Dependencies flow inward: `cli` → `service` → (`task`, `storage` interface,
`clock` interface). The service never imports a concrete repository — only
the interface.

## Adding a new storage backend

1. Implement `storage.Repository` in a new file or sub-package.
2. Reuse the contract suite by following the pattern in
   `internal/storage/memory_test.go` (a `runRepositoryContract` helper that
   takes a factory).
3. Wire the new backend behind a flag in `main.go` if you want it as a runtime
   choice — the `service` layer requires no change.

## Commit style

We recommend [Conventional Commits](https://www.conventionalcommits.org)
(`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`) but it is not
enforced.

## Before opening a PR

```bash
make test
make lint
```

CI runs the same checks on `ubuntu-latest` and `macos-latest`.
