# Phase 0 Research: Interactive Task Picker (TUI)

All Technical Context items were resolvable from the codebase and the spec's
Clarifications session; there are no open `NEEDS CLARIFICATION` markers. This file
records the design decisions that shape Phase 1.

> **Version note**: Bubble Tea / Bubbles / Lip Gloss APIs drift between releases.
> Pin exact versions in `go.mod` and verify the specific symbols below
> (`tea.WithContext`, `tea.WithAltScreen`, `list.Model`, `textinput.Model`,
> `FinalModel` return from `Program.Run`) against the versions actually resolved
> before relying on them.

## Decision 1 — UI stack: Bubble Tea + Bubbles + Lip Gloss

- **Decision**: Build the picker with `charmbracelet/bubbletea` (event loop,
  alt-screen, terminal teardown), `charmbracelet/bubbles` (`list` and
  `textinput` components), and `charmbracelet/lipgloss` (styling/highlighting).
- **Rationale**: Explicitly requested. Pure Go (keeps Principle V: `CGO_ENABLED=0`,
  cross-platform, single static binary). `bubbles/list` already provides
  navigation, pagination, and incremental filtering, minimizing custom code.
  Lip Gloss honors `NO_COLOR` and adapts to terminal capabilities, satisfying
  FR-015/FR-021 with little effort.
- **Alternatives considered**:
  - *Keep `fzf`* — rejected by FR-020 (external dependency, no detail pane, no
    in-process styling control).
  - *Hand-rolled raw-terminal UI* — rejected: re-implements what Bubble Tea
    already does (resize handling, alt-screen, input parsing) and is harder to
    keep portable and test.
  - *`promptui` / `survey`* — rejected: less suited to a full-screen, filterable
    list-with-detail layout and less actively aligned with the "modern" goal.

## Decision 2 — Run model: tear down the TUI, then delegate to `execute()`

- **Decision**: The Bubble Tea program records the selected task name in its model
  and calls `tea.Quit`. `choose.go` runs `program.Run()` **synchronously**, reads
  the selection from the returned final model, and — if non-empty — calls the
  existing `execute(opts, runefile, append([]string{picked}, args...))`. The task
  therefore runs **after** the TUI has fully exited and restored the terminal.
- **Rationale**: This is the simplest way to satisfy Q4/FR-008 (full native
  fidelity), FR-009 (native `Ctrl-C`), and FR-010 (existing `CodeFor` exit-code
  mapping) — the task inherits the real terminal exactly as a direct
  `rune <task>` would. It avoids `tea.ExecProcess`/output-capture entirely, so
  interactive subprocesses (REPLs, editors, prompts) keep working, and there is
  no risk of the picker's render state corrupting task output. It also matches the
  current `--choose` semantic of "run one task, then exit" (FR-010: no return to
  list).
- **Alternatives considered**:
  - *`tea.ExecProcess` to run the task and return to the list* — rejected per Q4:
    more complex, fragile for arbitrary/interactive tasks, and contradicts the
    "no return to list" decision.
  - *Capture task output into a Bubble Tea viewport pane* — rejected: breaks
    interactive tasks and risks output-fidelity divergence from direct runs.

## Decision 3 — Activation, TTY guard, and the FR-012 opt-out

- **Decision**: Activation is opt-in via `--choose` only (FR-019); bare `rune` is
  unchanged and the picker is never auto-launched. `choose.go` guards on an
  interactive terminal: the picker runs only when **stdin and stdout are TTYs**
  (using `mattn/go-isatty`, mirroring `useColor()` in `cmd/rune/root.go`). When
  `--choose` is invoked without an interactive terminal (piped/redirected/CI),
  Rune returns the existing clear usage error
  (`"--choose requires an interactive terminal"`) instead of rendering a broken
  UI — it does **not** silently fall back to running a task.
- **Rationale**: Because the picker is opt-in, FR-012's "explicit opt-out that
  forces non-interactive behavior" is satisfied by *not passing `--choose`* — the
  default is already non-interactive. No separate `--no-tui` flag is introduced
  (keeps the surface minimal, Principle IV). The non-TTY guard preserves US3:
  capturing `rune --choose` in a script yields a deterministic error, never a
  control-sequence-polluted stream.
- **Alternatives considered**:
  - *Add a `--no-tui` flag / `RUNE_NO_TUI` env* — deferred as unnecessary under
    opt-in activation; can be revisited if auto-launch is ever introduced.
  - *Silently run the first/only task when `--choose` hits a non-TTY* — rejected:
    surprising and untestable; an explicit error is clearer (Principle II).

## Decision 4 — Filtering over name + description

- **Decision**: The filter matches typed text against both the task **name** and
  its **description** (first line of `ast.Task.Doc`, via the existing `firstLine`
  helper), highlighting the matched span (FR-003, Q2). Implemented using
  `bubbles/list`'s filtering with a custom `FilterValue()`/match that spans both
  fields, or a custom filter function if finer highlight control is needed.
- **Rationale**: Matches the discovery goal and the line-based behavior of the
  `fzf` picker being replaced. Data is already available from `loadModule`.
- **Alternatives considered**: name-only (rejected by Q2); fuzzy vs substring —
  substring is sufficient and predictable for ~hundreds of tasks; fuzzy can be a
  later refinement without changing the contract.

## Decision 5 — Task list source & visibility

- **Decision**: Build picker items from the loaded module's tasks, excluding tasks
  where `IsPrivate()` is true or that do not match the current OS (`osMatches`,
  reused from `run.go`). Each item carries `Name` and `firstLine(Doc)`; the detail
  pane may show the full `Doc`.
- **Rationale**: Reuses the exact selection rules already applied by `--list` and
  shell completion, so the picker shows the same tasks users already see
  (consistency; no new visibility logic to test).

## Decision 6 — Styling, color, and graceful degradation

- **Decision**: Lip Gloss styles defined in `internal/tui/styles.go`, gated on the
  resolved color decision passed in from `Options.Color` / `NO_COLOR`. On a
  no-color terminal the picker renders plain (no escape styling) but fully usable
  (FR-015). Layout uses the window size from `tea.WindowSizeMsg`; when the
  terminal is too small, the detail pane collapses and the list stays usable
  rather than crashing (FR-017).
- **Rationale**: Satisfies FR-015/FR-021/FR-017 with Bubble Tea's built-in resize
  handling and Lip Gloss's color adaptation.

## Decision 7 — Testing strategy (test-first, Docker-only)

- **Decision**: Primary tests are **table-driven `Update` tests** on the pure
  model: feed `tea.Msg` values (key presses, `WindowSizeMsg`) and assert state
  (filtered items, highlighted index, selection, quit). No extra test dependency
  (`teatest`) is added — the model is a pure function, so direct `Update`
  assertions are deterministic and fast. Binary-level integration tests assert:
  (a) `--choose` on a non-TTY returns the usage error and exit 2; (b)
  non-interactive paths (`--list`, direct run, piped) remain byte-identical
  (existing golden suite). All run under `docker-compose run --rm test`.
- **Rationale**: Avoids flaky full-render golden snapshots while still covering
  behavior; aligns with Principle VI and the Docker-only constraint.
- **Alternatives considered**: `charmbracelet/x/exp/teatest` golden snapshots —
  deferred; useful for end-to-end render checks but adds a dependency and
  snapshot brittleness not warranted for the initial slice.

## Decision 8 — Dependency & portability impact

- **Decision**: Add the three Charm modules as direct dependencies; run the
  `build` (static cross-OS) and `release-dryrun` gates to confirm binary size and
  cross-compilation remain acceptable.
- **Rationale**: Required for FR-021; the libraries are pure Go and widely used on
  all three target OSes. Binary-size growth is the main tradeoff and is acceptable
  for the UX gain; it is measured, not assumed (Principle VIII).
