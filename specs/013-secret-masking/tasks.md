# Tasks: Secret Masking & Sanitization

**Input**: Design documents from `/specs/013-secret-masking/`

**Prerequisites**: plan.md, spec.md, research.md (D1–D8), data-model.md, contracts/secret-masking.md, quickstart.md

**Tests**: INCLUDED — Constitution VI mandates test-first (Red-Green-Refactor). Every implementation task is preceded by a failing-test task. All test commands run inside Docker (`docker-compose run --rm test go test ./...`), never on the host.

**Organization**: Tasks are grouped by user story. The writer-wrapping choke point is shared, so it lands with US1 (MVP); US2 and US3 layer surfaces and author controls on top, each independently testable. Integration tests are split into one file per story so `[P]` markers never collide on a file.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1 (agent/MCP masking), US2 (terminal parity), US3 (author declarations)

## Path Conventions

Single Go module at repo root; engine code under `internal/`, binary tests under `test/integration/`, docs fixtures under `docs/` (validated by `test/docs`). Layout per plan.md Project Structure.

---

## Phase 1: Setup

**Purpose**: Scaffold the one structural addition (justified in plan.md Complexity Tracking)

- [X] T001 Create `internal/mask` package scaffold: `internal/mask/doc.go` with package comment stating purpose (emission-time secret masking), invariants (immutable Set, bounded carry, concurrent-safe Writer, flush only when no producer can be writing, empty set ⇒ never installed), and the fixed constants (placeholder `***`, MinLen 4)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The `mask` package itself — Set derivation and streaming Writer. Everything else consumes these.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete and green.

- [X] T002 [P] Write failing unit tests for Set derivation in `internal/mask/set_test.go`: built-in pattern matching (case-insensitive substring: TOKEN, SECRET, PASSWORD, PASSWD, APIKEY, API_KEY, PRIVATE_KEY, ACCESS_KEY, CREDENTIAL, AUTH), declared names, exemptions (exempt beats pattern, does not beat declaration), MinLen-4 exclusion, multi-line value split into per-line entries **only** (no whole-value entry; per data-model.md derivation step 3), dedup, empty-env/empty-set cases, MaxEntryLen derived from line-sized entries
- [X] T003 [P] Write failing unit tests for Writer in `internal/mask/writer_test.go`: single/multiple occurrence replacement with `***`, occurrences split across consecutive `Write` calls (byte-at-a-time worst case), leftmost-longest overlap (one secret nested in another never exposes a fragment), carry never exceeds MaxEntryLen−1, Flush emits carry verbatim and clears (callers may only flush when no producer is mid-stream — research.md D4 flush safety rule), Close flushes, concurrent `Write`s from multiple goroutines are safe and never interleave partial masks (rules: data-model.md → MaskWriter, contract §2.2)
- [X] T004 Implement Set in `internal/mask/set.go`: `NewSet(env []string, declared, exempt []string) *Set` per data-model.md derivation steps 1–4; make T002 pass
- [X] T005 Implement Writer in `internal/mask/writer.go`: `NewWriter(w io.Writer, s *Set) *Writer` (io.WriteCloser + Flush), mutex-protected bounded carry per research.md D4; make T003 pass
- [X] T006 Verify foundation in Docker: `docker-compose run --rm test go test ./internal/mask/...` and `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./internal/mask/...` both green

**Checkpoint**: `internal/mask` is complete, race-clean, and unconsumed — user stories can begin.

---

## Phase 3: User Story 1 - Agent runs a task that leaks credentials, output arrives masked (Priority: P1) 🎯 MVP

**Goal**: An agent calling a task via MCP receives `***` in place of every secret value — the core promise (SC-002). Installs the single writer-wrap choke point that all later stories reuse.

**Independent Test**: quickstart.md Scenario 1 — stdio MCP `tools/call` on a secret-printing task; assert the tool result contains `***` and never the raw value.

### Tests for User Story 1 (write first, must fail)

- [X] T007 [P] [US1] Write failing integration test in `test/integration/secret_masking_mcp_test.go` (new file, `runWithStdin` harness pattern from `test/integration/harness_test.go`): MCP stdio `tools/call` on a task that echoes and `env`-dumps `API_TOKEN` returns masked stdout/stderr sections in the tool result text; raw value absent from the entire response (contract §3, SC-002)
- [X] T008 [P] [US1] Write failing integration test in `test/integration/secret_masking_exec_test.go` (new file): direct CLI run of a python-executor task (interp path) printing a pattern-matched env var emits `***` on stdout — proves masking is executor-independent (contract §2.3 item 1). First check how existing executor tests handle interpreter availability; guard with a `t.Skip` when `python3` is absent from PATH so CI environments without it stay green
- [X] T009 [US1] Build the per-run secret set in `internal/cli/run.go`: after `buildEnv` (≈line 594 call site in `execute`), collect effective env + union of all tasks' `[env("K","V")]` pairs and construct `mask.NewSet` (built-in patterns only until US3 wires declarations)
- [X] T010 [US1] Install the choke point in `internal/cli/run.go`: when the set is non-empty, wrap `Options.Stdout`/`Options.Stderr` with `mask.NewWriter` at engine construction. Implement the **flush safety rule** (research.md D4): flush a wrapper only where no producer can still be writing to it — after a body line's process exits when no `[parallel]` group is in flight, and at run end after the scheduler has joined every task; never on a timer or unconditionally per line. Empty set ⇒ writers untouched (FR-008 structural guarantee)
- [X] T011 [US1] Cover the MCP adapter path in `internal/cli/mcp.go`: `mcpAdapter.Call` (≈lines 57–91) builds its engine with `bytes.Buffer` writers — route it through the same wrap (shared helper) so buffers only ever hold masked text and `mcpserver/formatResult` needs no change; make T007/T008 pass
- [X] T012 [US1] Checkpoint in Docker: `docker-compose run --rm test go test ./internal/... ./test/integration/...` green, including race run on `./internal/cli/...`

**Checkpoint**: MVP — an agent can no longer receive raw credentials from any task output. Demo via quickstart.md Scenario 1.

---

## Phase 4: User Story 2 - Human terminal runs are masked the same way (Priority: P2)

**Goal**: Byte-for-byte the same masking guarantee on the terminal: task passthrough, echoed command lines, dry-run lines, interrupted runs, Rune's own emissions — plus proof that the task env is untouched (FR-011), that masking has no off switch (FR-006), and that secret-free files are byte-identical (SC-003).

**Independent Test**: quickstart.md Scenarios 2 and 5 — terminal run of a secret-printing task shows `***` everywhere including the echoed command; existing golden/styling suites pass unmodified.

**Note**: T013–T019 all extend `test/integration/secret_masking_cli_test.go` (one file), so they are intentionally **not** `[P]` — write them in sequence or as one sitting.

### Tests for User Story 2 (write first, must fail where behavior is new)

- [X] T013 [US2] Write failing integration tests in `test/integration/secret_masking_cli_test.go` (new file): (a) terminal run env-dump masked on stdout, (b) failing task's stderr masked with exit code preserved, (c) same secret value multiple times and across multiple output lines all masked (spec US1/AS3 via terminal, US2/AS1)
- [X] T014 [US2] Write failing integration test in `test/integration/secret_masking_cli_test.go`: un-suppressed (no `@`) command line interpolating a secret echoes with `***` on stderr; `@` and `set quiet` suppression semantics unchanged (contract §2.3 item 2; echo site: `internal/runtime/shell/shell.go:101-107`)
- [X] T015 [US2] Write failing integration test in `test/integration/secret_masking_cli_test.go`: `--dry-run` "would run" lines with interpolated secrets are masked (contract §2.3 item 3)
- [X] T016 [US2] Write failing integration test in `test/integration/secret_masking_cli_test.go`: a task that streams a secret and then sleeps is interrupted (context timeout / process kill) mid-stream — every byte emitted before the interrupt is masked, and no unmasked window appears at teardown (spec Edge Cases → interrupted tasks, FR-003)
- [X] T017 [US2] Write failing integration test in `test/integration/secret_masking_cli_test.go` for FR-011: the task itself compares `$API_TOKEN` against the known raw value (e.g. `test "$API_TOKEN" = "hunter2..." && echo match`) and prints `match` — proving the process env carries the *real* value while every emission is masked
- [X] T018 [US2] Write integration test in `test/integration/secret_masking_cli_test.go` for FR-006 / contract §2.6: plausible disable attempts (e.g. `RUNE_NO_MASK=1`-style env vars, absence of any CLI disable flag in `--help` output) leave masking fully active — the only opt-out is per-variable `set unmasked`
- [X] T019 [US2] Write byte-invariance guard in `test/integration/secret_masking_cli_test.go`: a secret-free Runefile run under the new binary produces output byte-identical to a table of expected literals (styled + unstyled); confirm existing `test/integration/styling_test.go`, `internal/cli` goldens, and `test/corpus` pass with zero fixture edits (SC-003)

### Implementation for User Story 2

- [X] T020 [US2] Close any gaps T013–T019 reveal in `internal/cli/run.go` (status/warning/dry-run/confirm emission sites ≈lines 252–349) — expected mostly no-ops since all sites write through the wrapped `opts.Stderr`; verify the flush safety rule holds under `[parallel]` groups (no flush while any group member still streams) and that interrupt teardown flushes only after producers have exited
- [X] T021 [US2] Checkpoint in Docker: full suite `docker-compose run --rm test go test ./...` green with zero golden-file modifications

**Checkpoint**: Terminal and agent surfaces provably share one masking behavior; env untouched; no off switch.

---

## Phase 5: User Story 3 - Author declares additional secrets beyond the defaults (Priority: P3)

**Goal**: `set secrets := [...]` / `set unmasked := [...]` give authors the last word, with full static validation (FR-005, FR-009).

**Independent Test**: quickstart.md Scenarios 3 and 4 — innocently-named declared variable masked; pattern-matched exempted variable unmasked; malformed/conflicting declarations fail before execution with positioned diagnostics.

### Tests for User Story 3 (write first, must fail)

- [ ] T022 [P] [US3] Write failing unit tests in `internal/config/settings_test.go`: `set secrets`/`set unmasked` resolve to `Settings.Secrets`/`Settings.Unmasked` via list evaluation; malformed elements yield positioned diagnostics; same name in both lists yields an error diagnostic with related span, emitted by `ResolveSettings` (contract §1; data-model.md validation rules)
- [ ] T023 [P] [US3] Write failing integration tests in `test/integration/secret_masking_settings_test.go` (new file): (a) declared innocent-name var masked, (b) `unmasked` exempts a pattern-matched name, (c) name absent from env is inert, (d) `set secrets := [42]`-style malformed value → positioned diagnostic + exit 3 + nothing executed, (e) same name in both lists → positioned error citing both spans + exit 3, (f) `rune analyze` flags `set secert` typo as RUNE2008 (quickstart Scenarios 3–4)
- [ ] T024 [P] [US3] Register `secrets` and `unmasked` in `internal/language/builtin.go` `builtinSettings` (unlocks analyzer/LSP recognition; RUNE2008 for typos comes free via `language.IsValid`)

### Implementation for User Story 3

- [ ] T025 [US3] Add `Secrets`/`Unmasked` fields + switch cases to `internal/config/settings.go` `ResolveSettings` using existing `evalList`; add the both-lists conflict diagnostic (primary + related span) there — execution-path enforcement per data-model.md (`rune analyze` surfacing is a non-v1 follow-up); make T022 pass
- [ ] T026 [US3] Wire `Settings.Secrets`/`Settings.Unmasked` into the `mask.NewSet` call from T009 in `internal/cli/run.go` (CLI and MCP paths share it); make T023 pass
- [ ] T027 [US3] Update grammar-drift fixtures: add both settings to `testdata/corpus/full.rune` and deliberately regenerate `testdata/corpus/full.ast` (never hand-edit); add a formatter case to `testdata/fmt/` if list-setting formatting needs pinning
- [ ] T028 [US3] Checkpoint in Docker: full suite green; `go run ./cmd/rune analyze` accepts the new settings and flags typos

**Checkpoint**: All three stories independently functional; author control complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Performance guardrail, documentation (Constitution: surface changes carry their docs — same PR), final gates.

- [ ] T029 [P] Add benchmark comparing masked vs unmasked writer on a 10 MB stream in `internal/mask/writer_bench_test.go`; document result against SC-004 (≤10% overhead); optimize only if the profile demands it (Constitution VIII)
- [ ] T030 [P] Update `docs/GRAMMAR.md` (settings grammar entries for `secrets`/`unmasked`) and `docs/runefile.md` (settings reference rows)
- [ ] T031 [P] Write `docs/how-to/secret-masking.md`: pattern list, `***` placeholder, MinLen 4, multi-line behavior, guarantee boundary/non-goals per FR-010 + contract §2.4; add pitfall cross-link in `docs/how-to/settings-and-dotenv.md`
- [ ] T032 [P] Update `docs/mcp.md`: agent-facing guarantee note (tool results masked identically to terminal; no agent-facing off switch, FR-006)
- [ ] T033 [P] Create runnable example `docs/examples/secret-masking/Runefile` + `docs/examples/secret-masking/README.md` (picked up by the `test/docs` harness as a fixture)
- [ ] T034 Execute all quickstart.md scenarios 1–7 end-to-end and record outcomes in the PR description
- [ ] T035 Full gate set: `docker-compose run --rm test go test ./...`, `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...`, `go run ./cmd/rune lint`, `go run ./cmd/rune docs-check`, `go run ./cmd/rune release-dryrun`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none — start immediately
- **Foundational (Phase 2)**: after T001 — BLOCKS all user stories
- **US1 (Phase 3)**: after Phase 2 — installs the shared choke point (T009–T011)
- **US2 (Phase 4)**: after Phase 2; its tests exercise the choke point from US1, so run after T010 lands (tests T013–T019 can be *written* any time after Phase 2)
- **US3 (Phase 5)**: after Phase 2; T026 touches the `mask.NewSet` call created in T009, so implementation follows US1; T022–T024 are independent of US1/US2
- **Polish (Phase 6)**: docs tasks (T030–T033) only need US3's surface to exist; T034/T035 need everything

### Within Each User Story

- Failing tests before implementation (Red-Green-Refactor, Constitution VI)
- `internal/mask` (models) → CLI wiring (services) → MCP/terminal surfaces (endpoints)
- Docker checkpoint task closes every phase

### Parallel Opportunities

- Phase 2: T002 ∥ T003 (different files)
- US1: T007 ∥ T008 (different test files)
- US2: T013–T019 share one test file — deliberately sequential (see phase note)
- US3: T022 ∥ T023 ∥ T024 (different files)
- Polish: T029–T033 all parallel (disjoint files)
- After Phase 2, US3's settings plumbing (T022, T024, T025) can proceed in parallel with US1's wiring — only T026 waits on T009

---

## Parallel Example: User Story 3

```bash
# After Phase 2 completes, launch together:
Task: "Failing unit tests for secrets/unmasked resolution in internal/config/settings_test.go"          # T022
Task: "Failing integration tests for declarations in test/integration/secret_masking_settings_test.go"  # T023
Task: "Register secrets/unmasked in internal/language/builtin.go"                                        # T024
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 (T001) → Phase 2 (T002–T006): the `mask` package, fully unit-tested and race-clean
2. Phase 3 (T007–T012): choke-point wiring + MCP end-to-end proof
3. **STOP and VALIDATE**: quickstart Scenario 1 — an agent demonstrably cannot receive a raw credential. This alone is shippable value.

### Incremental Delivery

1. MVP (above) → 2. US2 proves terminal parity, env non-mutation, always-on, and byte-invariance (SC-003) → 3. US3 adds author declarations + grammar fixtures → 4. Polish lands benchmark + the doc set the Constitution requires in the same PR. Each checkpoint leaves the suite green and the previous stories intact.

### Notes

- The entire feature ships as one PR (Constitution: surface changes carry docs + fixtures together); the story checkpoints are commit/validation boundaries within the branch, not separate releases
- Never hand-edit goldens (`full.ast`, styling fixtures) — regenerate deliberately and review the diff
- Commit after each task or logical group; every commit must pass the Docker test run (global policy)
