---
description: "Task list for Styled CLI Output & Friendlier Help"
---

# Tasks: Styled CLI Output & Friendlier Help

**Input**: Design documents from `specs/008-styled-cli-output/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included — the constitution mandates test-first (Principle VI) and spec
FR-022 explicitly requires styled-vs-plain tests. Write tests first; they must FAIL
before the implementation task in the same story.

**Organization**: Tasks grouped by user story. The shared theme + per-stream color
decision are **Foundational** (block every styling story).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on incomplete tasks)
- **[Story]**: US1–US6 from spec.md (Setup/Foundational/Polish have no story label)
- All Go commands run **inside Docker**: `docker-compose run --rm test go test ./...`

## Path Conventions

Single-project Go CLI. Engine packages locked (Principle IV); the only new package
is the leaf `internal/style`. Paths are repo-relative.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the invariance baseline before any styling lands.

- [X] T001 Confirm required deps are present in `go.mod` (no new ones): `github.com/charmbracelet/lipgloss`, `github.com/fatih/color`, `github.com/mattn/go-isatty`; record the resolved palette codes (170/245/241 + red/yellow/green) decision from research.md D7 as a comment anchor in the upcoming `internal/style/style.go`.
- [X] T002 Establish the pre-feature plain-output baseline: run `docker-compose run --rm test go test ./...` and confirm the current golden/integration suite is green, so any post-change byte drift is attributable to this feature.

**Checkpoint**: Baseline green; ready to build the shared foundation.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The single semantic theme and the per-stream color decision. Every
user story depends on this phase.

**⚠️ CRITICAL**: No styling story (US1, US3, US5, US6) can begin until this completes.

- [X] T003 [P] Write failing unit tests for the theme in `internal/style/style_test.go`: assert that with `New(false, w)` every role's `Render(s) == s` (no ANSI, byte-identical) and with `New(true, w)` output contains SGR escapes but identical *visible* text and width (FR-003, FR-012).
- [X] T004 Create the leaf package `internal/style/style.go`: define the palette once and `type Theme` with roles `Error, Warning, Success, TaskName, Heading, Muted, Locator, Caret`, plus `func New(enabled bool, w io.Writer) Theme`. Build styles from an explicit `lipgloss.NewRenderer(w)` with forced profile (ANSI256 when enabled, Ascii when disabled) per research.md D2; disabled ⇒ all roles are zero/plain styles (mirror `internal/tui/styles.go`). Make T003 pass.
- [X] T005 [P] Write failing unit tests for color resolution in `cmd/rune/root_test.go` (or `color_test.go`): table-drive the precedence matrix from `contracts/color-flag.md` — `never`→off, `always`→on, `NO_COLOR`→off (auto only), else `isatty(stream)`; assert per-stream independence and that `FORCE_COLOR`/`CLICOLOR`/`CLICOLOR_FORCE` are ignored.
- [X] T006 Add the global `--color` string flag (default `auto`) in `cmd/rune/root.go`; parse/validate into a `ColorMode` (auto|always|never), erroring with a clear message + non-zero exit on any other value (FR-006, FR-009).
- [X] T007 Replace `useColor()` with a per-stream `resolveColor(mode, stream)` in `cmd/rune/root.go` implementing the precedence in `contracts/color-flag.md`; set `fatih/color.NoColor` consistently so forced color survives a pipe. Make T005 pass.
- [X] T008 In `internal/cli/dispatch.go`, replace the single `Options.Color` field with `ColorStdout` and `ColorStderr` bools (resolved against `cmd.OutOrStdout()` / `cmd.ErrOrStderr()` in `PersistentPreRunE`); add helpers to build a stdout/stderr `style.Theme`. Update the existing diagnostic call site `internal/cli/run.go` to use the stderr theme (temporary shim is fine until T024).

**Checkpoint**: Theme + per-stream decision exist and are unit-tested; styling stories can begin in parallel.

---

## Phase 3: User Story 1 - Scannable task list (Priority: P1) 🎯 MVP

**Goal**: `rune --list` emphasizes task names, styles group headings, dims docs — on a color stdout — with columns and bytes unchanged when plain.

**Independent Test**: `rune --list` on a forced-color stream shows distinct name/heading/doc roles; `rune --list | cat` is byte-identical to today with the `#` column aligned.

- [X] T009 [US1] Write a failing styled-list test in `test/integration/us1_test.go` (or a new `cli_list_test.go`): with `--color=always` piped, assert ANSI is present around task names/headings and that stripping ANSI yields the exact pre-feature plain bytes (proves zero-width emphasis, FR-013/SC-002).
- [X] T010 [US1] Implement styling in `listTasks` in `internal/cli/run.go`: apply `Heading` to `Available tasks:` and `[group]`, `TaskName` to names, `Muted` to `# doc`; compute right-padding from the **visible rune width** (`utf8.RuneCountInString`) instead of `%-*s` on the colorized value, keeping the plain branch exactly as today. Make T009 pass.

**Checkpoint**: `--list` is styled on a TTY and provably byte-identical when plain — MVP shippable.

---

## Phase 4: User Story 2 - Scripts and CI see unchanged output (Priority: P1)

**Goal**: A reusable guardrail proving every styled surface is byte-for-byte plain under the three "off" triggers, and that `--color=always` is the only path that emits ANSI through a pipe.

**Independent Test**: The invariance matrix test passes for `--list` and diagnostics now (extended to status/echo when US3 lands), under piped / `NO_COLOR=1` / `--color=never`.

- [X] T011 [US2] Add a shared invariance helper in `test/integration/harness_test.go`: given a command + Runefile, capture stdout/stderr under (a) piped, (b) `NO_COLOR=1`, (c) `--color=never`, and assert all three are byte-identical to each other and contain no ESC (`0x1b`) byte.
- [X] T012 [US2] Add the invariance matrix test (`test/integration/us2_test.go`): apply the T011 helper to `rune --list` and to a diagnostic-producing Runefile; add a forced-color case asserting `--color=always | cat` *does* contain ESC for `--list` (FR-010, FR-022, SC-001, SC-004). (Re-run after US3/US5 to cover status/echo/diagnostics.)

**Checkpoint**: The byte-invariance guardrail exists and is wired to the surfaces shipped so far.

---

## Phase 5: User Story 3 - Legible run output (Priority: P2)

**Goal**: Status labels colored by meaning; command echo and cache notices dimmed; warnings in the warning role — text and stream unchanged.

**Independent Test**: A run that executes one task and serves another from cache shows `running:` active, `cached:` muted, echo dimmed (forced color); piped run is byte-identical to today.

- [X] T013 [US3] Write a failing test in `test/integration/us3_test.go`: with `--color=always`, assert the `running:`/`cached:`/`would run:`/`would skip (cached):` labels and the cache-write warning carry the expected roles, and that ANSI-stripped output equals the pre-feature plain bytes; pipe the same run and assert byte-identical (FR-014, FR-015).
- [X] T014 [US3] Thread the stderr `style.Theme` into the engine in `internal/cli/run.go` and style the status/cache lines (`running:` → Success/active, `cached:`/`would run:`/`would skip (cached):` → Muted, cache-write `warning:` → Warning). Make the label assertions in T013 pass.
- [X] T015 [US3] Add an optional `EchoStyle` (muted) to `shell.Options` in `internal/runtime/shell/shell.go` and apply it to the echoed command line **after** the `NoEcho`/`Quiet` suppression check, so suppressed lines never appear (FR-016); wire the stderr theme from `run.go`. (Interp executor echoes no per-line commands — out of scope.) Make T013's echo assertion pass.

**Checkpoint**: Normal runs are legible on a TTY and unchanged when piped. Re-run T012.

---

## Phase 6: User Story 4 - Explicit color control (Priority: P2)

**Goal**: `--color=auto|always|never` behaves per the contract across both streams; invalid values error cleanly.

**Independent Test**: `--color=always | cat` colors; `--color=never` on a TTY does not; `--color=sometimes` exits non-zero without running.

- [X] T016 [US4] Write failing acceptance tests in `test/integration/us4_test.go`: `--color=always` through a pipe emits ANSI on `--list` (stdout) and on a diagnostic (stderr); `--color=never` emits none on a forced TTY-like stream; `--color=sometimes` produces a clear error, non-zero exit, and no task output (FR-007, FR-008, FR-009).
- [X] T017 [US4] Verify/finish per-stream independence: confirm a stdout-only pipe leaves stderr's `auto` decision intact (mixed-redirection edge case) in `cmd/rune/root.go` resolution; make T016 pass. (Mechanism built in Phase 2; this hardens and proves it.)

**Checkpoint**: Color control fully honored and tested across streams.

---

## Phase 7: User Story 5 - Sharper diagnostics (Priority: P3)

**Goal**: Diagnostics draw their colors from the shared theme and add locator/caret emphasis, with the plain golden byte-identical and caret columns unchanged.

**Independent Test**: Colored diagnostic emphasizes severity/locator/caret; plain rendering matches `testdata/diag/render.golden` exactly and the caret sits under the same columns.

- [X] T018 [US5] Update `internal/diag/render_test.go`: keep the existing color-off golden assertion (must stay byte-identical, FR-018) and add a color-on case (theme enabled) asserting severity/caret carry roles while the caret count/position is unchanged (SC-003). Write before refactor; it should fail to compile against the new signature.
- [X] T019 [US5] Refactor `internal/diag/render.go` (`Render`/`RenderAll`/`caretUnderline`) to take a `style.Theme` instead of `useColor bool`, applying `Error`/`Warning` to severity, `Caret` to the caret run, and optional `Locator` emphasis to `file:line:col`; the disabled-theme path must emit the exact current plain bytes. Update the call site in `internal/cli/run.go` (stderr theme). Make T018 pass.

**Checkpoint**: Diagnostics single-sourced through the theme; plain golden unchanged. Re-run T012.

---

## Phase 8: User Story 6 - Friendlier help with examples (Priority: P3)

**Goal**: Redesigned `--help` with grouped sections + worked examples; colored headings on a TTY, ANSI-free and informative when piped. The new plain help is the deliberately-reviewed baseline.

**Independent Test**: `rune --help` shows Usage/Commands/Flags/Examples with an example per workflow; `rune --help | cat` is ANSI-free, and `--color=always --help` stripped equals the plain help.

- [X] T020 [US6] Update `test/integration/cli_help_test.go`: assert the redesigned plain `--help` contains the grouped section headings and ≥1 worked example for each of run-a-task, `--list`, `--choose`, `serve` (FR-019, FR-020, SC-006), and contains no ESC byte when piped; add a `--color=always` case asserting headings carry the `Heading` role (verified via strip-invariance).
- [X] T021 [US6] Implement the help redesign in a new `cmd/rune/help.go`: a custom root `HelpFunc` (subcommands keep Cobra's default) with grouped sections + examples and a heading colorizer gated on the stdout color decision; wire via `applyHelp(root)` in `cmd/rune/main.go`. Make T020 pass.
- [X] T022 [US6] Confirm the new `--help` is the reviewed baseline. No golden file is introduced — `--help` is verified by section/example substrings plus strip-invariance (styled-stripped == plain); the existing `serve --help` test stays green (subcommands untouched).

**Checkpoint**: Help is friendly and color-aware; its new plain form is the reviewed baseline.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, docs, and the full guardrail sweep.

- [X] T023 [P] Run the complete invariance matrix (T011/T012 helper) across **all** styled surfaces (`--list`, status/echo/cache, diagnostics) under piped / `NO_COLOR` / `--color=never`, plus `--color=always` ANSI-present, confirming SC-001/SC-004/SC-005 (exit codes unchanged).
- [X] T024 Remove any temporary shim from T008; confirm no surface re-implements color detection and the palette lives only in `internal/style` (SC-007) — grep for stray `fatih/color`/`isatty` usage outside `cmd/rune` and `internal/style`.
- [X] T025 [P] Update user-facing docs (README / docs CLI reference) for the `--color` flag and the redesigned help; run `docker-compose run --rm docs test` (docs-verify gate) so examples/links pass.
- [X] T026 [P] Run `docker-compose run --rm test golangci-lint run` (zero issues, gofumpt/goimports clean) and the static `CGO_ENABLED=0` build on the three OSes via the `build` gate.
- [X] T027 Execute `specs/008-styled-cli-output/quickstart.md` end-to-end against a built binary to validate all acceptance scenarios.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: no dependencies.
- **Foundational (Phase 2)**: depends on Setup; **BLOCKS** US1, US3, US5, US6.
- **User Stories (Phases 3–8)**: all depend on Foundational.
  - US1 (P1) and the US2 guardrail come first; US2's full coverage extends as US3/US5 land.
  - US4 (P2) mechanism is built in Foundational; its phase is acceptance hardening.
- **Polish (Phase 9)**: depends on all desired stories.

### User Story Dependencies

- **US1 (P1)**: after Foundational. Independent.
- **US2 (P1)**: after Foundational; its matrix references US1/US3/US5 surfaces as they ship (re-run T012/T023).
- **US3 (P2)**: after Foundational. Independent of US1.
- **US4 (P2)**: after Foundational (flag/decision already built there).
- **US5 (P3)**: after Foundational. Touches only `internal/diag` + its call site.
- **US6 (P3)**: after Foundational (needs `ColorStdout`). Touches only `cmd/rune`.

### Within Each Story

- Test task precedes implementation (must fail first).
- Foundational `internal/style` (T004) before any role usage.
- `Options` per-stream fields (T008) before surface styling.

### Parallel Opportunities

- T003 and T005 (different test files) can run in parallel.
- Once Foundational completes, US1, US3, US5, US6 touch disjoint files
  (`run.go` listTasks vs `run.go` engine lines — same file, so US1↔US3 serialize;
  US5 = `internal/diag`, US6 = `cmd/rune` — fully parallel with each other).
- Polish T023/T025/T026 are independent ([P]).

---

## Parallel Example: after Foundational

```bash
# US5 and US6 touch disjoint packages — run in parallel:
Task: "T018/T019 diagnostics theme refactor in internal/diag/"
Task: "T020/T021 help redesign in cmd/rune/help.go"
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Phase 1 Setup → 2. Phase 2 Foundational (theme + per-stream `--color`) →
3. Phase 3 US1 (`--list`) → **STOP & VALIDATE** styled-vs-plain → demo.

### Incremental Delivery

Foundational → US1 (+US2 guardrail) → US3 → US4 → US5 → US6, validating byte
invariance (re-run T012/T023) after each surface lands. Each story adds value
without changing any plain output (except the intentional US6 help baseline).

---

## Notes

- [P] = different files, no dependency. US1 and US3 both edit `internal/cli/run.go`, so they serialize.
- Constitution Principle VI: write each test task first and confirm it fails before implementing.
- Goldens are regenerated deliberately (T022) and never hand-edited to pass.
- The `--choose` picker (`internal/tui`) is out of scope and untouched.
