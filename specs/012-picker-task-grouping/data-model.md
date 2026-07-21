# Data Model: Grouped Sections in the Interactive Task Picker

This feature adds no persisted or long-lived data — it reorganizes an existing
in-memory projection (`PickerItem`, built from `ast.Task`) for one rendering
pass. The entities below are in-process view types, scoped to the picker's
lifetime.

**Revised during implementation**: the original design (below, superseded)
modeled a section header as its own `list.Item` (`sectionHeaderItem`),
inserted into the list ahead of each section's tasks. Reading `bubbles/list`'s
source while implementing this surfaced a real bug in that approach —
`list.Model.SetItems` re-triggers the library's own async fuzzy-filter
whenever a filter is active, which would immediately strip a freshly
reinserted header back out (see `research.md` R4 for the full trace). The
design was revised (with sign-off) to the shapes below: `PickerItem` gained a
`Section` field, and there is no separate header item type at all — a
`sectionDelegate` (`internal/tui/delegate.go`) draws the header as a
decorative line, computed at render time from adjacency in the list's own
`VisibleItems()`.

## PickerSection

Groups a section's tasks under a shared label; the shape `pickerItems`
produces instead of a flat `[]tui.PickerItem`, and what `tui.New` accepts.

| Field | Type              | Notes |
|-------|-------------------|-------|
| Name  | `string`          | Group name from `group(...)`; `""` for the ungrouped section. |
| Items | `[]tui.PickerItem`| Tasks in this section, in Runefile order — same per-task shape as today (`Name`, `Desc`, `Doc`), plus the new `Section` field below (set by `New`, not by the caller). |

**Ordering rule** (FR-002): sections are produced in the order their `Name`
first occurs among visible tasks — identical to `internal/cli/run.go`'s
`visibleTasksByGroup` helper (the single source of truth `--list` and the
picker both call). The `""` (ungrouped) section follows the same
first-occurrence rule as every other group name; it is not special-cased to
always sort first or last.

**Validation / invariants**:
- No section is empty at construction time — a `Name` only appears if at
  least one visible task carries it.
- Every visible task (per the existing `IsPrivate()` / `osMatches` filter)
  appears in exactly one section — grouping partitions, it never drops or
  duplicates a task (FR-007, SC-003).

## PickerItem.Section (new field on the existing type)

`tui.PickerItem` gains one field: `Section string` — the name of the
`PickerSection` this item came from (`""` for the ungrouped section). It is
set once, by `tui.New`, when it flattens `[]PickerSection` into the list's
underlying `[]list.Item`; callers building a `PickerItem` (e.g.
`internal/cli/choose.go`'s `pickerItems`) never set it themselves.

## sectionDelegate (internal to `internal/tui`, `delegate.go`)

Wraps `list.DefaultDelegate`. Not a `list.Item` — a `list.ItemDelegate`. Adds
no new selectable row; it only changes how existing `PickerItem` rows are
drawn.

| Field     | Type                    | Notes |
|-----------|-------------------------|-------|
| (embeds)  | `list.DefaultDelegate`  | The unmodified default rendering for the item's own title/description row. |
| styles    | `Styles`                | Supplies the `Header` style. |
| grouped   | `bool`                  | Whether the Runefile defines ≥1 group; when `false`, `Height()`/`Render()` behave identically to a plain `list.DefaultDelegate` (FR-004/SC-002). |

**Behavioral invariants**:
- `Height()` returns the embedded delegate's height unchanged when
  `grouped == false`; plus exactly one extra line when `grouped == true` (a
  single delegate-wide constant, since `list.ItemDelegate.Height()` cannot
  vary per item).
- `Render()` writes that extra line — the section's name, styled with
  `Header`, when the item is the first *visible* item of its section (by
  comparing against `m.VisibleItems()[index-1]`); a blank line otherwise,
  including for every item of the `""` (ungrouped) section, which never gets
  a header of its own (matching `--list`'s treatment) — then delegates the
  item's own row to the embedded `DefaultDelegate` untouched.
- Never produces a selectable row of its own (FR-005 holds structurally, not
  by a skip-check — see `research.md` R3): the list's `Items()` contain only
  `PickerItem`s, so every navigation key can only select a real task.
- Header visibility under filtering (FR-006) needs no separate state:
  because the header is computed from `VisibleItems()` at render time, a
  section with zero surviving matches simply has no visible item left to
  draw a header above.

## Relationship to existing types

```text
ast.Task ──(pickerItems, via visibleTasksByGroup)──> []tui.PickerSection
                     │
                     ▼ (internal/tui, tui.New)
     flatten to []PickerItem, each tagged with .Section = its section's Name
                     │
                     ▼ (rendered by sectionDelegate, not stored as items)
     "group name" or blank line, drawn above the item when it starts a
     section — derived from adjacency in m.list.VisibleItems() every render
```

No changes to `ast.Task` or `ast.AttrGroup` were required. `tui.PickerItem`
gained one field (`Section`); no other existing type changed shape.
