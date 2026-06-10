# Feature Specification: Modern CLI Interface

**Feature Branch**: `005-cobra-cli-adoption`

**Created**: 2026-06-10

**Status**: Draft

**Input**: User description: "i want to use cobra as lib for cli interface — do research and use golang skill (cobra user guide, cobra.dev shell-completion). so want to have a modern cli interface for app."

## Clarifications

### Session 2026-06-10

- Q: When a task's name collides with a built-in command, how is the task invoked? → A: Via the `--` separator — `rune -- <task> [args…]` forces task interpretation (built-in commands still take precedence otherwise).
- Q: Should dynamic task-name completion show each task's documentation summary? → A: Yes — completion includes each task's doc summary as the completion description for shells that display it (e.g. zsh, fish).
- Q: How should `serve` and `mcp` appear, given both must keep working? → A: `serve` is the canonical command (stdio by default, `--http` for HTTP); `mcp` is a documented alias for stdio serving.

## User Scenarios & Testing *(mandatory)*

Rune today exposes a single root command that hand-dispatches its built-in
capabilities (`mcp`, `serve`, `completion`) by inspecting positional arguments,
hand-parses `serve`'s options in a manual loop, and wraps completion in a custom
helper. The capabilities work, but they are **undiscoverable** (not listed in
help), **undocumented per-command** (no `serve --help`), and **inconsistent**
(server flags are unvalidated and untab-completable). This feature delivers a
modern, self-describing command-line experience while preserving every existing
behavior the tool's users and downstream CI pipelines depend on.

### User Story 1 - Discover and understand every command from help (Priority: P1)

A developer who has never used Rune runs `rune --help`, sees every built-in
command with a one-line description, then runs `rune serve --help` (or any
command) and learns its purpose, every flag, and at least one usage example —
without reading source code or external docs.

**Why this priority**: Self-description is the defining trait of a modern CLI and
the foundation every other improvement builds on. Shipping only this already
turns Rune's hidden dispatch into a discoverable, documented surface — a viable,
demonstrable improvement on its own.

**Independent Test**: Run `rune --help` and confirm all built-in commands appear
with descriptions and that user tasks are clearly distinguished with a pointer to
how they are listed; run `rune <command> --help` for each built-in command and
confirm usage, flags, and an example are present.

**Acceptance Scenarios**:

1. **Given** a terminal in any directory, **When** the user runs `rune --help`,
   **Then** the output lists every built-in command (serve/MCP, completion,
   version, help) with a short description and explains how to discover user
   tasks (e.g., points to `--list`).
2. **Given** any built-in command, **When** the user runs `rune <command> --help`,
   **Then** the output shows its usage line, every flag with a description, and at
   least one concrete example.
3. **Given** the user runs `rune version` (and, separately, `rune --version`),
   **Then** both report the same version and build commit.

---

### User Story 2 - Shell completion, including live task names (Priority: P2)

A developer enables completion for their shell (bash, zsh, fish, or PowerShell)
once, then relies on TAB to complete Rune's built-in commands, its global flags,
and — crucially — the **task names defined in the current Runefile**, which are
dynamic and project-specific.

**Why this priority**: Completion is a high-value modern-CLI affordance, and
dynamic task-name completion is the part users cannot get any other way. It
depends on the discoverable command structure from P1 but delivers independent,
demonstrable value.

**Independent Test**: Install the generated completion script for a shell, type
`rune ` and press TAB to see built-in commands and the current Runefile's task
names; type `rune --` and press TAB to see global flags.

**Acceptance Scenarios**:

1. **Given** a supported shell, **When** the user runs the documented completion
   command for that shell, **Then** a valid completion script is written to
   stdout and the command's help explains how to install it for that shell.
2. **Given** completion is installed and the working directory contains a valid
   Runefile, **When** the user presses TAB after `rune `, **Then** the
   suggestions include the built-in commands and the non-private task names from
   that Runefile, with each task's doc summary shown as its description in shells
   that support it.
3. **Given** completion is installed and no Runefile is found (or the Runefile has
   parse errors), **When** the user presses TAB after `rune `, **Then** completion
   degrades gracefully — it suggests built-in commands and offers no task names,
   and never emits an error or stack trace into the shell session.

---

### User Story 3 - Idiomatic, validated subcommands and friendly errors (Priority: P3)

A developer running the MCP/server capability gets first-class, validated flags
with help and completion, and anyone who mistypes a command or passes a bad flag
combination gets a concise, helpful error instead of a silent failure or a wall of
usage text.

**Why this priority**: This polishes correctness and ergonomics on top of the
discoverable structure (P1). It is the lowest priority because the capabilities
already function; this story makes their edges trustworthy and friendly.

**Independent Test**: Run the server command with each flag and with an invalid
combination; mistype a command name and confirm a "did you mean" style
suggestion; pass an unknown flag and confirm a concise error without a full usage
dump.

**Acceptance Scenarios**:

1. **Given** the server command, **When** the user runs it with its documented
   flags (e.g., HTTP mode, address, token file), **Then** the flags are parsed and
   validated, and mutually-incompatible or malformed combinations produce a clear
   message and a usage-error exit code.
2. **Given** a mistyped command (e.g., `rune serv`), **When** it is run, **Then**
   the error suggests the closest valid command (e.g., "did you mean serve?").
3. **Given** an unknown flag on a built-in command, **When** it is run, **Then**
   the error is concise and does not print the full usage/help text on every
   error.

---

### Edge Cases

- **Task name collides with a built-in command** (a Runefile task literally named
  `serve`, `mcp`, `completion`, `version`, or `help`): the built-in command takes
  precedence, and the task remains reachable via the `--` separator
  (`rune -- <task> [args…]`). Note: a task named `version` or `help` is runnable
  today as a bare positional; once those names become subcommands, the `--` form is
  the compatibility path that keeps such tasks reachable.
- **Trailing task flags that resemble Rune's own flags** (e.g., `rune build
  --watch` where `--watch` is meant for the task): everything after the task name
  is passed through to the task untouched; Rune does not interpret it as a global
  flag.
- **`rune` invoked with no arguments**: behaves as it does today (no regression).
- **Non-interactive / piped output**: color is disabled and output stays
  machine-friendly; diagnostics never contaminate stdout.
- **Interrupt during a running task** (SIGINT/SIGTERM): the run is cancelled and
  the process exits with the established signal exit code.
- **Completion requested for an unsupported shell**: a clear error naming the
  supported shells, not a panic.

## Requirements *(mandatory)*

### Functional Requirements

**Command discovery & help**

- **FR-001**: The CLI MUST present each built-in capability (server/MCP,
  completion, version, help) as a named, discoverable command listed in the
  top-level help output.
- **FR-002**: Each built-in command MUST provide its own help text containing a
  usage line, a description, every flag with a description, and at least one
  concrete example.
- **FR-003**: Top-level help MUST clearly distinguish built-in commands from
  user-defined tasks and MUST tell the user how to list available tasks.
- **FR-004**: The CLI MUST provide a `version` command that reports the same
  version and build information as the `--version` flag.
- **FR-005**: Help and usage text MUST be available for every command via a
  standard `--help`/`-h` flag and via a `help` command.

**Task invocation (preserved behavior)**

- **FR-006**: Task names MUST remain positional arguments; `rune <task> [args…]`
  MUST execute the named task.
- **FR-007**: All arguments and flags that appear after the task name MUST be
  passed to the task untouched and MUST NOT be interpreted as Rune's own flags.
- **FR-008**: Built-in command names are reserved and MUST take precedence over an
  identically-named task. To keep every task reachable, the CLI MUST treat a `--`
  separator as forcing task interpretation: `rune -- <task> [args…]` MUST run the
  named task even when its name matches a reserved command.

**Shell completion**

- **FR-009**: The CLI MUST generate shell-completion scripts for bash, zsh, fish,
  and PowerShell via a documented command, and that command's help MUST include
  per-shell installation instructions.
- **FR-010**: Generated completions MUST cover built-in commands and global flags.
- **FR-011**: Generated completions MUST dynamically suggest the non-private task
  names defined in the current Runefile, and MUST include each task's documentation
  summary as the completion description for shells that display descriptions (e.g.,
  zsh, fish).
- **FR-012**: When no Runefile is found or the Runefile cannot be parsed,
  completion MUST degrade gracefully — suggesting built-in commands, offering no
  task names, and never emitting errors into the shell session.
- **FR-013**: Requesting completion for an unsupported shell MUST produce a clear
  error that names the supported shells.

**Flags & errors**

- **FR-014**: The server command's options MUST be first-class, documented,
  validated flags (with completion where a fixed set of values applies), replacing
  ad-hoc manual argument parsing.
- **FR-015**: Invalid or mutually-incompatible flag combinations MUST produce a
  clear error message and the established usage-error exit code.
- **FR-016**: Unknown commands MUST, where a close match exists, produce a "did you
  mean" style suggestion.
- **FR-017**: Errors MUST be concise and MUST NOT print the full usage/help text on
  every failure.

**Preserved contract (non-regression)**

- **FR-018**: The established exit-code contract MUST be preserved unchanged
  (success, static-validation failure, usage error, task failure, and
  signal-termination codes).
- **FR-019**: stdout MUST remain reserved for program/task output suitable for
  piping; all of Rune's own diagnostics, logs, and banners MUST go to stderr.
- **FR-020**: Static validation diagnostics MUST continue to render with
  `file:line:col` spans, and the existing suppression of the failure banner for
  `[no-exit-message]` tasks MUST be preserved.
- **FR-021**: Color behavior MUST be preserved: honored only on a TTY, disabled
  when `NO_COLOR` is set or color is globally disabled.
- **FR-022**: Interrupt handling (SIGINT/SIGTERM cancels the run) MUST be preserved.

**Command identity**

- **FR-023**: The stdio MCP server MUST be invocable as both `rune serve` (its
  default mode) and `rune mcp`; `serve` is the canonical, primary command and `mcp`
  MUST be presented in help as a documented alias for stdio serving.

### Key Entities

- **Built-in Command**: A named, reserved capability of the tool (server/MCP,
  completion, version, help). Attributes: name, short and long description, set of
  flags, usage examples.
- **Task Invocation**: A dynamic, positional request to run a Runefile task.
  Resolved at runtime from the discovered Runefile; not part of the fixed command
  set.
- **Completion Specification**: The per-shell completion output plus the dynamic
  provider that supplies current task names; must function with or without a
  resolvable Runefile.
- **Flag**: A named option on a command. Attributes: long name, optional
  shorthand, value type, default, required/exclusive relationships, optional
  value-completion source.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of built-in capabilities are discoverable from `rune --help`
  (no capability is reachable only through undocumented argument inspection).
- **SC-002**: Every built-in command's `--help` includes a usage line, all of its
  flags, and at least one example (verifiable per command).
- **SC-003**: Completion scripts install and function on all four supported shells,
  and after `rune ` + TAB the suggestions include both built-in commands and the
  current Runefile's non-private task names (with task descriptions shown in shells
  that support them).
- **SC-004**: Using only `rune --help` output, a first-time user can identify how
  to start the server/MCP capability and how to list tasks, without consulting
  external documentation.
- **SC-005**: A mistyped built-in command (e.g., `rune serv`) yields a suggestion
  of the intended command.
- **SC-006**: The entire existing CLI behavior suite (stdout/stderr, exit codes,
  task argument pass-through, signal handling, validation diagnostics) passes
  unchanged — zero behavioral regressions in the established contract.

## Assumptions

- **Framework**: Per the user's explicit direction, the CLI is built on the Cobra
  command framework, which is already a project dependency (`github.com/spf13/cobra`
  in `go.mod`). This feature adopts Cobra idiomatically (real subcommands, native
  flag binding, built-in completion generators) rather than introducing a new
  dependency. Shell completion uses Cobra's built-in generators for the four shells
  it supports, per the referenced cobra.dev shell-completion guide.
- **Configuration scope**: Viper-style layered configuration (env/file/flag
  precedence) is **not** introduced. Rune's configuration model is the Runefile and
  its settings; adding a separate config-layering system is out of scope for this
  feature.
- **Precedence default**: Built-in command names take precedence over
  identically-named tasks; the `--` separator (`rune -- <task>`) is the documented
  escape hatch that keeps such tasks reachable (see FR-008).
- **Backward compatibility is binding**: Per the project constitution's
  backward-compatibility promise, the externally observable contract (exit codes,
  stdout/stderr discipline, task pass-through, diagnostics, color, signal handling)
  is sacrosanct and is preserved exactly. The existing integration suite is the
  guardrail for this.
- **Single static binary preserved**: No new runtime dependency that would break
  single-static-binary distribution or cross-compilation is introduced.

## Out of Scope

- Viper / layered env+file configuration for the CLI.
- New end-user features beyond making existing capabilities idiomatic and
  discoverable (e.g., no new scaffolding commands).
- Any change to the Runefile language, task semantics, or the existing interactive
  task picker (`--choose`).
- Changes to the MCP protocol surface itself (this feature concerns how the server
  capability is invoked and documented from the CLI, not the protocol).
