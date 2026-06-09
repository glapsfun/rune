# Feature Specification: Rune — A Shared Task Runner for Humans and AI Agents

**Feature Branch**: `001-rune-task-runner`

**Created**: 2026-06-08

**Status**: Draft

**Input**: User description: DSL-driven, Go-native task runner positioned as a spiritual successor to `just`, with first-class multi-language task bodies and native AI-agent / Model Context Protocol (MCP) integration.

## Overview

Rune is a command-line task runner: a single tool that lets a team capture every
project command (lint, test, build, deploy, codegen, release, and the long tail of
miscellaneous scripts) in one readable, version-controlled file, then run any of them by
name. Unlike a build system, Rune does not guess whether work is "up to date" — asking for
a task runs it. Its defining purpose is that the **same** task definitions are usable two
ways from one source of truth: a person runs them from the terminal, and an AI agent or IDE
discovers and runs the exact same tasks through a standard agent interface. There is no
divergence between "what the docs say," "what CI does," and "what an agent does."

## Clarifications

### Session 2026-06-08

- Q: For exposing tasks to AI agents/IDEs, what transport & access scope should v1 target? → A: Local (same-machine) interface is always available; a remote network endpoint is opt-in, binds to localhost by default, and requires an explicit auth token before any task is callable. (Note: the AI-agent capability works by driving an installed AI agent CLI — e.g., Claude CLI, Codex, Copilot — to run the task's prompt.)
- Q: What provider model should the AI-agent task type use in v1? → A: Drive an installed agent CLI (e.g. `claude`, `codex`, `copilot`) via a configurable command; the CLI handles its own authentication. The provider interface stays open to other backends (e.g. direct hosted APIs) later.
- Q: How is a task classified as "destructive" for agent gating? → A: Author-declared only — a task is gated from unattended agent execution iff the author marks it (confirm/destructive attribute). Private tasks are never exposed; other non-private tasks are agent-callable. Operators MAY layer a stricter allow-list on top. No content heuristics.
- Q: What block syntax should task bodies use? → A: Significant indentation (`make`/`just`-style), consistent within a task; mixing tabs and spaces within a single body is a located error. (Not braces/`begin`-`end`.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Author and run project commands from one readable file (Priority: P1)

A developer collects a project's everyday commands into a single task file at the repo
root. Each task has a short name and a body of commands. The developer runs a task by name
(`rune build`, `rune test`), passes parameters where a task accepts them, and declares that
some tasks depend on others so prerequisites run first. Running with no task name runs a
configured default. Running the tool anywhere inside the project finds the nearest task file
by searching upward, so the command works from any subdirectory.

**Why this priority**: This is the irreducible core — a usable command runner. On its own it
already replaces a pile of ad-hoc shell scripts and a `Makefile`, and every other capability
builds on it. It is the MVP.

**Independent Test**: Write a task file with two tasks where one depends on the other, run
the dependent task by name, and confirm both run in the correct order, each exactly once,
with parameters substituted correctly and a non-zero exit code if any command fails.

**Acceptance Scenarios**:

1. **Given** a task file defining a `build` task, **When** the developer runs the tool with
   `build`, **Then** the build task's commands execute and the tool exits 0 on success.
2. **Given** a task `deploy` that lists `build` and `test` as prerequisites, **When** the
   developer runs `deploy`, **Then** `build` and `test` run first (each once) and `deploy`
   runs only if they succeed.
3. **Given** a task that accepts a parameter with a default value, **When** the developer
   runs it without that argument, **Then** the default is used; **When** they pass an
   argument, **Then** the argument value is used.
4. **Given** the developer is in a nested subdirectory of the project, **When** they run the
   tool, **Then** it locates and uses the nearest task file found by searching upward.
5. **Given** a configured default task, **When** the tool is run with no task name, **Then**
   the default task runs; **When** run with `--list`, **Then** the available tasks (with
   their documentation) are shown instead of running anything.

---

### User Story 2 - Catch mistakes before anything runs (Priority: P2)

Before executing a single command, Rune analyzes the whole task file and refuses to run if
it finds a problem: a reference to a task that does not exist, a variable that was never
defined, a circular chain of dependencies, or a task invoked with the wrong number of
arguments. Each problem is reported with the exact file, line, and column and a visual
pointer to the offending text. If analysis fails, nothing executes and no side effects occur.

**Why this priority**: "Errors are a feature." Precise, up-front validation is a primary
reason people trust and prefer this class of tool, and it prevents half-run pipelines that
leave a project in a broken state. It depends on Story 1's file existing, so it follows it.

**Independent Test**: Introduce, one at a time, an unknown dependency, an undefined variable
reference, a dependency cycle, and a wrong-arity call; confirm that in each case the tool
exits non-zero, runs nothing, and prints a message naming the exact location of the problem.

**Acceptance Scenarios**:

1. **Given** a task that depends on a name with no matching task, **When** the file is run or
   validated, **Then** the tool reports an "unknown task" error with file/line/column and
   does not execute anything.
2. **Given** a body or expression that references an undefined variable, **When** the file is
   processed, **Then** the tool reports the undefined name and its location.
3. **Given** tasks whose dependencies form a cycle, **When** the file is processed, **Then**
   the tool reports the cycle and lists the chain of task names involved.
4. **Given** a task that requires two parameters invoked with one, **When** it is requested,
   **Then** the tool reports the arity mismatch before running and exits non-zero.
5. **Given** any of the above errors, **When** reported, **Then** the output includes a
   caret/underline pointing at the relevant span in the source line.

---

### User Story 3 - Run task bodies in multiple languages (Priority: P3)

A task author chooses the language a task body is written in: the default shell, or an
explicitly declared interpreter such as Python or Node, or a custom interpreter. Values
defined in the task file are substituted into the body before it runs. A shell task behaves
identically across operating systems without the author special-casing platforms.

**Why this priority**: Multi-language bodies as a first-class concept (not a shebang hack)
is one of the three headline differentiators. It meaningfully broadens who and what the tool
serves, but the single-language runner (Stories 1–2) is independently valuable first.

**Independent Test**: Define one shell task and one Python task in the same file; run each;
confirm each executes under its declared runtime and that interpolated values appear
correctly in both.

**Acceptance Scenarios**:

1. **Given** a task with no declared executor, **When** it runs, **Then** its body executes
   as a shell script with consistent behavior across supported operating systems.
2. **Given** a task that declares a Python (or Node) executor, **When** it runs, **Then** its
   body executes under that interpreter.
3. **Given** a task body containing a `{{ ... }}` interpolation of a defined value, **When**
   the task runs, **Then** the resolved value is present in what executes.
4. **Given** a task that declares an interpreter not installed on the machine, **When** it
   runs, **Then** the tool reports a clear, actionable error and exits non-zero.

---

### User Story 4 - Share the same tasks with AI agents and IDEs (Priority: P4)

Rune can present its tasks to an external AI agent or IDE through a standard agent tool
interface: the agent can list the available (non-private) tasks, see each task's
description and the parameters it accepts, and invoke a task — receiving its output and
success/failure back. Additionally, a task itself can be an AI-agent task whose body is a
natural-language instruction; when run, an agent carries out that instruction and may call
other project tasks as tools. Agents get read-only access by default; tasks marked dangerous
require explicit opt-in or confirmation before an agent may run them. Secret values never
appear in task definitions, task descriptions, or anything exposed to an agent.

**Why this priority**: This is the project's reason to exist — one shared automation layer
for people and agents. It is delivered after the human-facing runner is solid and trusted,
because the agent surface mirrors the same tasks and validation.

**Independent Test**: Start Rune's agent-facing mode against a task file; from an external
client, list the tasks and confirm names, descriptions, and parameter shapes are accurate;
invoke a safe task and confirm output/result are returned; confirm a task marked dangerous
is not run without explicit authorization.

**Acceptance Scenarios**:

1. **Given** a task file with several non-private tasks, **When** an external agent connects,
   **Then** it can enumerate those tasks with their descriptions and parameter definitions,
   and private tasks are not listed.
2. **Given** an agent invokes a listed task with valid arguments, **When** it runs, **Then**
   the task executes through the same engine as the CLI and the output plus success/failure
   are returned to the agent.
3. **Given** a task marked as requiring confirmation/dangerous, **When** an agent attempts to
   run it without authorization, **Then** the tool refuses or requires explicit approval per
   policy.
4. **Given** an AI-agent task whose body is an instruction, **When** it runs, **Then** an
   agent performs the instruction, may call other allowed tasks, and its result becomes the
   task's outcome; by default the agent may only call non-dangerous tasks.
5. **Given** credentials are supplied only through the environment, **When** tasks are
   exposed to an agent, **Then** no secret value appears in any task description, schema, or
   listing.

---

### User Story 5 - CI/CD ergonomics, speed, and composition (Priority: P5)

For automation and larger projects, Rune offers operational conveniences: list tasks
(grouped, with docs), preview what a run would do without executing it, emit a
machine-readable description of the task file, run independent prerequisites in parallel,
optionally skip a task when its declared inputs and outputs are unchanged since the last run,
and return predictable exit codes. Task files can be composed — splicing in another file's
definitions, or loading another file as a namespaced sub-group of tasks.

**Why this priority**: These make Rune pleasant in CI and at scale, but they are refinements
on top of a working, trustworthy runner and are not required for initial value.

**Independent Test**: Run the list and preview/dry-run commands and confirm no commands
execute during preview; configure a task with declared inputs/outputs, run it twice without
changing inputs, and confirm the second run is skipped; confirm exit codes reflect outcomes.

**Acceptance Scenarios**:

1. **Given** any task file, **When** the developer requests a listing, **Then** non-private
   tasks are shown with their documentation and groups; **When** they request a dry run,
   **Then** the planned actions are reported without executing them.
2. **Given** a task that declares its inputs and outputs for caching, **When** it is run a
   second time with unchanged inputs and existing outputs, **Then** it is skipped and the
   skip is reported; **When** an input changes, **Then** it runs again.
3. **Given** a task whose prerequisites are independent and marked for parallel execution,
   **When** it runs, **Then** those prerequisites run concurrently and the bound on
   concurrency is respected.
4. **Given** a task file that imports another file or loads a namespaced module, **When**
   tasks are listed or run, **Then** the composed/namespaced tasks are available and
   addressable.
5. **Given** a request for machine-readable output, **When** issued, **Then** the tool emits
   a structured representation of the parsed task file suitable for tooling.

---

### Edge Cases

- **No task file found**: running the tool where no task file exists in the current
  directory or any ancestor produces a clear "no task file found" message and a non-zero exit.
- **Dependency cycle**: reported before execution with the full cycle path (Story 2).
- **Wrong arity / missing required parameter**: reported before execution (Story 2).
- **Interpreter not installed**: a task declaring an unavailable interpreter fails with an
  actionable message naming the missing runtime (Story 3).
- **Name collisions across composed files**: importing two definitions of the same task name
  is reported as a conflict rather than silently picking one.
- **Operating-system-filtered task on the wrong OS**: a task restricted to one OS is excluded
  from listing/dispatch on others; requesting it explicitly on the wrong OS reports why it
  is unavailable.
- **Cache inputs missing or outputs deleted**: a cached task whose declared outputs are gone,
  or whose declared input set cannot be resolved, runs rather than being skipped.
- **Agent CLI missing or unauthenticated**: an AI-agent task run when its configured agent
  CLI is not installed or not signed in fails with a clear, actionable error naming the
  missing/unauthenticated tool, never by leaking or inventing credentials.
- **Dangerous task requested by an agent**: refused or gated per policy (Story 4).
- **Inconsistent body indentation**: a task body that mixes indentation styles in a way the
  language forbids is reported as a located error, not run ambiguously.
- **First command in a task fails**: execution stops at the first failure (unless a line is
  explicitly marked to continue on error) and the failing task, line, and exit code are
  surfaced.

## Requirements *(mandatory)*

### Functional Requirements

**Authoring & discovery**

- **FR-001**: The system MUST read task definitions from a single project task file and MUST
  locate it by searching the current directory and ancestors; users MUST be able to point at
  a specific file explicitly.
- **FR-002**: Users MUST be able to define named tasks, each with a body of one or more
  commands, and run any task by name from the command line. Task bodies MUST use significant
  indentation that is consistent within a task; mixing tabs and spaces within a single body
  MUST be reported as a located error rather than executed ambiguously.
- **FR-003**: The system MUST support task parameters, including parameters with default
  values and a trailing variadic parameter, and MUST substitute provided/default values into
  the task body.
- **FR-004**: The system MUST support prerequisite dependencies that run before a task and
  post-step hooks that run after a task only on its success.
- **FR-005**: Within a single invocation, a given task with a given set of arguments MUST run
  at most once (memoized), even if reachable through multiple dependency paths.
- **FR-006**: Users MUST be able to define reusable values (variables) and reference them in
  task bodies and in other values; the system MUST support overriding these values from the
  command line at run time.
- **FR-007**: The system MUST support a small, non-looping expression capability for values
  and interpolations (text concatenation, path joining, conditional selection, equality and
  pattern comparison, and a fixed set of built-in helper functions for environment lookup,
  OS/architecture info, path manipulation, string manipulation, file existence/reading, and
  hashing). The expression capability MUST NOT provide general-purpose loops or recursion.
- **FR-008**: Users MUST be able to attach documentation to a task (via comments or an
  attribute) that surfaces in the task listing.
- **FR-009**: The system MUST support task attributes including at minimum: hide-from-listing
  (private), require-confirmation (dangerous), grouping, run-prerequisites-in-parallel,
  operating-system restriction, working-directory override, per-task environment values, and
  declared inputs/outputs for caching.
- **FR-010**: The system MUST support project-level settings including at minimum: a default
  task, loading an environment file, exporting defined values to task bodies, a working
  directory, and quiet operation.
- **FR-011**: The system MUST support composition: splicing in another task file's
  definitions, and loading another task file as a namespaced group addressable by a prefix.

**Validation & diagnostics**

- **FR-012**: The system MUST validate the entire task file before executing anything, and
  MUST detect and report: references to unknown tasks, references to undefined variables,
  circular dependencies (including the cycle path), and parameter arity mismatches.
- **FR-013**: Every parse and validation error MUST identify the file, line, and column and
  MUST visually indicate the offending span in the source.
- **FR-014**: If validation fails, the system MUST NOT execute any task body or produce any
  task side effects, and MUST exit with a non-zero status.

**Execution**

- **FR-015**: The system MUST treat every task as always-run; it MUST NOT skip work based on
  file timestamps. Skipping MUST occur only for tasks that explicitly opt into input/output
  caching, and a skip MUST be reported to the user.
- **FR-016**: The default task body MUST execute as a shell script whose behavior is identical
  across the supported operating systems without author-side platform special-casing.
- **FR-017**: A task MUST be able to declare an alternate executor (e.g., Python, Node, or a
  custom interpreter) under which its body runs; the system MUST report a clear error if the
  required interpreter is unavailable.
- **FR-018**: The system MUST stop a task at its first failing command by default, surface the
  failing task/line/exit code, and provide a way to mark individual command lines to continue
  on error and to suppress command echo.
- **FR-019**: The system MUST run independent prerequisites concurrently when a task opts into
  parallel execution, respecting a sensible concurrency bound.
- **FR-020**: For caching-enabled tasks, the system MUST decide to skip or run based on a
  content fingerprint of the declared inputs together with the task body and resolved values,
  and MUST run the task if declared outputs are absent.
- **FR-021**: The system MUST return predictable, documented exit codes that distinguish
  success, task failure, and usage/validation errors.

**Operational tooling**

- **FR-022**: Users MUST be able to list available tasks (excluding private ones) with their
  documentation and groups, and to preview (dry-run) what an invocation would do without
  executing it.
- **FR-023**: The system MUST be able to emit a machine-readable representation of the parsed
  task file for external tooling.
- **FR-024**: The system MUST provide helpful top-level help and version information.

**AI / agent integration**

- **FR-025**: The system MUST be able to expose its non-private tasks to an external AI agent
  or IDE as discoverable, invokable tools, where each tool carries the task's name, its
  documentation as a description, and a parameter definition derived from the task's
  parameters.
- **FR-026**: When an agent invokes an exposed task, the system MUST execute it through the
  same engine used by the command line and MUST return the task's output and success/failure
  to the caller.
- **FR-027**: The system MUST support an AI-agent task type whose body is a natural-language
  instruction; running it MUST drive a configured agent that may call other permitted tasks
  and whose result becomes the task's outcome. In v1 the agent is an **installed agent CLI**
  (e.g. `claude`, `codex`, `copilot`) invoked via a configurable command; the agent-provider
  interface MUST remain open to other backends. If the configured agent CLI is absent or
  unauthenticated, the task MUST fail with a clear, actionable error.
- **FR-028**: Destructiveness MUST be author-declared: a task is gated from unattended agent
  execution if and only if its author marks it (the confirmation/destructive attribute). The
  system MUST NOT infer destructiveness from task contents. Private tasks MUST NOT be exposed
  to agents at all; other non-private tasks are agent-callable, and gated tasks MUST require
  explicit opt-in or approval before an agent may run them. Operators MAY additionally
  restrict agent-callable tasks via an explicit allow-list.
- **FR-029**: The system MUST NOT read model/agent credentials from the task file; when driving
  an agent CLI, authentication is delegated to that CLI's own session/credentials, and any
  other provider's credentials MUST come from the environment only. Secret values MUST NEVER
  appear in task descriptions, parameter schemas, or listings exposed to agents.
- **FR-030**: The agent/model integration MUST be vendor-neutral — the system MUST allow
  different agent providers to be configured without favoring or hard-requiring any single
  vendor.
- **FR-031**: The system MUST always offer a same-machine (local) interface for exposing
  tasks to agents/IDEs. Any remote (network-reachable) endpoint MUST be opt-in (off by
  default), MUST bind to localhost by default, and MUST require an explicit authentication
  token before any task may be listed or invoked over it.

**Compatibility**

- **FR-032**: The system MUST be distributed as a single self-contained executable that runs
  on Linux, macOS, and Windows.
- **FR-033**: Changes to task-file interpretation MUST preserve backward compatibility: an
  existing task file MUST continue to behave the same across tool upgrades, and any breaking
  change MUST be opt-in per file.

### Key Entities

- **Task file**: The single source-of-truth document for a project, discovered by walking up
  the directory tree; contains settings, variables, and tasks.
- **Task**: A named unit of work with an optional doc, parameters, an executor, prerequisites,
  post-step hooks, attributes, and a body.
- **Parameter**: A named input to a task; may have a default value; the last one may accept
  one-or-more / zero-or-more values.
- **Dependency / Post-hook**: A reference from one task to others that run before it, or after
  it on success.
- **Variable**: A named, statically evaluated value usable in expressions, bodies, and other
  values; overridable at run time.
- **Setting**: A project-level configuration (default task, environment file, export, working
  directory, quiet, etc.).
- **Attribute**: A declarative annotation on a task (private, confirm, group, parallel, OS
  filter, working directory, env, cache inputs/outputs, executor override).
- **Executor**: The runtime a task body executes under (shell by default; Python, Node, custom
  interpreter; or AI agent).
- **Module / Import**: A composition unit — spliced definitions or a namespaced group from
  another task file.
- **Cache fingerprint**: The stored content signature (inputs + body + resolved values) used
  to decide whether a caching-enabled task may be skipped.
- **Exposed tool**: The agent-facing representation of a task (name, description, parameter
  schema, destructiveness/network hints).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer who already knows `make` or `just` can read an unfamiliar Rune task
  file and correctly predict what each task does within 30 seconds per task, with no external
  reference.
- **SC-002**: 100% of the defined validation error classes (unknown task, undefined variable,
  dependency cycle, arity mismatch) are reported before any side effect occurs, each with the
  exact file/line/column.
- **SC-003**: From a standing start, a new user can install the tool, write a three-line task,
  and run it successfully in under 5 minutes.
- **SC-004**: A task file that uses only the default shell executor produces identical observable
  results across Linux, macOS, and Windows in at least 99% of a representative compatibility
  test suite, with no platform-specific edits.
- **SC-005**: An external AI agent or IDE can discover and successfully run a project task with
  no project-specific glue code written by the user.
- **SC-006**: A caching-enabled task re-run with unchanged inputs completes in under 10% of its
  original execution time (i.e., the skip is effectively instantaneous) and is reported as
  skipped.
- **SC-007**: Zero secret values appear in any task listing, description, or agent-facing schema
  across the test suite (measured: 0 occurrences).
- **SC-008**: Exit codes correctly reflect outcome (success vs. task failure vs. usage/validation
  error) in 100% of tested scenarios.
- **SC-009**: Independent prerequisites marked for parallel execution complete measurably faster
  than the same set run sequentially on a multi-core machine (wall-clock reduction observable
  and bounded by available cores).
- **SC-010**: In 100% of tested scenarios, a remote agent endpoint rejects every list/invoke
  attempt that lacks a valid auth token, and by default is not reachable from non-localhost
  addresses.

## Assumptions

- **Naming is provisional**: the working name "Rune" (file name "Runefile", extension ".rune")
  is used throughout pending the final naming decision; the spec's substance does not depend on
  the final name.
- **Body syntax** uses significant indentation (familiar to `make`/`just` users), with a strict
  rule against mixing indentation styles within a task body.
- **Expression capability stays small and total** — no user-authored loops or recursion in the
  task-file language; non-trivial logic lives in task bodies.
- **Alternate-language bodies run via the user's installed interpreters** (which must be present
  on the machine/PATH); the tool does not bundle Python/Node runtimes.
- **The AI-agent task type drives an installed agent CLI** (Claude CLI, Codex, Copilot, or a
  compatible tool) via a configurable command; that CLI supplies its own authentication. Rune
  never reads credentials from the task file, agent access defaults to read-only, and the
  provider interface stays open to other backends (e.g. direct hosted APIs) later.
- **Caching is opt-in per task** and based on content hashing of declared inputs/outputs plus the
  task body and resolved values; cache state is stored in a project-local location.
- **Backward compatibility is a standing promise** ("no breaking 2.0"; breaking changes opt-in per
  file), consistent with the project constitution.
- **Delivery is phased**: the human-facing command runner with static validation (Stories 1–2)
  is the MVP; multi-language bodies (Story 3) follow; the AI/agent surface (Story 4) and advanced
  CI/CD ergonomics and composition (Story 5) complete the v1 identity.
