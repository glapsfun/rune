# Feature Specification: Grouped Sections in the Interactive Task Picker

**Feature Branch**: `012-interactive-task-picker`

**Created**: 2026-07-20

**Status**: Draft

**Input**: User description: "Interactive Discovery: Add a flag like --choose to open a fuzzy-finder list of available tasks, helping humans navigate large task sets more efficiently" — narrowed, after discovering `--choose` already exists (shipped in `007-interactive-tui`), to the specific gap: the picker shows every task as one flat list, while `rune --list` already organizes tasks into named sections via the Runefile's `group(...)` attribute. This feature brings that same sectioning into the picker so it stays navigable as the task set grows.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Scan a large task set by category (Priority: P1)

A developer working in a repository whose Runefile organizes tasks into named
groups (e.g. "build", "test", "release") opens the interactive picker. Instead
of one long, undifferentiated list, they see the same tasks organized into
labeled sections that mirror the groups they already know from `rune --list`,
so they can jump straight to the category they care about instead of reading
every task name in the file.

**Why this priority**: This is the entire point of the feature — without
visible sections, grouping delivers no value. It is also independently
shippable: rendering sections and letting the user scroll/select within them
is a complete, demonstrable improvement on its own.

**Independent Test**: Open `--choose` against a Runefile with tasks split
across three or more groups plus a few ungrouped tasks. Verify section
headers appear, each task is listed under the correct header, and the header
order matches `rune --list`'s order for the same Runefile.

**Acceptance Scenarios**:

1. **Given** a Runefile whose tasks are tagged into groups "build", "test",
   and "release", with a few tasks left ungrouped, **When** the user opens
   `--choose`, **Then** the picker shows a labeled section per group — in the
   same order those groups first appear in the Runefile — plus a section for
   the ungrouped tasks, matching how `rune --list` presents the same file.
2. **Given** a Runefile where no task has a group assigned, **When** the user
   opens `--choose`, **Then** the picker renders exactly as it does today: a
   single flat list with no section headers.

---

### User Story 2 - Navigate and filter without breaking on section boundaries (Priority: P2)

A developer moves through the grouped picker with the arrow keys and also
types a search query to narrow the list. Section headers behave as visual
labels only — the cursor never lands on one — and typing a filter still
narrows down to matching tasks, now with any section that has no remaining
matches simply disappearing instead of showing an empty, header-only section.

**Why this priority**: Sectioning is only useful if it doesn't get in the way
of the two things a picker is for — moving through the list and searching it.
This story makes sure the P1 feature is actually usable, not just visible.

**Independent Test**: In a grouped picker, press the down arrow repeatedly
from the first to the last task and confirm the highlight only ever rests on
a task row. Then type a query that matches tasks in only one of several
groups and confirm the other groups' headers disappear along with their
tasks.

**Acceptance Scenarios**:

1. **Given** the picker is showing multiple section headers and task rows,
   **When** the user presses the down (or up) arrow repeatedly, **Then** the
   selection highlight moves only between task rows, transparently skipping
   over header rows, and it never selects a header.
2. **Given** the picker is showing multiple sections, **When** the user types
   a filter query that matches tasks in only one section, **Then** only that
   section's header and its matching tasks remain visible, and sections left
   with zero matches (headers included) are hidden.
3. **Given** the user selects a task that belongs to any section and confirms
   it, **When** the picker hands off, **Then** the selected task runs exactly
   as it does today, unaffected by which section it was displayed under.

---

### Edge Cases

- What happens when the same group name is used by tasks that are not
  contiguous in the Runefile (e.g. group "build" tasks interleaved with group
  "test" tasks)? All tasks sharing a group name are collected into that one
  section, positioned at the group's first occurrence — matching `rune --list`.
- What happens when a filter query leaves exactly one match in a group? The
  section header still renders above that single remaining task; only
  sections with zero remaining matches are hidden.
- What happens when there are more sections and tasks than fit on screen? The
  picker's existing scrolling behavior applies; no new pagination or
  section-jump mechanism is introduced by this feature.
- What happens when a group name is unusually long? It renders using the same
  presentation the section headers already use elsewhere; no new truncation
  rule is introduced beyond what the picker's terminal-width handling already
  does.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When the Runefile assigns at least one visible task to a named
  group, the picker MUST display tasks organized into labeled sections, one
  per group name, instead of a single flat list.
- **FR-002**: Section order in the picker MUST match `rune --list`'s
  ordering rule: a section appears in the position where its group name
  first occurs among the Runefile's visible tasks.
- **FR-003**: Tasks with no group assigned MUST be presented together in
  their own section, consistent with how `rune --list` presents ungrouped
  tasks.
- **FR-004**: When no visible task has a group assigned, the picker MUST
  render as a single flat list with no section headers, identical to its
  current behavior — grouping must introduce zero visual or behavioral change
  for Runefiles that don't use groups.
- **FR-005**: Section headers MUST be presentational only: keyboard
  navigation (moving the selection up or down) MUST move between task rows
  only, skipping over header rows without ever selecting one.
- **FR-006**: Filtering/searching MUST continue to match against task name
  and description exactly as it does today; a section whose tasks are all
  filtered out MUST have its header hidden along with them.
- **FR-007**: Which tasks are eligible to appear in the picker (private-task
  exclusion, current-OS matching) MUST be unchanged by this feature —
  grouping only reorganizes presentation, it does not add, remove, or hide
  any task that would otherwise be shown.
- **FR-008**: Selecting and confirming a task MUST run that task exactly as
  selection does today, regardless of which section it was displayed under.

### Key Entities

- **Section**: A labeled visual grouping in the picker, derived from a task's
  existing `group(...)` attribute. Holds an ordered set of tasks and renders
  as a header followed by its member tasks; carries no data of its own beyond
  its name and membership.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a Runefile with five or more task groups, a user can
  identify which section any given task belongs to by reading only the
  section headers — without having to scan unrelated tasks' names first.
- **SC-002**: For a Runefile where no task belongs to a group, the picker's
  appearance and behavior are unchanged from before this feature — a direct
  before/after comparison shows zero difference.
- **SC-003**: Every task that appears in `rune --list` for a given Runefile
  also appears, exactly once, somewhere in the picker after grouping is
  introduced — the total task count never changes due to sectioning.
- **SC-004**: Searching within a grouped picker narrows results with no
  perceptible delay, and no empty, header-only section is ever left on
  screen.

## Assumptions

- The meaning and syntax of the `group(...)` attribute are unchanged; this
  feature only changes how the picker *presents* groups that `rune --list`
  already understands.
- Filtering continues to match against task name and description only, as it
  does today; matching a query against a group/section name is a possible
  future enhancement but is not part of this feature.
- No new CLI flag is introduced — sectioning is automatic whenever `--choose`
  is used and the Runefile defines at least one group, and is a no-op
  (identical to today) otherwise.
- Section headers reuse the picker's existing color theme and styling
  conventions (matching the heading style `rune --list` already uses) rather
  than introducing a new visual style.
- The picker's existing scrolling behavior is sufficient for large,
  many-section task sets; this feature does not add pagination, collapsible
  sections, or a "jump to next section" keybinding.
