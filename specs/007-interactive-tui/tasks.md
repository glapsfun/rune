---

description: "Task list for Interactive Task Picker (TUI)"
---

# Tasks: Interactive Task Picker (TUI)

**Input**: Design documents from `specs/007-interactive-tui/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tui-picker.md, quickstart.md

**Tests**: INCLUDED. The project constitution (Principle VI — Test-First, Multi-Layer
Verification) mandates Red-Green-Refactor, so test tasks precede implementation within
each story. All tests run **inside Docker** (`docker-compose run --rm test ...`), never
on the host.

**Organization**: Tasks are grouped by user story for independent implementation and
testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1 / US2 / US3 (maps to spec.md user stories)
- Exact file paths are included in every task

## Path Conventions

Single-project Go CLI. New presentation package at `internal/tui/`; wiring seam at
`internal/cli/choose.go`. Engine packages are untouched (Constitution Principle IV).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add dependencies and create the new package skeleton

- [X] T001 Add Charm dependencies (`github.com/charmbracelet/bubbletea`, `.../bubbles`, `.../lipgloss`) via `go get`, then `go mod tidy`; verify `CGO_ENABLED=0 go build ./cmd/rune` still succeeds — updates `go.mod`, `go.sum`
- [X] T002 [P] Create the `internal/tui` package with a package doc comment in `internal/tui/doc.go` (states: pure presentation layer, no task execution)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared building blocks every story's `--choose` path depends on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T003 [P] Add an interactive-terminal guard helper (stdin **and** stdout are TTYs, via `mattn/go-isatty`, mirroring `useColor()` in `cmd/rune/root.go`) in `internal/cli/choose.go`
- [X] T004 [P] Define `PickerItem` (fields `Name`, `Desc`, `Doc`) implementing `bubbles/list.Item` — `Title()`=Name, `Description()`=Desc, `FilterValue()` spanning name+description — in `internal/tui/item.go`
- [X] T005 Build the filtered item set in `internal/cli/choose.go`: from `loadModule`, include tasks where `!IsPrivate()` and `osMatches(t, runtime.GOOS)`, mapping each to `tui.PickerItem{Name, firstLine(Doc), Doc}` in Runefile order (depends on T004)

**Checkpoint**: Dependencies present, `tui` package exists, item projection + TTY guard ready

---

## Phase 3: User Story 1 - Browse, find, and run a task interactively (Priority: P1) 🎯 MVP

**Goal**: A styled full-screen picker lists non-private tasks; the user navigates,
incrementally filters over name+description (match highlighted), reads the highlighted
task's docs, and selects one to run.

**Independent Test**: In a repo with several tasks, `rune --choose` on a TTY → filter to
a task by typing → see its description → press Enter → the selected task runs and its
name matches the selection.

### Tests for User Story 1 ⚠️ (write first, must FAIL before implementation)

- [X] T006 [US1] Table-driven `Update`/state-transition tests in `internal/tui/picker_test.go`: navigation (`↑`/`↓`, `j`/`k`) moves highlight; typing narrows over name **and** description; `Esc` clears filter; `Enter` sets `selected` + returns `tea.Quit`; `q`/`Ctrl-C` cancels with empty `selected`; `tea.WindowSizeMsg` updates layout

### Implementation for User Story 1

- [X] T007 [P] [US1] Implement Lip Gloss styles gated on a color flag (plain rendering when `NO_COLOR`/color disabled; highlight active item and matched filter span) in `internal/tui/styles.go`
- [X] T008 [US1] Implement the picker `Model`/`Update`/`View` in `internal/tui/picker.go`: `bubbles/list` + filter input over name+description with match highlight, detail pane showing full `Doc`, navigation, selection capture into `selected`, and `WindowSizeMsg` layout that collapses the detail pane when too small (FR-001/002/003/004/017); makes T006 pass (depends on T004, T007)
- [X] T009 [US1] Wire `internal/cli/choose.go` to run the Bubble Tea program (alt-screen) over the built items, read the final model's `selected`, and run nothing on empty selection / cancel (FR-005) (depends on T005, T008)
- [X] T010 [US1] Empty-list guard in `internal/cli/choose.go`: return the usage error `no tasks to choose from` (exit 2) when no selectable tasks exist; never start the program (FR-007)
- [X] T011 [US1] Static-error-first in `internal/cli/choose.go`: ensure `loadModule` diagnostics render and the picker is NOT started on parse/analyze errors (exit 3) — covers spec Edge Case (Constitution Principle II)

**Checkpoint**: `rune --choose` opens the styled picker, filters, shows details, and runs the chosen task — MVP complete and independently testable

---

## Phase 4: User Story 2 - Run the selected task with native, full-fidelity output (Priority: P2)

**Goal**: On selection, the picker tears down and hands the terminal to the existing
execution path; the task runs natively (output, color, interactive subprocesses,
`Ctrl-C`), and Rune exits with the task's code without returning to the list.

**Independent Test**: Select a task that emits output over a few seconds → picker
disappears cleanly → output identical to a direct run → `Ctrl-C` interrupts it → Rune's
exit code equals the task's exit code.

### Tests for User Story 2 ⚠️ (write first, must FAIL before implementation)

- [X] T012 [US2] Test the selection→execution delegation in `internal/cli/choose_test.go`: given a (injected) selected task name, `chooseAndRun` calls `execute(opts, runefile, append([]string{picked}, args...))`, forwards pass-through `args` (FR-006/Q3), and its returned error maps via `CodeFor` to the task's exit code; cancellation (empty selection) returns nil/exit 0

### Implementation for User Story 2

- [X] T013 [US2] Run the program synchronously with `tea.WithContext(opts.ctx())` and ensure terminal teardown completes **before** invoking `execute()`; treat context cancellation (SIGINT) as cancellation (exit 130) in `internal/cli/choose.go` (FR-008/009, depends on T009)
- [X] T014 [US2] On non-empty selection, delegate to `execute(opts, runefile, append([]string{picked}, args...))` and return its result so the existing pipeline + `CodeFor` mapping apply; do NOT re-open the picker afterward in `internal/cli/choose.go` (FR-006/010, depends on T009)

**Checkpoint**: Selecting a task runs it with full fidelity and the correct exit code; US1 + US2 both work

---

## Phase 5: User Story 3 - Non-interactive usage is completely unchanged (Priority: P1)

**Goal**: Every non-interactive/piped/CI invocation behaves exactly as before; the picker
never activates outside an explicit `--choose` on a TTY; the old `fzf`/numbered picker is
removed.

**Independent Test**: Run `rune --list`, `rune <task>`, `rune <task> | cat`, and
`rune --choose | cat` under a non-TTY environment → outputs and exit codes match the
pre-feature baseline, and `--choose` on a non-TTY errors clearly (exit 2) with no UI bytes.

### Tests for User Story 3 ⚠️ (write first, must FAIL before implementation)

- [X] T015 [P] [US3] Integration test in `internal/cli/choose_test.go`: `--choose` with a non-interactive stream returns `--choose requires an interactive terminal` (exit 2) and emits no terminal-control bytes
- [X] T016 [P] [US3] Golden/integration coverage in `test/` (extend existing suite): `--list`, a direct task run, `--dry-run`, `--dump`, `--summary`, and piped output are byte-for-byte identical to baseline and never open the picker (FR-014)

### Implementation for User Story 3

- [X] T017 [US3] Non-TTY guard in `internal/cli/choose.go`: when `--choose` is given without an interactive terminal, return the usage error (exit 2) before building items or starting the program (uses T003); makes T015 pass
- [X] T018 [US3] Remove the superseded picker code (`pickTask`, `pickWithFzf`, `pickBuiltin`, and the `fzf`/`bufio` plumbing) from `internal/cli/choose.go` so the built-in picker is the single interactive path (FR-020)

**Checkpoint**: One interactive path; non-interactive behavior provably unchanged; all three stories functional

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Docs, gates, and final validation

- [X] T019 [P] Update user-facing `--choose` documentation (modern picker behavior, key bindings, opt-in only, `fzf` removed) under `docs/`, with matching `test/docs` fixtures so the `docs-verify` gate passes
- [X] T020 [P] Verify `NO_COLOR` plain rendering and tiny-terminal graceful degradation against the picker (FR-015/FR-017) — add/extend assertions in `internal/tui/picker_test.go`
- [X] T021 Run the full gate set in Docker: `docker-compose run --rm test go test ./...`, `-race`, `golangci-lint run`, static `CGO_ENABLED=0` build (3 OSes), `golden` no-drift, and `goreleaser release --snapshot` (binary-size check)
- [X] T022 Execute the `quickstart.md` manual interactive validation checklist (steps 1–8) in a real terminal

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately (T001 → enables everything; T002 [P])
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Stories (Phase 3–5)**: All depend on Foundational
  - US1 (P1) is the MVP and should be built first
  - US2 (P2) depends on US1's program-run wiring (T009)
  - US3 (P1) depends only on Foundational (T003) + US1's choose.go scaffold; its removal task (T018) should land after US1/US2 wiring is in place
- **Polish (Phase 6)**: Depends on all desired stories complete

### Within Each User Story

- Tests written first and FAIL before implementation (Constitution VI)
- `internal/tui` model (T008) before choose.go wiring (T009)
- choose.go is a single file: T003, T005, T009, T010, T011, T013, T014, T017, T018 are **sequential** (same file)

### Parallel Opportunities

- T002, T003, T004 are different files → can run in parallel (T005 waits on T004)
- T007 (styles.go) parallel with T006 (picker_test.go) before T008
- T015 and T016 are different files → parallel
- T019 and T020 are different files → parallel

---

## Parallel Example: Foundational

```bash
# Different files, no interdependencies:
Task: "T003 TTY guard helper in internal/cli/choose.go"
Task: "T004 PickerItem type in internal/tui/item.go"
# Then T005 (depends on T004) builds the item set in internal/cli/choose.go
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Phase 1: Setup (T001–T002)
2. Phase 2: Foundational (T003–T005)
3. Phase 3: User Story 1 (T006–T011)
4. **STOP and VALIDATE**: `rune --choose` browses, filters, shows docs, runs the selection
5. Demo MVP

### Incremental Delivery

1. Setup + Foundational → ready
2. US1 → test → demo (MVP: pick-and-run works)
3. US2 → test → demo (native handoff + correct exit codes)
4. US3 → test → demo (fzf removed, non-interactive provably unchanged)

---

## Notes

- [P] = different files, no dependencies on incomplete tasks
- `internal/cli/choose.go` is touched by many tasks → those are intentionally **not** [P]
- The `internal/tui` package must stay free of execution logic so `Update` is a pure,
  table-testable function (no `teatest` dependency)
- All tests run in Docker; never run the suite on the host (global policy + Constitution)
- Commit after each task or logical group; stop at any checkpoint to validate independently
