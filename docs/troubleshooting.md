# Troubleshooting

> What Rune does when something is wrong, and how to fix it. Most mistakes are caught
> **before any task runs** (Principle: *errors are a feature*), with a precise
> `file:line:column` and a caret-underlined span. New here? See [Getting started](getting-started.md).

## Exit codes at a glance

| Code | Meaning |
|------|---------|
| `0` | All requested tasks succeeded. |
| `1` | A task body failed at runtime. |
| `2` | Usage error — no Runefile found, unknown task at the command line, or bad arguments. |
| `3` | Static validation error (parse/analyze) — **nothing executed**. |
| `130` | Interrupted (SIGINT / Ctrl-C). |

Static errors (`3`) are detected up front, so a broken Runefile never leaves your project
half-built. See the [CLI reference](cli.md#exit-codes).

## Common errors

### "unknown task" (as a dependency) — exit 3

A task depends on a name that doesn't exist. Caught during analysis; nothing runs.

```text
Runefile:1:4: error: unknown task: missing
1 | a: missing
  |    ^^^^^^^
```

**Fix:** define the task, or correct the typo in the dependency list.

### "unknown task" (at the command line) — exit 2

You asked to run a task that isn't defined.

```text
rune: unknown task: nope
```

**Fix:** run `rune --list` to see the available tasks.

### "undefined variable" — exit 3

An interpolation or expression references a name that was never defined.

```text
Runefile:2:11: error: undefined variable: nope
2 |     @echo {{nope}}
  |           ^^^^^^^^
```

**Fix:** define it with `name := "value"`, pass it as a parameter, or override it on the
command line (`rune task NAME=value`). See [Settings & dotenv](guides/settings-and-dotenv.md).

### "dependency cycle" — exit 3

Tasks depend on each other in a loop. Rune reports the full cycle path.

```text
Runefile:3:1: error: dependency cycle: a → b → a
3 | b: a
  | ^
```

**Fix:** break the loop — extract the shared work into a third task both can depend on. See
[Dependencies & post-hooks](guides/dependencies-and-hooks.md).

### Argument count mismatch (arity) — exit 3

A task (or a dependency) is called with the wrong number of arguments.

```text
Runefile:3:7: error: task "greet" expects at least 1 argument(s), got 0
3 | main: greet
  |       ^^^^^
```

**Fix:** pass the required argument — e.g. `(greet "Ada")` in the dependency, or give the
parameter a default. See [Parameters](guides/parameters.md).

### "no Runefile found" — exit 2

Rune walks up from the current directory looking for a `Runefile`; none was found.

```text
rune: no Runefile found in this directory or any parent
```

**Fix:** run from inside your project, or point at a file with `rune --file path/to/Runefile`.

### A task body failed — exit 1

A command in the body exited non-zero. Rune stops at the first failure and names the task, the
failing line, and the exit status.

```text
rune: task "boom" failed at "exit 7": exit status 7
```

**Fix:** address the failing command. Prefix a line with `-` to deliberately ignore its error.

### Missing interpreter

A task with a `(python)`, `(node)`, or custom executor needs that runtime on `PATH`. If it's
absent, the task fails with an actionable error naming the missing interpreter (exit `1`).

**Fix:** install the interpreter, or run in an environment that has it. Note the
[container image](docker.md) ships only the static binary — see its caveat. See
[Executors](guides/executors.md).

### A cached task ran when you expected a skip

`[cache]` re-runs (rather than falsely skipping) when a declared **output is missing** or an
**input can't be resolved** — by design. Clear stale fingerprints with `rune --clear-cache`.
See [Caching](guides/caching.md).

## Still stuck?

- `rune --dry-run <task>` shows the resolved plan without running anything.
- `rune --dump --format json` prints the parsed Runefile for inspection.
- Check the [CLI reference](cli.md) and the [language guide](runefile.md).
