# Research: Grouped Sections in the Interactive Task Picker

## R1 — How does `rune --list` compute group membership and ordering?

**Decision**: Reuse `internal/cli/run.go`'s `listTasks` grouping logic verbatim
as the source of truth: iterate `f.Tasks` in file order; skip private tasks and
tasks that fail `osMatches`; read each task's group via
`t.Attr(ast.AttrGroup)` (empty string if absent); record each group name the
first time it's seen, in a `order []string` slice; bucket tasks into
`groups map[string][]row`. The picker will call the *same* grouping helper
(extracted to a shared, package-visible function) rather than reimplementing
the rule, so the two surfaces can never drift.

**Rationale**: FR-002/FR-003 require the picker's section order and ungrouped
handling to match `--list` exactly. The existing code already implements this
rule correctly and is covered by existing tests (`internal/cli/run_test.go`
golden files); extracting rather than duplicating it eliminates an entire
class of future divergence bugs.

**Alternatives considered**:
- *Reimplement grouping inside `internal/tui` or `internal/cli/choose.go`* —
  rejected: duplicates logic that must stay in lockstep with `--list`, and any
  future `group(...)` semantic change would need two updates instead of one.
- *Sort groups alphabetically instead of by first occurrence* — rejected: it
  would make the picker's section order diverge from `--list`'s for the same
  Runefile, which is the exact inconsistency this feature exists to remove.

## R2 — How to render non-selectable section headers inside `bubbles/list`?

**Decision** (superseded once, see below): render each section's group name
as a *decorative line* drawn by a custom `list.ItemDelegate` — `sectionDelegate`,
wrapping `list.DefaultDelegate` — immediately above whichever item is the
first *visible* item of its section. "First visible item of a section" is
determined at render time by comparing the item at `index` against
`m.VisibleItems()[index-1]`'s `Section` field (a new field added to
`PickerItem` itself, set once by `New` when it flattens `[]PickerSection`).
No separate header `list.Item` exists. Since `list.ItemDelegate.Height()` is
a single constant for every item in a list (not settable per item), the
delegate reserves exactly one extra line for *every* item when grouping is
active (blank when the item doesn't start a new section, the group name when
it does) and reports the embedded delegate's original height, unchanged,
when grouping is inactive (FR-004/SC-002).

**Original decision (rejected after prototyping — see R3/R4 below for why)**:
introduce a second `list.Item` implementation, a header item, and build the
picker's item slice as a flat sequence of `[header, task, task, …, header,
task, …]`. This is documented here because it's a natural first idea for
"sectioned lists in bubbles/list" and the reason it doesn't work is
non-obvious from the library's public API alone — it only surfaces once you
read `SetItems`'s implementation (R4).

**Rationale**: `bubbles/list` (v1.0.0, the version already vendored) has no
built-in concept of sections or grouped rendering — it operates on a flat
`[]list.Item` rendered through a single `ItemDelegate`. A decorative line
drawn by the delegate needs no new dependency (Principle V) and, as a
byproduct, makes FR-005 (headers are never selectable) hold *structurally*:
since a header is never a `list.Item`, there is nothing for any cursor-moving
key — including Home/End/PageUp/PageDown — to ever select.

**Alternatives considered**:
- *Header as a fake, non-filterable `list.Item`* (the original decision,
  above) — rejected: see R4, where a real, demonstrated bug in that approach
  (not a hypothetical) is documented.
- *Render sections as separate stacked `list.Model` instances* — rejected:
  multiplies keyboard/filter/scroll state management (each sub-list would need
  its own cursor, and moving the "current" list on boundary crossing is far
  more error-prone than a single delegate's adjacency check).
- *Fork/wrap `bubbles/list` or switch libraries* — rejected: no new dependency
  is justified by this feature; the existing library supports this pattern via
  its already-open `ItemDelegate` extension point.

## R3 — How does this keep header rows non-selectable (FR-005), including jump keys?

**Decision**: No special-case code is needed at all. Under R2's design there
is no header `list.Item` in the list — every entry in `m.list.Items()` is a
real `PickerItem` — so every key the underlying list binds (up/down/j/k,
Home/End, PageUp/PageDown, and any future binding) can only ever move the
cursor between real tasks. This was originally planned as an explicit
post-move "skip the header, nudge one more step" check in `Model.Update`
(and an earlier draft of that check only covered up/down/j/k, missing
Home/End/PageUp/PageDown entirely — a real gap caught in `/speckit-analyze`
before it shipped). R2's redesign removes the entire problem instead of
enumerating and testing every key that could trigger it.

**Rationale**: A structural guarantee ("there is nothing to select") is
strictly more robust than a procedural one ("we remember to intercept every
key that could select it"), and is immune to bubbles/list adding a new
navigation binding in a future version — no matching change would be needed
in the picker.

**Alternatives considered**:
- *Post-move skip-check on an allow-list of keys* — rejected: this is exactly
  the design that was caught missing Home/End/PageUp/PageDown during
  analysis; broadening it to "any key" is possible but strictly more code
  than R2's approach for the same guarantee.
- *Disable up/down entirely when a header would be hit, requiring a distinct
  "jump" key* — rejected: not requested (spec's Assumptions explicitly scope
  out a jump-to-next-section key) and would make navigation feel broken
  compared to today's picker.

## R4 — How do headers behave under the existing incremental filter (FR-006)?

**Decision**: `sectionDelegate.Render` derives the header line from
`m.VisibleItems()` — the library's own, already-filtered result — at render
time, every time. There is no separate header state to keep in sync with the
filter; a section's header simply cannot appear once none of its tasks
survive filtering, because there is no longer any visible item for the
delegate to draw it above.

**Rejected approach and the bug that ruled it out**: the original plan (R2)
was to keep headers as filterable-but-unmatchable `list.Item`s and, on every
filter-state change, call `m.list.SetItems(...)` to rebuild the header+task
sequence from the currently-matching tasks. Reading `bubbles/list`'s actual
source (`list.go:385-393`) shows `SetItems` **re-triggers the library's own
async fuzzy-filter** on whatever is passed to it whenever a filter is
currently active. Since a header's `FilterValue()` must never match search
text (or it would incorrectly surface on unrelated queries), the very next
filter pass would immediately strip out the headers `SetItems` had just
reinserted — an unfixable oscillation, not a detail to patch. This is a
concrete, demonstrated defect in the original design, not a hypothetical
alternative.

**Rationale**: Once R2 removed headers from the filterable item pool
entirely, this whole class of bug disappears — there is nothing to
re-synchronize with the filter, because the header was never data the filter
could act on in the first place.

**Alternatives considered**:
- *Give headers a `FilterValue()` equal to their group name, so typing the
  group name surfaces it* — rejected independently of the bug above: spec
  Assumptions explicitly scope out filtering by group name for this feature.

## R5 — Testing approach

**Decision**: Extend the existing `internal/tui/picker_test.go` and
`internal/cli/choose_test.go`'s test coverage with: (a) unit tests over the
new shared grouping helper (equivalence with `--list`'s grouping for a
matrix of Runefiles — ordering, ungrouped bucket, non-contiguous reuse), (b)
delegate-level unit tests (`internal/tui/delegate_test.go`) asserting header
placement by adjacency and a byte-for-byte comparison against
`list.DefaultDelegate`'s own `Render` output for the ungrouped case
(SC-002 — plain assertions, not golden files: this package has no existing
golden-file infrastructure), (c) a `bubbletea` model test driving `Update`
with every cursor-moving key (including jump keys) as a regression guard,
and (d) a `list.Model.SetFilterText`-driven test asserting only the matching
section's tasks remain in `VisibleItems()`. This follows the project's
existing Test-First, layered-verification approach (Constitution Principle
VI).

**Rationale**: The riskiest behaviors are the ones a human tester wouldn't
casually notice — a header stealing a cursor position, or the "no groups"
path silently changing pixel-for-pixel output — so those get direct,
deterministic tests rather than only manual verification.

**Alternatives considered**:
- *Only manual/exploratory testing* — rejected: violates Constitution
  Principle VI and this project's Docker-run `go test ./...` gate.
