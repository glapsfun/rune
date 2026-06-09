# Feature Specification: Idiomatic Go Refactor — Skill-Governed Review & Refactoring

**Feature Branch**: `003-idiomatic-go-refactor`

**Created**: 2026-06-08

**Status**: Draft

**Input**: User description: "i want to do refactoring with applying best practisk for designe golang project use all avalable golang skill to apply best practics for rune project do reviwe and refactroing structure project"

## Overview

This feature does a **code-level review and refactoring** of the Rune Go codebase to bring
it to idiomatic best practices, governed by the project's bundled Go engineering skills
(Constitution **Principle VIII**): code style, naming, concurrency, design patterns,
performance, and the `golang-pro` umbrella. It is the substantive *code-quality* work that
the prior hardening feature (`002-best-practices-refactor`) deliberately deferred — `002`
covered docs/CI/Docker/repo hygiene and was explicitly behavior-preserving with **no Go
logic changed**; **this** feature reviews and improves the Go code and internal structure
itself.

The work remains **externally behavior-preserving** (the constitution's tool
backward-compatibility promise): Runefile semantics, CLI contracts, exit codes,
diagnostics, and MCP tool behavior do not change. Internal, unexported APIs and
within-package organization may change; the constitution-locked package layout
(Principle IV) is preserved.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Skill-governed code review (Priority: P1)

A maintainer wants to know exactly where the codebase diverges from the project's binding
Go engineering skills. They get a prioritized, traceable review report: every finding cites
the skill rule it breaks, the precise location, a severity, and a recommended fix.

**Why this priority**: You cannot refactor well without first knowing what to fix. A
systematic, skill-mapped audit is independently valuable (it guides reviewers and
contributors) and is the foundation every later story builds on.

**Independent Test**: Run the review against a clean checkout. Success = a report that
covers every Go package, and where each finding names a specific skill + rule, a
`file:line` location, a severity, and a concrete recommendation.

**Acceptance Scenarios**:

1. **Given** the Go codebase, **When** the review runs, **Then** every Go package under
   `cmd/`, `internal/`, and `mcpserver/` is assessed against each bundled `golang-*` skill.
2. **Given** a finding, **When** a maintainer reads it, **Then** it names the skill and rule
   violated, the location, a severity, and a recommended fix.
3. **Given** the full finding set, **When** it is presented, **Then** findings are ordered by
   severity: correctness/safety > design > style/naming/docs > performance.
4. **Given** the review, **When** it covers concurrency, **Then** it explicitly checks
   goroutine ownership/exit, `ctx.Done()` in blocking selects, channel direction/ownership,
   and bounded concurrency.

---

### User Story 2 - Remediate correctness & safety findings (Priority: P1)

A maintainer applies the fixes for the highest-severity findings — concurrency safety,
error handling, resource lifecycle, and context propagation — without changing what Rune
does.

**Why this priority**: These are the findings that affect correctness and reliability of a
tool other pipelines depend on. Fixing them is the core payoff of the review and is
independently shippable.

**Independent Test**: After remediation, the race detector and leak checks are clean and the
entire existing behavioral suite passes with no golden file regenerated.

**Acceptance Scenarios**:

1. **Given** a concurrency finding, **When** it is fixed, **Then** the goroutine has a clear
   owner and guaranteed exit, every blocking `select` observes `ctx.Done()`, and concurrency
   is bounded.
2. **Given** an error-handling finding, **When** it is fixed, **Then** the error is handled
   explicitly and wrapped with `%w` (no silent `_` discards); `panic` remains only for
   impossible-state bugs.
3. **Given** a resource/lifecycle finding, **When** it is fixed, **Then** an opened resource
   is `defer`-closed and every external/blocking call carries a timeout or cancellation.
4. **Given** all correctness fixes, **When** the suite runs, **Then** `go test -race ./...`
   passes with zero data races, concurrency tests report zero goroutine leaks, and **no**
   golden file is regenerated to accommodate changed behavior.

---

### User Story 3 - Idiomatic design & structural refactor (Priority: P2)

A maintainer applies the design-level findings: idiomatic constructors, dependency
injection, elimination of `init()`/global state, interface placement, and cleaner
single-responsibility file/package organization — all within the locked package layout.

**Why this priority**: Improves long-term maintainability and contributor onboarding.
Valuable but lower-risk and lower-urgency than correctness fixes.

**Independent Test**: Design findings are resolved and the suite still passes unchanged;
the locked package layout is intact and the public/CLI surface is unchanged.

**Acceptance Scenarios**:

1. **Given** a constructor with many parameters or optional config, **When** refactored,
   **Then** it uses functional options (or a documented justification for not doing so).
2. **Given** `init()` functions or mutable package globals, **When** refactored, **Then**
   they are replaced by explicit, injectable constructors (or explicitly justified).
3. **Given** the package layout, **When** files are reorganized, **Then** the
   constitution-locked package names (Principle IV) are unchanged and each package keeps a
   single clear responsibility.
4. **Given** the public surface, **When** the refactor lands, **Then** the `mcpserver` API
   and the CLI contract are backward-compatible and unnecessary exports are unexported.

---

### User Story 4 - Encode the skills as automated gates (Priority: P2)

A maintainer expands the automated linter set so the skill rules are enforced
mechanically — future drift is caught in CI, not by memory.

**Why this priority**: Makes the review durable. Without automated enforcement, the
codebase drifts back. Depends on nothing but the existing CI.

**Independent Test**: The expanded linter set runs clean on the refactored code and, on a
representative seeded violation, flags it.

**Acceptance Scenarios**:

1. **Given** the linter configuration, **When** it is expanded, **Then** it adds checks that
   encode skill rules (e.g., error-wrapping, context usage, common-idiom checks) on top of
   the existing set.
2. **Given** the expanded gate, **When** it runs on the refactored code, **Then** it reports
   zero violations.
3. **Given** a change that reintroduces a skill violation, **When** CI runs, **Then** the
   gate fails and blocks merge.

---

### User Story 5 - Benchmark-gated performance review (Priority: P3)

A maintainer reviews performance on the hot paths, adds benchmarks, and applies only
optimizations proven by measurement.

**Why this priority**: Performance matters for a tool in CI pipelines, but Principle VIII
forbids unprofiled optimization. This is the lowest priority and the most guarded.

**Independent Test**: Benchmarks exist for the identified hot paths; any performance change
carries a recorded `benchstat` improvement.

**Acceptance Scenarios**:

1. **Given** the hot paths (e.g., lexing, parsing, evaluation, scheduling), **When** the
   review runs, **Then** benchmarks are added for them.
2. **Given** a proposed optimization, **When** it is merged, **Then** it carries a
   `benchstat` result proving the win, an explanatory comment, and a `perf(scope):` commit.
3. **Given** a speculative/unprofiled optimization, **When** proposed, **Then** it is
   rejected (no benchmark, no merge).

---

### Edge Cases

- **Behavior drift during refactor**: a change alters an exit code, stdout, a diagnostic
  span, or an MCP tool schema — the behavioral/golden/integration/MCP tests must catch it
  and block merge; intentional behavior change is out of scope (backward-compat promise).
- **Locked-layout temptation**: a structural cleanup wants to rename a constitution-locked
  package (`token`, `lexer`, `ast`, `parser`, `analyzer`, `eval`, `runtime/*`, `mcpserver`).
  This is forbidden without a constitution amendment; reorganization stays within the layout.
- **Skill vs. constitution conflict**: a skill recommendation contradicts the constitution —
  the constitution governs, and the conflict is documented.
- **Public-API pressure**: a finding would require changing the public `mcpserver` API — the
  refactor must preserve backward compatibility or the finding is deferred, not forced.
- **Unprofiled optimization**: a "faster" change without a benchmark — rejected per
  Principle VIII.
- **Goroutine leak surfaced**: leak detection flags a goroutine without a guaranteed exit —
  treated as a correctness finding, not a nit.
- **Large mechanical refactor hides a real change**: refactors are kept small and reviewable
  so the test suite can serve as a reliable behavior oracle.

## Requirements *(mandatory)*

### Functional Requirements

**Review (US1)**

- **FR-001**: The work MUST produce a code-review report assessing every Go package under
  `cmd/`, `internal/`, and `mcpserver/` against each bundled `golang-*` skill.
- **FR-002**: Each finding MUST cite the specific skill and rule, a `file:line` location, a
  severity, and a recommended fix.
- **FR-003**: Findings MUST be prioritized by severity in the order: correctness/safety >
  design > style/naming/documentation > performance.
- **FR-004**: The review MUST explicitly cover: concurrency (goroutine ownership and
  guaranteed exit, `ctx.Done()` in blocking selects, channel direction/ownership, bounded
  concurrency); error handling (explicit handling, `%w` wrapping, no silent `_` discards,
  `panic` only for impossible states); design (functional options, no `init()`/globals,
  `defer Close()`, timeouts and `context.Context`-first parameters on external/blocking
  calls, accept-interfaces/inject-dependencies); style and clarity; naming; and
  documentation.

**Correctness & safety remediation (US2)**

- **FR-005**: All correctness/safety findings MUST be remediated.
- **FR-006**: Every goroutine MUST have a clear owner and a guaranteed exit (context
  cancellation, channel close, or `WaitGroup`); every blocking `select` MUST include
  `ctx.Done()`; channel parameters MUST be directional; concurrency MUST be bounded.
- **FR-007**: Every error MUST be handled explicitly (no silent `_` discards of actionable
  errors); errors MUST be wrapped with `fmt.Errorf("…: %w", err)` **wherever the error chain
  should be preserved**. Intentional non-wraps (sentinel errors, deliberately-opaque
  top-level messages) are allowed when documented. `panic` MUST be reserved for
  impossible-state bugs.
- **FR-008**: Every external or blocking call MUST carry a timeout or cancellation, and every
  opened resource MUST be `defer`-closed immediately after opening.
- **FR-009**: Refactoring MUST be externally behavior-preserving — no change to Runefile
  semantics, CLI contracts, exit codes, diagnostics, or MCP tool behavior; the existing
  suite MUST pass with **no** golden file regenerated.
- **FR-010**: `go test -race ./...` MUST pass with zero data races; concurrency-bearing
  tests SHOULD guard against goroutine leaks.

**Design & structure (US3)**

- **FR-011**: Constructors SHOULD use functional options where they exceed ~4 parameters or
  carry optional configuration; `init()` functions and mutable package globals MUST be
  eliminated in favor of explicit, injectable constructors (or explicitly justified).
- **FR-012**: Dependencies MUST be injected via interfaces defined at the consumer; domain
  logic MUST stay free of framework dependencies.
- **FR-013**: Files and packages MUST be organized for single responsibility **within** the
  constitution-locked package layout; the locked package names MUST NOT change.
- **FR-014**: The public surface MUST be minimized — unexport by default **within
  `internal/*`**. The already-public `mcpserver` API and the CLI contract are **frozen**:
  they MUST remain backward-compatible and are **NOT** subject to the unexport pass (a
  currently-exported `mcpserver`/`cmd` symbol MUST NOT be unexported).

**Automated enforcement (US4)**

- **FR-015**: The linter configuration MUST be expanded with checks that encode the skill
  rules, and the refactored codebase MUST pass the expanded set with zero violations.
- **FR-016**: The expanded gate MUST run in CI on every change and block merge on violation.

**Performance (US5)**

- **FR-017**: A performance review MUST identify the hot paths and add benchmarks for them.
- **FR-018**: Any performance optimization MUST be justified by a `benchstat` measurement
  proving the win, carry an explanatory comment, and ship under a `perf(scope):` commit;
  unprofiled/speculative optimizations MUST NOT merge.

**Cross-cutting**

- **FR-019**: Where a skill and the constitution conflict, the constitution governs, and the
  conflict MUST be documented.
- **FR-020**: Exported identifiers MUST carry doc comments per the documentation standard.

### Key Entities *(deliverable artifacts)*

- **Code-review report**: the inventory of findings (skill, rule, `file:line`, severity,
  recommended fix), prioritized by severity.
- **Remediation changesets**: behavior-preserving commits, each traceable to a finding.
- **Expanded linter configuration + CI gate**: the automated encoding of the skill rules.
- **Benchmark suite + benchstat records**: hot-path benchmarks and the evidence behind any
  performance change.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of Go packages are covered by the review, and every finding maps to a
  skill rule, a location, and a severity.
- **SC-002**: 100% of correctness/safety (highest-severity) findings are resolved.
- **SC-003**: The full test suite passes with the race detector enabled with zero data
  races, and concurrency tests report zero goroutine leaks.
- **SC-004**: The refactor changes no observable behavior — the entire existing behavioral,
  integration, golden, and corpus suite passes with **no** golden file regenerated.
- **SC-005**: The expanded automated quality gate passes with zero violations and runs on
  every change.
- **SC-006**: Zero `init()` functions remain, and no **mutable** package-global state
  remains. Read-only globals (lookup tables, sentinel errors, compiled regexes) are
  acceptable when documented as immutable.
- **SC-007**: 100% of exported identifiers carry a doc comment.
- **SC-008**: Every identified hot path has a benchmark, and every merged performance change
  has a recorded `benchstat` improvement (zero unprofiled optimizations merged).
- **SC-009**: The public `mcpserver` API and the CLI contract are unchanged, verified by the
  existing contract/integration tests passing without modification.
- **SC-010**: Every remediation commit is traceable to the review finding and skill rule it
  addresses.

## Assumptions

- **Rubric**: The binding standard is the bundled `golang-*` skill set named in Constitution
  Principle VIII — `golang-pro`, `golang-code-style`, `golang-concurrency`,
  `golang-design-patterns`, `golang-performance` — plus the naming, lint, and documentation
  guidance they reference.
- **Behavior-preserving externally**: Runefile semantics, CLI contracts, exit codes,
  diagnostics, and MCP tool behavior are unchanged. Internal/unexported APIs and
  within-package file organization may change.
- **Locked layout respected**: The constitution-locked package layout (Principle IV) is
  preserved; no locked package is renamed (that would require a constitution amendment).
- **Test suite is the oracle**: The existing unit/golden/integration/corpus/fuzz suite is
  the behavior-preservation oracle; per global policy it runs inside Docker.
- **Performance is gated**: Performance optimizations require `benchstat` proof; pure perf
  work is the lowest priority (P3) and only measured wins merge.
- **Builds on `002`**: The expanded linter set extends the `.golangci.yml` and CI introduced
  by feature `002-best-practices-refactor`.
- **No product changes**: No new features, DSL/grammar changes, executors, or CLI commands.

## Out of Scope

- New features, Runefile/DSL/grammar changes, new executors, or new CLI commands.
- Any change to observable behavior (forbidden by the backward-compatibility promise unless
  opt-in, which is out of scope here).
- Renaming constitution-locked packages or restructuring beyond the locked layout (would
  require a separate constitution amendment).
- Speculative performance optimization without benchmarks.
- Documentation/CI/Docker/release work already delivered by `002-best-practices-refactor`
  (this feature only *extends* the linter gate).
