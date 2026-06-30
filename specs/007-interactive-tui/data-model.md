# Phase 1 Data Model: Interactive Task Picker (TUI)

The picker introduces **no persistent data** and **no new domain types in the
engine**. It defines small in-memory view types in `internal/tui` that project
the already-parsed `ast.Task` set into a selectable list. All entities below are
process-local and discarded when the picker exits.

## Entity: PickerItem

A single selectable row, projected from an `ast.Task`.

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| `Name` | `string` | `ast.Task.Name` | The task name; what gets executed on selection. |
| `Desc` | `string` | `firstLine(ast.Task.Doc)` | One-line description shown in the row. |
| `Doc`  | `string` | `ast.Task.Doc` | Full documentation shown in the detail pane. |

**Derivation rules**:
- Included only if `!task.IsPrivate()` **and** `osMatches(task, runtime.GOOS)`
  (same rules as `--list` / completion).
- Order follows the Runefile's task order (as returned by the loaded module),
  matching `--list`.

**FilterValue**: matching is performed over `Name` + `Desc` (FR-003); the matched
span is highlighted in the rendered row.

## Entity: Model (Bubble Tea state)

The full picker state; a pure value updated by `Update(msg) (Model, tea.Cmd)`.

| Field | Type | Purpose |
|-------|------|---------|
| `list` | `bubbles/list.Model` | Holds items, cursor/highlight, pagination, and the filter input. |
| `items` | `[]PickerItem` | The full, unfiltered backing set (for the detail pane + reset). |
| `width`, `height` | `int` | Latest terminal size from `tea.WindowSizeMsg`; drives layout/collapse. |
| `selected` | `string` | Task name chosen on confirm; `""` means "cancelled / nothing selected". |
| `styles` | `Styles` | Lip Gloss styles, pre-resolved for the current color mode. |

**State transitions** (driven by `Update`):

| From → To | Trigger | Effect |
|-----------|---------|--------|
| browsing → browsing | ↑/↓ (and `j`/`k`) | Move highlight; detail pane reflects new highlight. |
| browsing → filtering | `/` or typing (per list config) | Open/extend filter; list narrows over name+desc. |
| filtering → browsing | `Esc` / clear filter | Restore full list. |
| browsing → **selected** | `Enter` on a highlighted item | Set `selected = item.Name`; return `tea.Quit`. |
| any → **cancelled** | `q` / `Ctrl-C` / `Esc` at top level | Leave `selected = ""`; return `tea.Quit`. |
| any → any | `tea.WindowSizeMsg` | Recompute `width`/`height`; collapse detail pane if too small. |

**Terminal lifecycle**: the program is started with alt-screen; Bubble Tea
restores the screen and cursor on `Quit` and on error (FR-016). Empty item set is
handled by the adapter *before* starting the program (see Validation), so the
model never renders an empty selectable list.

## Entity: Selection (adapter-level result)

Not a struct — the outcome the adapter reads from the final model:

- `selected == ""` → user cancelled; `chooseAndRun` returns `nil` (exit 0), runs
  nothing.
- `selected != ""` → `chooseAndRun` calls
  `execute(opts, runefile, append([]string{selected}, args...))`, forwarding any
  command-line pass-through `args` (FR-006/Q3). The task then flows through the
  unchanged pipeline and `CodeFor` exit mapping.

## Validation & guards (enforced in `internal/cli/choose.go`)

1. **Static errors first**: `loadModule` runs parse/compose/analyze; on errors it
   renders diagnostics and returns `ValidationError` (exit 3) — the picker is not
   started (Principle II; spec Edge Cases).
2. **TTY guard**: if stdin/stdout is not an interactive terminal, return the usage
   error `"--choose requires an interactive terminal"` (exit 2). No UI is started.
3. **Empty list**: if no non-private, OS-matching tasks exist, return the usage
   error `"no tasks to choose from"` (FR-007). No UI is started.
4. **Context**: the program is run with `tea.WithContext(opts.ctx())` so an
   external SIGINT cancels the picker cleanly (exit treated as cancellation).

## Relationships

```text
loadedModule.file.Tasks ──filter(IsPrivate, osMatches)──▶ []PickerItem ──▶ Model.list
                                                                              │
                                                            Enter ───────────▶ selected (Name)
                                                                              │
                                          chooseAndRun ──execute(picked, args)──▶ existing pipeline
```
