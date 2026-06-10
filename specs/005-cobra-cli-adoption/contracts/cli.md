# CLI Contract: Modern CLI Interface

**Feature**: `005-cobra-cli-adoption` · **Date**: 2026-06-10

This is the externally observable contract for the Cobra-based CLI. It supersedes the
ad-hoc dispatch described in `rune/main.go` while preserving every behavior in
`specs/001-rune-task-runner/contracts/cli.md`. Anything not listed here is unchanged.

## Command tree

```
rune [global flags] [VAR=VALUE ...] [TASK [ARGS...]] ...   # root: task dispatcher
rune serve [--http] [--addr ADDR] [--token-file PATH] [--mcp]
rune mcp                                                   # alias of `serve` (stdio)
rune version
rune completion {bash|zsh|fish|powershell}                 # Cobra auto-generated
rune help [command]                                        # Cobra auto-generated
rune --help | -h            # any command
rune --version              # parity with `rune version`
```

**Discoverability (FR-001/SC-001)**: `rune --help` lists `serve`, `version`,
`completion`, `help`, and points to `--list` for tasks. Every built-in is visible
(none `Hidden`). `mcp` appears as an alias of `serve`.

## Routing & precedence

1. The target command is chosen from the **first non-flag token**.
2. If it matches a command name or alias (`serve`/`mcp`/`version`/`completion`/`help`),
   that command runs (**built-in precedence**).
3. Otherwise the **root task dispatcher** runs: tokens are `VAR=VALUE` overrides and
   task names with positional args, bound by arity. Args after a task name are passed
   to the task untouched (`SetInterspersed(false)` retained). (FR-006/FR-007)
4. **Escape hatch (FR-008)**: `rune -- <name> [args…]` always runs `<name>` as a task,
   even when it collides with a reserved command.

## `serve` / `mcp`

| Flag | Type | Default | Meaning |
|------|------|---------|---------|
| `--http` | bool | false | Serve over Streamable HTTP instead of stdio. |
| `--addr` | string | "" | HTTP listen address (requires `--http`). |
| `--token-file` | string | "" | Bearer-token file for HTTP (requires `--http`). |
| `--mcp` | bool | false | Accepted for clarity; MCP is the only protocol. |

- Default (no `--http`) = stdio MCP server. `rune mcp` == `rune serve` (stdio).
- `--addr` or `--token-file` without `--http` → **usage error, exit 2** (FR-015).
- Behavior delegates to the unchanged `cli.ServeMCP(opts, useHTTP, addr, tokenFile)`.
- Secrets/tokens are read from file/env only — never echoed in help or logs
  (Constitution VII).

## Shell completion (FR-009–FR-012, SC-003)

- `rune completion <shell>` writes a completion script to stdout for bash, zsh, fish,
  powershell; its `--help` includes per-shell install instructions (Cobra default).
- Unsupported shell → clear error naming supported shells (Cobra default; FR-013).
- Dynamic completion protocol: `rune __complete <args> <toComplete>` (hidden) returns
  candidates + a directive line `:<n>`.
  - `rune <TAB>` (i.e. `rune __complete ""`) → built-in command names **and** the current
    Runefile's non-private, OS-matching **task names with their doc summaries**.
  - Directive is `ShellCompDirectiveNoFileComp` (no file completion for task names).
  - No Runefile / parse error → built-in commands only, **no error emitted** (FR-012).
  - The completion code never writes to stdout itself.

## Errors (FR-016/FR-017)

- Unknown task/command → `rune: unknown task: <tok>` and, when a close match exists,
  `(did you mean "<nearest>"?)`; candidates include task names and reserved command
  names. Exit 2.
- `SilenceUsage`/`SilenceErrors` are set: errors are concise; full usage is shown only
  via `--help`.

## Exit codes (FR-018 — unchanged, authoritative in `internal/cli/exit.go`)

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Task body failed |
| 2 | Usage error / no Runefile / unknown task / bad args / invalid serve flags |
| 3 | Static parse/analyze error — nothing executed |
| 130 | Interrupted (SIGINT/SIGTERM) |

## I/O & environment (FR-019/FR-021/FR-022 — unchanged)

- stdout: program/task output only (pipe-safe). stderr: Rune's diagnostics, banners,
  prompts, `running:`/`cached:` lines.
- Color on stderr only when a TTY, `NO_COLOR` unset, and color enabled.
- Static diagnostics render with `file:line:col` carets; `[no-exit-message]` suppresses
  the trailing banner (not the exit code).
- SIGINT/SIGTERM cancels the run via the shared context → exit 130.

## Backward-compatibility assertions (regression guardrail)

The existing end-to-end suite MUST stay green unchanged (SC-006). In particular:
`rune <task>`, `rune --list`, `rune --dry-run`, `rune --fmt`, `rune mcp`,
`rune serve --http --addr … --token-file …`, task arg pass-through, exit codes, and
stdout/stderr separation behave exactly as before.
