# Feature Specification: Styled CLI Output & Friendlier Help

**Feature Branch**: `008-styled-cli-output`

**Created**: 2026-06-30

**Status**: Draft

**Input**: User description: "Give Rune's non-interactive CLI output a modern, colorful, polished look without changing what scripts, pipes, and CI see — styled `--list`, status/echo/cache lines, sharper diagnostics, an explicit color-control flag, and a friendlier `--help` with examples."

## Clarifications

### Session 2026-06-30

- Q: How does the byte-for-byte invariance reconcile with the intentional `--help` redesign? → A: Invariance applies to every surface except `--help`/usage; the redesigned plain `--help` becomes a new, deliberately reviewed baseline (golden/snapshot regenerated). Colored vs plain help differ only by ANSI.
- Q: Which inputs does the color decision honor? → A: Only `NO_COLOR`, the `--color=auto|always|never` flag, and TTY auto-detection. `FORCE_COLOR`, `CLICOLOR`, and `CLICOLOR_FORCE` are explicitly NOT honored; `--color=always` is the force-on path.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scannable task list (Priority: P1)

A developer runs `rune --list` in their terminal to discover what tasks a Runefile
offers. Today the listing is flat, uncolored text; task names, group headings, and
doc summaries all read the same. With styling, task names stand out, group headings
are visually distinct, and the trailing doc summaries recede so the eye lands on the
names first.

**Why this priority**: `--list` is the most frequently read everyday surface and the
primary way users discover tasks. It is currently fully plain, so it has the highest
readability payoff and is the natural MVP.

**Independent Test**: Run `rune --list` against a Runefile with grouped and ungrouped
tasks on a color-capable terminal; confirm names, group headings, and doc text are
visually differentiated. Pipe the same command to a file and confirm the bytes match
today's plain output exactly.

**Acceptance Scenarios**:

1. **Given** a Runefile with grouped, ungrouped, and documented tasks, **When** `rune --list` runs on a color-capable TTY with `NO_COLOR` unset, **Then** task names are emphasized, group headings are styled, and doc summaries are dimmed — while the text content, order, column alignment, and indentation are unchanged from the plain listing.
2. **Given** the same Runefile, **When** `rune --list` output is redirected to a file or pipe, **Then** the output is byte-for-byte identical to the current (pre-feature) plain listing.

---

### User Story 2 - Scripts and CI see unchanged output (Priority: P1)

A maintainer relies on Rune in CI and in shell scripts, and the test suite asserts
exact stdout/stderr via golden files. Styling must never leak ANSI escapes, shift
columns, or reorder content into any captured, redirected, or non-interactive stream.

**Why this priority**: This is the hard guardrail the whole feature is gated on. If
styling changes scriptable output, it breaks CI, downstream parsers, and the golden
suite. It must hold for every styled surface, so it is co-equal priority with the
first visible win.

**Independent Test**: For every styled command, capture stdout and stderr while not
attached to a TTY (piped) and again with `NO_COLOR=1`, and with `--color=never`;
compare against the pre-feature output and confirm zero byte differences.

**Acceptance Scenarios**:

1. **Given** any Rune command whose output is styled on a TTY, **When** that command runs with stdout/stderr not attached to a terminal, **Then** no ANSI escape sequences are emitted and the output is byte-for-byte identical to the pre-feature release.
2. **Given** any styled command, **When** `NO_COLOR` is set to any non-empty value, **Then** no color is emitted on any stream regardless of TTY status.
3. **Given** the styled diagnostics, **When** rendered in color, **Then** the `file:line:col` prefix, the source line, and the caret-span underline occupy the exact same columns as the plain rendering (color adds no width).

---

### User Story 3 - Legible run output (Priority: P2)

A developer runs tasks and watches the per-task status, command echo, and cache-hit
notices stream by on stderr. Styling makes the status verbs (`running:`, `cached:`,
`would run:`) semantically colored, dims the echoed command lines and cache notices
so real output stands out, and renders warnings in a warning color.

**Why this priority**: This is the second-most-read everyday surface. It builds on the
shared theme established by P1 and improves the signal-to-noise ratio of normal runs,
but the tool is fully usable without it.

**Independent Test**: Run a task that hits the cache and a task that executes, on a
TTY; confirm the status verbs are colored by meaning (e.g. cache-hit muted, execution
neutral/active, failure in the error color) and command echo is dimmed. Pipe the same
runs and confirm byte-identical plain output.

**Acceptance Scenarios**:

1. **Given** a task that runs and a task that is served from cache, **When** executed on a color TTY, **Then** the `running:` / `cached:` / `would run:` / `would skip (cached):` labels are styled by semantic role and the echoed command and cache-hit notice are dimmed — with identical text and stream assignment to today.
2. **Given** the `@` echo-suppression sigil and the `--quiet` flag, **When** a task runs, **Then** the existing suppression behavior is unchanged; styling never causes a suppressed line to appear.
3. **Given** a cache-write failure warning, **When** emitted on a color TTY, **Then** it is rendered in the shared warning color, matching the warning role used by diagnostics.

---

### User Story 4 - Explicit color control (Priority: P2)

A user wants to force color on through a pager or CI log viewer that strips TTY
detection, or force it off entirely. A global `--color=auto|always|never` flag gives
deliberate control beyond the automatic `NO_COLOR` + TTY behavior.

**Why this priority**: A standard CLI convenience that complements the automatic
behavior. Valuable but not required for the core styling to ship; `auto` preserves
today's behavior as the default.

**Independent Test**: Run a styled command with `--color=always` through a pipe and
confirm ANSI is present; run with `--color=never` on a TTY and confirm no ANSI;
run with `--color=auto` (or no flag) and confirm TTY-based behavior.

**Acceptance Scenarios**:

1. **Given** `--color=always`, **When** a styled command is piped or redirected, **Then** color is emitted despite the absence of a TTY.
2. **Given** `--color=never`, **When** a styled command runs on a color-capable TTY, **Then** no color is emitted.
3. **Given** `--color=auto` or no `--color` flag, **When** a styled command runs, **Then** color is decided by TTY detection and `NO_COLOR` exactly as in the pre-feature release.
4. **Given** an invalid value such as `--color=sometimes`, **When** the command is invoked, **Then** Rune reports a clear error and exits non-zero without running tasks.

---

### User Story 5 - Sharper diagnostics (Priority: P3)

A user hits a Runefile error and reads the analyzer diagnostic. Diagnostics already
color the severity word and caret; this story refines the emphasis (e.g. highlighting
the `file:line:col` locator and brightening the caret span) so the eye jumps to the
location and the offending span — without ever shifting a column.

**Why this priority**: Diagnostics are already partly styled and functional, so this is
a refinement rather than a gap. "Errors are a feature," so alignment integrity is a
hard constraint on any change here.

**Independent Test**: Render a diagnostic with color on and off; confirm the colored
version emphasizes the locator and caret span while the plain version is byte-identical
to today's golden file, and the caret aligns in both.

**Acceptance Scenarios**:

1. **Given** a diagnostic with a caret span, **When** rendered in color, **Then** the locator and/or caret span are emphasized using the shared theme and the column layout is identical to the plain rendering.
2. **Given** the plain (non-color) diagnostic, **When** rendered, **Then** it matches the existing diagnostic golden output byte-for-byte.

---

### User Story 6 - Friendlier help with examples (Priority: P3)

A new user runs `rune --help` (or `rune <command> --help`) to learn the tool. The help
is redesigned into grouped, plain-language sections with concrete, runnable examples
for the common workflows (running a task, listing tasks, the interactive picker,
serving over MCP), with colored headings on a TTY and clear plain text when piped.

**Why this priority**: Improves first-run discoverability. Independent of the coloring
work in structure, and the tool is fully usable with today's default auto-generated
help, so it is lowest priority.

**Independent Test**: Run `rune --help` and a subcommand `--help`; confirm grouped
sections, plain-language descriptions for each flag and command, and at least one
worked example per common workflow. Confirm help still renders informatively when
piped (no required color).

**Acceptance Scenarios**:

1. **Given** `rune --help` on a TTY, **When** displayed, **Then** it shows grouped sections (usage, common commands, flags, examples) with colored headings and plain-language descriptions.
2. **Given** `rune --help` piped to a file, **When** displayed, **Then** the same structured content is present with no ANSI escapes and remains readable.
3. **Given** a user who has never used Rune, **When** they read `--help`, **Then** they can find how to run a task, list tasks, and launch the picker from worked examples without consulting external docs.

---

### Edge Cases

- **Dumb / unknown terminals**: when the terminal cannot render styling, behavior degrades to plain text (treated as `auto` with no TTY benefit) rather than emitting broken escapes.
- **Mixed stream redirection**: when one stream is a TTY and the other is redirected (e.g. stdout piped, stderr to terminal), each surface's color decision uses the TTY status of the stream it writes to, so stdout-bound `--list` and stderr-bound status lines decide independently.
- **`NO_COLOR` vs `--color=always`**: an explicit `--color` value is a deliberate override and takes precedence over `NO_COLOR` and TTY detection; `--color=auto` defers to `NO_COLOR` then TTY.
- **Very long task names / doc summaries**: styling must not change the existing column-width and truncation behavior of the listing.
- **Diagnostics with multi-byte or wide characters in the source line**: caret alignment must remain correct in both colored and plain renderings.
- **Empty / no-task Runefile, or all tasks filtered out**: the styled listing falls back to the same content the plain listing shows today.

## Requirements *(mandatory)*

### Functional Requirements

#### Shared theme & color decision

- **FR-001**: The system MUST define a single, shared, semantic theme with named roles — at minimum: error, warning, success, task-name, heading/group, and muted/meta — defined in one place and reused across every styled surface.
- **FR-002**: Every styled surface MUST derive its styling from a single color decision; no surface may independently re-implement TTY/`NO_COLOR` detection.
- **FR-003**: When color is disabled, every theme role MUST degrade to plain, unstyled text (the role becomes a no-op), producing output with no ANSI escape sequences.
- **FR-004**: The color decision for a surface MUST consider the TTY status of the stream that surface writes to (stdout for `--list` and `--help`; stderr for status, echo, cache, and diagnostics).
- **FR-005**: The system MUST honor `NO_COLOR` (any non-empty value disables color) and the existing global color toggle, preserving current behavior under `--color=auto`. The color decision's inputs are limited to `NO_COLOR`, the `--color` flag, and TTY status; `FORCE_COLOR`, `CLICOLOR`, and `CLICOLOR_FORCE` MUST NOT be honored (use `--color=always` to force color on).

#### Color control flag

- **FR-006**: The system MUST provide a global `--color` flag accepting `auto`, `always`, and `never`, defaulting to `auto`.
- **FR-007**: `--color=auto` MUST reproduce the pre-feature automatic behavior (color only on a TTY with `NO_COLOR` unset). `--color=always` MUST force color on even on a non-TTY/redirected stream. `--color=never` MUST force color off even on a TTY.
- **FR-008**: An explicit `--color=always|never` MUST take precedence over `NO_COLOR` and TTY detection.
- **FR-009**: An invalid `--color` value MUST produce a clear error and a non-zero exit with no task execution.

#### Non-interactive invariance (the guardrail)

- **FR-010**: When color is off (non-TTY, `NO_COLOR` set, or `--color=never`), the stdout and stderr of every command MUST be byte-for-byte identical to the pre-feature release — no ANSI, no added/removed whitespace, no column shifts, no reordering. **Exception**: `rune --help`/usage output is intentionally redesigned (User Story 6); its new plain form is a deliberately reviewed new baseline, and colored vs plain help differ only by ANSI.
- **FR-011**: Styling MUST NOT change any command's exit codes, stream assignment (which content goes to stdout vs stderr), text content, or machine-readable formats.
- **FR-012**: Styling MUST NOT alter column alignment, indentation, or width of any surface; color may only add zero-width visual emphasis.

#### `--list`

- **FR-013**: On a color-enabled stdout, `rune --list` MUST emphasize task names, distinctly style group headings, and dim doc summaries, while preserving the exact text, order, indentation, and column widths of the plain listing.

#### Run output (status / echo / cache)

- **FR-014**: On color-enabled stderr, per-task status labels (`running:`, `cached:`, `would run:`, `would skip (cached):`) MUST be styled by semantic role, and echoed command lines and cache-hit notices MUST be dimmed (muted role).
- **FR-015**: Cache-write and other run warnings MUST use the shared warning role, consistent with diagnostics.
- **FR-016**: Existing echo-suppression behavior (the `@` sigil and `--quiet`) MUST be unchanged by styling.

#### Diagnostics

- **FR-017**: Diagnostics MUST continue to color the severity word and caret using the shared theme, and MAY add emphasis to the `file:line:col` locator and caret span, with zero change to column layout.
- **FR-018**: The plain (non-color) diagnostic rendering MUST remain byte-for-byte identical to the current diagnostic golden output.

#### `--help`

- **FR-019**: `rune --help` and subcommand help MUST present grouped sections (usage, commands, flags, examples) with plain-language descriptions for each command and flag.
- **FR-020**: Help MUST include at least one concrete, runnable example for each common workflow: running a task, listing tasks, launching the interactive picker, and serving over MCP.
- **FR-021**: Help headings MAY be colored on a TTY but MUST remain fully informative and ANSI-free when piped or redirected.

#### Verification

- **FR-022**: The feature MUST include tests proving (a) styled output on a forced-color terminal for each styled surface and (b) byte-identical plain output when piped, when `NO_COLOR` is set, and when `--color=never`, runnable inside the repo's Docker test harness.

### Key Entities *(include if feature involves data)*

- **Theme**: the single source of truth mapping semantic roles (error, warning, success, task-name, heading, muted/meta) to visual styling; produces plain no-op styling when color is off.
- **Color decision**: the resolved on/off determination for a given output stream, derived from the `--color` flag, `NO_COLOR`, the global color toggle, and that stream's TTY status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For every command except `--help`/usage, with output piped/redirected, with `NO_COLOR` set, or with `--color=never`, stdout and stderr are byte-for-byte identical to the pre-feature release (verified by the golden and integration suites; 100% of compared bytes match). The redesigned `--help` is verified against its new reviewed baseline instead.
- **SC-002**: On a color-capable terminal, `rune --list` visually distinguishes task names, group headings, and doc summaries (each rendered with a distinct theme role), while column alignment is unchanged.
- **SC-003**: Diagnostic caret spans remain column-aligned with the offending source in both colored and plain renderings (alignment offset difference is zero).
- **SC-004**: `--color=always` emits color through a non-TTY stream and `--color=never` emits no color on a TTY, in 100% of styled surfaces.
- **SC-005**: Exit codes for every command path are unchanged from the pre-feature release.
- **SC-006**: A first-time user can determine how to run a task, list tasks, and launch the picker solely from `rune --help`, evidenced by a worked example present for each of those workflows.
- **SC-007**: All semantic styling is defined in exactly one place and reused; no output surface re-implements color detection.

## Assumptions

- "Pre-feature release" output is the current `main` behavior as captured by the existing golden files and integration assertions; these define the byte-for-byte baseline.
- Diagnostics already colorize the severity word and caret today; this feature refines emphasis and brings them under the shared theme rather than introducing color there from scratch.
- Exact color values (palette codes) are an implementation detail to be fixed in `plan.md`; the spec fixes only the semantic roles and that styling must be restrained and consistent.
- The shared theme and color decision will reuse the already-present styling dependencies (Lip Gloss / fatih-color, go-isatty) rather than adding new ones.
- The interactive `--choose` picker is out of scope; it already has its own TTY-gated styling and is unaffected.
- Stdout remains reserved for task output and `--list`/`--help` content; stderr remains Rune's own message channel — this stream split is unchanged.
- `--color=auto` is the default so that omitting the flag preserves today's behavior exactly.
