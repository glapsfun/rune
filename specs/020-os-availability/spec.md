# Feature Specification: OS Availability Enforcement

**Feature Branch**: `020-os-availability`

**Created**: 2026-08-12

**Status**: Draft

**Input**: User description: "Conditional Task Availability: Use attributes
like [linux] or [windows] to not just enable tasks for the host OS, but to
hide irrelevant tasks from the AI agent entirely, ensuring it doesn't attempt
to run platform-incompatible commands."

## Feature Summary

Rune already parses the OS attributes `[linux]`, `[macos]`, `[windows]`, and
`[unix]`, and already hides mismatched tasks from `rune --list`, the TUI
picker, and shell completion. But the availability rule stops there:

1. **The MCP server ignores OS attributes.** An AI agent connected to
   `rune mcp` (or `rune serve`) on Linux still sees every `[windows]` task as
   a callable tool, and can invoke platform-incompatible commands.
2. **Nothing is enforced at run time.** `rune setup-win` on Linux runs the
   Windows commands anyway. The os-filtering example's documentation claims
   an OS-restricted task "only shows and runs on that OS" — the "runs" half
   is currently false.
3. The availability predicate is private to the CLI listing code, is
   duplicated nowhere else, and has no test coverage.

This feature makes OS availability a single, enforced property of a task:
one shared predicate applied to every surface — listing, completion, the MCP
tool set, direct invocation, and dependency resolution — so that a task
declared for another OS is invisible to agents and impossible to run,
while cross-platform dispatch through dependencies keeps working.

No new syntax is introduced. The existing attributes and their semantics
(no OS attribute ⇒ available everywhere; multiple OS attributes combine as
OR; `unix` ⇒ every platform except Windows) are unchanged.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - AI Agents Never See Foreign-OS Tasks (Priority: P1)

An AI agent connects to a project's Runefile through Rune's MCP server on a
Linux host. The Runefile contains `[windows]`-only tasks. The agent's tool
list contains no trace of those tasks, so it cannot attempt to run
platform-incompatible commands — the failure mode is prevented, not merely
discouraged.

**Why this priority**: This is the reported feature request. MCP is the
agent-facing surface, and today it is the only listing surface that leaks
OS-mismatched tasks.

**Independent Test**: Serve a Runefile containing tasks for the host OS,
tasks for a different OS, and unrestricted tasks over MCP; list the
server's tools and confirm only host-OS and unrestricted tasks appear.

**Acceptance Scenarios**:

1. **Given** a Runefile with a `[windows]` task and the MCP server running
   on Linux, **When** the agent lists available tools, **Then** the
   Windows-only task is absent from the tool list.
2. **Given** the same Runefile, **When** the agent lists tools, **Then**
   unrestricted tasks and `[linux]`/`[unix]` tasks are all present.
3. **Given** a task carrying both `[linux]` and `[windows]`, **When** the
   agent lists tools on either platform, **Then** the task is present
   (multiple OS attributes combine as OR).
4. **Given** the MCP allowlist / destructive-task gate is configured,
   **When** it evaluates the task set, **Then** it reasons over the same
   OS-filtered set the agent sees.

---

### User Story 2 - Foreign-OS Tasks Cannot Be Run (Priority: P2)

A user (or an agent that guessed a task name) explicitly invokes an
OS-mismatched task by name. Rune refuses before executing anything, with an
error naming the task, the OS it requires, and the current host OS.

**Why this priority**: Hiding without enforcement is a leaky abstraction —
anything that knows a task's name can still run incompatible commands.
Enforcement also makes the existing documentation's "only shows and runs on
that OS" claim true.

**Independent Test**: On a host whose OS does not match a task's OS
attribute, run that task by name and confirm Rune exits non-zero with the
availability error and executes none of the task's body, dependencies, or
hooks.

**Acceptance Scenarios**:

1. **Given** a `[windows]` task on a Linux host, **When** the user runs it
   by name, **Then** Rune fails with an error identifying the task, the
   required OS, and the host OS, exits non-zero, and executes nothing.
2. **Given** a `[unix]` task on a macOS host, **When** the user runs it,
   **Then** it runs normally (`unix` matches every platform except
   Windows).
3. **Given** several tasks named in one invocation where any one is
   OS-mismatched, **When** the user runs them, **Then** the invocation
   fails up front and no task executes.

---

### User Story 3 - Cross-Platform Dispatch Through Dependencies (Priority: P3)

A Runefile author writes one public `setup` task that depends on
`setup-linux` (`[linux]`), `setup-mac` (`[macos]`), and `setup-win`
(`[windows]`). On any host, running `setup` silently skips the mismatched
dependencies and runs only the matching one, then the body of `setup`
itself.

**Why this priority**: Without this, runtime enforcement (Story 2) would
make OS attributes unusable in dependency graphs — every cross-platform
Runefile would break. Skipping is what turns the attributes into a dispatch
mechanism.

**Independent Test**: Run the dispatcher task on any host and confirm
exactly one OS-specific dependency executed, the others were skipped
without error, and the dispatcher's own body ran.

**Acceptance Scenarios**:

1. **Given** a task whose dependencies carry different OS attributes,
   **When** it runs, **Then** only host-matching dependencies execute,
   mismatched ones are skipped silently, and the run succeeds.
2. **Given** a task whose dependencies are *all* OS-mismatched, **When** it
   runs, **Then** every dependency is skipped and the task's own body still
   runs successfully.
3. **Given** a task with an OS-mismatched post-hook, **When** it runs,
   **Then** the post-hook is skipped under the same rule as dependencies.
4. **Given** a mismatched dependency is skipped, **When** the run
   completes, **Then** the skip caused no error, no non-zero exit, and the
   skipped task's body was never executed.

---

### User Story 4 - Machine-Readable Availability (Priority: P4)

Tooling that consumes `rune --dump --format json` (editor integrations,
agent frameworks, CI scripts) can read a computed `available` field per
task, alongside the raw attributes, without re-implementing Rune's OS
semantics.

**Why this priority**: Small additive convenience; everything else in this
feature works without it.

**Independent Test**: Dump a Runefile containing matched, mismatched, and
unrestricted tasks and confirm each task's `available` value reflects the
host OS.

**Acceptance Scenarios**:

1. **Given** a `[windows]` task dumped on Linux, **When** the JSON is read,
   **Then** the task appears with `available: false` and its attributes
   still listed.
2. **Given** an unrestricted task, **When** dumped on any host, **Then**
   `available` is `true`.

---

### Edge Cases

- A task with no OS attribute is available everywhere — unchanged.
- Multiple OS attributes on one task (one line or several) combine as OR;
  `[unix]` plus `[windows]` therefore means available everywhere.
- `unix` matches every platform except Windows — not merely Linux and
  macOS.
- A private (`[private]` or `_`-prefixed) OS-matched task stays hidden from
  all listings and MCP exactly as today; privacy and availability filters
  compose.
- An OS-mismatched task referenced as a dependency is skipped (Story 3),
  but the same task invoked directly by name errors (Story 2) — the two
  behaviors are intentional and must both hold for the same task.
- Skipping a dependency must not disturb the rest of the dependency
  order or the post-hook sequence.
- `rune --list`, the picker, and shell completion keep their current
  filtering behavior (regression guard).
- Existing Runefiles where an OS-mismatched dependency previously *ran* will
  change behavior: the dependency is now skipped. This is the intended fix
  of a defect, and is called out in the changelog as a behavior change.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Task OS availability MUST be computed by one authoritative
  rule applied uniformly by every surface (list, picker, completion, MCP,
  run-time), preserving today's semantics: no OS attribute ⇒ available;
  multiple OS attributes OR together; `unix` ⇒ any platform except
  Windows.
- **FR-002**: The MCP server MUST NOT expose OS-mismatched tasks as tools,
  on both the stdio (`rune mcp`) and HTTP (`rune serve`) transports; the
  MCP authorization/allowlist layer MUST operate on the same filtered set.
- **FR-003**: Direct invocation of an OS-mismatched task MUST fail before
  any execution, with a non-zero exit and an error naming the task, its
  required OS(es), and the host OS.
- **FR-004**: OS-mismatched dependencies and post-hooks MUST be silently
  skipped during graph resolution; a skip is not an error, and the
  depending task's remaining dependencies, body, and hooks proceed
  normally.
- **FR-005**: The JSON dump output MUST include a computed boolean
  `available` field per task, evaluated against the host OS, alongside the
  existing raw attributes.
- **FR-006**: Documentation MUST be updated so every claim about OS
  attributes is true after this change: the MCP docs state that only
  host-OS-available non-private tasks are exposed, and the os-filtering
  example's "shows and runs" description matches actual behavior. The
  documentation test suite stays green.
- **FR-007**: Availability MUST be evaluatable for any named operating
  system, not only the one the evaluation runs on, so every platform's
  behavior can be verified from any development machine.

### Non-Functional Requirements

- **NFR-001**: No new Runefile syntax, attribute, flag, or configuration is
  introduced; existing Runefiles keep parsing and formatting byte-identically.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a given host, the MCP tool list contains 100% of available
  non-private tasks and zero OS-mismatched tasks — verified by an automated
  test with mixed-OS fixtures.
- **SC-002**: Direct invocation of an OS-mismatched task executes zero
  commands and exits non-zero with the availability error — verified for
  single- and multi-task invocations.
- **SC-003**: The cross-platform dispatch pattern (one task depending on
  per-OS tasks) runs exactly the host-matching dependency and succeeds on
  every supported operating system — verified by tests that evaluate
  availability for each supported OS.
- **SC-004**: The availability rule has automated coverage for: no
  attributes, single match, single mismatch, multi-attribute OR, and
  `unix` on Linux, macOS, and Windows (its first-ever test coverage).
- **SC-005**: Zero regressions in `--list`, picker, and completion
  filtering; the documentation test suite (`rune docs-check`) passes with
  the updated docs.

## Assumptions

- The current OR semantics for multiple OS attributes and the
  `unix` = not-Windows rule are intentional and are preserved, now
  documented and tested rather than changed.
- Silently skipping mismatched dependencies (no notice printed) is the
  chosen behavior; verbose/debug output MAY mention skips, but the default
  human-visible output does not.
- The dependency-skip behavior change (mismatched deps used to run) is
  acceptable as a pre-1.0 minor-version change with a changelog note; no
  compatibility flag is provided.
- No escape hatch (e.g. `--ignore-os`) is provided to force-run a
  mismatched task; if a genuine need appears later it can be added as a
  separate feature.
- Architecture (`arch()`), environment, or expression-based availability
  conditions are out of scope; only the four existing OS attributes govern
  availability.
