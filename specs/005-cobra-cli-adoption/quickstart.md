# Quickstart & Validation: Modern CLI Interface

**Feature**: `005-cobra-cli-adoption` · **Date**: 2026-06-10

Runnable scenarios that prove the feature works end-to-end. Details live in
[contracts/cli.md](./contracts/cli.md) and [data-model.md](./data-model.md); this is a
validation guide, not implementation.

## Prerequisites

- Go 1.25 toolchain (for `go build`/`go run` manual checks).
- Docker + standalone `docker-compose` (the project runs **all tests in Docker**, never
  on the host — Constitution VI / project policy).
- A Runefile in the working directory (the repo root `Runefile` works: it defines
  `fmt`, `lint`, `test`, `build`, …).

## Build

```sh
go build -o /tmp/rune ./rune        # or: rune build  (dogfooded task)
```

## Automated tests (authoritative — run in Docker)

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

**Definition of done**: the full suite is green with `-race` (SC-006: no behavioral
regression), and the new unit + integration tests below pass.

## Manual validation scenarios

### S1 — Commands are discoverable with help (US1 / FR-001..FR-005, SC-001/SC-002/SC-004)

```sh
/tmp/rune --help          # lists serve, version, completion, help; points to --list for tasks
/tmp/rune serve --help    # usage, every flag, ≥1 example; shows "Aliases: serve, mcp"
/tmp/rune version         # same string as:
/tmp/rune --version
```

Expected: all built-ins listed; each `--help` shows flags + an example; `version` and
`--version` agree.

### S2 — Dynamic task-name completion with descriptions (US2 / FR-009..FR-012, SC-003)

```sh
# Drive Cobra's dynamic-completion protocol directly (what the shell calls):
/tmp/rune __complete ""           # built-in commands + Runefile task names (with descriptions)
/tmp/rune __complete "te"         # narrows to e.g. test / test-race
/tmp/rune completion zsh | head   # a valid zsh script; `completion --help` shows install steps

# In a directory with no Runefile (or a broken one):
cd /tmp && /tmp/rune __complete ""   # built-in commands only, NO error line, exit 0
```

Expected: task names appear with their doc summaries; the trailing `:<directive>` line is
`ShellCompDirectiveNoFileComp`; graceful degradation when no/invalid Runefile.

### S3 — Built-in precedence and the `--` escape hatch (FR-006/FR-008)

```sh
# Given a Runefile that defines a task literally named `serve`:
/tmp/rune serve            # runs the SERVER (built-in wins)
/tmp/rune -- serve         # runs the TASK named serve
/tmp/rune build --watch    # `--watch` is passed to the build task untouched
```

### S4 — serve flags & validation (US3 / FR-014/FR-015)

```sh
/tmp/rune serve --http --addr :7777 --token-file ./tok   # HTTP server
/tmp/rune mcp                                            # stdio (alias)
/tmp/rune serve --addr :7777 ; echo $?                   # 2 (—addr without —http: usage error)
```

### S5 — Did-you-mean & concise errors (US3 / FR-016/FR-017)

```sh
/tmp/rune serv ; echo $?       # rune: unknown task: serv (did you mean "serve"?) ; exit 2
/tmp/rune tset ; echo $?       # suggests nearest task (e.g. "test") ; exit 2, no usage dump
```

### S6 — Contract preservation (SC-006 / FR-018..FR-022)

```sh
/tmp/rune --list               # task listing on stdout
/tmp/rune ; echo $?            # no task + no default → usage error, exit 2
/tmp/rune <task> | cat         # stdout pipe-safe; diagnostics on stderr only
```

## Test coverage to add (TDD — write failing first)

- **Unit (`internal/cli`)**: `TaskCandidates` (filters private/OS, graceful nil on
  missing/broken Runefile); `nearest`/Levenshtein suggestion table tests; serve
  conditional-requires validation.
- **Integration (compiled binary)**: S1 help discoverability; S2 `__complete` output +
  graceful degradation; S3 precedence + `--` escape; S4 serve flags + exit 2; S5
  did-you-mean + exit 2; S6 listing/pipe/exit codes. Each asserts **stdout, stderr, and
  exit code**.
- **Cross-platform**: CI runs the suite on Linux, macOS, Windows with `-race`.
