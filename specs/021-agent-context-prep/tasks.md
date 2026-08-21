# Tasks: Agent Context Prep (`[context]`)

**Input**: Design documents from `/specs/021-agent-context-prep/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md,
contracts/context-attribute.md, quickstart.md

**Authoritative detail**: [implementation-detail.md](implementation-detail.md)
(cited below as **ID-1 … ID-6**) carries the exact test code, implementation
code, and expected red/green outputs for every task. Follow it verbatim;
this file adds speckit ordering, story mapping, and parallelism.

**Tests**: Included — the constitution (Principle VI) mandates test-first
Red-Green-Refactor, and every spec success criterion is test-backed. Each
story's test tasks MUST be written and observed failing before their
implementation tasks.

**All test/build commands run inside Docker**: `docker-compose run --rm test
go test ./<pkg>/...` (never on the host).

**Organization**: Grouped by user story (US1–US4 from spec.md) after a
foundational parser phase all stories share.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1 (MCP instructions, P1) · US2 (agent prompt, P2) ·
  US3 (never blocks, P3) · US4 (static validation, P4)

## Path Conventions

Existing repo layout (plan.md Project Structure): `internal/…`, `mcpserver/`,
`test/integration/`, `testdata/fmt/`, `docs/`.

---

## Phase 1: Setup

No setup tasks — existing repository, no new dependencies, harness already in
place. (Branch `021-agent-context-prep` is checked out; Docker test harness
is the standing prerequisite.)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: `[context]` must parse before any surface, validator, or doc can
exist. Blocks all user stories. (= **ID-1**)

- [ ] T001 Write failing parser tests for the `[context]` attribute (bare form
      and `[context, private]` combination) in
      `internal/parser/attribute_context_test.go`; run
      `docker-compose run --rm test go test ./internal/parser/ -run TestParseContext -v`
      and observe FAIL with `unknown attribute "context"` (ID-1 steps 1–2)
- [ ] T002 Add `AttrContext = "context"` to the attribute-kind constants in
      `internal/ast/ast.go` and add `ast.AttrContext` to the bare-attribute
      case in `internal/parser/attribute.go`; rerun the T001 command to GREEN
      (ID-1 steps 3–4)
- [ ] T003 Add formatter golden fixture `testdata/fmt/context-attr.rune`,
      generate `testdata/fmt/context-attr.fmt` via
      `docker-compose run --rm test go test ./internal/formatter/ -run TestFmtGolden -update`,
      eyeball that `[context]` round-trips unchanged, and confirm
      `docker-compose run --rm test go test ./internal/formatter/ -v` passes
      (ID-1 steps 5–6)
- [ ] T004 Commit: `feat(parser): accept the [context] task attribute`
      (ID-1 step 7)

**Checkpoint**: `[context]` parses and formats; all stories unblocked.

---

## Phase 3: User Story 1 — MCP Agents Start Informed (P1) 🎯 MVP

**Goal**: An MCP client's initialize result carries the hook's masked output
as `instructions`; the hook never appears as a tool. (= **ID-3** core +
**ID-4**)

**Independent Test**: Serve a Runefile whose `[context]` task emits a marker;
initialize an in-memory MCP session and find the marker in
`InitializeResult().Instructions`; confirm the hook is absent from the tool
list (spec US1; quickstart §3).

### Tests (write first, observe RED)

- [ ] T005 [US1] Write failing tests for `gatherContext` success, no-hook, and
      OS-mismatch cases (`TestGatherContextSuccess`, `TestGatherContextNoHook`,
      `TestGatherContextOSMismatchTreatedAsAbsent`) in
      `internal/cli/contextprep_test.go`; expect compile failure —
      `gatherContext` undefined (ID-3 steps 1–2, first three tests)
- [ ] T006 [P] [US1] Write failing MCP tests: `TestInstructionsDelivered` /
      `TestNoInstructionsByDefault` in `mcpserver/server_test.go` (reads
      `cs.InitializeResult().Instructions` over the in-memory transport) and
      `TestAdapterExcludesContextTask` in `internal/cli/mcp_test.go`
      (ID-4 steps 1–2)

### Implementation

- [ ] T007 [US1] Create `internal/cli/contextprep.go`: `contextTask(f, goos)`
      lookup honoring `AvailableOn`, `(a *mcpAdapter) gatherContext(ctx,
      stderr) (string, bool)` over the masked `Call()` path,
      `var contextTimeout = 10 * time.Second`, `const contextMaxBytes = 8 *
      1024`, trim + cap + `[truncated]` marker, degrade notice + stderr
      warning (ID-3 step 3, full listing); T005 tests GREEN
- [ ] T008 [US1] Add `Instructions string` to `mcpserver.Options` and pass
      `&mcp.ServerOptions{Instructions: opts.Instructions}` in `New` in
      `mcpserver/server.go`; exclude `t.Attr(ast.AttrContext) != nil` tasks in
      `mcpAdapter.Tasks()` in `internal/cli/mcp.go`; in `ServeMCP`
      (`internal/cli/serve.go`) hoist `ctx := opts.ctx()`, gather once, pass
      `Instructions` (ID-4 step 3); T006 tests GREEN
- [ ] T009 [US1] Run full `mcpserver` + `internal/cli` suites
      (`docker-compose run --rm test go test ./mcpserver/ ./internal/cli/ -v`)
      as regression guard, then commit:
      `feat(mcp): deliver [context] hook output as server instructions`
      (ID-3 step 5 + ID-4 steps 4–5)

**Checkpoint**: US1 fully functional — quickstart §3 verifiable. MVP done.

---

## Phase 4: User Story 2 — Agent-Executor Prompt Prefix (P2)

**Goal**: Every `(agent)` task invocation gathers fresh context and prepends
it to the provider prompt under the fixed delimiters. (= **ID-5**)

**Independent Test**: With an `sh` stub agent that prints `$1`, run an
`(agent)` task and read the assembled prompt from stdout (spec US2;
quickstart §4).

### Tests (write first, observe RED)

- [ ] T010 [US2] Write failing unit tests `TestBuildAgentPromptWithContext` /
      `TestBuildAgentPromptWithoutContext` (exact string equality on the
      delimiters) in `internal/cli/agentprompt_test.go`; expect compile
      failure — `buildAgentPrompt` undefined (ID-5 steps 1–2)

### Implementation

- [ ] T011 [US2] Add `buildAgentPrompt(contextText string, hasContext bool,
      body string) string` and rework `executeAgent` in
      `internal/cli/agentexec.go` to build the adapter once, gather per
      invocation, and reuse the adapter for the callback endpoint (ID-5
      step 3, full listing); T010 GREEN
- [ ] T012 [US2] Write end-to-end integration test
      `TestContextPrefixReachesAgentPrompt` in
      `test/integration/context_test.go` using the `sh`-stub `agent_cmd`
      Runefile (ID-5 step 5, first test); run
      `docker-compose run --rm test go test ./test/integration/ -run TestContext -v`
      to GREEN
- [ ] T013 [US2] Commit:
      `feat(agent): prepend [context] hook output to agent-executor prompts`
      (ID-5 step 7, first commit — degrade integration test lands in US3)

**Checkpoint**: US2 independently demonstrable via quickstart §4.

---

## Phase 5: User Story 3 — Context Prep Never Blocks (P3)

**Goal**: Failure, timeout, oversized output, and secrets all degrade safely;
no agent surface ever blocks. (= **ID-3** fault-injection tests + **ID-5**
degrade integration test)

**Independent Test**: Point the hook at `@exit 7`, `@sleep 5` (with shrunken
test timeout), 9 000-byte output, and a secret-bearing env var; every
session/run still succeeds with the specified notice/truncation/masking
(spec US3; quickstart §5).

### Tests (these ARE the story — implementation landed in T007)

- [ ] T014 [P] [US3] Add fault-injection tests to
      `internal/cli/contextprep_test.go`: `TestGatherContextFailureDegrades`
      (exact notice line + stderr warning), `TestGatherContextTimeoutDegrades`
      (override `contextTimeout`; if it hangs ~5 s instead of degrading, stop
      and investigate context propagation in `internal/runtime` before
      proceeding), `TestGatherContextTruncates` (cap + `[truncated]` suffix),
      `TestGatherContextMasksSecrets` (`deriveMaskSet`; `***` present, secret
      absent — mirror the nearest existing masking test if env-name
      heuristics differ) (ID-3 step 1, last four tests)
- [ ] T015 [P] [US3] Add integration test `TestContextHookFailureStillRunsAgent`
      in `test/integration/context_test.go`: failing hook → exit 0, notice in
      prompt, warning on stderr (ID-5 step 5, second test)
- [ ] T016 [US3] Run `docker-compose run --rm test go test ./internal/cli/
      ./test/integration/ -run 'TestGatherContext|TestContext' -v` to GREEN
      (fix only if a genuine defect surfaces — expected to pass against T007's
      implementation), then commit:
      `test(context): fault-injection coverage for the [context] hook`

**Checkpoint**: Spec SC-003 satisfied; degrade paths proven end-to-end.

---

## Phase 6: User Story 4 — Authoring Errors Caught Statically (P4)

**Goal**: Duplicate hook, `[confirm]` combo, defaultless parameter, and
`agent`-executor misuse fail analysis with RUNE2007 diagnostics (LSP gets
them via the shared analyzer). (= **ID-2**)

**Independent Test**: Analyze six fixture sources and assert each specific
diagnostic (spec US4; quickstart §6).

### Tests (write first, observe RED)

- [ ] T017 [US4] Write analyzer tests in `internal/analyzer/context_test.go`:
      valid singleton (+OS attrs) clean; duplicate, `[confirm]`, required
      param, `(agent)` executor each rejected; defaulted param allowed
      (ID-2 steps 1–2); expect the four rejection tests RED

### Implementation

- [ ] T018 [US4] Add `checkContext()` to `internal/analyzer/analyzer.go`
      (called from `Analyze` after `collect()`; duplicate diagnostic carries a
      related location naming the first hook; add `fmt` import) per ID-2
      step 3 full listing; run
      `docker-compose run --rm test go test ./internal/analyzer/ -v` to GREEN
- [ ] T019 [US4] Commit: `feat(analyzer): validate the [context] hook contract`
      (ID-2 step 5)

**Checkpoint**: All four stories complete; spec FR-001…FR-009 covered.

---

## Phase 7: Polish & Cross-Cutting

**Purpose**: Documentation (FR-010), full-suite verification, spec success
criteria sign-off. (= **ID-6**)

- [ ] T020 [P] Document the attribute: `docs/GRAMMAR.md` AttrItem alternative,
      `docs/runefile.md` attribute list + paragraph, `docs/mcp.md` new
      "Project context for agents" section with fenced `rune` example
      (ID-6 steps 1–3, exact copy provided)
- [ ] T021 Verify docs harness:
      `docker-compose run --rm test go test ./test/docs/...` (fenced blocks
      parse, links resolve) (ID-6 step 4)
- [ ] T022 Full suite + race + lint:
      `docker-compose run --rm test go test ./...`,
      `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...`,
      and `golangci-lint run` — zero issues (constitution VIII)
- [ ] T023 Walk quickstart.md §§2–6 manually once (author hook → CLI run →
      MCP initialize probe → agent prefix → degrade → misuse diagnostic) and
      check off spec SC-001…SC-005; commit:
      `docs: document the [context] agent context-prep attribute` (ID-6
      step 6)

---

## Dependencies & Execution Order

```text
Phase 2 (Foundational: T001→T002→T003→T004)
    │
    ├──► Phase 3 US1 (T005,T006 → T007 → T008 → T009)   ── MVP
    │        │
    │        ├──► Phase 4 US2 (T010 → T011 → T012 → T013)   [needs gatherContext from T007]
    │        └──► Phase 5 US3 (T014,T015 → T016)            [tests T007/T011 behavior; T015 needs T011]
    │
    └──► Phase 6 US4 (T017 → T018 → T019)               [independent of US1–US3]
                 │
Phase 7 (Polish: T020 → T021 → T022 → T023)  [after all stories]
```

- US4 depends only on the Foundational phase — it can run in parallel with
  US1–US3 (different packages).
- US3's unit tests (T014) exercise code landed in US1 (T007); its integration
  test (T015) needs US2's wiring (T011).
- Within every story: tests RED before implementation (constitution VI).

## Parallel Opportunities

- **T005 ∥ T006**: different test files/packages (`internal/cli` vs
  `mcpserver`).
- **Whole of US4 (T017–T019) ∥ US1–US3**: `internal/analyzer` touches no file
  the other stories touch.
- **T014 ∥ T015**: different files (`internal/cli` unit vs `test/integration`).
- **T020 ∥ T022 prep**: docs edits are independent of the code suites.

## Implementation Strategy

**MVP first**: Foundational + US1 (T001–T009) yields the headline capability
— MCP agents start informed — independently shippable and demonstrable via
quickstart §3. Then US2 (prompt prefix), US3 (resilience proof), US4 (static
validation), Polish. One commit per story slice as listed; every commit
leaves the tree green (tests run in Docker before each).
