# Contract: styled output surfaces

For each surface: the **plain** form (color OFF) is the invariant baseline; the
**styled** form (color ON) adds only zero-width emphasis via `internal/style`
roles. Text, order, indentation, columns, and stream are identical in both.

## `--list` (stdout, gated by ColorStdout)

Plain (unchanged from today):
```
Available tasks:
  [build]
    compile   # build the binary
    test      # run the suite
    lint
```

Styled — same bytes plus roles:
- `Available tasks:` → Heading
- `[build]` group header → Heading
- task name (`compile`, `test`, `lint`) → TaskName (bold/accent)
- `# build the binary` doc → Muted

**Alignment rule (FR-012/SC-002)**: the `#` column is padded by the *visible*
name width (rune count), not the styled byte length; every `#` stays aligned in
both modes. Tasks with no doc render exactly as today.

## Run status / cache (stderr, gated by ColorStderr)

| Line (text unchanged) | Role |
|-----------------------|------|
| `running: <task>` | Success/active |
| `cached: <task>` | Muted |
| `would run: <task>` | Muted |
| `would skip (cached): <task>` | Muted |
| `warning: failed to write cache for <task>: <err>` | Warning |

## Command echo (stderr, gated by ColorStderr)

- Each non-suppressed echoed command line → Muted.
- `@`-prefixed (NoEcho) and `--quiet` lines are **not printed at all** — checked
  before styling (FR-016). Styling never reveals a suppressed line.

## Diagnostics (stderr, gated by ColorStderr)

Plain golden (`testdata/diag/render.golden`) is **byte-for-byte unchanged**
(FR-018):
```
Runefile:2:11: error: undefined variable: nope
2 |     @echo {{nope}}
  |           ^^^^^^^^
```
Styled — same layout, roles applied:
- `error`/`warning` severity word → Error / Warning
- caret run `^^^^^^^^` → Caret (same count/position; SC-003)
- optional: `Runefile:2:11` locator → Locator (emphasis only, no width change)

## `--help` / usage (stdout, gated by ColorStdout) — **redesigned**

New structure (the new plain baseline; a new golden/snapshot):
```
Rune — a shared task runner for humans and AI agents.

Usage:
  rune [global flags] [VAR=VALUE ...] [TASK [ARGS...]]

Common commands:
  rune --list            List the tasks in your Runefile
  rune <task> [args]     Run a task
  rune --choose          Pick a task interactively
  rune serve             Expose tasks to AI agents over MCP

Examples:
  rune --list                 # see available tasks
  rune build                  # run the 'build' task
  rune build --watch          # pass flags through to the task
  rune -- test                # run a task whose name shadows a builtin
  rune --choose               # interactive picker
  rune serve --http :7000     # serve over MCP

Flags:
  -f, --file string   use a specific Runefile instead of upward discovery
      --color string  when to colorize output: auto|always|never (default "auto")
  ...
```
- Section **headings** (`Usage:`, `Common commands:`, `Examples:`, `Flags:`) →
  Heading role when ColorStdout is on; body stays plain.
- Piped/`NO_COLOR`/`--color=never`: identical structure, **no ANSI**, fully
  readable (FR-021).
- Each common workflow (run task, `--list`, `--choose`, `serve`) has ≥1 worked
  example (FR-020, SC-006).
