# Feature Specification: Interactive Task Picker (TUI)

**Feature Branch**: `007-interactive-tui`

**Created**: 2026-06-30

**Status**: Draft

**Input**: User description: "Redesign Rune's CLI user experience to feel modern and interactive by adding a full-screen interactive task picker, without breaking Rune's existing scriptable, non-interactive behavior. Rune is a task runner: a repo-root `Runefile` defines dev tasks, run via `rune <task>` / `rune --list`. The interactive experience must be additive — non-interactive, piped, and CI usage must be unchanged. At minimum: browse the tasks from the Runefile, search/filter, view a task's details, and run a selected task while watching its output. Activation must be TTY-aware with a graceful fallback and an explicit opt-out."

## Clarifications

### Session 2026-06-30

- Q: What is the picker's CLI entry point (opt-in surface)? → A: Reuse the existing `--choose` flag as the only entry point; no new subcommand.
- Q: What does the incremental filter match against? → A: Both the task name and its description, highlighting where matched.
- Q: How are arguments passed to the task selected in the picker? → A: CLI args supplied before launch are forwarded to the picked task; no in-picker argument entry.
- Q: How does the selected task run and display its output? → A: The picker tears down and hands off to the real terminal; the task runs with full native fidelity (like `rune <task>`), then Rune exits with the task's status. No embedded output pane; no return to the list.
- Q: What is the canonical term for the feature? → A: "interactive task picker" (or "the picker") — normalized throughout; "browser" retired.

### Session 2026-06-30 (post-implementation)

- Q: How is the FR-012 "explicit opt-out" satisfied — a dedicated flag, or the opt-in design? → A: The opt-out is simply omitting `--choose`. The picker is opt-in (FR-019), so the default is already non-interactive; no separate `--no-tui` flag is introduced.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Browse, find, and run a task interactively (Priority: P1)

A developer working in a repository runs Rune and is presented with a full-screen,
navigable list of the tasks defined in the project's Runefile. They can move
through the list, type to filter it down to the task they want, read a short
description of the highlighted task, and launch it — all without having to first
recall or type the task's exact name.

**Why this priority**: This is the core of the feature and the primary reason a
user would reach for the interactive experience. It replaces the "remember the
exact task name, or run `--list` then re-type it" loop with point-and-run
discovery. Shipping only this story already delivers a usable, valuable product.

**Independent Test**: In a repository with a Runefile containing several tasks,
launch the interactive picker on an interactive terminal, filter to a task by
typing part of its name, select it, and confirm the task runs and the chosen
task's name matches the selection. Fully testable on its own.

**Acceptance Scenarios**:

1. **Given** a Runefile with multiple non-private tasks and an interactive
   terminal, **When** the user launches the interactive picker, **Then** all
   non-private tasks are listed and one is highlighted by default.
2. **Given** the picker is open, **When** the user types characters, **Then**
   the list narrows to tasks whose name or description matches (matched portion
   highlighted), and clearing the filter restores the full list.
3. **Given** a task is highlighted, **When** the user views it, **Then** the
   task's documentation/description is shown alongside the list.
4. **Given** a task is highlighted, **When** the user confirms the selection,
   **Then** that task is executed and its exit status is reported the same way
   as a direct `rune <task>` invocation.
5. **Given** the picker is open, **When** the user cancels/quits without
   selecting, **Then** no task runs and Rune exits with a success status and no
   error.
6. **Given** a Runefile with no runnable (non-private) tasks, **When** the user
   launches the picker, **Then** a clear "no tasks to choose from" message is
   shown and nothing runs.

---

### User Story 2 - Run the selected task with native, full-fidelity output (Priority: P2)

After the developer selects a task, the picker tears itself down, restores the
terminal, and runs the task exactly as a direct `rune <task>` invocation would —
so output streams natively, colors and progress display correctly, interactive
subprocesses keep working, and `Ctrl-C` interrupts the task normally. When the
task finishes, Rune exits with the task's status.

**Why this priority**: Closing the loop from "pick" to "run" is what makes the
picker actually useful, and handing the terminal directly to the task guarantees
fidelity for arbitrary (possibly interactive) tasks. It depends on US1 existing
first, so it ships second.

**Independent Test**: Select a task that emits output over a few seconds; confirm
the picker UI disappears cleanly, the output is identical to running the task
directly, `Ctrl-C` interrupts it, and Rune's exit code equals the task's exit
code.

**Acceptance Scenarios**:

1. **Given** a task is selected, **When** it runs, **Then** the picker UI is torn
   down, the terminal is restored, and the task's output streams natively and
   identically to a direct `rune <task>` invocation.
2. **Given** a selected task is running (the picker already torn down), **When**
   the user presses `Ctrl-C`, **Then** the task is interrupted exactly as it
   would be during a direct invocation and a non-success status is surfaced.
3. **Given** a task has finished (success or failure), **When** it completes,
   **Then** Rune exits with the task's exit code and does not return to the
   picker list.

---

### User Story 3 - Non-interactive usage is completely unchanged (Priority: P1)

A script, CI pipeline, or piped command invokes Rune exactly as before and
receives byte-for-byte the same output, behavior, and exit codes — the
interactive experience never activates and never injects cursor movement,
colors, or prompts into captured output.

**Why this priority**: This is a hard guardrail, not a nice-to-have. Rune is
dogfooded in its own CI and relied on by scripts; any regression here breaks
existing users. It is co-equal P1 with US1: the feature is only acceptable if
this holds.

**Independent Test**: Run representative existing invocations (`rune <task>`,
`rune --list`, `rune --dry-run`, `rune <task> | cat`, and the same under a
CI-like non-TTY environment) and confirm output and exit codes are identical to
the pre-feature baseline (golden comparison).

**Acceptance Scenarios**:

1. **Given** output is piped or redirected (not a terminal), **When** any Rune
   command runs, **Then** the interactive experience does not activate and
   output matches the existing plain behavior exactly.
2. **Given** an explicit task and arguments are supplied, **When** Rune runs,
   **Then** the task runs directly without ever showing the interactive
   picker.
3. **Given** the user does not pass `--choose`, **When** Rune runs on a
   terminal, **Then** the picker does not activate and Rune behaves exactly as
   before (the opt-out is simply omitting `--choose`).
4. **Given** `NO_COLOR` is set or color is disabled, **When** the interactive
   experience runs, **Then** it renders without color/escape styling while
   remaining usable.
5. **Given** a CI environment is detected, **When** Rune runs with no task,
   **Then** the interactive picker does not auto-launch.

---

### Edge Cases

- **No terminal / piped input or output**: the interactive experience must not
  activate; fall back to existing behavior.
- **Terminal too small**: the layout must degrade gracefully (e.g., collapse the
  detail pane) rather than corrupt the display or crash.
- **Very large task lists**: the list must remain navigable and filterable
  without noticeable lag.
- **Filter matches nothing**: show an empty-state message; confirming does
  nothing.
- **Task whose name collides with a built-in command**: selecting it from the
  picker must still run the task (not the built-in).
- **Interrupt signal (Ctrl-C)** while the picker is open vs. while a task is
  running: the first cancels selection; the second interrupts the task. Both
  must leave the terminal in a clean state (cursor restored, no stuck raw mode).
- **Runefile resolution failure or analyzer error**: surface the same
  diagnostics Rune already produces, without entering the interactive UI in a
  broken state.
- **`--list`, `--dry-run`, `--dump`, `--summary` and similar inspection flags**:
  these are non-interactive and must never trigger the picker.

## Requirements *(mandatory)*

### Functional Requirements

#### Interactive browsing & selection

- **FR-001**: The system MUST provide a full-screen interactive task picker that
  lists the non-private tasks defined in the resolved Runefile.
- **FR-002**: Users MUST be able to navigate the task list and change the
  highlighted task using the keyboard.
- **FR-003**: Users MUST be able to filter/search the list incrementally by
  typing, narrowing it to tasks whose name **or** description matches the typed
  text (with the matched portion highlighted), and clear the filter to restore
  the full list.
- **FR-004**: The picker MUST display the highlighted task's documentation /
  description so the user can understand a task before running it.
- **FR-005**: Users MUST be able to confirm a selection to run the highlighted
  task, and MUST be able to cancel/quit the picker without running anything.
- **FR-006**: When a task is selected and run, the system MUST execute it through
  the same task-execution path as a direct `rune <task>` invocation, producing
  the same result and exit status. Any arguments supplied on the command line
  before the picker launched (e.g. `rune --choose -- --watch`) MUST be forwarded
  to the selected task; the picker MUST NOT provide an in-session field for
  entering arguments.
- **FR-007**: When the Runefile defines no non-private tasks, the picker MUST
  show a clear empty-state message and run nothing.

#### Running & output (US2)

- **FR-008**: On selection, the system MUST tear down the picker UI and restore
  the terminal before running the task, then run the selected task with direct
  terminal access so its output streams natively and with full fidelity
  (identical to a direct `rune <task>` invocation — no capture or buffering into
  an embedded pane).
- **FR-009**: A task run from the picker MUST be interruptible via the standard
  `Ctrl-C` signal exactly as during a direct invocation, with the resulting
  non-success status surfaced.
- **FR-010**: The final exit status of a task run from the picker MUST be
  reflected in Rune's overall process exit code, matching direct invocation; the
  picker MUST NOT return to the task list after the task completes.

#### Activation, fallback & compatibility (US3)

- **FR-011**: The interactive picker MUST activate only when standard
  input/output is an interactive terminal; in any non-TTY, piped, redirected, or
  CI context it MUST NOT activate.
- **FR-012**: The opt-out from the interactive picker is *not running with
  `--choose`*. Because the picker is opt-in (FR-019), the default behavior on any
  invocation without `--choose` is already non-interactive; the system MUST NOT
  require a separate opt-out flag, and none is introduced.
- **FR-013**: When an explicit task (and any arguments) is supplied, the system
  MUST run it directly and MUST NOT show the interactive picker.
- **FR-014**: Existing non-interactive commands, output formats, and exit codes
  (including `--list`, `--dry-run`, `--dump`, `--summary`, direct task runs, and
  piped output) MUST remain byte-for-byte unchanged.
- **FR-015**: The interactive experience MUST honor color-disabling conventions
  (e.g., `NO_COLOR`, disabled color) and remain usable without color.
- **FR-016**: On exit — whether by completion, cancellation, error, or interrupt
  — the system MUST restore the terminal to a clean state (cursor visible, no
  residual raw-mode or alternate-screen artifacts).
- **FR-017**: The interactive picker MUST degrade gracefully when the terminal
  is too small to show the full layout, without crashing or corrupting the
  display.
- **FR-018**: A task whose name collides with a built-in command MUST still be
  runnable when selected from the picker.

#### Activation policy & scope

- **FR-019**: The interactive picker MUST be opt-in only. When invoked with no
  task and no inspection flag on an interactive terminal, the system MUST NOT
  auto-launch the picker; bare `rune` behavior is unchanged. The picker activates
  only when the user explicitly requests it via the existing `--choose` flag. No
  new subcommand is introduced for this feature.
- **FR-020**: The new built-in interactive picker MUST fully own interactive task
  selection, replacing the existing external fuzzy-finder integration and the
  minimal numbered-prompt fallback. After this feature there is exactly one
  interactive selection experience.
- **FR-021**: The interactive picker MUST present a modern, visually styled,
  colorful interface by default (consistent theming, clear highlighting of the
  active item and matched filter text), while honoring the color-disabling rules
  in FR-015 so it remains usable without color.

### Key Entities *(include if feature involves data)*

- **Task**: A named unit of work defined in the Runefile. Relevant attributes:
  name, human-readable documentation/description, and visibility (private tasks
  are excluded from the picker).
- **Task list**: The ordered, filterable collection of selectable (non-private)
  tasks derived from the resolved Runefile for the current working directory.
- **Selection**: The user's chosen task plus any pass-through arguments, handed
  to the existing execution path.
- **Run session**: A single in-progress task execution surfaced in the
  interactive view, characterized by its live output stream, in-progress state,
  and final exit status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user who does not remember a task's exact name can locate and
  launch it from the interactive picker in under 10 seconds in a project with
  up to ~50 tasks.
- **SC-002**: 100% of existing non-interactive invocations (direct task runs,
  `--list`, `--dry-run`, `--dump`, `--summary`, and piped/redirected output)
  produce output and exit codes identical to the pre-feature baseline.
- **SC-003**: The interactive picker never activates in non-TTY, piped, or CI
  contexts — verified across all such contexts with zero false activations.
- **SC-004**: After every exit path (completion, cancel, error, interrupt) the
  terminal is left in a clean, usable state in 100% of tested cases.
- **SC-005**: Incremental filtering returns the narrowed list with no perceptible
  lag (updates feel instant) for task lists up to ~500 entries.
- **SC-006**: A task run from the picker yields the same final exit code as the
  equivalent direct `rune <task>` invocation in 100% of tested cases.

## Assumptions

- **CLI framework**: The existing Cobra-based command layer (confirmed in the
  repository) is reused; the interactive picker is wired into it as an additive
  path, not a rewrite of the command surface.
- **Interactive UI technology**: The interactive experience is built with the
  Charm stack named by the requester (Bubble Tea + Bubbles + Lip Gloss). Concrete
  component choices and package boundaries are deferred to the implementation
  plan; this spec stays behavior-focused.
- **Scope anchor**: The task picker (browse → filter → detail → run) is
  the primary deliverable. Watching task output in-session (US2) is included;
  editing the Runefile, editing configuration, or other interactive authoring
  flows are **out of scope** for this feature.
- **Existing behavior is the baseline**: Rune already detects TTY and honors
  `NO_COLOR`, already exposes an interactive picker via `--choose`, and already
  runs tasks through a single execution path. This feature reuses the TTY/color
  detection and the single execution path, and **replaces** the current
  `--choose` implementation (external fuzzy-finder + numbered fallback) with the
  new styled built-in picker (FR-020).
- **Business logic stays separate**: Runefile parsing, analysis, and task
  execution remain independent of the presentation layer; the interactive view
  consumes them through existing interfaces.
- **Testing**: Interactive-view logic is verified via state-transition tests, and
  the unchanged non-interactive paths are protected by Rune's existing golden /
  integration suite. Tests run inside the project's Docker harness per repository
  policy.
- **Single-select**: The picker runs one selected task per session (consistent
  with current `--choose`); multi-select / task chaining from the picker is out
  of scope for this feature.
