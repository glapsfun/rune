# Contract: Grouped Sections in the Interactive Task Picker

This contract **extends** `specs/007-interactive-tui/contracts/tui-picker.md`
(the existing `--choose` contract, unchanged and still in force) with the
sectioning behavior added by this feature. It does not introduce a new flag,
subcommand, or invocation — `rune --choose` remains the only entry point.

## Section derivation

| Runefile state | Picker rendering |
|-----------------|-------------------|
| No visible task carries a `group(...)` attribute | Flat list, no section headers — **byte-for-byte identical** to the pre-feature picker (SC-002). |
| ≥1 visible task carries a `group(...)` attribute | Tasks render under labeled section headers; one header per distinct group name, plus one unlabeled section for any remaining ungrouped tasks. |
| Same group name reused non-contiguously in the Runefile | All tasks sharing that name appear together in one section, positioned at the group's first occurrence — matching `rune --list`. |

Section order and membership **must** match `rune --list`'s output for the
same Runefile (FR-002/FR-003): running `rune --list` and `rune --choose` (with
no filter typed) against the same Runefile shows every task under the same
relative section grouping in both.

## In-picker behavior (additions to the existing contract)

| Action | Key(s) | Result |
|--------|--------|--------|
| Move highlight past a section boundary | any navigation key (`↑`/`↓`, `j`/`k`, Home/End, PageUp/PageDown) | Highlight lands only on task rows; header lines are decorative and are never a selectable stop (FR-005). |
| Filter while sectioned | type (or `/` then type) | Matching continues to check task name/description only (unchanged from the base contract); any section left with zero matching tasks has its header hidden along with them (FR-006). |
| Confirm a task in any section | `Enter` | Runs exactly as the base contract describes — section membership has no effect on execution (FR-008). |

All other rows of the base contract's "In-picker behavior", "Activation
matrix", "Selection → execution handoff", "Exit codes", and "Compatibility
guarantees" tables are unchanged by this feature.

## Non-goals (explicit)

- No filtering/matching against a group/section name (Assumptions).
- No collapsible sections, section-jump keybinding, or new pagination model
  (Assumptions) — the existing list's scrolling handles overflow.
- No new CLI flag — sectioning is automatic and derives entirely from
  existing `group(...)` attributes already understood by `rune --list`.
