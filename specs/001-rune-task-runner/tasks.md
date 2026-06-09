---
description: "Task list for Rune — A Shared Task Runner for Humans and AI Agents"
---

# Tasks: Rune — A Shared Task Runner for Humans and AI Agents

**Input**: Design documents from `/specs/001-rune-task-runner/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅ (cli, grammar, mcp-tools, cache-fingerprint), quickstart.md ✅

**Tests**: INCLUDED. Constitution Principle VI ("Test-First, Multi-Layer Verification") is NON-NEGOTIABLE and plan.md commits to TDD (table-driven unit tests, golden files for AST/formatter, binary-level integration tests, Go native fuzz targets, and a compatibility corpus). Test tasks are therefore generated and MUST be written first (Red) within each story.

**Organization**: Tasks are grouped by user story (P1–P5). The delivered MVP is **US1 + US2** (a `just`-class shell runner with up-front static validation), per plan.md phased delivery. US3–US5 complete the v1 identity.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: US1–US5 (user-story phases only; Setup/Foundational/Polish carry no label)
- Exact file paths are included in every task

## Path Conventions

Single Go module rooted at the repo (per plan.md "Project Structure"). One binary `cmd/rune`; compiler/runtime logic in `internal/` focused packages; public `mcpserver/`; fixtures in `testdata/`; binary-level integration tests in `test/`.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, dependencies, and CI/release scaffolding.

- [X] T001 Initialize the Go module (`go 1.24` minimum) and create the package directory skeleton per plan.md (`cmd/rune/`, `internal/{token,lexer,ast,parser,analyzer,diag,eval,config,dotenv,cli}`, `internal/runtime/{scheduler,shell,interp,agent}`, `internal/cache`, `mcpserver/`, `testdata/`, `test/`, `docs/`) in `go.mod`
- [X] T002 Add and pin primary dependencies in `go.mod` (`mvdan.cc/sh/v3`, `github.com/modelcontextprotocol/go-sdk`, `github.com/spf13/cobra`, `github.com/fsnotify/fsnotify`, `golang.org/x/sync/errgroup`, `github.com/fatih/color`) — MCP SDK added at US4
- [X] T003 [P] Configure `gofmt`/`go vet`/golangci-lint in `.golangci.yml`
- [X] T004 [P] Create CI workflow (lint + full test suite matrix on Linux/macOS/Windows + fuzz smoke) in `.github/workflows/ci.yml`
- [X] T005 [P] Create cross-platform static-binary release config (`CGO_ENABLED=0`, Linux/macOS/Windows × amd64/arm64, checksums) in `.goreleaser.yaml`
- [X] T006 [P] Add `.gitignore` excluding `.rune/` (cache) and `dist/` (build output) in `.gitignore`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared compiler primitives and CLI scaffold that every user story depends on.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T007 [P] Implement `Position` (byte offset + 1-based line/col) and `Span` (`{file, start, end}`) in `internal/token/position.go`
- [X] T008 [P] Define token kinds (keywords `set`/`import`/`mod`, sigils `@ - + * :=` `:` `&&` `( )` `[ ]`, `INDENT`/`DEDENT`/`NEWLINE`, string/comment literals, operators `+ / == != =~`) in `internal/token/token.go`
- [X] T009 Define all AST node types with a `Span` on every node (`File`, `Setting`, `Assignment`, `Task`, `Param`, `DepCall`, `Attribute`, `BodyLine`, `Import`, `Mod`, and `Expr` variants: `StringLit`/`Concat`/`PathJoin`/`Conditional`/`Comparison`/`RegexMatch`/`FuncCall`/`VarRef`/`ParamRef`) in `internal/ast/ast.go` (depends on T007)
- [X] T010 Implement the diagnostic model (`Diagnostic{severity, span, message, snippet}`) and the renderer producing `file:line:col` + source line + caret-underline span (Principle II) in `internal/diag/diagnostic.go` and `internal/diag/render.go` (depends on T007)
- [X] T011 [P] Implement Runefile discovery: walk current dir + ancestors for `Runefile`/`.runefile` (case-insensitive), `--file` override, "no Runefile found" → exit 2 in `internal/config/discover.go` (depends on T010)
- [X] T012 [P] Define exit-code constants (0 success, 1 task failure, 2 usage/no-Runefile/unknown-task, 3 static-validation, 130 interrupt) and the central error→exit-code mapping in `internal/cli/exit.go`
- [X] T013 Scaffold the cobra root command with the dynamic task dispatcher entrypoint, `-f/--file`, `--version`, `-h/--help`, stdout/stderr stream discipline (Rune messages → stderr) and NO_COLOR-aware color init in `cmd/rune/main.go` (depends on T011, T012)

**Checkpoint**: Compiler primitives + CLI skeleton ready — user-story implementation can begin.

---

## Phase 3: User Story 1 - Author and run project commands from one readable file (Priority: P1) 🎯 MVP

**Goal**: A usable command runner — define named tasks with bodies, run any by name, pass params (with defaults/variadic), declare dependencies that run first (each once), a default task, and upward Runefile discovery from any subdirectory.

**Independent Test**: Write a Runefile with two tasks where one depends on the other; run the dependent task; confirm both run in the correct order, each exactly once, params substituted, non-zero exit if a command fails.

### Tests for User Story 1 (write first — Red) ⚠️

- [X] T014 [P] [US1] Lexer table-driven unit tests + golden token streams (settings, assignment, task signature, params, deps, body lines, `INDENT`/`DEDENT`, tab/space-mix located error) in `internal/lexer/lexer_test.go` and `testdata/lexer/`
- [X] T015 [P] [US1] Parser unit tests + golden AST dumps for a representative Runefile in `internal/parser/parser_test.go` and `testdata/parser/`
- [X] T016 [P] [US1] Expression (Pratt) parser + evaluator unit tests (string lits, concat `+`, path-join `/`, var/param refs, conditionals, comparisons `== != =~`, builtins) in `internal/eval/eval_test.go`
- [X] T017 [P] [US1] Scheduler unit tests (topo order, run-once memoization across multiple dependency paths, fail-fast) in `internal/runtime/scheduler/scheduler_test.go`
- [X] T018 [P] [US1] Integration tests (compiled binary) for Quickstart Scenario 1 — `rune` (default `greet`), `rune greet Ada`, `rune build` (dep runs once), `rune --list` — asserting stdout/stderr/exit in `test/integration/us1_test.go` and `testdata/us1/Runefile`
- [X] T019 [P] [US1] Go native fuzz targets for the lexer and parser in `internal/lexer/fuzz_test.go` and `internal/parser/fuzz_test.go`

### Implementation for User Story 1

- [X] T020 [US1] Implement the Pike-style state-function lexer: comments, settings, names, sigils, string literals (single/double/triple with de-dent), `{{ }}` interpolation spans, and `INDENT`/`DEDENT`/`NEWLINE` with consistent-indentation enforcement (tab/space mix within one body = located error) in `internal/lexer/lexer.go` (depends on T008, T010)
- [X] T021 [US1] Implement the recursive-descent parser for declarative items: `Setting`, `Assignment`, `Task` signature (name, params, executor), `Deps`, `&&` post-hooks, `Attribute` blocks, and body lines (`@`/`-` sigils); attach doc from the preceding comment run in `internal/parser/parser.go` (depends on T009, T020)
- [X] T022 [US1] Implement the Pratt expression sub-parser (concat `+`, path-join `/`, comparisons `== != =~`, `if/else if/else`, function calls, grouping) feeding the same AST in `internal/parser/expr.go` (depends on T021)
- [X] T023 [P] [US1] Implement the total tree-walking evaluator: `Expr` → string/bool `Value`; param/variable resolution; run-time variable overrides (`NAME=VALUE` / `--set`) consulted before file assignments in `internal/eval/eval.go` (depends on T009)
- [X] T024 [P] [US1] Implement `{{ … }}` body interpolation (with `{{{{` brace escaping) using the evaluator in `internal/eval/interp.go` (depends on T023)
- [X] T025 [US1] Implement the builtin function registry (FR-007: `env`, `os`, `arch`, `os_family`, `num_cpus`, path ops `join`/`clean`/`extension`/`file_name`/`file_stem`/`parent_dir`/`absolute_path`, string ops `uppercase`/`lowercase`/`trim`/`replace`/`replace_regex`/case-conversions, `path_exists`, `read`, `sha256`/`sha256_file`, `uuid`, `datetime`, `error`, `require`/`which`, `quote`) — total, no loops/recursion — in `internal/eval/builtins.go` (depends on T023)
- [X] T026 [US1] Implement variable-assignment evaluation and minimal settings resolution (`default`, `working-directory`, `quiet`, `export`) needed to run in `internal/config/settings.go` (depends on T023)
- [X] T027 [US1] Implement the dependency scheduler: build the DAG from deps/post-hooks, topo-sort, memoize `(namespace, task, canonical-args)` run-once (FR-005), sequential execution with fail-fast in `internal/runtime/scheduler/scheduler.go` (depends on T021)
- [X] T028 [US1] Implement the default `(sh)` executor via `mvdan.cc/sh/v3` (`syntax` parse + `interp` run; NEVER the system shell — Principle V), streaming stdout/stderr, honoring `@` (no-echo) / `-` (continue-on-error) per-line sigils, surfacing failing task/line + exit code in `internal/runtime/shell/shell.go` (depends on T024)
- [X] T029 [US1] Wire the CLI dispatcher: resolve Runefile → lex → parse → run task(s) by name, default task on no-arg, `VAR=VALUE` overrides, task chaining; add `--list` rendering of non-private tasks with docs in `internal/cli/dispatch.go` and `cmd/rune/main.go` (depends on T027, T028)
- [X] T030 [US1] Implement param binding (required / defaulted / `+`one-or-more / `*`zero-or-more) for CLI invocations and dep calls, sufficient for the happy path in `internal/cli/args.go` (depends on T029)

**Checkpoint**: US1 fully functional — a working shell task runner, independently testable.

---

## Phase 4: User Story 2 - Catch mistakes before anything runs (Priority: P2)

**Goal**: Whole-file static validation before any execution — unknown task, undefined variable, dependency cycle (with path), and arity mismatch — each reported with `file:line:col` + caret span; on failure nothing runs and exit is non-zero (3).

**Independent Test**: Introduce, one at a time, an unknown dependency, an undefined variable, a dependency cycle, and a wrong-arity call; confirm each exits non-zero, runs nothing, and prints the exact location with a caret.

### Tests for User Story 2 (write first — Red) ⚠️

- [X] T031 [P] [US2] Analyzer unit tests: undefined var, unknown task/dep, dependency cycle (full path), arity mismatch, duplicate setting, variadic-last + defaulted-after-required ordering, import collision, body indentation mix in `internal/analyzer/analyzer_test.go`
- [X] T032 [P] [US2] Diagnostic renderer golden tests: `file:line:col` + caret underline for each error class in `internal/diag/render_test.go` and `testdata/diag/`
- [X] T033 [P] [US2] Integration tests (compiled binary) for Quickstart Scenario 2 — `rune a` (unknown task `b`), `rune c` (self-cycle), `rune greet` (undefined var) → exit 3, nothing runs, located message + caret in `test/integration/us2_test.go` and `testdata/us2/`

### Implementation for User Story 2

- [X] T034 [US2] Implement the semantic analyzer (runs before any execution): resolve every `VarRef`/`ParamRef`; resolve every `DepCall` + CLI-invoked task name; detect cycles and report the full path; check arity at every call/dep site; enforce one-setting-per-name, variadic-last, defaulted-after-required in `internal/analyzer/analyzer.go` (depends on T021, T010)
- [X] T035 [US2] Gate execution on analysis: emit ALL diagnostics, execute nothing, exit 3 (FR-014) in `internal/cli/dispatch.go` (depends on T034)
- [X] T036 [US2] Ensure lexer/parser errors also flow through the same spanned diagnostic path and exit 3 in `internal/lexer/lexer.go` and `internal/parser/parser.go` (depends on T034)

**Checkpoint**: MVP complete (US1 + US2) — a `just`-class shell runner with trustworthy up-front validation.

---

## Phase 5: User Story 3 - Run task bodies in multiple languages (Priority: P3)

**Goal**: A task author selects the body language — default shell, or declared `python`/`node`/custom interpreter — with values interpolated into the body and an actionable error if the interpreter is missing.

**Independent Test**: Define one shell task and one Python task in the same file; run each; confirm each executes under its declared runtime with interpolated values present.

### Tests for User Story 3 (write first — Red) ⚠️

- [X] T037 [P] [US3] Executor-selection + temp-file interpreter unit tests (python/node/custom, interpolation into body, missing-interpreter error) in `internal/runtime/interp/interp_test.go`
- [X] T038 [P] [US3] Integration tests (compiled binary) for Quickstart Scenario 3 — `rune analyze` (python), `rune bundle` (node), missing-interpreter actionable error + exit 1 in `test/integration/us3_test.go` and `testdata/us3/`

### Implementation for User Story 3

- [X] T039 [US3] Define the `Executor` abstraction and selection (sh default vs `python`/`node`/custom vs `[script("…")]`), resolving the interpreter command from `set python`/`set node`/custom in `internal/runtime/executor.go` (depends on T028)
- [X] T040 [US3] Implement the temp-file + exec interpreter executor: write the interpolated body to a temp file (exec bit on Unix), exec the configured interpreter, stream output, clean up in `internal/runtime/interp/interp.go` (depends on T039, T024)
- [X] T041 [US3] Implement missing/unavailable-interpreter detection → clear actionable error naming the runtime; exit 1 (FR-017) in `internal/runtime/interp/interp.go` (depends on T040)
- [X] T042 [US3] Honor `set shell := [...]` override (switch that task to temp-file + exec of the named shell) in `internal/runtime/shell/shell.go` (depends on T039)

**Checkpoint**: US3 functional — multi-language bodies as a first-class concept.

---

## Phase 6: User Story 4 - Share the same tasks with AI agents and IDEs (Priority: P4)

**Goal**: Expose non-private tasks to external agents/IDEs as discoverable, invokable MCP tools (run through the same engine); support an `(agent)` task type that drives an installed agent CLI; gate destructive tasks; keep secrets out of everything agent-facing; local always available, remote opt-in/localhost/token-gated.

**Independent Test**: Start agent-facing mode; from a client, list tasks (names/descriptions/param shapes accurate, private hidden); invoke a safe task and get output/result; confirm a `[confirm]` task is refused without authorization.

### Tests for User Story 4 (write first — Red) ⚠️

- [X] T043 [P] [US4] MCP server unit tests: task→tool mapping (namespacing `mod__task`, description from doc, `inputSchema` from params, `destructiveHint` from `[confirm]`, `openWorldHint`), private tasks excluded, no secret values in any field in `mcpserver/server_test.go`
- [X] T044 [P] [US4] Authorization unit tests: private not exposed; non-destructive callable; destructive requires approval (`--yes`/allow-list/confirm) else refused; operator allow-list narrowing in `mcpserver/authz_test.go`
- [X] T045 [P] [US4] Transport/auth tests: stdio always available; HTTP opt-in binds `127.0.0.1`; list/call without a valid token rejected; non-localhost requires explicit `--addr` (SC-010) in `mcpserver/transport_test.go`
- [X] T046 [P] [US4] Agent-executor unit tests: prompt interpolation, in-process MCP exposure of only allowed tasks, missing/unauthenticated CLI → actionable error, never invents credentials in `internal/runtime/agent/agent_test.go`
- [X] T047 [P] [US4] Integration tests for Quickstart Scenario 4 — list tools, call `logs` via shared engine, `clean` destructive gating, HTTP without token rejected, no secret leakage, `rune triage` agent task in `test/integration/us4_test.go` and `testdata/us4/`

### Implementation for User Story 4

- [X] T048 [US4] Implement the MCP server: expose each non-private task as a tool (`mcp.AddTool`) with derived `inputSchema` (defaulted→optional, required→required, variadic→array), description, and annotations (`destructiveHint`/`openWorldHint`); namespace submodule tasks as `mod__task` in `mcpserver/server.go` (depends on T034)
- [X] T049 [US4] Implement the tool-call handler: validate args → run through the SAME scheduler as the CLI → return `{stdout, stderr, exitCode}`; guarantee no secret value surfaces in name/description/schema/result (FR-029, SC-007) in `mcpserver/handler.go` (depends on T048, T027)
- [X] T050 [US4] Implement authorization (Q3/FR-028): private excluded; non-destructive callable; destructive (`[confirm]`) gated behind explicit approval; operator allow-list narrowing in `mcpserver/authz.go` (depends on T048)
- [X] T051 [US4] Implement transports: stdio (always) + Streamable HTTP/SSE (opt-in, binds `127.0.0.1`, bearer token required before any list/call; non-localhost needs explicit `--addr`) in `mcpserver/transport.go` (depends on T048)
- [X] T052 [US4] Define the vendor-neutral `Provider` interface (`Run(ctx, prompt, toolSession, opts) -> (finalText, toolTrace, error)`) in `internal/runtime/agent/provider.go`
- [X] T053 [US4] Implement the agent-CLI provider: resolve prompt interpolation, start the in-process MCP server with only allowed tasks, invoke the configured agent CLI (`set agent_cmd := ["claude","-p"]` / codex / copilot), capture final output as the task result; missing/unauthenticated CLI → actionable error with no invented credentials (FR-027) in `internal/runtime/agent/agent.go` (depends on T052, T048, T040)
- [X] T054 [US4] Add `rune mcp` and `rune serve --mcp [--http --addr 127.0.0.1:PORT --token-file PATH]` subcommands (reserved names; warn on task shadowing) and wire the `(agent)` executor into the runtime in `cmd/rune/main.go` and `internal/runtime/executor.go` (depends on T051, T053)
- [X] T076 [US4] Add an author-declared `[network]` attribute that sets the MCP `openWorldHint` annotation (mirrors `[confirm]`→`destructiveHint`; no content heuristics — Principle VII); add it to `contracts/grammar.md`, `Attribute.kind` in `internal/ast/ast.go`, the parser, and tool derivation in `mcpserver/server.go` (resolves analyze finding U1; depends on T021, T048)

**Checkpoint**: US4 functional — the shared human + agent automation layer (v1 identity reached).

---

## Phase 7: User Story 5 - CI/CD ergonomics, speed, and composition (Priority: P5)

**Goal**: Operational conveniences — grouped listing, dry-run/preview, machine-readable dump, parallel prerequisites, opt-in input/output caching with visible decisions, predictable exit codes, and task-file composition (`import` splice + `mod` namespace).

**Independent Test**: Run list and dry-run (confirm nothing executes); configure a caching task, run it twice unchanged (second is skipped + reported), then change an input (it re-runs); confirm exit codes reflect outcomes.

### Tests for User Story 5 (write first — Red) ⚠️

- [X] T055 [P] [US5] Cache fingerprint unit tests: SHA-256 over sorted input file hashes + body + resolved vars + executor identity; skip iff hash matches AND all outputs exist; missing output forces run; corruption = miss in `internal/cache/cache_test.go`
- [X] T056 [P] [US5] Parallel scheduler tests: `[parallel]` independent deps run concurrently, errgroup bound = `num_cpus()`, run-once preserved under concurrency, first-error cancellation in `internal/runtime/scheduler/parallel_test.go`
- [X] T057 [P] [US5] Composition tests: `import` splice + collision conflict reported; `mod` namespace addressing (`name::task` / `name task`) in `internal/config/compose_test.go`
- [X] T058 [P] [US5] Integration tests for Quickstart Scenario 5 — `build-cached` 1st/2nd run (`running`/`cached`, <10% time), touch input → re-run, `rune checks` parallel, `--dry-run`, `--dump --format json` in `test/integration/us5_test.go` and `testdata/us5/`

### Implementation for User Story 5

- [X] T059 [US5] Implement cache fingerprinting + storage: compute the SHA-256 fingerprint, store JSON under `.rune/cache/<sanitized-key>.json`, skip-or-run decision (hash match + outputs exist), update on success, treat corruption/unreadable as a miss in `internal/cache/cache.go` (depends on T025, T027)
- [X] T060 [US5] Wire `[cache(inputs=[…], outputs=[…])]` into the scheduler with visible `cached:`/`running:` stderr notices — no silent skipping, no timestamp decisions (Principle I) in `internal/runtime/scheduler/scheduler.go` (depends on T059)
- [X] T061 [US5] Implement bounded parallel execution for `[parallel]` deps via `errgroup.SetLimit(num_cpus())`, preserving memoization + fail-fast cancellation in `internal/runtime/scheduler/parallel.go` (depends on T027)
- [X] T062 [US5] Implement `import` (splice, optional `import?`, name collision = reported conflict) and `mod` (namespaced child module inheriting parent env, own settings) resolution in `internal/config/compose.go` (depends on T021, T034)
- [X] T063 [US5] Implement `--dry-run` (print resolved plan + would-be cache decision, run nothing) and `--summary` (task names, one per line) in `internal/cli/dispatch.go` and `cmd/rune/main.go` (depends on T060)
- [X] T064 [US5] Implement `--dump [--format json]` machine-readable parse output in `internal/cli/dump.go` (depends on T021)
- [X] T065 [US5] Implement task listing with `[group]` grouping + OS-filter exclusion (`[linux]`/`[macos]`/`[windows]`/`[unix]`, private hidden) and the "unavailable on this OS" explanation when explicitly requested in `internal/cli/dispatch.go` (depends on T029)
- [X] T066 [US5] Implement `--watch` re-run on file changes via `fsnotify` in `internal/cli/watch.go` (depends on T029)
- [X] T067 [US5] Implement `[working-directory]`/`no-cd`/`[env]` per-task attributes and `set dotenv` loading via the `mvdan/sh` shell package in `internal/runtime/executor.go` and `internal/dotenv/dotenv.go` (depends on T039)
- [X] T077 [US5] Implement `rune --clear-cache` to remove the project-local `.rune/cache/` directory in `internal/cache/cache.go` and `cmd/rune/main.go` (resolves analyze finding G2; depends on T059)

**Checkpoint**: US5 functional — Rune is pleasant in CI and at scale.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Hardening, documentation, and v1 finishing touches that span stories.

- [X] T068 [P] Generate `docs/GRAMMAR.md` from `contracts/grammar.md` and add the per-PR sync discipline note (Constitution: DSL changes update GRAMMAR.md + fixtures) in `docs/GRAMMAR.md`
- [X] T069 [P] Add the compatibility-corpus guard test (re-parse known fixtures; fail on silent grammar drift — Principle VI / FR-033) in `test/corpus/corpus_test.go` and `testdata/corpus/`
- [X] T070 [P] Generate cobra shell completions (bash/zsh/fish) in `cmd/rune/completion.go`
- [X] T071 [P] Implement `--choose` interactive task picker (shell out to `fzf` if present, else a minimal built-in picker) in `internal/cli/choose.go`
- [X] T072 [P] Implement `--quiet` / `--yes` global behaviors end-to-end and consume the `[no-exit-message]` attribute (from T081) for error-banner suppression (does not hide the exit code) in `internal/cli/dispatch.go`
- [X] T073 [P] Finalize `.goreleaser.yaml` multi-platform build + checksums and verify the `CGO_ENABLED=0` single static binary on every target (FR-032) in `.goreleaser.yaml`
- [X] T074 [P] Scaffold the backward-compatibility version pragma (`set rune_version`) and a test that default interpretation is unchanged under upgrades (FR-033) in `internal/config/version.go`
- [X] T075 Run the full `quickstart.md` validation (all 5 scenarios) against the compiled binary on Linux/macOS/Windows and reconcile any gaps (per `quickstart.md`)
- [X] T078 [P] Formatter golden tests for canonical `--fmt` output (Constitution Principle VI mandates golden files for formatter output) in `internal/cli/fmt_test.go` and `testdata/fmt/` (resolves analyze finding C1)
- [X] T079 Implement `--fmt` canonical Runefile rewrite (stable formatting of settings/assignments/tasks/attributes/bodies) in `internal/cli/fmt.go` (resolves analyze finding C1; depends on T021, T078)
- [X] T080 [P] Implement SIGINT handling: cancel the scheduler context, terminate child processes, and exit 130 (per `contracts/cli.md` exit codes) in `internal/cli/dispatch.go` (resolves analyze finding G1; depends on T027)
- [X] T081 Add the `[no-exit-message]` attribute to the grammar and `Attribute.kind` (sync `contracts/grammar.md` + `data-model.md`) and parse it in `internal/ast/ast.go` and `internal/parser/parser.go` (resolves analyze finding I1; depends on T021)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **User Stories (Phases 3–7)**: All depend on Foundational completion.
  - US1 (P1) is the base layer; US2–US5 build on US1's lexer/parser/AST/scheduler.
  - Recommended order is priority order (US1 → US2 → US3 → US4 → US5).
- **Polish (Phase 8)**: Depends on the desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: Depends only on Foundational. The irreducible core / MVP base.
- **US2 (P2)**: Builds on US1's parser + AST (analyzes the parsed file before US1's scheduler runs). Independently testable via error-injection.
- **US3 (P3)**: Builds on US1's executor + interpolation (adds the `Executor` abstraction). Independently testable with a python+shell file.
- **US4 (P4)**: Builds on US1's scheduler (shared engine) and US2's analyzer (validated tools) + US3's interpreter exec (agent driver). Independently testable via an MCP client.
- **US5 (P5)**: Builds on US1's scheduler + US3's executor. Independently testable via caching/parallel/dump/compose scenarios.

### Within Each User Story

- Tests are written FIRST and MUST fail before implementation (Constitution Principle VI — Red-Green-Refactor).
- Lexer → Parser → Evaluator → Scheduler → Executor → CLI wiring.
- Story complete before moving to the next priority.

### Parallel Opportunities

- Setup: T003, T004, T005, T006 in parallel.
- Foundational: T007 + T008 in parallel; T011 + T012 in parallel (after T010).
- All test tasks within a story (`[P]`) run in parallel and before that story's implementation.
- US1 implementation: T023 + T024 are parallel-friendly relative to the parser line (different files).
- Once Foundational is done and US1 lands, US2/US3 can be staffed in parallel by different developers; US4/US5 follow once their US1/US2/US3 dependencies are in.

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together (write first, expect Red):
Task: "Lexer unit tests + golden token streams in internal/lexer/lexer_test.go"
Task: "Parser unit tests + golden AST dumps in internal/parser/parser_test.go"
Task: "Expression parser + evaluator unit tests in internal/eval/eval_test.go"
Task: "Scheduler unit tests in internal/runtime/scheduler/scheduler_test.go"
Task: "Scenario-1 integration tests in test/integration/us1_test.go"
Task: "Fuzz targets in internal/lexer/fuzz_test.go and internal/parser/fuzz_test.go"

# Parallel-friendly implementation files (different packages):
Task: "Evaluator in internal/eval/eval.go"
Task: "Body interpolation in internal/eval/interp.go"
```

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1 (Setup) and Phase 2 (Foundational).
2. Complete Phase 3 (US1) → **STOP and VALIDATE** the Independent Test (a working shell runner).
3. Complete Phase 4 (US2) → validate static-validation error injection (exit 3, located diagnostics, nothing runs).
4. This is the shippable MVP — a `just`-class runner with up-front validation (plan.md phased delivery).

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. US1 → test independently → demo (MVP base).
3. US2 → test independently → demo (MVP complete).
4. US3 → multi-language bodies → demo.
5. US4 → AI/agent + MCP surface → demo (v1 identity).
6. US5 → CI/CD ergonomics + composition → demo.
7. Polish (Phase 8) → harden, document, release.

### Parallel Team Strategy

After Foundational completes: one developer drives US1 (the shared base) to completion, then US2/US3 can proceed in parallel, with US4 (depends on US2/US3) and US5 (depends on US3) following.

---

## Notes

- `[P]` = different files, no dependency on incomplete tasks.
- `[Story]` label maps each task to a user story for traceability (Setup/Foundational/Polish carry none).
- Every user story is independently completable and testable.
- Verify tests FAIL before implementing (Principle VI — non-negotiable).
- The default `(sh)` executor MUST NOT shell out to the system shell (Principle V).
- The expression sublanguage MUST stay total — no loops/recursion (Principle III).
- Every static-detectable error MUST carry `file:line:col` + caret and be reported before any side effect (Principle II).
- Secrets come from the environment / agent CLI session only — never the Runefile, never an agent-facing field (Principle VII).
- Any PR that changes DSL surface MUST update `docs/GRAMMAR.md` + golden/integration fixtures in the same PR.
- **T076–T081** were added by `/speckit-analyze` remediation (findings C1/G1/U1/I1/G2); they are placed within their logical phases — execution order follows each task's stated dependencies, not its numeric ID.
