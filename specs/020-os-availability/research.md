# Research: OS Availability Enforcement

**Feature**: 020-os-availability | **Date**: 2026-08-12

No `NEEDS CLARIFICATION` markers existed in the Technical Context — the
codebase already contains the attribute grammar, the predicate logic, and
all affected surfaces, so research consisted of locating them and deciding
where each new behavior belongs. Decisions below.

## D1: Where the availability predicate lives

- **Decision**: Move the `osMatches` logic (currently
  `internal/cli/run.go:729`, unexported) to `internal/ast` as
  `(*Task).AvailableOn(goos string) bool`, plus a small
  `(*Task).OSFilters() []string` helper returning the declared OS attribute
  kinds (needed for the error message and reusable by future tooling).
  Delete `osMatches`; all call sites (`hasVisibleTasks`,
  `visibleTasksByGroup`, `complete.go`) switch to the method.
- **Rationale**: `ast.Task` already owns task classification
  (`IsPrivate()`, `Attr()`); every consumer (CLI, scheduler, MCP adapter,
  potentially analyzer/LSP later) already imports `internal/ast`; the
  method takes `goos` as a parameter, satisfying FR-007 (any-OS evaluation
  from any host) exactly like the existing injectable `eval.Scope.GOOS`.
  Note `Attr(kind)` returns only the first match, which is why the
  predicate loops over `Attributes` itself — OR semantics need all of them.
- **Alternatives considered**: (a) a new `internal/availability` package —
  rejected: one function does not justify a package, and the layout is
  constitutionally locked; (b) keep it in `internal/cli` and export —
  rejected: `internal/runtime/scheduler` must not import the CLI layer
  (dependency direction would invert).

## D2: How dependency/post-hook skipping is implemented

- **Decision**: Extend the scheduler's `Engine` interface
  (`internal/runtime/scheduler/scheduler.go:19`) with
  `Available(task *ast.Task) bool`. In `runDep` (used by both the serial
  loop and `runDepsParallel`, and by the post-hook loop), after
  `ResolveDep` returns the target, return `nil` (skip) when
  `!engine.Available(target)`. The skip does not write a memo entry — the
  memo table records only executed outcomes.
- **Rationale**: `runDep` is the single choke point every dependency and
  post-hook flows through, so one check covers serial deps, `[parallel]`
  deps, and post-hooks (FR-004, Story 3 scenario 3). Putting the decision
  behind the `Engine` interface keeps the scheduler free of OS knowledge
  and keeps the predicate injectable for tests. The CLI `engine` struct
  (sole implementation, also used by the MCP `Call` path) implements it as
  `t.AvailableOn(e.goos)` with `goos` defaulting to `runtime.GOOS`.
- **Alternatives considered**: (a) sentinel return from `ResolveDep` (nil
  task = skip) — rejected: overloads an error-reporting path with control
  flow and hides the rule; (b) pre-filtering `task.Deps` in the CLI before
  scheduling — rejected: post-hooks and transitively-resolved deps would
  need duplicate filtering, and the scheduler's dry-run/summary plan modes
  would disagree with real runs; (c) skipping inside `state.run` — works,
  but `runDep` is earlier and avoids touching cycle-detection bookkeeping.
- **Interaction with cycle detection**: skipping happens before `run` is
  entered, so a skipped task contributes no chain entry; a cycle that is
  only reachable through an unavailable task is therefore not reported on
  this host. That matches "the task does not exist here" semantics and the
  static analyzer still catches structural cycles independently.

## D3: Where direct invocation is rejected, and how it reports

- **Decision**: Gate in `resolveRoots` (`internal/cli/run.go:201`), which
  turns CLI invocations into scheduler roots: any root task failing
  `AvailableOn` aborts the whole invocation with a `ValidationError`
  (exit code 3 — `ExitValidation`, "nothing executed"), message shaped as
  `task "NAME" is not available on HOST_OS (requires OS1 or OS2)`.
- **Rationale**: `resolveRoots` runs before the scheduler starts, so a
  multi-task invocation with one bad root executes nothing (Story 2
  scenario 3). `ValidationError`/exit 3 is the established class for
  "statically detectable, zero side effects" failures (Principle II).
  Because `--dry-run` and `--summary` also flow through `resolveRoots`,
  plan modes reject unavailable roots too — consistent with "this task does
  not exist on this platform".
- **Alternatives considered**: (a) reject in `splitArgs` where task
  existence is checked — rejected: that path reports "unknown task", which
  is a worse diagnostic than naming the OS requirement; (b) treat the root
  like a dep and skip it silently — rejected by spec (FR-003): explicit
  requests must fail loudly.

## D4: MCP surface — filter and defense-in-depth

- **Decision**: `mcpAdapter.Tasks()` (`internal/cli/mcp.go:34`) skips tasks
  failing `AvailableOn` exactly as it skips `IsPrivate()`. Additionally,
  `mcpAdapter.Call()` re-checks availability and returns the same
  "not available on HOST_OS" error for a mismatched name.
- **Rationale**: `Tasks()` feeds both tool registration
  (`mcpserver.New`) and the authz allowlist (`mcpserver/authz.go`), so one
  filter fixes both (FR-002, Story 1 scenario 4). The `Call()` check is
  defense-in-depth: `a.tasks` still contains every task by name, and a
  client could call a tool name it learned elsewhere (stale tool list,
  guessed name). Story 2 covers "an agent that guessed a task name".
- **Alternatives considered**: filtering `a.tasks` itself — rejected:
  unavailable tasks must remain resolvable as *dependencies* (to be
  skipped deliberately), and private tasks are already callable as deps by
  design; the map must stay complete.

## D5: `--dump` availability field

- **Decision**: `taskDTO` (`internal/cli/dump.go:31`) gains
  `Available bool` with JSON tag `available` (no `omitempty` — `false` is
  the signal). `toDTO` takes the target OS as a parameter; the call site
  passes the engine's `goos`.
- **Rationale**: dump remains a *complete* description of the file (all
  tasks stay listed, matching the existing `private: true` treatment), and
  consumers get the computed verdict without re-implementing OR/`unix`
  semantics (FR-005, Story 4). Parameterizing `toDTO` keeps it a pure
  function for table tests.
- **Alternatives considered**: omitting unavailable tasks from dump —
  rejected: dump is the machine view of the whole file (formatter/JSON
  parity), and Story 4 requires the mismatched task to appear with
  `available: false`.

## D6: What deliberately does NOT change

- **Parser/lexer/formatter/LSP/VS Code grammar**: attribute set unchanged;
  `docs/GRAMMAR.md` untouched (no DSL surface change).
- **`mcpserver/` public package**: its `Engine`/`TaskInfo` contract already
  expresses "the tasks to expose"; filtering happens in the adapter.
- **No skip notice in default output** (spec Assumptions): silent skip;
  verbose/debug output may mention it later, not in scope.
- **No `--ignore-os` escape hatch** (spec Assumptions): deferred.
- **`os()`/`arch()` builtins and `eval.Scope`**: unrelated to availability;
  unchanged.

## Existing-behavior notes (verified in code)

- `osMatches` semantics preserved verbatim: no OS attrs → `true`; OR across
  all OS attrs; `unix` → `goos != "windows"`; `macos` → `darwin`
  (run.go:729–758).
- Current call sites are visibility-only: `run.go:652`, `run.go:668`,
  `complete.go:41`. MCP (`mcp.go:37`) filters only private. Nothing gates
  execution today.
- Scheduler memoizes per `(namespace, task, canonical-args)`; skip-before-
  `run` leaves the memo table untouched (D2).
- `ValidationError` → `ExitValidation = 3` via `CodeFor`
  (`internal/cli/exit.go`).
