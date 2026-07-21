# Implementation Plan: Grouped Sections in the Interactive Task Picker

**Branch**: `012-interactive-task-picker` | **Date**: 2026-07-20 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/012-picker-task-grouping/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

`rune --choose` currently renders every visible task as one flat list, while
`rune --list` already organizes the same tasks into labeled sections via each
task's `group(...)` attribute. This plan adds that same sectioning to the
picker: a shared grouping step (reused from `--list`'s existing logic)
partitions tasks into ordered sections, and `internal/tui`'s Bubbles-list-based
picker renders each section as a non-selectable header row followed by its
tasks — skipped transparently during keyboard navigation and hidden entirely
when a filter leaves it with zero matches. Runefiles with no groups render
exactly as today (no headers, no behavior change). No new dependency, flag, or
subcommand is introduced.

## Technical Context

**Language/Version**: Go 1.25 (per `go.mod`)

**Primary Dependencies**: `github.com/charmbracelet/bubbletea` v1.3.10,
`github.com/charmbracelet/bubbles` v1.0.0 (`list` package), and
`github.com/charmbracelet/lipgloss` v1.1.0 — all already vendored for the
existing `--choose` picker (`internal/tui`); no new dependency is added.

**Storage**: N/A — in-memory projection of the already-parsed Runefile AST for
the lifetime of one picker invocation.

**Testing**: `go test ./...` (race-enabled) inside the project's Docker
harness (`docker-compose run --rm test go test ./...`), per this repo's
Docker-only testing policy. `internal/tui`'s existing pattern — plain Go
assertions driving `Model.Update` directly, no golden files — is extended as-
is; `internal/cli`'s grouping helper gets a new test file (`run_test.go`
does not exist yet) rather than extending an existing one (see `research.md`
R5).

**Target Platform**: Cross-platform CLI — Linux, macOS, Windows (CI matrix
already covers all three); no platform-specific behavior is introduced.

**Project Type**: Single Go CLI project (existing `internal/` package
layout) — this feature touches `internal/cli` (grouping/projection) and
`internal/tui` (rendering/navigation) only.

**Performance Goals**: Filtering and section recomputation must stay
imperceptible (SC-004) at the scale this feature targets — no async work, no
new I/O; grouping is a single linear pass over already-in-memory tasks.

**Constraints**: Zero behavioral or visual change when no task has a
`group(...)` attribute (FR-004/SC-002); zero new CLI surface (no new flag);
section order/membership must stay derived from, and provably consistent
with, `--list`'s existing grouping rule (FR-002/FR-003) rather than a
second, independently-maintained implementation of it.

**Scale/Scope**: Runefiles with dozens of groups and hundreds of tasks — the
same order of magnitude `--list` already renders correctly today; no new
scale ceiling is introduced.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|-----------|-------|--------|
| I. Command Runner, Not a Build System | No caching/skipping semantics touched; feature is purely presentational. | PASS |
| II. Errors Are a Feature | No new error paths; existing `--choose` activation-matrix errors (non-TTY, no tasks) are unchanged. | PASS |
| III. Minimal, Total DSL | No DSL/grammar change — reuses the existing `group(...)` attribute as-is; no new attribute, syntax, or semantics. | PASS |
| IV. Hand-Written Front End, Idiomatic Go | No parser/lexer change. New code lives in the existing `internal/cli` and `internal/tui` packages — no new top-level package. | PASS |
| V. Boringly Portable | No new dependency; pure Go; no OS-specific code. | PASS |
| VI. Test-First, Multi-Layer Verification | Plan requires unit tests for the grouping helper, Bubble Tea model tests for navigation-skip, and a golden/byte-identical test for the no-groups case (`research.md` R5) before implementation is considered done. | PASS |
| VII. AI-Native, Secure by Default | No MCP/agent-surface change; `--choose` is already excluded from the agent-facing tool surface. | PASS |
| VIII. Go Engineering Discipline | Extends existing idiomatic patterns (`list.Item`, `ItemDelegate`); no `init()`/globals; no new external calls needing context/timeout (nothing async is added). | PASS |

No violations — Complexity Tracking table is not needed.

## Project Structure

### Documentation (this feature)

```text
specs/012-picker-task-grouping/
├── plan.md                          # This file (/speckit-plan command output)
├── research.md                      # Phase 0 output
├── data-model.md                    # Phase 1 output
├── quickstart.md                    # Phase 1 output
├── contracts/
│   └── tui-picker-grouping.md       # Phase 1 output — extends specs/007-interactive-tui's contract
└── tasks.md                         # Phase 2 output (/speckit-tasks command — NOT created by /speckit-plan)
```

### Source Code (repository root)

Existing single-project Go layout (Constitution Principle IV — locked package
layout); this feature only touches two already-existing packages, adding no
new ones:

```text
internal/
├── cli/
│   ├── run.go          # existing --list grouping logic, extracted into the
│   │                   # shared visibleTasksByGroup helper (research.md R1)
│   ├── run_test.go     # NEW FILE — unit tests for visibleTasksByGroup
│   ├── choose.go       # pickerItems(...) returns []tui.PickerSection
│   │                   # (built from the shared helper) instead of a flat
│   │                   # []tui.PickerItem
│   └── choose_test.go  # extended: grouping-equivalence, no-groups shape,
│                       # and non-first-section execution tests
└── tui/
    ├── item.go          # PickerItem gains a Section field; new PickerSection
    │                    # type (data-model.md)
    ├── delegate.go       # NEW FILE — sectionDelegate wrapping
    │                    # list.DefaultDelegate: draws the header line as
    │                    # decoration (adjacency in VisibleItems()), never as
    │                    # a selectable list.Item (research.md R2-R4, revised
    │                    # from the original fake-header-item design after a
    │                    # real bug was found in that approach)
    ├── delegate_test.go  # NEW FILE — header-placement and byte-identical-
    │                    # when-ungrouped assertions
    ├── picker.go         # New(...) accepts []PickerSection, flattens and
    │                    # tags each item, constructs the list with
    │                    # newSectionDelegate(...) — no Update changes needed
    ├── styles.go         # add a Header style, reusing --list's existing
    │                    # heading treatment rather than inventing a new one
    └── picker_test.go    # existing plain-Go model tests (no golden files);
                          # migrated to the section-based New(...) signature,
                          # extended with a filter-narrows-to-one-section test
                          # and a nav-across-every-key regression guard
```

**Structure Decision**: No new package is introduced. Grouping/ordering logic
is added to `internal/cli` (where `--list`'s equivalent logic already lives,
per Constitution Principle IV's locked layout) and consumed by
`internal/tui` for rendering — the same split of responsibility the existing
`--choose` implementation already uses (`internal/cli/choose.go` builds
`tui.PickerItem`s; `internal/tui` renders and drives them).

## Complexity Tracking

No Constitution Check violations — this table is intentionally left empty.
