# Feature Specification: Colorful CLI Output Everywhere

**Feature Branch**: `014-colorful-cli-output`

**Created**: 2026-07-22

**Status**: Draft

**Input**: User description: "do colorfull all cli output / reviwe cli commands and args / i gues if found a bug / not correct work"

## Overview

Rune already has a consistent color scheme for parts of its terminal output (task
listings, run-status lines, command echoes, diagnostics, the root help screen).
But the coverage is incomplete, and the result reads as broken: the single most
important message Rune ever prints — the failure banner (`rune: …`) — appears in
plain text on a color terminal, while less important messages around it are
colored. Warnings are styled in one place and plain in another. The standalone
analysis command prints the *same* diagnostics as a normal run, but flat and
uncolored. Help screens are styled for the root command only; every subcommand's
help looks like it belongs to a different tool.

This feature finishes the job: every human-facing message Rune prints uses the
one shared color scheme, warnings and errors are visually consistent wherever
they appear, and a full review of the command and flag surface confirms help and
output behave uniformly. Machine-readable outputs and piped/plain output remain
byte-for-byte unchanged.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Failures and warnings are visually consistent (Priority: P1)

A user runs a task that fails. Today the failure banner (`rune: task "build"
failed …`) prints as plain text even though the terminal supports color, while a
lesser cache warning two lines above it is colored. The user wants every error
and warning — the failure banner, dependency-cycle errors, watch-mode errors,
the minimum-version override warning — rendered with the same error/warning
colors already used elsewhere, so failures are instantly scannable in a long
transcript.

**Why this priority**: This is the reported bug ("not correct work"). Errors are
the most important output the tool produces, and they are currently the *least*
styled. Rune's constitution treats good error presentation as part of the
product.

**Independent Test**: On a color terminal, run a task that fails, a Runefile
with a dependency cycle, and a run with `--ignore-version`. Every error line and
warning line uses the shared error/warning colors. Pipe the same commands to a
file: output is byte-identical to today's plain output.

**Acceptance Scenarios**:

1. **Given** a color-capable terminal, **When** a task fails and Rune prints its
   failure banner, **Then** the banner uses the shared error color, and the
   non-zero exit code is unchanged.
2. **Given** a color-capable terminal, **When** Rune prints the
   minimum-version-override warning, **Then** the warning label uses the same
   warning color as the existing cache-write warning.
3. **Given** a color-capable terminal and watch mode, **When** a rebuild fails,
   **Then** watch-mode banners and error lines use the shared theme.
4. **Given** output redirected to a file or `--color=never` or `NO_COLOR`,
   **When** any of the above messages are printed, **Then** they contain no
   color/escape sequences and are byte-identical to current plain output.
5. **Given** a failure message containing a masked secret, **When** the banner
   is printed in color, **Then** the secret is still masked as `***`.

---

### User Story 2 - Every remaining status message joins the theme (Priority: P2)

A user interacts with the rest of the CLI surface: the bare `rune` overview, the
`[confirm]` prompt, `--fmt`'s "formatted:" notice, `--clear-cache`'s "cleared:"
notice, the server startup banners, and the standalone analysis command. Today
these are plain while structurally identical messages (`running:`, `cached:`)
are colored. The user wants all of Rune's own status/informational messages
styled consistently, and the standalone analysis command to present diagnostics
in the same rich format (severity color, location emphasis, caret span) as a
normal run does.

**Why this priority**: Delivers the literal request — "colorful all cli output"
— and removes the mixed styled/plain look that makes the tool feel unfinished.
Depends on nothing in Story 1 but is less critical than fixing error visibility.

**Independent Test**: On a color terminal, exercise each surface (bare `rune`,
confirm prompt, `--fmt`, `--clear-cache`, server startup, standalone analysis)
and verify each message uses the shared theme; the analysis command's
diagnostics match the run-time diagnostic rendering. Piped output for each is
byte-identical to today, except the analysis command whose plain layout is
allowed to change to match the run-time diagnostic layout.

**Acceptance Scenarios**:

1. **Given** a color terminal, **When** the user runs bare `rune`, **Then** the
   version header and the embedded task list share the same theme (no plain
   header above a styled list).
2. **Given** a color terminal, **When** the standalone analysis command reports
   diagnostics, **Then** each diagnostic is rendered identically to how the same
   Runefile error is rendered by a normal run (severity color, location, caret).
3. **Given** a color terminal, **When** a `[confirm]` task prompts, **Then** the
   prompt is styled and still accepts the same input behavior.
4. **Given** machine-readable modes (JSON output modes, dump mode, editor/agent
   protocols), **When** any command runs, **Then** their output contains no
   color/escape sequences.

---

### User Story 3 - Help is uniform across every command (Priority: P3)

A user runs `--help` on subcommands (server, version, analysis, completion,
editor integration). Today only the root help screen is styled and grouped;
subcommand help falls back to a different, uncolored layout. The user wants
every help screen to share the root help's structure and colors, and a review of
all commands/flags to confirm descriptions are present, accurate, and
consistently worded.

**Why this priority**: Completes the "review cli commands and args" request.
Cosmetic-plus-audit; valuable but not blocking daily use.

**Independent Test**: Run help for the root command and every subcommand on a
color terminal; all screens share heading style and section grouping. Piped help
output remains free of escape sequences.

**Acceptance Scenarios**:

1. **Given** a color terminal, **When** the user requests help for any
   subcommand, **Then** the help screen uses the same heading colors and section
   layout as root help.
2. **Given** the full command/flag inventory, **When** the audit completes,
   **Then** every command and flag has a description, and wording/casing is
   consistent across commands (audit findings recorded and fixed or explicitly
   deferred).

---

### Edge Cases

- Output is piped, redirected, or `--color=never` / `NO_COLOR` is set: every
  newly styled surface must degrade to exactly its current plain text
  (byte-identical), preserving golden-file stability — except surfaces this spec
  explicitly reformats (standalone analysis diagnostics), whose new plain form
  becomes the golden output.
- `--color=always` with piped output: newly styled surfaces emit color, matching
  the existing behavior of already-styled surfaces.
- Machine-readable outputs (JSON modes, dump, shell-completion scripts,
  editor/agent protocol streams, agent-captured task output) must never contain
  escape sequences regardless of color mode.
- A task name containing wide or multibyte characters in the styled task list:
  the description column must align identically in colored and plain modes
  (fixes a latent width-calculation mismatch between the two rendering paths).
- Secret masking and coloring combine: masked values stay masked in every newly
  colored message; color sequences must not split a secret in a way that defeats
  masking.
- Interactive picker and task list must present the same accent colors (one
  scheme, not two near-copies that can drift).
- Very narrow terminals or dumb terminals (`TERM=dumb`): existing degradation
  behavior applies unchanged to new surfaces.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Every error message Rune prints on its own behalf — including the
  top-level failure banner, dependency-cycle errors, and watch-mode errors —
  MUST use the shared error styling when color is enabled for that stream.
- **FR-002**: Every warning Rune prints MUST use the shared warning styling; no
  warning may be plain while another is styled.
- **FR-003**: All remaining informational/status messages Rune prints for
  humans (bare-command overview header, confirm prompt, format notice,
  cache-clear notice, server startup banners, watch-mode banners) MUST use the
  shared theme when color is enabled.
- **FR-004**: The standalone analysis command MUST render diagnostics in the
  same format and styling as diagnostics rendered during a normal run,
  including severity coloring, location emphasis, and caret spans.
- **FR-005**: Help screens for the root command and every subcommand MUST share
  the same section grouping and heading styling.
- **FR-006**: When color is disabled (non-terminal stream, `--color=never`, or
  `NO_COLOR`), all output MUST be free of escape sequences and byte-identical
  to the pre-feature plain output, except surfaces explicitly reformatted by
  FR-004 and FR-005, whose plain layout may change once and is then locked.
- **FR-007**: Machine-readable outputs (JSON modes, dump output, completion
  scripts, editor/agent protocol streams, and task output captured for agents)
  MUST never contain color escape sequences under any color mode.
- **FR-008**: The existing color controls (`--color auto|always|never`,
  `NO_COLOR`, per-stream terminal detection) MUST govern every newly styled
  surface with no additional switches introduced.
- **FR-009**: Colored and plain renderings of the task list MUST produce
  identical column alignment for any task name, including names containing
  multibyte or wide characters.
- **FR-010**: The interactive picker and the task list MUST draw their accent
  and muted colors from a single shared definition so the palettes cannot
  drift apart.
- **FR-011**: Secret masking MUST apply to every newly styled message exactly
  as it does to unstyled output; styling must never expose a masked value.
- **FR-012**: A documented audit of all commands, flags, and arguments MUST be
  produced, confirming each has a description and consistent wording; every
  inconsistency found is either fixed in this feature or explicitly listed as
  deferred.
- **FR-013**: Exit codes, message ordering, and stream assignment
  (stdout/stderr) of every touched message MUST remain unchanged.

### Key Entities

- **Semantic theme**: the single shared set of named color roles (error,
  warning, success, task name, heading, muted, location, caret) that all
  human-facing output draws from.
- **Output surface**: any distinct place Rune emits text (failure banner, task
  list, help screen, analysis report, watch banner, prompt, server banner,
  machine formats); each surface is classified as human-facing (themed) or
  machine-facing (never themed).
- **Color decision**: the per-stream on/off outcome derived from flag, `NO_COLOR`,
  and terminal detection, applied uniformly across all surfaces.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of Rune's human-facing message surfaces (per the audit
  inventory) render with the shared theme on a color terminal — verified by an
  enumerated checklist of every surface.
- **SC-002**: With color disabled, output across the full test corpus is
  byte-identical to the pre-feature output, except the explicitly reformatted
  analysis and subcommand-help surfaces — zero unintended golden-file changes.
- **SC-003**: Zero escape sequences appear in any machine-readable output mode
  across the full test corpus, under all three color modes.
- **SC-004**: A user scanning a failed run's transcript can locate the failure
  banner by color alone; every error/warning line in the corpus uses the
  error/warning roles (no plain-text stragglers — count is zero in the audit).
- **SC-005**: Help screens for 100% of commands share the same layout and
  heading treatment.
- **SC-006**: The command/flag audit document exists, covers every command and
  flag, and each finding is marked fixed or deferred.

## Assumptions

- The user's reported bug ("colors not correct work") is the inconsistent
  coverage documented above — most visibly the plain failure banner next to
  colored status lines — rather than a single crash. The audit (FR-012) will
  surface anything else.
- The existing color scheme, palette, and control flags from the earlier styled
  output feature are kept as-is; this feature extends coverage, it does not
  redesign the palette.
- "All CLI output" means all *human-facing* output. Machine formats (JSON,
  dump, completion scripts, editor/agent protocols) are deliberately excluded
  and must stay plain, per the project's AI-native and portability principles.
- Changing the plain-text layout is acceptable only where this spec calls for
  format unification (standalone analysis diagnostics, subcommand help); all
  other plain output is frozen byte-for-byte to protect the golden-file test
  corpus.
- No new flags or configuration are wanted; the existing `--color` /
  `NO_COLOR` controls are sufficient.
- The interactive picker keeps its own widget-level look; only its palette
  source is unified with the shared theme.
