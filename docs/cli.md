# CLI Reference

> The complete command-line surface — every flag, subcommand, and exit code. New here? Start
> with [Getting started](getting-started.md); for the file format see the
> [language guide](runefile.md).

```text
rune [global flags] [VAR=VALUE ...] [TASK [ARGS...]] ...
```

Rune discovers a `Runefile` (or `.runefile`) by walking up from the current directory,
then parses and statically validates it before running anything. You can chain multiple
tasks and interleave `VAR=VALUE` variable overrides between them.

## Running tasks

```sh
rune                 # run the default task (set default := "...")
rune build           # run the `build` task
rune build test      # run `build`, then `test`
rune greet Ada       # run `greet` passing "Ada" as its first argument
rune build target=release   # override the variable `target` for this run
```

- Positional `VAR=VALUE` tokens override variables defined in the Runefile. They may be
  placed before or between task names.
- Arguments after a task name are bound to that task's parameters.
- Dependencies run before their dependent task, each at most once per invocation.

### Validating without running

There is **no separate `check` subcommand** — Rune always parses and statically validates a
Runefile before running anything. To validate *without* running, use a run-nothing flag:

```sh
rune --list            # validate + list tasks (no task argument needed)
rune --dry-run build   # validate + show the resolved plan for `build`
rune --dump            # validate + print the parsed Runefile
```

A clean exit `0` means the Runefile is valid; exit `3` means a static error (with a
`file:line:col` span — see [Troubleshooting](troubleshooting.md)).

## Global flags

| Flag | Description |
|------|-------------|
| `-f`, `--file <path>` | Use a specific Runefile instead of upward discovery. |
| `--list` | List non-private tasks with their docs, then exit (runs nothing). |
| `--dry-run` | Print the resolved execution plan, then exit (runs nothing). |
| `--summary` | Print the task names that would run, one per line. |
| `--dump` | Emit the parsed Runefile as canonical text (or JSON with `--format json`). |
| `--format <fmt>` | Output format for `--dump` (currently `json`). |
| `--set <NAME=VALUE>` | Override a variable. The positional `NAME=VALUE` form (above) is the primary mechanism. |
| `--watch` | Re-run the requested task(s) when files change. |
| `--choose` | Interactive task picker. |
| `--yes` | Auto-approve `[confirm]` tasks (non-interactive). |
| `--quiet` | Suppress command echo. |
| `--fmt` | Rewrite the Runefile in canonical formatting. |
| `--clear-cache` | Remove the project-local `.rune/cache` directory. |
| `--version` | Print version and commit. |
| `-h`, `--help` | Show help. |

### Examples

```sh
rune --list                      # what can I run here?
rune --dry-run deploy            # what would `deploy` do?
rune --summary deploy            # just the task names, for scripting
rune --dump --format json        # machine-readable parse of the Runefile
rune -f ci/Runefile test         # use an explicit file
rune --fmt                       # canonically reformat the Runefile in place
rune --watch test                # re-run tests on file changes
rune --yes clean                 # run a [confirm] task without prompting
rune --clear-cache               # drop the [cache] fingerprints
```

## Subcommands

Rune reserves a few names so they never silently shadow a task:

| Command | Description |
|---------|-------------|
| `rune mcp` | Start the MCP server over **stdio** (for local agents/IDEs). |
| `rune serve [--http] [--addr <addr>] [--token-file <file>]` | Start the MCP server; `--http` enables the Streamable HTTP transport, `--addr` sets the bind address, `--token-file` supplies the bearer token. |
| `rune completion [bash\|zsh\|fish\|powershell]` | Print a shell-completion script (defaults to `bash`). |

See the [MCP guide](mcp.md) for exposing tasks to AI agents and the security model.

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All requested tasks succeeded. |
| `1` | A task body failed. |
| `2` | Usage error — no Runefile found, unknown task, or bad arguments. |
| `3` | Static validation error (parse/analyze) — nothing was executed. |
| `130` | Interrupted (SIGINT / Ctrl-C). |

Rune's own messages go to **stderr**, so a task's `stdout` stays clean for piping.

## Output & color

Color is emitted on a TTY unless `NO_COLOR` is set or output is redirected. Validation
errors are rendered with `file:line:col` and a caret-underlined source span.

## See also

- [Runefile language guide](runefile.md) — the file format these commands run.
- [Guides](guides/README.md) — task-oriented deep dives per capability.
- [Troubleshooting](troubleshooting.md) — what each error and exit code means.
