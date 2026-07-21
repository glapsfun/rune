<!-- SPECKIT START -->
Active feature plan: `specs/012-picker-task-grouping/plan.md` (Grouped Sections
in the Interactive Task Picker — `rune --choose` now mirrors `rune --list`'s
existing `group(...)`-based sectioning so large task sets stay navigable. A
shared grouping/ordering helper (`visibleTasksByGroup`, extracted from
`internal/cli/run.go`'s `--list` logic) is the single source both surfaces
derive section order/membership from, so they can never drift. `choose.go`'s
`pickerItems` returns `[]tui.PickerSection` instead of a flat `[]tui.PickerItem`.
Headers are **not** `list.Item`s: Bubbles' `list.Model.SetItems` re-triggers its
own async fuzzy-filter on whatever it's given, which would strip a header-as-item
right back out the moment a filter is active — a real bug found while
implementing the original design. Instead, `internal/tui/delegate.go`'s
`sectionDelegate` (wrapping `list.DefaultDelegate`) draws the group name as a
decorative line above whichever *visible* item starts a new section — derived
at render time from adjacency in `m.list.VisibleItems()` — so there is no header
row for `picker.go`'s `Update` to skip, and no separate recompute step on filter
changes. Runefiles with no `group(...)` attributes render byte-identical to the
pre-feature picker (verified: `sectionDelegate`'s output matches
`list.DefaultDelegate`'s exactly when ungrouped). No new package, dependency,
flag, or subcommand — reuses the existing `internal/cli` / `internal/tui` split
and the already-vendored `bubbles`/`bubbletea`/`lipgloss`. No Constitution
violations. Read the plan, `research.md`, `data-model.md`, and
`contracts/tui-picker-grouping.md` (which extends, not replaces, the base
`--choose` contract in `specs/007-interactive-tui/contracts/tui-picker.md`) for
details.
<!-- SPECKIT END -->

## Development workflow

Rune dogfoods itself: the repo-root `Runefile` defines the dev tasks. Run `rune --list`
(or `go run ./cmd/rune --list`) to see them — `fmt`, `lint`, `test`, `test-race`, `build`,
`docker`, `docs-check`, `release-dryrun`.

Tests run **inside Docker**, never on the host (per global policy and the lack of a compose
plugin — use standalone `docker-compose`):

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

See `CONTRIBUTING.md` for the full workflow and CI gate set.
