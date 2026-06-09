# Contract: CLI Surface

**Feature**: 001-rune-task-runner | **Date**: 2026-06-08

The command line is Rune's primary human interface. Task names are **dynamic** (whatever the
Runefile defines); global flags are fixed. Invocation shape:

```
rune [GLOBAL FLAGS] [VAR=VALUE ...] [TASK [TASK ARGS...]] [TASK2 ...]
```

- Running with no task → the configured default task (`set default`), else an error listing
  available tasks.
- `VAR=VALUE` before/among tasks overrides Runefile variables (FR-006).
- Multiple tasks may be chained in one invocation; each runs once (memoized, FR-005).

## Global flags

| Flag | Behavior | Req |
|------|----------|-----|
| `-f, --file PATH` | Use a specific Runefile instead of upward discovery | FR-001 |
| `--list` | List non-private tasks with docs + groups; run nothing | FR-022 |
| `--dry-run` | Print the resolved execution plan; run nothing | FR-022 |
| `--summary` | Print task names that would run, one per line | FR-022 |
| `--dump [--format json]` | Emit the parsed file (canonical text, or JSON) | FR-023 |
| `--fmt` | Rewrite the Runefile in canonical formatting | (Phase 1 tooling) |
| `--set NAME VALUE` | Override a variable (alt to `NAME=VALUE`) | FR-006 |
| `--watch` | Re-run on file changes | FR (Phase 6) |
| `--choose` | Interactive task picker | FR (Phase 8) |
| `--yes` | Auto-approve `[confirm]` tasks (CI) | FR-009 |
| `--quiet` | Suppress command echo | FR-010/FR-018 |
| `--version` | Print version and exit | FR-024 |
| `-h, --help` | Print help and exit | FR-024 |

## Subcommands (reserved, non-task)

| Command | Behavior | Req |
|---------|----------|-----|
| `rune mcp` / `rune serve --mcp` | Start the MCP server (stdio default) | FR-025 |
| `rune serve --mcp --http [--addr 127.0.0.1:PORT] --token-file PATH` | Opt-in remote endpoint, localhost-bound, token-required | FR-031 |

Reserved names (`mcp`, `serve`) MUST NOT shadow tasks silently; a task with a reserved name is
reachable via an explicit run form and a warning is emitted.

## Task discovery (FR-001)

Search the current directory and ancestors for `Runefile` / `.runefile` (case-insensitive);
use the nearest. `--file` overrides. A `path/task` form may `cd` into a subdirectory before
running. If none found: error "no Runefile found" + exit code 2.

## Exit codes (FR-021)

| Code | Meaning |
|------|---------|
| 0 | All requested tasks succeeded |
| 1 | A task body failed (non-zero from a command/interpreter/agent) |
| 2 | Usage error / no Runefile / unknown task / bad arguments |
| 3 | Static validation error (parse/analyze) — nothing executed |
| 130 | Interrupted (SIGINT) |

A failing task reports the failing task name, the offending body line, and the underlying exit
code. `[no-exit-message]`-style suppression hides the trailing error banner but not the code.

## Output streams

- Task body stdout/stderr stream through to the user's stdout/stderr unchanged.
- Rune's own messages (echoed command lines, "cached"/"running" notices, diagnostics) go to
  stderr so stdout stays clean for piping.
- Color is applied only to a TTY and respects `NO_COLOR`.

## Acceptance mapping

US1 → default/list/run/params/chaining. US2 → exit 3 + diagnostics. US5 → `--list`,
`--dry-run`, `--summary`, `--dump --format json`, `--watch`.
