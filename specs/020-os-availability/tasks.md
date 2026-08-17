# Tasks: OS Availability Enforcement

**Input**: Design documents from `/specs/020-os-availability/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: REQUIRED — the constitution (Principle VI) mandates test-first
Red-Green-Refactor. Every story writes its failing tests before the
implementation that turns them green. All Go tests run inside Docker:
`docker-compose run --rm test go test ./...` — never on the host.

**Organization**: Tasks are grouped by user story (spec.md priorities) so
each story is an independently testable increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1 (MCP hiding), US2 (direct-invoke error), US3 (dep skip), US4 (dump field)

## Phase 1: Setup

**Purpose**: Confirm a green baseline so every later red test is
attributable to this feature.

- [X] T001 Run the full suite in Docker (`docker-compose run --rm test go test ./...`) and `golangci-lint run` on branch `020-os-availability`; both must be clean before any change

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared availability predicate (FR-001/FR-007) and the
injectable host-OS plumbing every story consumes.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 [P] RED: add table tests for `(*Task).AvailableOn` and `(*Task).OSFilters` in `internal/ast/availability_test.go` — cover the full truth table from data-model.md (no attrs; each single OS on linux/darwin/windows; `linux,windows` OR; `unix,windows` = everywhere; attrs split across multiple lines; OSFilters source order and empty case). Package currently has no tests; file will not compile until T003 (that is the red state)
- [X] T003 GREEN: implement `AvailableOn(goos string) bool` and `OSFilters() []string` on `*Task` in `internal/ast/ast.go`, semantics copied verbatim from `osMatches` (`internal/cli/run.go:729`); T002 passes
- [X] T004 Switch existing call sites to the new method and delete `osMatches`: `hasVisibleTasks` and `visibleTasksByGroup` in `internal/cli/run.go`, completion filter in `internal/cli/complete.go`; full suite stays green (SC-005 regression: `--list`, picker, completion behavior unchanged)
- [X] T005 Plumb the injectable host OS: add a `goos string` field to the CLI `engine` struct (`internal/cli/run.go`) and to `mcpAdapter` (`internal/cli/mcp.go`), defaulted to `runtime.GOOS` at every construction site; no behavior change yet, suite green

**Checkpoint**: One authoritative predicate exists, is fully unit-tested,
and both execution engines carry a test-injectable host OS.

---

## Phase 3: User Story 1 — AI Agents Never See Foreign-OS Tasks (Priority: P1) 🎯 MVP

**Goal**: The MCP tool list (stdio `rune mcp` and HTTP `rune serve`)
exposes only non-private tasks available on the host OS; the authz
allowlist reasons over the same filtered set (FR-002).

**Independent Test**: Serve a mixed-OS Runefile over MCP and confirm the
tool list contains host-OS and unrestricted tasks only (spec Story 1;
quickstart Scenario 5).

### Tests for User Story 1 (write first, must FAIL)

- [X] T006 [P] [US1] RED: in `internal/cli/mcp_test.go` (create or extend the existing MCP adapter test file), build a `mcpAdapter` over a fixture with `[windows]`, `[linux]`, `[linux, windows]`, unrestricted, and private tasks; with `goos: "linux"` assert `Tasks()` omits the `[windows]` task, includes the `[linux]`, OR-combined, and unrestricted tasks, and still omits private (Story 1 scenarios 1–3). Both transports (`rune mcp`, `rune serve`) build this same adapter, so adapter-level assertions cover both — note this in the test comment

### Implementation for User Story 1

- [X] T007 [US1] GREEN: in `mcpAdapter.Tasks()` (`internal/cli/mcp.go`), skip tasks failing `t.AvailableOn(a.goos)` alongside the existing `IsPrivate()` skip; T006 passes
- [X] T008 [US1] Assert the filtered set is what tool registration AND authz consume, over the T006 fixture: (a) build `mcpserver.New(adapter, ...)` and assert no tool is registered for the OS-mismatched task; (b) exercise the authz layer (`mcpserver/authz.go` — allowlist/destructive gating built from `Engine.Tasks()`) directly and assert the mismatched task can be neither listed nor allowlisted (Story 1 scenario 4)

**Checkpoint**: An agent connected on this host cannot see a
platform-incompatible task. MVP delivered.

---

## Phase 4: User Story 2 — Foreign-OS Tasks Cannot Be Run (Priority: P2)

**Goal**: Direct invocation of an OS-mismatched task fails before any
execution with a `ValidationError` (exit 3) naming task, required OS(es),
and host OS; an MCP call naming such a task gets the same refusal (FR-003).

**Independent Test**: `rune setup-win` on a non-Windows host exits 3 with
the availability message and executes nothing (quickstart Scenario 2).

### Tests for User Story 2 (write first, must FAIL)

- [X] T009 [P] [US2] RED: CLI-level tests in `internal/cli/run_test.go` (or the existing file covering `resolveRoots`/`execute`) with injected `goos`: (a) invoking a mismatched task exits `ExitValidation` (3) with a message containing the task name, its required OS list, and the host OS in **attribute vocabulary** (`macos`, never `darwin` — contract availability.md); (b) `rune ok-task bad-task` executes nothing — not even `ok-task` (Story 2 scenario 3); (c) `--dry-run` and `--summary` of a mismatched task also fail (research D3); (d) a `[unix]` task runs on darwin (Story 2 scenario 2)
- [X] T010 [P] [US2] RED: in `internal/cli/mcp_test.go`, `Call()` on an OS-mismatched task name returns the availability error (not "unknown task") and schedules nothing (contract mcp-surface.md defense-in-depth)
- [X] T011 [P] [US2] RED: binary-level integration test in `test/integration` (constitution VI layer — real binary, real GOOS, no injection): a fixture Runefile with a task restricted to an OS that is never the CI host's counterpart is invoked directly; assert stderr carries the availability message, exit code is 3, and nothing executed. Use an OS-complement fixture (e.g. a `[windows]` task asserted on linux/darwin hosts and a `[unix]` task asserted on windows hosts) so the test is meaningful on all three CI platforms

### Implementation for User Story 2

- [X] T012 [US2] GREEN: gate roots in `resolveRoots` (`internal/cli/run.go`): for each root task failing `AvailableOn(e.goos)`, return a `ValidationError` shaped `task "NAME" is not available on HOST (requires F1 or F2)` using `OSFilters()`, where HOST is rendered via a small display helper mapping GOOS→attribute vocabulary (`darwin`→`macos`, else as-is; contract availability.md); T009 and T011 pass
- [X] T013 [US2] GREEN: re-check availability at the top of `mcpAdapter.Call()` (`internal/cli/mcp.go`) with the same message; keep `unknown task:` for genuinely unknown names; T010 passes

**Checkpoint**: Nothing — human, script, or agent — can execute a task on
a platform it excludes.

---

## Phase 5: User Story 3 — Cross-Platform Dispatch Through Dependencies (Priority: P3)

**Goal**: OS-mismatched dependencies and post-hooks are silently skipped
during graph resolution, enabling one dispatcher task with per-OS deps
(FR-004).

**Independent Test**: A `setup` task depending on `[unix]`/`[windows]`
halves runs exactly the matching half plus its own body (quickstart
Scenario 3).

### Tests for User Story 3 (write first, must FAIL)

- [X] T014 [P] [US3] RED: scheduler tests in `internal/runtime/scheduler/scheduler_test.go` using a fake engine whose `Available` rejects chosen tasks: (a) serial dep skipped, remaining deps keep order; (b) `[parallel]` dep skipped; (c) post-hook skipped (Story 3 scenario 3); (d) all deps skipped ⇒ body still executes (scenario 2); (e) skipped target never reaches `Execute` and leaves no memo entry. Existing fake engines gain an `Available` method to compile — that compile failure plus the new assertions are the red state
- [X] T015 [P] [US3] RED: CLI-level dispatch test in `internal/cli/run_test.go` with injected `goos`: dispatcher task with `[unix]` + `[windows]` deps runs only the matching dep then its own body, exit 0, no skip notice in default output (spec Assumptions: silent skip)

### Implementation for User Story 3

- [X] T016 [US3] GREEN: add `Available(task *ast.Task) bool` to the `Engine` interface in `internal/runtime/scheduler/scheduler.go` and skip in `runDep` after `ResolveDep` (return nil, no memo entry, chain untouched — research D2); update the package doc comment; T014 passes
- [X] T017 [US3] GREEN: implement `Available` on the CLI `engine` (`internal/cli/run.go`) as `task.AvailableOn(e.goos)`; T015 passes and the same behavior flows to MCP `Call` (shared scheduler)

**Checkpoint**: The dispatch pattern works; direct-invoke error (US2) and
dep skip coexist for the same task (spec Edge Cases).

---

## Phase 6: User Story 4 — Machine-Readable Availability (Priority: P4)

**Goal**: `rune --dump --format json` reports a computed `available`
boolean per task, all tasks still listed (FR-005).

**Independent Test**: Dump a mixed-OS Runefile and check each task's
`available` against the host (quickstart Scenario 4).

### Tests for User Story 4 (write first, must FAIL)

- [ ] T018 [P] [US4] RED: dump tests in `internal/cli/dump_test.go` (create or extend): with a goos parameter, a `[windows]` task dumps `"available": false` on linux and `true` on windows, unrestricted tasks always `true`, unavailable and private tasks remain present, and the raw OS names stay in `attributes` (contract dump-schema.md — verify the actual attribute rendering and align the contract's example if it differs)

### Implementation for User Story 4

- [ ] T019 [US4] GREEN: add `Available bool` with JSON tag `available` (no omitempty) to `taskDTO` in `internal/cli/dump.go`, parameterize `toDTO` with the target OS, pass the engine/host `goos` at the call site; T018 passes

**Checkpoint**: All four stories functional and independently verified.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Make every documentation claim true (FR-006), record the
behavior change, and run the full constitutional gate set.

- [ ] T020 [P] Update `docs/runefile.md` attributes section: OS attributes now document visibility AND enforcement — hidden from list/picker/completion/MCP, direct invocation errors, dependencies/post-hooks skip silently, OR semantics, `unix` = everything except Windows
- [ ] T021 [P] Update `docs/mcp.md` exposure rule: "Every non-private task **available on the host OS** is exposed as an MCP tool" (contract mcp-surface.md)
- [ ] T022 [P] Update `docs/examples/os-filtering/`: add a dispatcher task (deps on the per-OS setup tasks) to the `Runefile` and rewrite `README.md` so the "only shows and runs on that OS" claim is true and the skip/dispatch pattern is demonstrated
- [ ] T023 Run the docs harness in Docker: `docker-compose run --rm test go test ./test/docs/...` — examples validate and run, fenced blocks parse, links resolve (SC-005)
- [ ] T024 Execute quickstart.md scenarios 1–5 manually with a built binary (`go build -trimpath -o dist/rune ./cmd/rune`) and confirm expected outputs
- [ ] T025 Full gate set: `golangci-lint run` clean, `docker-compose run --rm test go test ./...`, `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...`, `goreleaser check` untouched-config sanity — all green
- [ ] T026 Record the behavior change for the changelog (git-cliff reads commits): the feature commit body must note that OS-mismatched dependencies are now skipped and direct invocation now errors — previously both executed — as an intentional pre-1.0 behavior change (plan Complexity Tracking)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none — start immediately
- **Foundational (Phase 2)**: T002 → T003 → T004; T005 after T003 (independent of T004). BLOCKS all stories
- **US1 (Phase 3)**: needs T003 + T005 (predicate + adapter goos)
- **US2 (Phase 4)**: needs T003 + T005; independent of US1 (different concern, small shared file `mcp.go` — see notes)
- **US3 (Phase 5)**: needs T003 + T005; scheduler work (T014/T016) is independent of US1/US2
- **US4 (Phase 6)**: needs T003 + T005 only; fully independent of US1–US3
- **Polish (Phase 7)**: T020–T022 after all stories; T023 after T022; T024–T026 last

### User Story Dependencies

- No story depends on another story's tasks. US1 is the MVP.
- Shared-file caution: US1 (T007), US2 (T013) both edit `internal/cli/mcp.go`; US2 (T012), US3 (T017) both edit `internal/cli/run.go`. Parallelize across stories only with separate branches/worktrees; sequential P1→P4 avoids conflicts entirely.

### Within Each User Story

- RED tests precede GREEN implementation (constitution Principle VI) —
  verify each red task actually fails/breaks compilation before
  implementing.
- Run the suite in Docker after every GREEN task.

### Parallel Opportunities

- T002 (ast tests) alongside T001 baseline.
- After Phase 2: T006, T009, T010, T011, T014, T015, T018 (all RED tests across five distinct files) can be written in parallel — T006/T010 share `mcp_test.go` and T009/T015 share `run_test.go`, so same-file pairs belong to one author.
- Polish docs T020, T021, T022 in parallel (three separate docs).

---

## Parallel Example: post-Foundational RED wave

```bash
# All failing tests, five distinct files, no shared state:
Task: "T006 MCP Tasks() filter test in internal/cli/mcp_test.go"
Task: "T009 direct-invoke error tests in internal/cli/run_test.go"
Task: "T010 MCP Call() defense test in internal/cli/mcp_test.go"    # same file as T006 — same author
Task: "T011 binary-level availability-error test in test/integration/"
Task: "T014 scheduler skip tests in internal/runtime/scheduler/scheduler_test.go"
Task: "T015 CLI dispatch test in internal/cli/run_test.go"           # same file as T009 — same author
Task: "T018 dump availability test in internal/cli/dump_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 (baseline) + Phase 2 (predicate, call-site swap, goos plumbing)
2. Phase 3: MCP filtering — the feature as originally requested
3. **STOP and VALIDATE**: quickstart Scenario 5 + `--list` regression
4. Ship if desired: agents already can't see foreign-OS tasks

### Incremental Delivery

- +US2 → nothing mismatched can execute (closes the leaky abstraction)
- +US3 → dispatch pattern works (makes enforcement ergonomic)
- +US4 → tooling sees computed verdicts
- Polish → docs truthful, gates green, changelog note

Sequential single-developer order T001→T026 is conflict-free by
construction.

---

## Notes

- Total: 26 tasks (Setup 1, Foundational 4, US1 3, US2 5, US3 4, US4 2, Polish 7)
- Exact existing test-file names may differ (`run_test.go` etc.) — extend
  the file that already covers that unit; create it only if none exists
- Tests run in Docker exclusively (constitution Engineering Constraints)
- Commit after each GREEN task or logical group; conventional-commit
  messages (git-cliff changelog)
