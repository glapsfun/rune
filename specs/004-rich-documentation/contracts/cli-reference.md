# Contract: CLI Reference Surface

**Feature**: `004-rich-documentation`

The authoritative CLI surface that `docs/cli.md` MUST mirror exactly (FR-013). Transcribed
from `cmd/rune/main.go` (flags + reserved subcommands) and `internal/cli/exit.go` (exit codes)
as of this plan. The harness's drift check (D7) fails if `docs/cli.md` and the binary disagree.

> This is a snapshot of the **existing** surface for documentation accuracy. This feature does
> NOT add, remove, or change any flag, subcommand, or exit code (FR-024). If the binary changes
> later, both the binary and this reference move together under the normal CI gate.

## Invocation

```text
rune [global flags] [VAR=VALUE ...] [TASK [ARGS...]] ...
```

- Global flag parsing stops at the first positional (`SetInterspersed(false)`), so flags after
  a task name pass through to the task untouched.
- Rune's own messages go to **stderr**; task stdout stays clean for piping (golang-cli
  discipline; `SilenceUsage`/`SilenceErrors` are set).
- Running with no task runs the configured default; `--list` lists tasks and runs nothing.

## Global flags (from `cmd/rune/main.go`)

| Flag | Shorthand | Type | Meaning |
|------|-----------|------|---------|
| `--file` | `-f` | string | Use a specific Runefile instead of upward discovery |
| `--list` | | bool | List non-private tasks with docs; run nothing |
| `--dry-run` | | bool | Print the resolved execution plan; run nothing |
| `--summary` | | bool | Print task names that would run, one per line |
| `--dump` | | bool | Emit the parsed Runefile (canonical text, or JSON) |
| `--format` | | string | Output format for `--dump` (`json`) |
| `--set` | | string[] | Override a variable: `--set NAME VALUE` |
| `--watch` | | bool | Re-run on file changes |
| `--choose` | | bool | Interactive task picker |
| `--yes` | | bool | Auto-approve `[confirm]` tasks |
| `--quiet` | | bool | Suppress command echo |
| `--fmt` | | bool | Rewrite the Runefile in canonical formatting |
| `--clear-cache` | | bool | Remove the project-local `.rune/cache` directory |
| `--version` | | bool | Print version + commit (Cobra built-in) |
| `--help` | `-h` | bool | Help (Cobra built-in) |

**Static validation (no `check` subcommand).** There is **no built-in `rune check`** — the
parser/analyzer always run before execution (Principle II), and validation *without* running is
obtained via the run-nothing flags above: **`--list`** (parse+analyze, list tasks, no task arg
needed — the harness's choice), `--dry-run` (parse+analyze + print the plan for a target task),
or `--dump` (emit the parsed Runefile). `docs/cli.md` MUST document these as the
"validate-without-running" path and MUST NOT imply a `check` subcommand exists. (A Runefile may
*define a task named* `check`; that is a project convention, not a CLI feature — finding F1.)

## Reserved subcommands (from `runServe` / `genCompletion`)

| Subcommand | Args | Meaning |
|------------|------|---------|
| `mcp` | — | Serve over stdio MCP (shorthand) |
| `serve` | `--http`, `--addr <a>`, `--token-file <f>`, `--mcp` | Serve MCP; HTTP endpoint is opt-in, localhost-bound, token-gated (Principle VII) |
| `completion` | `[bash\|zsh\|fish\|powershell]` | Emit a shell completion script |

These are handled explicitly so they never silently shadow a task of the same name —
document that nuance.

## Exit codes (from `internal/cli/exit.go`)

| Code | Name | When |
|------|------|------|
| `0` | success | All requested tasks succeeded |
| `1` | task failure | A task body failed |
| `2` | usage error | No Runefile / unknown task / bad args (also the default for unknown errors) |
| `3` | validation error | Static parse/analyze failure — **nothing executed** (Principle II) |
| `130` | interrupted | SIGINT / SIGTERM (signal 130) |

`docs/cli.md` MUST list exactly these five codes with these meanings (FR-013), and the
troubleshooting page MUST map its failure modes onto them (FR-011).

## Drift check (enforced by `test/docs`)

- Every flag in the table above MUST appear in the binary's `--help`; every flag in `--help`
  MUST be documented. Mismatch → test failure (D7).
- Exit-code list in `docs/cli.md` MUST equal the set `{0,1,2,3,130}`.
