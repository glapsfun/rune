# Contract: Interactive Task Picker (`--choose`)

This contract defines the observable behavior of the interactive task picker. It
**supersedes** the prior `--choose` implementation (external `fzf` + numbered
fallback) described in `specs/001-rune-task-runner/contracts/cli.md`; the flag,
its name, and its shell completion are unchanged — only the implementation and
the interactive experience change.

## Invocation

```text
rune --choose                 # open the picker, run the selected task
rune --choose -- --watch      # forward `--watch` to whatever task is selected
```

- `--choose` is the **only** entry point (FR-019). There is no new subcommand.
- Bare `rune` (no task, no flags) does **not** open the picker.

## Activation matrix

| Condition | Behavior | Exit |
|-----------|----------|------|
| `--choose`, interactive TTY (stdin+stdout), ≥1 selectable task | Picker opens | per selected task |
| `--choose`, **not** a TTY (piped/redirected/CI) | Error: `--choose requires an interactive terminal` | 2 |
| `--choose`, TTY, **no** selectable tasks | Error: `no tasks to choose from` | 2 |
| `--choose`, Runefile has static errors | Diagnostics rendered; picker not opened | 3 |
| no `--choose` | Existing behavior, picker never involved | unchanged |

"Selectable task" = non-private (`IsPrivate()` false) and OS-matching
(`osMatches`).

## In-picker behavior

| Action | Key(s) | Result |
|--------|--------|--------|
| Move highlight | `↑`/`↓`, `j`/`k` | Highlight moves; detail pane updates |
| Filter | type (or `/` then type) | List narrows to items whose **name or description** matches; matched span highlighted |
| Clear filter | `Esc` | Full list restored |
| Confirm | `Enter` | Selected task name captured; picker tears down |
| Cancel | `q`, `Ctrl-C` | No selection; picker tears down; nothing runs |

- The highlighted task's documentation (full `Doc`) is visible in a detail pane.
- On a non-color terminal (`NO_COLOR` or color disabled) the picker renders
  without ANSI styling but remains fully usable.
- On a terminal too small for the full layout, the detail pane collapses; the
  list stays usable; no crash or corruption.

## Selection → execution handoff

1. On `Enter`, the picker program exits (terminal + cursor restored).
2. Rune runs the selected task via the **same execution path** as
   `rune <task>`, forwarding any pass-through `args` supplied after `--`.
3. The task has **direct terminal access**: output streams natively, color and
   progress display correctly, interactive subprocesses work, and `Ctrl-C`
   interrupts the task as in a direct invocation.
4. Rune exits with the task's status (`CodeFor` mapping). The picker does **not**
   re-open after the task completes.

## Exit codes (unchanged mapping)

| Outcome | Code |
|---------|------|
| Cancelled in picker (nothing run) | 0 |
| Selected task succeeds | 0 |
| Selected task body fails | 1 |
| `--choose` non-TTY / no tasks | 2 |
| Static Runefile error (picker not opened) | 3 |
| Interrupted (SIGINT) during picker or task | 130 |

## Compatibility guarantees (US3 / FR-014)

- No change to any non-interactive command output, format, or exit code.
- The picker emits its UI only to an interactive terminal; it never injects
  cursor movement, colors, or prompts into piped/redirected/CI output.
- `--list`, `--dry-run`, `--dump`, `--summary`, and direct task runs never open
  the picker.
