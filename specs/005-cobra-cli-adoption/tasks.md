# Tasks: Modern CLI Interface

**Input**: Design documents from `/specs/005-cobra-cli-adoption/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, quickstart.md

**Tests**: REQUIRED. Constitution Principle VI (Test-First, Multi-Layer) is NON-NEGOTIABLE
and quickstart.md enumerates the tests to write. Every story writes failing tests first.

**Organization**: Tasks are grouped by user story. Stories are independently testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1 / US2 / US3 (Setup, Foundational, Polish carry no story label)
- All paths are repository-relative. Tests run **in Docker only** (`docker-compose run --rm test …`).

## Conventions for this feature

- Binary lives at `cmd/rune/` (relocated from `rune/` — see plan.md path-reconciliation).
- `internal/cli` MUST NOT import `cobra` (Constitution VIII). `package main` (cmd/rune)
  adapts `cli` types to Cobra types.
- No `os.Exit` inside any `RunE`; return `cli.*` error types and let `main` map exit codes
  via `cli.CodeFor` (golang-cli skill).
- Behavior contract (exit codes, stdout/stderr, task pass-through, diagnostics, color,
  signals) is preserved exactly (SC-006).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Reconcile the binary location so the build tooling and test harnesses work.

- [x] T001 Relocate the `main` package to `cmd/rune/` to match all build tooling and the `golang-cli` `cmd/<app>/` layout: `git mv rune/main.go cmd/rune/main.go` and `git mv rune/completion.go cmd/rune/completion.go`; remove the now-empty `rune/` dir; confirm `go build ./cmd/rune` succeeds (it fails today — `./cmd/rune` is referenced by Runefile, `.goreleaser.yaml`, `Dockerfile`, `test/integration/harness_test.go`, `test/docs/harness_test.go` but does not exist).
- [x] T002 Establish the red baseline in Docker: `docker-compose run --rm test go test ./...` — confirm `test/integration` and `test/docs` now **build** after T001, and record which (if any) tests are red before changes so regressions are detectable (SC-006 guardrail).

**Checkpoint**: `go build ./cmd/rune` works; the test harnesses build; baseline captured.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Replace hand-rolled dispatch with an idiomatic Cobra command tree —
**behavior-preserving** — so every story has real subcommands to build on.

**⚠️ CRITICAL**: No user-story work can begin until this phase is complete.

- [x] T003 Create `cmd/rune/root.go` with `newRootCmd(opts *cli.Options, version, commit string) *cobra.Command`: move root construction out of `main.go` — all global flags (`-f/--file`, `--list`, `--dry-run`, `--summary`, `--dump`, `--format`, `--set`, `--watch`, `--choose`, `--yes`, `--quiet`, `--fmt`, `--clear-cache`), `SilenceUsage:true`, `SilenceErrors:true`, `Version`, `Flags().SetInterspersed(false)`, `SetOut/SetErr`, and `RunE` that populates `opts` and calls `cli.Run(opts, args)` (the task path). Behavior identical to today.
- [x] T004 Slim `cmd/rune/main.go` to: `signal.NotifyContext`, build `cli.Options`, `newRootCmd(...)`, register subcommands (T005, T006), `root.ExecuteContext(ctx)`, and the existing terminal error-banner mapping (suppress for `*cli.ValidationError` and silent `*cli.TaskFailure`) + `os.Exit(cli.CodeFor(err))`. Depends on T003.
- [x] T005 [P] Create `cmd/rune/serve.go` with `newServeCmd(opts *cli.Options) *cobra.Command`: `Use:"serve"`, `Aliases:[]string{"mcp"}`, flags `--http` (bool), `--addr` (string), `--token-file` (string), `--mcp` (bool), and `RunE` that calls `cli.ServeMCP(opts, http, addr, tokenFile)`. Replaces the manual `runServe` arg loop from the old `main.go`. (Flag *validation* is added in US3.) Depends on T003.
- [x] T006 [P] Create `cmd/rune/version.go` with `newVersionCmd(version, commit string) *cobra.Command` printing the version+commit string to `cmd.OutOrStdout()`. Depends on T003.
- [x] T007 In the root/main wiring, remove the manual `mcp`/`serve`/`completion` positional dispatch from the old `RunE`, and **delete the custom `genCompletion`** (the relocated `cmd/rune/completion.go`) — Cobra auto-adds the `completion` command once subcommands exist. Verify `rune --help` lists subcommands and `rune completion bash|zsh|fish|powershell` emits scripts. Depends on T004, T005, T006.

**Checkpoint**: `rune <task>`, `rune mcp`, `rune serve --http …`, `rune --list`, exit codes,
and stdout/stderr behave exactly as before; existing suite green in Docker.

---

## Phase 3: User Story 1 - Discover & understand every command from help (Priority: P1) 🎯 MVP

**Goal**: `rune --help` lists every built-in command with a description and points to
`--list` for tasks; each `rune <cmd> --help` shows usage, flags, and an example;
`rune version` matches `rune --version`.

**Independent Test**: Run `rune --help`, `rune serve --help`, `rune version`,
`rune --version`; confirm discoverability, per-command help with examples, and version parity.

### Tests for User Story 1 ⚠️ (write first, must FAIL)

- [ ] T008 [P] [US1] Integration test `test/integration/cli_help_test.go` (uses the existing `run`/`writeRunefile` harness): `rune --help` lists `serve`, `version`, `completion`, `help` and references `--list`; `rune serve --help` shows the flags, an example, and `Aliases:` includes `mcp`; `rune version` stdout == `rune --version` stdout; assert exit codes.

### Implementation for User Story 1

- [ ] T009 [US1] In `cmd/rune/root.go`, add `Short`, `Long`, and `Example` to the root command; ensure top-level help distinguishes built-in commands from tasks and tells users to run `rune --list` (FR-003).
- [ ] T010 [P] [US1] In `cmd/rune/serve.go`, add `Short`, `Long`, and `Example` covering both stdio (`rune mcp`/`rune serve`) and HTTP (`rune serve --http --addr …`) usage (FR-002).
- [ ] T011 [P] [US1] In `cmd/rune/version.go`, add `Short`, `Long`, and `Example`; route output through `cmd.OutOrStdout()` (FR-002).
- [ ] T012 [US1] Add a shared version formatter (e.g. `versionString(version, commit string)` in `cmd/rune/version.go`) used by both `newVersionCmd` and `root.Version` in `cmd/rune/root.go`, so `rune version` and `rune --version` are byte-identical (FR-004). Depends on T011.

**Checkpoint**: US1 fully functional — commands are discoverable and documented (SC-001/SC-002/SC-004).

---

## Phase 4: User Story 2 - Shell completion, including live task names (Priority: P2)

**Goal**: `rune completion {bash|zsh|fish|powershell}` installs; pressing TAB completes
built-in commands, global flags, and the current Runefile's task names with their doc
summaries; graceful when no/invalid Runefile.

**Independent Test**: `rune __complete ""` lists commands + task names (with descriptions);
`rune __complete ""` in an empty dir lists commands only with no error; `rune completion zsh`
emits a script whose `--help` documents installation.

### Tests for User Story 2 ⚠️ (write first, must FAIL)

- [ ] T013 [P] [US2] Unit test `internal/cli/complete_test.go`: `TaskCandidates` returns only non-private, OS-matching tasks each with its first doc line; returns `nil` (no error, no panic) when the Runefile is missing or unparseable.
- [ ] T014 [P] [US2] Integration test `test/integration/cli_completion_test.go`: `rune __complete ""` includes built-in command names and Runefile task names with descriptions and ends with the `ShellCompDirectiveNoFileComp` directive; in a dir with no/broken Runefile it returns commands only and emits **no** error line (FR-012); `rune completion zsh` exits 0 and prints a script.

### Implementation for User Story 2

- [ ] T015 [US2] Implement `internal/cli/complete.go`: type `TaskCandidate{Name, Doc string}` and `func TaskCandidates(opts Options) []TaskCandidate` — resolve (`config.Resolve`), parse + compose (`parser.Parse`/`config.Compose`), filter via `IsPrivate()`/`osMatches`, take `firstLine(t.Doc)`; **skip the analyzer**; return `nil` on any error. Cobra-free.
- [ ] T016 [US2] In `cmd/rune/root.go`, set `root.ValidArgsFunction` to call `cli.TaskCandidates(opts)` and map each to `cobra.CompletionWithDesc(c.Name, c.Doc)`, returning `cobra.ShellCompDirectiveNoFileComp`; never write to stdout from the function. Depends on T015.
- [ ] T017 [US2] Reconcile existing completion coverage in `test/docs/cli_test.go` (and any doc CLI reference) with Cobra's `completion <shell>` surface for bash/zsh/fish/powershell; update fixtures/assertions that assumed the old custom `genCompletion` (e.g. `rune completion` with no shell arg).

**Checkpoint**: US2 fully functional — dynamic, described task-name completion on all four shells (SC-003).

---

## Phase 5: User Story 3 - Idiomatic, validated subcommands & friendly errors (Priority: P3)

**Goal**: `serve` validates its flags; mistyped commands get a "did you mean" suggestion;
errors are concise (no full-usage dump); the `--` escape hatch runs a colliding task.

**Independent Test**: `rune serve --addr :7777` (no `--http`) → exit 2; `rune serv` →
"did you mean serve?" exit 2; `rune -- serve` runs the task named `serve`.

### Tests for User Story 3 ⚠️ (write first, must FAIL)

- [ ] T018 [P] [US3] Unit test `internal/cli/suggest_test.go`: table-driven `nearest` — `serv→serve`, `tset→test`, within-threshold matches, and no suggestion when distance exceeds the threshold.
- [ ] T019 [P] [US3] Unit test `cmd/rune/serve_test.go` (package main): serve flag validation — `--addr`/`--token-file` without `--http` returns a usage error; valid combinations pass.
- [ ] T020 [P] [US3] Integration test `test/integration/cli_errors_test.go`: `rune serv` → stderr contains `did you mean "serve"`, exit 2, **no** usage block; `rune serve --addr :1` → exit 2; with a Runefile defining a task named `serve`, `rune -- serve` runs the **task** (asserts task output) while `rune serve` runs the **server** (FR-008 routing — the one empirical risk from research.md D2).

### Implementation for User Story 3

- [ ] T021 [P] [US3] Add `Commands []string` to `cli.Options` in `internal/cli/dispatch.go`; populate it in `cmd/rune/main.go` with the reserved command names (`serve`, `mcp`, `completion`, `help`, `version`).
- [ ] T022 [P] [US3] Implement `internal/cli/suggest.go`: `nearest(token string, candidates []string) (string, bool)` using Levenshtein distance with threshold `min(2, ⌊len/3⌋)` — cobra-free.
- [ ] T023 [US3] In `internal/cli/args.go`, enhance the `splitArgs` `unknown task: <tok>` error to append `(did you mean "<nearest>"?)`, where candidates = Runefile task names ∪ `opts.Commands`. Keep exit code 2. Depends on T021, T022.
- [ ] T024 [US3] In `cmd/rune/serve.go`, add `RunE` validation: `--addr`/`--token-file` without `--http` → `&cli.UsageError{}` (exit 2, FR-015); register `--token-file` file completion via `RegisterFlagCompletionFunc`. Depends on T005.
- [ ] T025 [US3] Confirm `rune -- <task>` routes to the task path; if Cobra routes `rune -- serve` to the `serve` subcommand, detect `cmd.ArgsLenAtDash() == 0` in root `RunE` / adjust registration so the task path wins. Validated by T020.

**Checkpoint**: All three stories independently functional; built-in precedence + `--` escape verified.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T026 [P] Update CLI documentation for the new command surface and completion install steps: `docs/` CLI reference, `README.md`, `CONTRIBUTING.md` (technical-writer skill); keep `docs-check` green.
- [ ] T027 [P] Format & lint: `rune fmt` then `golangci-lint run` — gofumpt/goimports clean, no unhandled errors, no new globals/`init()` (Constitution VIII).
- [ ] T028 Full verification in Docker: `docker-compose run --rm test go test ./...` and `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...` — green on the whole suite (SC-006).
- [ ] T029 Execute quickstart.md scenarios S1–S6 against the built binary and confirm expected stdout/stderr/exit for each.
- [ ] T030 [P] Cross-platform CI smoke: confirm the suite builds/passes on Linux, macOS, Windows (Constitution VI) — completion scripts generate on all four shells.

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: start immediately; T001 → T002.
- **Foundational (Phase 2)**: after Setup. T003 → T004; T005, T006 [P] after T003; T007 after T004/T005/T006. **BLOCKS all stories.**
- **User Stories (Phase 3–5)**: all require Foundational complete. After that they are independent and may proceed in parallel or in priority order P1 → P2 → P3.
- **Polish (Phase 6)**: after the desired stories are complete.

### Story independence

- **US1 (P1)**: needs only Foundational. Pure help/version enrichment.
- **US2 (P2)**: needs only Foundational. Independent of US1/US3 (touches `complete.go` + root `ValidArgsFunction`).
- **US3 (P3)**: needs only Foundational. Independent of US1/US2 (touches `suggest.go`, `args.go`, `dispatch.go`, serve validation).

### Within each story

- Write tests first and see them FAIL, then implement (Constitution VI).
- `complete.go` (T015) before wiring `ValidArgsFunction` (T016).
- `Options.Commands` (T021) + `nearest` (T022) before the suggestion in `args.go` (T023).

### Parallel opportunities

- T005 ∥ T006 (foundational subcommand files).
- T010 ∥ T011 (US1 serve/version help).
- T013 ∥ T014 (US2 tests); T018 ∥ T019 ∥ T020 (US3 tests).
- T021 ∥ T022 ∥ T024 (US3 implementation, different files).
- Whole stories US1 ∥ US2 ∥ US3 once Foundational is done (different files, no cross-deps).

---

## Parallel Example: User Story 3

```bash
# Tests first (different files, all should FAIL):
Task: "Unit test nearest() in internal/cli/suggest_test.go"
Task: "Unit test serve flag validation in cmd/rune/serve_test.go"
Task: "Integration test errors+escape in test/integration/cli_errors_test.go"

# Then independent implementation files in parallel:
Task: "Add Options.Commands in internal/cli/dispatch.go"
Task: "Implement nearest() in internal/cli/suggest.go"
Task: "serve flag validation + token-file completion in cmd/rune/serve.go"
```

---

## Implementation Strategy

### MVP first (US1 only)

1. Phase 1 Setup (relocate to `cmd/rune/`).
2. Phase 2 Foundational (Cobra command tree, behavior-preserving).
3. Phase 3 US1 (discoverable, documented commands).
4. **STOP & VALIDATE**: `rune --help`, `rune serve --help`, `rune version` — demo MVP.

### Incremental delivery

- Setup + Foundational → behavior preserved, suite green.
- + US1 → discoverable help (MVP). + US2 → dynamic completion. + US3 → validation + friendly errors.
- Each story is demoable without breaking the previous.

---

## Notes

- [P] = different files, no incomplete-task dependency.
- The single empirical risk (Cobra routing of `rune -- serve`) is covered by T020/T025.
- Constitution VI is non-negotiable: tests precede implementation in every story.
- Commit after each task or logical group (when the user asks).
- `internal/cli` stays cobra-free; the framework boundary is the `cmd/rune` wiring.
