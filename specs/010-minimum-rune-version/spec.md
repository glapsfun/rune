# Feature Specification: Minimum Rune Version

**Feature Branch**: `010-minimum-rune-version`

**Created**: 2026-07-09

**Status**: Draft

**Input**: User description: "Feature 1: Minimum Rune Version — a Runefile can declare a minimum required Rune binary version; an older binary is rejected before any execution, with an actionable diagnostic. Includes a static-value guard, an emergency `--ignore-version` override, and `rune version` / `rune version --check` commands."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Reject an incompatible binary before execution (Priority: P1)

A project pins the minimum Rune runtime it needs in its root Runefile. A teammate, CI
runner, or agent that has an older Rune binary attempts to run a task. Instead of a
confusing parser error or subtly wrong behavior, they get an immediate, clear message
telling them exactly which version is required, which version they have, and where to
upgrade — and nothing is executed.

**Why this priority**: This is the core value of the feature. Without it, the whole
capability does not exist. It directly prevents the "works on my machine / fails in CI"
class of failure and turns a vague error into an actionable one.

**Independent Test**: Author a Runefile with `set minimum_version := "0.8.0"`, run it with
a simulated installed version of `0.7.2`, and confirm the run aborts with the required-vs-
installed diagnostic and a non-zero exit, before any task, shell, interpreter, or agent
runs.

**Acceptance Scenarios**:

1. **Given** a Runefile declaring `set minimum_version := "0.8.0"` and an installed Rune of
   `0.7.2`, **When** the user runs any task, **Then** Rune prints an error stating the file
   requires `>= 0.8.0`, shows the installed and required versions and an upgrade URL, points
   the caret at the version literal in the `set` statement, executes nothing, and exits
   non-zero.
2. **Given** the same Runefile and an installed Rune of `0.8.0`, **When** the user runs a
   task, **Then** the version check passes silently and the task runs normally.
3. **Given** the same Runefile and an installed Rune of `0.9.1`, **When** the user runs a
   task, **Then** the version check passes silently and the task runs normally.
4. **Given** a Runefile with no `minimum_version` setting, **When** the user runs a task,
   **Then** behavior is unchanged from today (no version gate applied).

---

### User Story 2 - Static-value guarding (Priority: P1)

The minimum-version declaration must be a fixed, statically knowable value. A project author
(or a reviewer) must be able to trust that the requirement can be determined without running
anything. If the author tries to compute the value dynamically (e.g. from an environment
variable), Rune rejects the Runefile with a clear diagnostic rather than silently deriving a
requirement at runtime.

**Why this priority**: The requirement must be checkable before semantic analysis and before
any expression evaluation. A dynamic value would defeat the purpose (you'd have to run
unsupported behavior to learn what's supported). This guard is inseparable from the P1
gate.

**Independent Test**: Author a Runefile that assigns `minimum_version` from a non-literal
expression and confirm Rune rejects it with the "must be a static semantic version"
diagnostic, pointing at the offending value.

**Acceptance Scenarios**:

1. **Given** `set minimum_version := "0.8.0"` (a string literal), **When** Rune reads the
   file, **Then** the value is accepted.
2. **Given** a `minimum_version` set from a non-literal expression (e.g. an `env(...)` call
   or a reference to another variable), **When** Rune reads the file, **Then** Rune reports
   `minimum_version must be a static semantic version`, points the caret at the value, and
   executes nothing.
3. **Given** `set minimum_version := "not-a-version"` (a literal that is not a valid semantic
   version), **When** Rune reads the file, **Then** Rune reports that the value is not a valid
   semantic version and executes nothing.

---

### User Story 3 - Inspect and check compatibility from the CLI (Priority: P2)

A developer or CI job wants to know which Rune version is installed and whether it satisfies
the current project's requirement — for humans on the terminal and for scripts that parse
output.

**Why this priority**: Valuable for diagnosability and CI gating, but the protective behavior
(P1) works without it. This is the ergonomic and automation layer on top.

**Independent Test**: Run `rune version` and confirm it reports the installed Rune version and
the Runefile language version; run `rune version --check` in a project with a `minimum_version`
and confirm it reports required, installed, and compatibility status; add `--json` and confirm
machine-readable output.

**Acceptance Scenarios**:

1. **Given** any installed Rune, **When** the user runs `rune version`, **Then** it prints the
   installed Rune version and the Runefile language version.
2. **Given** a project with `set minimum_version := "0.8.0"` and installed `0.8.3`, **When** the
   user runs `rune version --check`, **Then** it reports the required version, the installed
   version, and a "compatible" status.
3. **Given** the same project, **When** the user runs `rune version --check --json`, **Then** it
   prints a machine-readable object containing the installed version, the required version, a
   boolean compatibility flag, and the resolved Runefile path.
4. **Given** a project whose installed version does not satisfy the requirement, **When** the
   user runs `rune version --check`, **Then** it reports an incompatible status and exits
   non-zero, without running any task.
5. **Given** a directory with no Runefile, **When** the user runs `rune version --check`, **Then**
   it reports that no requirement is declared (or no Runefile found) and does not error as if
   incompatible.

---

### User Story 4 - Emergency override (Priority: P3)

Occasionally a user needs to run a project with a binary that does not meet the declared
minimum — for a quick local experiment or a break-glass situation. An explicit CLI flag lets
them bypass the gate, but it always announces loudly that the gate was bypassed, and it cannot
be turned on silently from within a Runefile.

**Why this priority**: A safety valve, not a primary path. It must exist so the gate is never a
hard lock-out, but it should be rare and visible.

**Independent Test**: Run a task with `--ignore-version` against a Runefile whose requirement is
not met, and confirm the task runs but a visible warning naming both versions is printed;
confirm no Runefile setting can enable this behavior.

**Acceptance Scenarios**:

1. **Given** a Runefile requiring `>= 0.8.0` and installed `0.7.2`, **When** the user runs with
   `--ignore-version`, **Then** a warning is printed naming the ignored requirement and the
   running version, and the requested task proceeds.
2. **Given** any Runefile, **When** an author attempts to enable version-ignoring from inside the
   Runefile, **Then** there is no mechanism to do so — the override is CLI-only.
3. **Given** non-interactive / agent (MCP) execution, **When** a project's requirement is not met,
   **Then** the override is disabled unless the operator has explicitly configured it outside the
   Runefile, and otherwise execution is refused with the standard incompatibility error.

---

### Edge Cases

- **Prerelease installed vs release required**: An installed `0.9.0-rc.1` does **not** satisfy a
  requirement of `0.9.0`. An installed `0.9.0` satisfies `0.9.0`.
- **Development builds**: A development build reports a base version plus commit metadata (e.g.
  `0.9.0-dev+13dbf54`); compatibility is decided from the base version by the same rules, and the
  version seen by the check is injectable by tests rather than read from an ambient global.
- **Requirement in an imported file**: A `minimum_version` declared in an imported/child Runefile
  does not silently override or relax the root project's effective requirement — the root Runefile
  owns the effective requirement.
- **Requirement equal to installed**: Installed version exactly equal to the requirement is
  compatible (`>=` semantics).
- **Malformed requirement literal**: A non-semantic-version literal is rejected at read time with a
  clear diagnostic, not treated as "always compatible."
- **No Runefile / no requirement**: Commands and normal execution behave as today; the gate is
  simply not applied.
- **Both `rune_version` and `minimum_version` present**: They are independent settings and are both
  honored without interfering with each other.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A Runefile MUST be able to declare a minimum required Rune runtime version via a
  dedicated setting whose public name is `minimum_version`.
- **FR-002**: The declared value MUST be interpreted as a lower bound with `>=` semantics (the file
  requires that version or newer).
- **FR-003**: Rune MUST compare the installed version against the requirement using Semantic
  Versioning precedence rules (including prerelease precedence).
- **FR-004**: Rune MUST evaluate the requirement after discovering root-Runefile settings and BEFORE
  interpreting imports, BEFORE semantic analysis, and BEFORE starting any task, shell, interpreter,
  or agent.
- **FR-005**: When the installed version does not satisfy the requirement, Rune MUST abort with a
  non-zero exit and execute nothing.
- **FR-006**: The incompatibility error MUST identify the required version, the installed version,
  and an upgrade location, and MUST include a `file:line:col` location with a caret span pointing at
  the offending version value in the source.
- **FR-007**: The `minimum_version` value MUST be a static literal semantic version; Rune MUST reject
  a value derived from any non-literal expression with the diagnostic `minimum_version must be a
  static semantic version`.
- **FR-008**: Rune MUST reject a `minimum_version` literal that is not a valid semantic version, with
  a clear diagnostic and no execution.
- **FR-009**: For the first release, only a single minimum-version constraint is supported; range or
  compound constraint syntax (e.g. `>=0.8,<1.0`, `^0.8`, `~0.8.1`) MUST NOT be accepted as
  `minimum_version` values.
- **FR-010**: A prerelease installed version MUST NOT satisfy a requirement equal to the
  corresponding release version (e.g. `0.9.0-rc.1` does not satisfy `0.9.0`).
- **FR-011**: The version used for comparison MUST be injectable by tests (an explicit test hook)
  rather than sourced only from an ambient global, so behavior is deterministic across release,
  prerelease, and development builds.
- **FR-012**: A `minimum_version` declared in an imported/child Runefile MUST NOT silently override or
  relax the root project's effective requirement; the root Runefile owns the effective requirement.
- **FR-013**: `minimum_version` MUST remain independent from the existing `rune_version` (language
  compatibility) setting; neither affects the other.
- **FR-014**: Rune MUST provide a CLI flag `--ignore-version` that bypasses the requirement gate for
  that invocation.
- **FR-015**: When `--ignore-version` is used and the requirement would otherwise fail, Rune MUST
  print a visible warning naming the ignored required version and the running version, then proceed.
- **FR-016**: The version-ignore behavior MUST NOT be enablable from within a Runefile (CLI-only).
- **FR-017**: For non-interactive / agent (MCP) execution, the ignore behavior MUST be disabled by
  default and only available when explicitly configured by the operator outside the Runefile.
- **FR-018**: Rune MUST provide a `rune version` command that prints the installed Rune version and
  the Runefile language version.
- **FR-019**: Rune MUST provide `rune version --check` that reports the project's required version, the
  installed version, and a human-readable compatibility status, resolving the applicable Runefile from
  the working directory.
- **FR-020**: `rune version --check` MUST exit non-zero when incompatible and MUST NOT run any task.
- **FR-021**: Rune MUST provide `rune version --check --json` producing machine-readable output that
  includes the installed version, the required version, a boolean compatibility flag, and the resolved
  Runefile path.
- **FR-022**: When no Runefile or no requirement is present, `rune version --check` MUST report the
  absence of a requirement rather than reporting incompatibility.
- **FR-023**: Existing Runefiles that do not declare `minimum_version` MUST behave exactly as before
  (no version gate, byte-identical existing behavior).

### Key Entities *(include if feature involves data)*

- **Minimum-version requirement**: The static semantic-version lower bound declared by the root
  Runefile via `minimum_version`. Attributes: the literal value, its source location (for
  diagnostics), and its origin (root vs imported).
- **Installed version**: The version of the running Rune binary, expressed as a semantic version,
  possibly with prerelease and/or build metadata for development builds. Injectable for testing.
- **Compatibility result**: The outcome of comparing installed vs required — a boolean status plus the
  two versions and the resolved Runefile path — surfaced to humans (error/status text) and machines
  (`--json`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Runefile can pin a minimum Rune version in a single line, and running it with an
  older binary is refused before any task, shell, interpreter, or agent starts — verified in
  automated tests for at least one older-than, equal-to, and newer-than case.
- **SC-002**: 100% of incompatibility errors identify the required version, the installed version, an
  upgrade location, and the exact source location (with caret) of the offending value.
- **SC-003**: A non-literal or non-semantic-version `minimum_version` is always rejected with a
  specific diagnostic and never results in execution.
- **SC-004**: Version comparison follows Semantic Versioning precedence in every tested case,
  including the prerelease-does-not-satisfy-release case.
- **SC-005**: `rune version` and `rune version --check` (including `--json`) work identically in local
  development and CI, and the `--json` output is stable enough to be parsed by scripts.
- **SC-006**: `--ignore-version` allows a bypass exactly when used, always emits a visible warning, and
  can never be triggered from within a Runefile or from default non-interactive/agent execution.
- **SC-007**: An imported file cannot change the root project's effective requirement — demonstrated by
  a test where a child declares a different (higher and lower) value and the root's value governs.
- **SC-008**: The feature ships with unit, integration, golden-diagnostic, and cross-platform tests,
  and all constitution quality gates (lint, race test on Linux/macOS/Windows, build, golden, fuzz,
  docs-verify, release-dryrun) pass.

## Assumptions

- **Public setting name is `minimum_version`.** The description offered `minimum_rune_version`,
  `min_version`, and `minimum_version`, and explicitly recommended `minimum_version`; that name is
  adopted. Aliases are out of scope for the first release unless raised during planning.
- **`>=` is the only supported operator** for the first release; the value is a bare semantic version,
  not a constraint expression. A future `rune_requirement` range syntax is explicitly deferred.
- **Semantic Versioning 2.0.0 precedence** governs comparison, including build-metadata being ignored
  for precedence and prerelease ordering below the associated release.
- **Development builds** encode a base release version plus commit metadata (e.g.
  `0.9.0-dev+13dbf54`); compatibility is judged from the base version under the same rules.
- **Root ownership.** The "root Runefile" is the entry Runefile Rune resolves for the invocation;
  imported files are children whose `minimum_version` does not govern the effective requirement.
- **MCP operator configuration** for allowing the override is provided out-of-band (e.g. operator-set
  configuration/environment at server startup), never from a Runefile; the precise mechanism is a
  planning detail.
- **Upgrade URL** points at the project's public releases page; the exact canonical URL is confirmed
  during planning (the description shows `github.com/glapsfun/rune/releases`).
- **No new DSL grammar surface beyond a `set` value.** `minimum_version` is a setting consumed like
  other `set` settings; it does not introduce new statement types.
