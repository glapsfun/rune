---

description: "Task list template for feature implementation"
---

# Tasks: Grouped Sections in the Interactive Task Picker

**Input**: Design documents from `/specs/012-picker-task-grouping/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/tui-picker-grouping.md, quickstart.md

**Tests**: Included — the project constitution (Principle VI: Test-First, Multi-Layer Verification) requires Red-Green-Refactor; test tasks are written to fail before their corresponding implementation task.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent implementation and testing of each story.

**Implementation-time design change**: while implementing the Foundational
phase, reading `bubbles/list`'s actual source (`SetItems` re-triggers its own
async fuzzy-filter, `list.go:385-393`) surfaced a real bug in the originally
planned "header as a fake, non-filterable `list.Item`" approach — the filter
would strip a freshly-reinserted header right back out. The design was
revised (with sign-off) to a `sectionDelegate` that draws the group name as a
decorative line above whichever *visible* item starts a new section, derived
at render time from adjacency in `m.list.VisibleItems()`. `PickerItem` gained
a `Section` field instead of `PickerSection` being flattened into fake header
items; `sectionHeaderItem` and the T018/T019 skip-check and filter-recompute
tasks were dropped as unnecessary — since there is no header `list.Item` to
land on, every navigation key (including Home/End/PageUp/PageDown) is
correct by construction. `research.md`, `data-model.md`, and this file are
updated below to match what was actually built.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Exact file paths are included in every task description

## Path Conventions

Single Go project, existing `internal/` layout (Constitution Principle IV — locked package layout). No new packages are introduced; every task below touches `internal/cli` and/or `internal/tui`.

---

## Phase 1: Setup

**Purpose**: Establish a clean baseline before any change, so the "no groups → byte-identical" requirement (FR-004/SC-002) has something concrete to compare against.

- [X] T001 Run `docker-compose run --rm test go test ./internal/cli/... ./internal/tui/...` from the repo root and confirm it passes on the unmodified tree (baseline for later byte-identical comparisons)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The single source of truth for group ordering/membership, plus the new view-layer types both user stories build on.

**⚠️ CRITICAL**: No user story task may begin until this phase is complete.

- [X] T002 [P] Add `internal/cli/run_test.go` with unit tests for a not-yet-existing `visibleTasksByGroup` helper: assert order-by-first-occurrence, correct handling of the ungrouped ("") bucket, and correct collection when the same group name is reused non-contiguously in the Runefile, across a small matrix of fixture Runefiles (research.md R1) — these tests must fail to compile/pass until T003 lands
- [X] T003 Extract the grouping/ordering logic currently inline in `listTasks` (`internal/cli/run.go`) into a standalone helper `visibleTasksByGroup(f *ast.File) (order []string, groups map[string][]*ast.Task)` that applies the existing private/OS-visibility filter; update `listTasks` to call it, with zero change to `--list`'s output (makes T002 pass; depends on T002 existing)
- [X] T004 [P] Add a `Section string` field to `PickerItem` and a `PickerSection` type (`Name string`, `Items []PickerItem`) to `internal/tui/item.go` per `data-model.md` (revised: `Section` replaces the fake-header-item design)
- [X] T005 [P] Add `internal/tui/delegate.go` with a `sectionDelegate` wrapping `list.DefaultDelegate`: `Height()` returns the embedded delegate's height plus one when grouped (else unchanged, preserving FR-004); `Render` writes a decorative line — the group name when the item is the first *visible* item of its section (via adjacency in `m.VisibleItems()`), blank otherwise, never for the ungrouped section — then delegates the row itself to the embedded default delegate unchanged (revised design; supersedes the originally planned `sectionHeaderItem` `list.Item`)
- [X] T006 [P] Add a `Header` style to `internal/tui/styles.go`, following the existing color/no-color `newStyles` convention already used by `Title`/`Detail`/`Help` (reuses `--list`'s accent color, `internal/style`'s `colorAccent` = `"170"`)

**Checkpoint**: Shared grouping source of truth and the picker's new building blocks (`PickerSection`, `sectionDelegate`, `Header` style) exist, but nothing consumes them yet.

---

## Phase 3: User Story 1 - Scan a large task set by category (Priority: P1) 🎯 MVP

**Goal**: `--choose` renders tasks organized into labeled sections that match `--list`'s existing `group(...)` order and membership; Runefiles with no groups render exactly as before.

**Independent Test**: Open `--choose` against a Runefile with 3+ groups plus ungrouped tasks — verify section headers, order, and membership match `rune --list` for the same file. Separately, against a Runefile with no `group(...)` attributes at all, verify the picker is unchanged from before this feature.

### Tests for User Story 1

- [X] T007 [P] [US1] Extend `internal/cli/choose_test.go` with a test asserting `pickerItems` partitions tasks into `[]tui.PickerSection` whose order and membership match `visibleTasksByGroup`'s output, for a Runefile mixing multiple `group(...)` tasks with ungrouped tasks (FR-001, FR-002, FR-003) — `TestPickerItems_MatchesListGrouping`
- [X] T008 [P] [US1] Add a test in `internal/cli/choose_test.go` asserting that for a Runefile with **no** `group(...)` attributes, `pickerItems` returns exactly one section (`Name == ""`) containing every visible task in file order — the pre-feature-equivalent shape (FR-004, SC-002) — `TestPickerItems_NoGroupsYieldsOneUnnamedSection`
- [X] T009 [P] [US1] Add tests asserting header placement (revised for the `sectionDelegate` design): `internal/tui/delegate_test.go`'s `TestSectionDelegate_HeaderOnlyAtSectionStart` (header exactly at each section's first item, including a mix of named sections plus one ungrouped section — spec.md US1 Acceptance Scenario 1), `TestSectionDelegate_UngroupedNeverGetsAHeader`, `TestSectionDelegate_NotGroupedMatchesDefaultDelegate` (byte-identical to `list.DefaultDelegate` when ungrouped — SC-002), and `TestSectionDelegate_GroupedAddsExactlyOneLine`; plus `internal/tui/picker_test.go`'s `TestNewFlattensSectionsTaggingEachItem` (FR-001, FR-003, SC-002)

### Implementation for User Story 1

- [X] T010 [US1] Change `pickerItems` in `internal/cli/choose.go` to call `visibleTasksByGroup` (T003) and return `[]tui.PickerSection` instead of a flat `[]tui.PickerItem`, preserving existing private/OS-visibility filtering; migrated the existing `TestPickerItems_FiltersPrivateAndKeepsBuiltinCollision` in `internal/cli/choose_test.go` to flatten the returned sections before asserting on names/descriptions (depends on T003; makes T007/T008 pass)
- [X] T011 [US1] Change `tui.New` in `internal/tui/picker.go` to accept `[]PickerSection`, flatten it into a `[]PickerItem` tagged with each item's section name, and construct the list with `newSectionDelegate(styles, grouped)` in place of `list.NewDefaultDelegate()`; migrated `internal/tui/picker_test.go`'s `testItems()`/`newSizedModel` helpers (and every test that calls them) to build a single-section `[]PickerSection` (depends on T004, T005; makes T009 pass)
- [X] T012 [US1] Update `chooseAndRun`/`runPicker` in `internal/cli/choose.go` to pass `pickerItems(mod)`'s new `[]tui.PickerSection` result directly into `tui.New` (depends on T010, T011)
- [X] T013 [US1] Wire `newSectionDelegate` into `tui.New`'s list construction (folded into T011 above — the delegate itself, including its `Header`-styled rendering, was built in T005)

**Checkpoint**: User Story 1 is fully functional and independently testable — grouped sections render with correct order/membership; the no-groups path is unchanged; T002 and T007–T009 all pass.

---

## Phase 4: User Story 2 - Navigate and filter without breaking on section boundaries (Priority: P2)

**Goal**: Keyboard navigation never highlights a header row, and filtering hides any section whose tasks are all filtered out — without changing how filtering or selection/execution work today.

**Independent Test**: In a grouped picker, press the down arrow from the first to the last task and confirm the highlight only ever rests on a task; type a filter that matches tasks in only one section and confirm the other sections' headers disappear with their tasks; select and confirm a task from any section and confirm it runs normally.

### Tests for User Story 2

- [X] T014 [P] [US2] Add a test driving `Update` with a scripted "down" sequence, asserting `SelectedItem()` is always a `PickerItem` at every step, including at every section boundary (FR-005) — `TestSelectionAlwaysAPickerItem` (`internal/tui/picker_test.go`)
- [X] T015 [P] [US2] Extend the same test to cover "up" and every other cursor-moving key the underlying list binds — Home, End, PageUp, PageDown (FR-005) — folded into `TestSelectionAlwaysAPickerItem`: with the `sectionDelegate` design there is no header `list.Item` to land on, so this is a structural regression guard rather than a skip-check test
- [X] T016 [P] [US2] Add a test that applies a filter matching tasks in only one of several sections and asserts no other section's tasks remain visible — FR-006's "header hidden along with its section" follows automatically, since a header is only ever drawn adjacent to a *visible* item — `TestFilterNarrowsToMatchingSectionOnly` (`internal/tui/picker_test.go`)
- [X] T017 [P] [US2] Add a test selecting a task belonging to a non-first, non-ungrouped section and confirming it runs exactly like any other task by name (FR-008) — `TestPickerItems_NonFirstSectionTaskIsSelectableAndRuns` (`internal/cli/choose_test.go`)

### Implementation for User Story 2

- [X] T018 ~~Add a post-move header-skip check to `Model.Update`~~ — **superseded**: not needed. Headers are never `list.Item`s under the revised design (T005), so there is nothing for any key — including Home/End/PageUp/PageDown, the exact gap flagged in `/speckit-analyze` — to land on and skip.
- [X] T019 ~~Add filter-driven section recompute to `Model`~~ — **superseded**: not needed. `sectionDelegate.Render` derives the header purely from `m.VisibleItems()` at render time, which already reflects whatever the built-in filter currently matches — there is no separate item sequence to rebuild, and no risk of the `SetItems`-retriggers-async-filter bug the original design had.

**Checkpoint**: User Story 2 is complete — navigation and filtering behave correctly around section boundaries; T014–T017 pass; User Story 1's behavior is unaffected.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and full-suite verification per the project's quality gates.

- [X] T020 [P] Update `docs/cli.md`'s `--choose` description to note that tasks are grouped into the same sections `--list` shows whenever the Runefile defines `group(...)` attributes
- [X] T021 Manually run the quickstart scenarios against a locally built `rune` binary — `--list` grouping confirmed correct (order/membership matches `visibleTasksByGroup`'s tests exactly); this also caught and fixed a syntax bug in `quickstart.md`'s own example Runefile (used a brace-based task syntax Rune doesn't have — corrected to the real colon-based syntax). Driving the actual interactive `--choose` session through a real pty was attempted (`script`) but the sandbox couldn't reliably deliver keystrokes to the raw-mode TUI (two independent hangs) — this could not be visually confirmed end-to-end; every behavior in the quickstart scenarios is otherwise covered deterministically by the 32 passing unit/model tests (T007–T017)
- [X] T022 Ran `docker-compose run --rm test go test ./...` (all packages pass), `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./internal/cli/... ./internal/tui/...` (clean), and `rune lint` (`go vet` + `golangci-lint run`, 0 issues after a `gofumpt` formatting fix to `picker_test.go`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — run first.
- **Foundational (Phase 2)**: Depends on Setup. BLOCKS both user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on User Story 2.
- **User Story 2 (Phase 4)**: Depends on Foundational **and** on User Story 1's implementation tasks (T011 builds the section-tagged items and delegate that US2's tests exercise) — it cannot be meaningfully tested until sections render. It does not require any US1 test task to have been written, only the US1 implementation tasks to have landed.
- **Polish (Phase 5)**: Depends on both user stories being complete.

### Within Each User Story

- Tests are written first and must fail before their corresponding implementation task lands (Constitution Principle VI).
- Within US1: section-building (T010) and list-flattening/delegate-wiring (T011) came before the call-site wiring (T012); T013 folded into T011 once the revised design made a separate render-wiring step unnecessary.
- Within US2: both originally-planned implementation tasks (T018, T019) turned out to be unnecessary under the revised design — see the design-change note above.

### Parallel Opportunities

- T002, T004, T005, T006 (Phase 2) touch different files/concerns and ran in parallel.
- T007, T008, T009 (US1 tests) touch different test files/cases and ran in parallel.
- T014, T015, T016, T017 (US2 tests) touch different test cases and ran in parallel.

---

## Parallel Example: Foundational Phase

```bash
Task: "Add internal/cli/run_test.go with grouping-order unit tests (T002)"
Task: "Add Section field + PickerSection type to internal/tui/item.go (T004)"
Task: "Add sectionDelegate to internal/tui/delegate.go (T005)"
Task: "Add Header style to internal/tui/styles.go (T006)"
```

## Parallel Example: User Story 1 Tests

```bash
Task: "Assert pickerItems grouping matches visibleTasksByGroup (T007)"
Task: "Assert no-groups pickerItems shape is pre-feature-equivalent (T008)"
Task: "Assert sectionDelegate header placement for single/mixed/multi-section input (T009)"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (blocks everything else).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run `quickstart.md` Scenarios 1 and 2 against a local build.
5. This alone delivers the feature's core value — sectioned browsing — and is demoable on its own.

### Incremental Delivery

1. Setup + Foundational → shared grouping helper and picker building blocks ready.
2. Add User Story 1 → validate independently (Scenarios 1–2) → this is the MVP.
3. Add User Story 2 → validate independently (Scenarios 3–4) → full feature complete.
4. Polish (docs, full quickstart pass, Docker test/lint gates) → ship.

### Notes

- [P] tasks touch different files/cases with no ordering dependency between them.
- Commit after each task or logical group, per this repo's normal workflow.
- Verify each test task's test fails before landing its paired implementation task (Red-Green-Refactor).
- Avoid: reintroducing a second, independently-maintained copy of `--list`'s grouping rule anywhere in `internal/tui` or `internal/cli/choose.go` — always go through `visibleTasksByGroup` (T003).
