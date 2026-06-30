# Quickstart & Validation: Interactive Task Picker (TUI)

This guide proves the feature works end-to-end and that non-interactive behavior
is unchanged. Details live in [contracts/tui-picker.md](./contracts/tui-picker.md)
and [data-model.md](./data-model.md); this file is the run/validation checklist.

## Prerequisites

- Repo checked out on branch `007-interactive-tui`.
- Docker available (tests run **only** inside the Docker harness per project
  policy — never on the host).
- A `Runefile` with a few non-private tasks (the repo-root `Runefile` works).

## Build

```sh
go build ./cmd/rune            # local sanity build (host)
# CI/portability gate (static, all OSes) is run by the `build` workflow gate
```

## Manual validation (interactive — run in a real terminal)

| # | Steps | Expected (maps to) |
|---|-------|--------------------|
| 1 | `rune --choose` | Full-screen styled picker lists non-private tasks; one highlighted (US1, FR-001/002) |
| 2 | Type part of a task name | List narrows over name **and** description; match highlighted (FR-003, Q2) |
| 3 | Highlight a task | Its documentation shows in the detail pane (FR-004) |
| 4 | Press `Enter` | Picker disappears; task runs with native output; Rune exits with the task's code (US2, FR-006/008/010) |
| 5 | `rune --choose` then `q` / `Ctrl-C` | Nothing runs; clean terminal; exit 0 (FR-005/016) |
| 6 | `rune --choose -- --watch` → select a task | Task runs with `--watch` forwarded (FR-006, Q3) |
| 7 | Shrink the terminal, `rune --choose` | Detail pane collapses; list still usable; no corruption (FR-017) |
| 8 | `NO_COLOR=1 rune --choose` | Picker renders without color, still usable (FR-015) |

## Automated validation (Docker)

```sh
# Unit: pure model Update/state-transition tests
docker-compose run --rm test go test ./internal/tui/...

# Wiring + guards (opt-in, non-TTY error, empty list, arg pass-through)
docker-compose run --rm test go test ./internal/cli/...

# Full suite + race (protects US3 / unchanged non-interactive paths)
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

### Non-interactive guardrails (US3 / FR-014) — must hold

| Command | Expected |
|---------|----------|
| `rune --choose \| cat` (non-TTY stdout) | `--choose requires an interactive terminal`; exit 2; no UI bytes |
| `rune --list` | Identical to pre-feature baseline (golden) |
| `rune <task>` | Identical to pre-feature baseline; picker never opens |
| `rune --dry-run` / `--dump` / `--summary` | Unchanged; picker never opens |

## Gate checklist (must pass to merge)

- [ ] `lint` — `golangci-lint run` clean (incl. new `internal/tui` package)
- [ ] `test` — full suite + `-race`, all three OSes
- [ ] `build` — static `CGO_ENABLED=0` binary builds on Linux/macOS/Windows
- [ ] `golden` — committed golden files match (no drift in non-interactive output)
- [ ] `docs-verify` — `--choose` docs updated and validated
- [ ] `release-dryrun` — `goreleaser release --snapshot` succeeds (binary size OK)
```
