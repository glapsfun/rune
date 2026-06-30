# Implementation Plan: Interactive Task Picker (TUI)

**Branch**: `007-interactive-tui` | **Date**: 2026-06-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/007-interactive-tui/spec.md`

## Summary

Replace Rune's current `--choose` implementation (external `fzf` + a minimal
numbered-prompt fallback) with a single, modern, styled in-terminal task picker
built on the Charm stack (Bubble Tea + Bubbles + Lip Gloss). The picker lists the
non-private tasks from the resolved Runefile, supports incremental filtering over
**name and description** (matched portion highlighted), shows the highlighted
task's documentation, and on selection **tears itself down and hands the terminal
to the existing `execute()` path** so the task runs with full native fidelity and
the established exit-code mapping. Activation is **opt-in only via `--choose`**;
bare `rune` is unchanged, and every non-interactive/piped/CI path is untouched.

Core technical approach: a pure, table-testable `internal/tui` package holds the
Bubble Tea `Model`/`Update`/`View`; `internal/cli/choose.go` becomes a thin
adapter that loads the module, guards on a TTY, runs the program to completion,
reads the selected task name from the returned model, and delegates to
`execute()`. No engine package changes; no new subcommand.

## Technical Context

**Language/Version**: Go 1.25.0

**Primary Dependencies**: `github.com/charmbracelet/bubbletea` (event loop /
alt-screen), `github.com/charmbracelet/bubbles` (`list`, `textinput`),
`github.com/charmbracelet/lipgloss` (styling). Existing: `spf13/cobra`,
`mattn/go-isatty`, `fatih/color`. The task-execution and Runefile-loading
internals (`internal/cli` `loadModule`, `execute`, `firstLine`, `osMatches`,
`ast.Task`) are reused unchanged.

**Storage**: N/A (tasks are read from the in-memory parsed Runefile).

**Testing**: `go test` inside the Docker harness (`docker-compose run --rm test`),
race build via `-race`. Picker logic verified by **table-driven `Update` tests**
on the pure model (deterministic, no extra test dependency); binary-level
integration tests assert the opt-in/fallback behavior and that non-interactive
output is unchanged (golden).

**Target Platform**: Linux, macOS, Windows — single static binary, `CGO_ENABLED=0`
(Bubble Tea/Lip Gloss are pure Go; no cgo).

**Project Type**: Single-project CLI (task runner).

**Performance Goals**: Incremental filter feels instant for lists up to ~500
tasks (SC-005); locate-and-launch in <10s for ~50 tasks (SC-001). Filtering is
in-memory substring/fuzzy over a slice — trivially within budget.

**Constraints**: Must not alter any existing non-interactive output, format, or
exit code (US3 / FR-014, golden-protected). Must honor `NO_COLOR` and degrade
without color (FR-015). Must restore the terminal on every exit path (FR-016).
Must degrade gracefully on tiny terminals (FR-017). Picker only runs on a TTY;
`--choose` in a non-interactive context errors clearly rather than rendering a
broken UI.

**Scale/Scope**: One new internal package (~1 model file + styles + tests) and a
rewritten `choose.go` adapter (~replacing ~70 lines). 3 new direct dependencies.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment | Status |
|-----------|-----------|--------|
| I. Command Runner, Not a Build System | Picker only selects a task and defers to the unchanged execution path; no caching/skip semantics touched. | ✅ Pass |
| II. Errors Are a Feature | `--choose` reuses `loadModule`, so static parse/analyze errors render with `file:line:col` and exit 3 **before** any UI; the picker never opens on a broken Runefile (spec Edge Cases). | ✅ Pass |
| III. Minimal, Total DSL | No DSL/grammar change. | ✅ Pass |
| IV. Hand-Written Front End, Idiomatic Go | Engine packages (`token`…`runtime`) are untouched. The new `internal/tui` package is **presentation**, not engine logic, so the locked engine layout is preserved. Business logic stays in `internal/cli`. | ✅ Pass |
| V. Boringly Portable | Charm libs are pure Go, `CGO_ENABLED=0`, and work on Linux/macOS/Windows; no system-shell or WSL dependency. | ✅ Pass (verify static cross-build in `build` gate) |
| VI. Test-First, Multi-Layer Verification | Red-Green-Refactor: write `Update` table tests first; add integration tests for opt-in/non-TTY behavior; existing golden/integration suite guards US3. | ✅ Pass |
| VII. AI-Native, Secure by Default | MCP surface unchanged. The picker shows only already-public task names/docs (same data as `--list`); no secrets surfaced. | ✅ Pass |
| VIII. Go Engineering Discipline | `golangci-lint` clean, errors wrapped with `%w`, the Bubble Tea program runs synchronously and returns (no leaked goroutine), and it honors `opts.Ctx` via `tea.WithContext`. | ✅ Pass |

**Engineering Constraints**: Docker-only testing honored; package layout
additive (no engine restructuring); backward compatible (`--choose` keeps its
flag and completion; only its *implementation* changes); no DSL surface change so
no `docs/GRAMMAR.md` impact (user-facing `--choose` doc/help updated with its
fixtures).

**Result**: PASS — no violations. Complexity Tracking left empty. Adding three
Charm dependencies is the minimal way to meet the "modern, colorful" requirement
(FR-021) and is justified in `research.md`.

## Project Structure

### Documentation (this feature)

```text
specs/007-interactive-tui/
├── plan.md              # This file
├── spec.md              # Feature specification (with Clarifications)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── tui-picker.md    # Picker behavior + key bindings contract
├── checklists/
│   └── requirements.md  # Spec quality checklist (from /speckit-specify)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
internal/
├── tui/                      # NEW — pure presentation layer (no execution)
│   ├── picker.go             # Bubble Tea Model/Update/View; filter over name+doc
│   ├── styles.go             # Lip Gloss styles; honor NO_COLOR / Options.Color
│   ├── item.go               # Item type (name, doc) feeding bubbles/list
│   └── picker_test.go        # Table-driven Update/state-transition tests
├── cli/
│   ├── choose.go             # REWRITTEN — TTY guard + run program + delegate to execute()
│   ├── choose_test.go        # NEW — opt-in/non-TTY/empty-list behavior
│   ├── dispatch.go           # unchanged (Run still routes opts.Choose → chooseAndRun)
│   ├── run.go                # unchanged (execute, firstLine, osMatches reused)
│   └── serve.go              # unchanged (loadModule reused)
└── ast/ast.go                # unchanged (Task.Name, Task.Doc, IsPrivate reused)

cmd/rune/root.go              # unchanged flag wiring; `--choose` help text reused

docs/                         # user-facing `--choose` doc updated (with test/docs fixtures)
test/                         # integration: --choose opt-in & non-TTY error (golden)
```

**Structure Decision**: Single-project CLI. The picker lives in a **new
`internal/tui` package** kept free of execution logic so it is unit-testable
(`Update` is a pure function of `(Model, Msg)`). `internal/cli/choose.go` is the
only wiring seam: it owns module loading, TTY detection, running the program, and
delegating the selection to the existing `execute()`. This preserves the locked
engine package layout (Principle IV) and the single execution path.

## Complexity Tracking

> No constitution violations — table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
