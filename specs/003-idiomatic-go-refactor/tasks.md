---
description: "Task list for Idiomatic Go Refactor — Skill-Governed Review & Refactoring"
---

# Tasks: Idiomatic Go Refactor — Skill-Governed Review & Refactoring

**Input**: Design documents from `/specs/003-idiomatic-go-refactor/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅ (review-rubric, preservation-invariants, lint-gate), quickstart.md ✅

**Tests**: No new **product** TDD tests — this is an **externally behavior-preserving** refactor whose oracle is the **existing** suite (unit/golden/integration/corpus/fuzz + mcpserver/cli contract tests), which MUST stay green **with no golden regenerated** (SC-004). Net-new **test code** is limited to `goleak` guards (US2) and hot-path **benchmarks** (US5). Per global policy the Go suite runs **inside Docker** (`docker-compose run --rm test …`); `-race` runs with `-e CGO_ENABLED=1`.

**Organization**: Tasks are grouped by user story (US1–US5). Priorities: US1 (review) and US2 (correctness) are **P1**; US3 (design) and US4 (gates) are **P2**; US5 (perf) is **P3**. Suggested MVP = **US1 + US2**.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1–US5 (user-story phases only; Setup/Foundational/Polish carry no label)
- Exact file paths are included in every task

## Path Conventions

Single Go module rooted at the repo. The **constitution-locked package layout (Principle IV) is preserved** — no locked package is renamed; reorganization is within-package only. The review report is `specs/003-idiomatic-go-refactor/review.md`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Tooling prerequisites.

- [X] T001 [P] Add `go.uber.org/goleak` as a test dependency (`go get go.uber.org/goleak@latest`; update `go.mod` + `go.sum`) — used by US2 leak guards (supports FR-010)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish the known-good behavior baseline that every refactor is measured against.

**⚠️ CRITICAL**: This baseline must be green before any remediation begins.

- [X] T002 Capture the behavior-preservation baseline: `docker-compose run --rm test go test ./...` and `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...` both green, and `git status --porcelain testdata/` empty — the oracle for SC-004/INV acceptance (`contracts/preservation-invariants.md`)

**Checkpoint**: known-green starting point recorded — remediation can be measured against it.

---

## Phase 3: User Story 1 - Skill-governed review (Priority: P1) 🎯 MVP

**Goal**: A complete, traceable findings report — every Go package assessed against every `golang-*` skill, each finding carrying skill + rule + `file:line` + severity + fix.

**Independent Test**: The report covers every package; findings are well-formed and ordered S1→S4 (SC-001). See `quickstart.md` §US1 and `contracts/review-rubric.md`.

- [X] T003 [US1] Create the review report `specs/003-idiomatic-go-refactor/review.md` with the finding schema, severity legend (S1–S4), and a per-package coverage checklist from `contracts/review-rubric.md`
- [X] T004 [P] [US1] Review the front-end packages (`internal/{token,lexer,ast,parser,analyzer,diag}`) against all `golang-*` skills; record findings (incl. the known dead `_ = lb/_ = rb` in `internal/parser/attribute.go:25-26`)
- [X] T005 [P] [US1] Review the eval/config packages (`internal/{eval,config,dotenv,cache}`); record findings (incl. `init()`+mutable `builtins` global in `internal/eval/builtins.go`; the `unsafe` global naming in `internal/cache/cache.go:118`)
- [X] T006 [P] [US1] Review the runtime packages (`internal/runtime{,/scheduler,/shell,/interp,/agent}`); record concurrency findings (scheduler `select`/`ctx.Done()`/bounding; `interp` temp-file resource handling)
- [X] T007 [P] [US1] Review `internal/cli` (11 files / 1857 LOC) and `cmd/rune`; record findings (silent `cache.Store` discard at `internal/cli/run.go:252`; file-responsibility/decomposition candidates; flag wiring)
- [X] T008 [P] [US1] Review `mcpserver` against all skills; record findings (goroutine ownership/exit + swallowed `srv.Serve` error in `transport.go`; confirm secure-by-default preserved per INV-6)
- [X] T009 [US1] Consolidate: audit the 7 `fmt.Errorf`-without-`%w` sites and triage every `_ =` discard (fix vs keep-with-comment, per research §6); finalize severities, order S1→S4, and confirm every package is covered (reviewed or "clean") (SC-001) — depends on T004–T008

**Checkpoint**: an actionable, traceable review exists; US2/US3/US5 draw their work from it.

---

## Phase 4: User Story 2 - Remediate correctness & safety (Priority: P1)

**Goal**: All S1 findings fixed; race- and leak-clean; zero behavior change.

**Independent Test**: `-race` + goleak clean; full suite green with no golden regenerated (SC-002/003/004). See `quickstart.md` §US2. Depends on US1 findings.

- [ ] T010 [P] [US2] Surface the swallowed server error in `mcpserver/transport.go` (capture/report `srv.Serve` on the failure path; preserve the shutdown contract — INV-6/INV-7)
- [X] T011 [P] [US2] Handle the silent `cache.Store` discard in `internal/cli/run.go:252` (log the failure; do NOT change caching or exit semantics — Principle I, INV-1/INV-2)
- [X] T012 [P] [US2] Resolve the dead `_ = lb/_ = rb` in `internal/parser/attribute.go:25-26` (use the spans or remove; corpus/parser goldens MUST NOT change — INV-5)
- [X] T013 [US2] Wrap actionable errors with `%w` (and use `errors.Is/As` where compared) at the `fmt.Errorf` sites not already touched by T010/T011; leave/comment intentional non-wraps — depends on T010, T011
- [X] T014 [US2] Verify/repair concurrency in `internal/runtime/scheduler/{scheduler,parallel}.go` and `internal/cli/watch.go`: every blocking `select` observes `ctx.Done()`, every goroutine has a guaranteed exit, fan-out is bounded (`errgroup.SetLimit`); **ensure a `cli` test actually exercises `watch` so the goleak guard (T015) covers the watcher goroutine — otherwise note the gap and rely on this manual review (L1)** (FR-006)
- [X] T015 [P] [US2] Add `goleak.VerifyTestMain(m)` `TestMain` to the `internal/runtime/scheduler`, `mcpserver`, and `internal/cli` test packages (FR-010) — depends on T001
- [X] T016 [US2] Verify US2: `docker-compose run --rm test go test ./...` green + `git status --porcelain testdata/` empty + `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...` green (0 races, 0 leaks) (SC-002/003/004) — depends on T010–T015

**Checkpoint**: the codebase is correctness/safety-clean with behavior unchanged.

---

## Phase 5: User Story 3 - Idiomatic design & structure (Priority: P2)

**Goal**: `init()`/globals removed, naming fixed, exported docs filled, files tidied within the locked layout — behavior-preserving.

**Independent Test**: zero unjustified `init()`/globals; locked layout intact; suite green, no golden regenerated (SC-004/006). See `quickstart.md` §US3. Depends on US1 findings.

- [X] T017 [US3] Replace the `init()`-populated mutable `builtins` global in `internal/eval/builtins.go` with `sync.OnceValue(newBuiltins)` (or evaluator-held map); `IsBuiltin`/`callBuiltin` use the accessor; the function set + error messages are identical (eval/diag/corpus goldens unchanged — INV-5) (FR-011, SC-006)
- [X] T018 [P] [US3] Rename the `unsafe` package-global in `internal/cache/cache.go:118` (shadows the stdlib `unsafe` package) to a descriptive name; update references (naming rule, FR-013)
- [ ] T019 [US3] Apply the remaining S2/S3 design findings from `review.md` (early-return / no `else` after terminal branch, explicit slice/map init, field-named composite literals, unexport unnecessary surface **in `internal/*` only — the public `mcpserver` API and `cmd` surface are frozen, INV-7/G1**) across the flagged packages
- [ ] T020 [US3] Where US1 flagged mixed responsibilities in `internal/cli`, split files **within** the package (package name unchanged, no exported-surface growth) per research §7 — depends on T019
- [ ] T021 [P] [US3] Add/complete doc comments on the exported identifiers flagged by the review (FR-020, SC-007)
- [X] T022 [US3] Verify US3: `grep -rn '^func init()' cmd internal mcpserver | grep -v _test` empty (or each justified); **document the remaining read-only package globals (lookup tables, sentinel errors, compiled regexes) as acceptable-immutable so SC-006's "justified" clause is concretely satisfied (A1)**; locked package list unchanged; `docker-compose run --rm test go test ./...` green with no golden regenerated **and the `mcpserver`/CLI contract tests pass unmodified (M1)** (SC-004/006/009) — depends on T017–T021

**Checkpoint**: the design is idiomatic; structure tidied within the locked layout; behavior unchanged.

---

## Phase 6: User Story 4 - Encode skills as gates (Priority: P2)

**Goal**: the expanded linter set passes clean and is enforced in CI.

**Independent Test**: `golangci-lint run` → 0 issues on the refactored tree; a seeded violation is flagged (SC-005). See `quickstart.md` §US4 and `contracts/lint-gate.md`. Best done after US2/US3 so the tree is already clean.

- [X] T023 [US4] Extend `.golangci.yml` with `errorlint`, `contextcheck`, `noctx`, `bodyclose`, `predeclared`, `wastedassign`, `revive` (curated rule subset), and `gocritic` (stable diagnostic/style tags) per `contracts/lint-gate.md` (FR-015)
- [X] T024 [US4] Run `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...` and fix any residual findings until it reports **0 issues** (SC-005) — depends on T023
- [X] T025 [US4] Confirm the gate is live: the existing CI `lint` job (`.github/workflows/ci.yml`) picks up `.golangci.yml` automatically; verify a deliberately-seeded violation (e.g. an unwrapped error) is flagged, then revert (FR-016) — depends on T024

**Checkpoint**: future skill drift fails CI automatically.

---

## Phase 7: User Story 5 - Benchmark-gated performance (Priority: P3)

**Goal**: hot-path benchmarks exist; only benchstat-proven optimizations merge.

**Independent Test**: benchmarks present for lexer/parser/eval/scheduler; a baseline is recordable; any perf change carries a benchstat win (SC-008). See `quickstart.md` §US5. Depends on US1 findings.

- [X] T026 [P] [US5] Add `Benchmark*` functions to `internal/lexer` and `internal/parser` (tokenize/parse a representative `testdata` Runefile) (FR-017)
- [ ] T027 [P] [US5] Add `Benchmark*` functions to `internal/eval` (expression + `{{…}}` interpolation) and `internal/runtime/scheduler` (DAG topo-sort + bounded fan-out) (FR-017)
- [X] T028 [US5] Record a baseline (`docker-compose run --rm test go test -run=xxx -bench=. -benchmem ./internal/lexer ./internal/parser ./internal/eval ./internal/runtime/scheduler` → `bench-before.txt`); merge only optimizations with a `benchstat` win, each under a `perf(scope):` commit with an explanatory comment (FR-018, SC-008) — depends on T026, T027

**Checkpoint**: a performance baseline exists and gates all future optimization.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: close out the review, validate end-to-end, final gate.

- [X] T029 [P] Finalize `review.md`: set each finding's `status` (`fixed` + commit SHA / `deferred(reason)` / `wontfix(justification)`); document any skill ⟂ constitution deferrals where the constitution governs (SC-010, FR-019)
- [ ] T030 Run the full `quickstart.md` validation (all US scenarios) and check off its Done-when matrix (validates SC-001–SC-010)
- [X] T031 Final green gate: expanded `golangci-lint run` → 0 issues + `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...` green + `git status --porcelain testdata/` empty (SC-003/004/005)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: T001 (goleak dep) — no dependencies.
- **Foundational (P2)**: T002 baseline — depends on a clean checkout; blocks all remediation.
- **US1 (P3, review)**: depends on Foundational. **Feeds US2/US3/US5** (their work lists).
- **US2 (P4, correctness)**: depends on US1 findings + T001 (goleak) + T002 (baseline).
- **US3 (P5, design)**: depends on US1 findings; best after US2 to avoid touching the same files twice.
- **US4 (P6, gates)**: depends on US2 + US3 (add linters only once the tree is clean — research §2).
- **US5 (P7, perf)**: depends on US1 findings; benchmarks independent of US2–US4.
- **Polish (P8)**: depends on all targeted stories.

### User Story Dependencies

- **US1 (P1)** is the spine — every other story consumes its findings.
- **US2 (P1)** depends on US1 (S1 findings) and Setup/Foundational.
- **US3 (P2)** depends on US1; sequenced after US2 (shared files).
- **US4 (P2)** depends on US2+US3 (clean tree before the gate goes red).
- **US5 (P3)** depends on US1; otherwise independent.

### Within Each Story

- US1: T003 (report skeleton) before the package reviews (T004–T008); T009 consolidates last.
- US2: T010/T011/T012/T015 are different files → parallelizable; T013 after T010/T011 (file overlap); T014 (scheduler+watch); T016 verifies last.
- US3: T017 (eval) and T018 (cache) independent; T019 broad; T020 after T019; T022 verifies last.
- US4: strictly sequential (T023 → T024 → T025; same config + dependency).
- US5: T026/T027 different files → parallel; T028 after both.

### Parallel Opportunities

- **US1 reviews (T004–T008)** are independent analyses by package group — split across reviewers; each appends its section to `review.md` (serialize the *write*, parallelize the *analysis*).
- **US2**: T010, T011, T012, T015 touch different files → run [P]; T014 also independent.
- **US3**: T018 and T021 run alongside T017.
- **US5**: T026 ∥ T027.
- **Cross-story**: US5 benchmarks can proceed in parallel with US2/US3 (different files), since they only add `Benchmark*` functions.

---

## Parallel Example: User Story 1 (reviews)

```text
# Independent package-group analyses (results merged into review.md):
Task: "Review front-end packages (token/lexer/ast/parser/analyzer/diag)"  # T004
Task: "Review eval/config/dotenv/cache"                                    # T005
Task: "Review runtime/{scheduler,shell,interp,agent}"                      # T006
Task: "Review internal/cli + cmd/rune"                                     # T007
Task: "Review mcpserver"                                                    # T008
# Then T009 consolidates + orders by severity.
```

## Parallel Example: User Story 2 (distinct files)

```text
Task: "Surface srv.Serve error in mcpserver/transport.go"     # T010
Task: "Log cache.Store failure in internal/cli/run.go"        # T011
Task: "Resolve dead _ = lb/_ = rb in parser/attribute.go"     # T012
Task: "Add goleak TestMain to scheduler/mcpserver/cli"        # T015
# Then T013 (errorf %w on remaining sites), T014 (concurrency), T016 (verify).
```

---

## Implementation Strategy

### MVP First (P1 baseline = US1 + US2)

1. Setup + Foundational (goleak dep; green baseline).
2. US1 (review) → **STOP & VALIDATE**: a complete, traceable findings report (SC-001). Valuable on its own.
3. US2 (correctness) → **STOP & VALIDATE**: S1 fixed, `-race`+goleak clean, no behavior change. P1 baseline complete.

### Incremental Delivery

1. US1 review → the work list and the headline value (the audit).
2. US2 → safer, correct core (the highest-impact fixes).
3. US3 → idiomatic design/structure.
4. US4 → durable enforcement (skills as gates).
5. US5 → performance baseline + proven optimizations.
6. Polish → finalize the report, full validation.

### Parallel Team Strategy

After US1: Dev A drives US2 (correctness), Dev B drives US5 (benchmarks, different files), Dev C prepares US3 design fixes. US4 integrates once US2+US3 land.

---

## Requirements Coverage (traceability)

| Req | Tasks | Req | Tasks |
|-----|-------|-----|-------|
| FR-001 | T003–T009 | FR-011 | T017 |
| FR-002 | T003–T009 | FR-012 | T019 |
| FR-003 | T009 | FR-013 | T018, T020 |
| FR-004 | T004–T008 | FR-014 | T019, T022 |
| FR-005 | T010–T014 | FR-015 | T023 |
| FR-006 | T014 | FR-016 | T025 |
| FR-007 | T013 | FR-017 | T026, T027 |
| FR-008 | T010, T014 | FR-018 | T028 |
| FR-009 | T016, T022 | FR-019 | T029 |
| FR-010 | T015, T016 | FR-020 | T021 |
| SC-001 | T009 | SC-006 | T017, T022 |
| SC-002 | T010–T016 | SC-007 | T021 |
| SC-003 | T016, T031 | SC-008 | T026–T028 |
| SC-004 | T016, T022, T031 | SC-009 | T016, T022, T031 (contract tests unmodified) |
| SC-005 | T024, T025, T031 | SC-010 | T029 |

---

## Notes

- [P] = different files, no dependency on an incomplete task. US1 reviews append to one `review.md` (analysis parallel, write serialized).
- **Behavior-preserving**: "done" means the existing suite stays green with **no golden regenerated** (SC-004); the `mcpserver`/CLI contract tests pass **unmodified** (SC-009).
- All Go test/bench runs go through Docker (`docker-compose run --rm test …`; `-race` adds `-e CGO_ENABLED=1`).
- Each remediation commit references its finding id from `review.md` (SC-010).
- No package in the constitution-locked layout is renamed (Principle IV); structural change is within-package only.
- Performance: no optimization merges without a `benchstat` win (Principle VIII).
