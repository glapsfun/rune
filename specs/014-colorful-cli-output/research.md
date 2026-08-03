# Research: Colorful CLI Output Everywhere

All unknowns from Technical Context resolved. Decisions numbered D1–D10 and
referenced from plan.md, data-model.md, and the contracts.

## D1 — Failure banner styling (main.go) with tolerant color resolution

**Decision**: Style the top-level banner `rune: <message>` in `cmd/rune/main.go`
by rendering the `rune:` prefix with the theme's `Error` role and leaving the
message text unstyled. Color is resolved *at banner time* with the same
tolerant pattern `applyHelp` already uses: read the `--color` flag off the root
command if it parsed (fall back to `auto` otherwise), then
`resolveColor(mode, streamIsTTY(os.Stderr))`.

**Rationale**: `PersistentPreRunE` (where `opts.ColorStdout/Stderr` are
normally resolved) does not run for flag-parse failures, and the banner prints
after `ExecuteContext` returns — so the banner cannot rely on `opts` being
populated. help.go:26-34 already solved this exact problem; reusing the pattern
keeps one idiom. Coloring only the `rune:` prefix (bold red) makes failures
scannable without turning multi-line wrapped messages into a red wall, and
matches how diagnostics color the severity word, not the whole message.
Masking is unaffected: `maskErr` already rewrites the error *message* before
the banner prints, and styling wraps the already-masked string.

**Alternatives considered**: (a) Whole-line red — rejected: long messages
become unreadable and it diverges from the diagnostic renderer's
severity-word-only convention. (b) Plumb a resolved theme out of `Options` —
rejected: unavailable on early flag errors, and threading state out of cobra's
error path is more code than re-resolving two booleans.

## D2 — Cycle error and watch banners reuse the same banner helper

**Decision**: Extract a tiny helper in `internal/cli` (e.g.
`printErrorBanner(opts, msg)`) that renders `rune:` via
`opts.themeStderr().Error` — used by the cycle-error site (run.go:533) and
watch's two error sites (watch.go:27,55). Watch's informational lines
("watching …", "change detected, re-running…") use the `Muted` role, matching
dry-run/cache notices.

**Rationale**: These sites sit inside `internal/cli` where `Options` *is*
populated, so they use the standard `themeStderr()` path; only the outermost
main.go banner needs D1's tolerant resolution. One helper keeps the banner
format literally identical everywhere (`rune: ` prefix, one space, message).

**Alternatives considered**: Routing all banners through main.go by returning
richer errors — rejected: watch must print per-iteration failures and keep
running; it cannot defer to the process-exit banner.

## D3 — Min-version override warning joins the Warning role

**Decision**: version_gate.go:35 becomes
`fmt.Fprintf(opts.Stderr, "%s: ignoring …", theme.Warning.Render("warning"))`,
mirroring the cache-write warning (run.go:280). Plain bytes unchanged
(`warning: ignoring Runefile minimum Rune version …`).

**Rationale**: FR-002 (no warning may be plain while another is styled); the
plain form already matches the cache warning's `warning: ` shape, so only the
label gains color. Zero golden churn (color-off output identical).

**Alternatives considered**: None viable — this is a one-line consistency fix.

## D4 — `rune analyze` text output adopts the run-path diagnostic renderer

**Decision**: `printAnalyzeText` (internal/cli/analyze.go:49-61) switches from
hand-formatted `loc: severity: message` lines to `diag.RenderAll(diags, src,
theme)` — the same renderer used by runs, `--fmt`, and the version gate — with
a source provider over the analyzed files, followed by the existing summary
line with severity words styled (`Error`/`Warning` roles) when their counts are
non-zero. Output stays on **stdout** (the analysis product), themed by
`themeStdout()`. The plain layout therefore changes once (gains caret spans and
the renderer's locator format) and is then frozen; `analyze --json` is
untouched.

**Rationale**: Spec FR-004 requires format parity with run diagnostics;
Constitution II says diagnostics with caret spans are the product. The shared
analysis service already guarantees the *set* of diagnostics matches the LSP;
this makes the *presentation* match the CLI's own run path too. The spec's
FR-006 exception explicitly authorizes this one plain-layout change; goldens
are regenerated deliberately in their own commit (Constitution VI).

**Alternatives considered**: (a) Colorize the existing one-line format —
rejected: creates a *third* diagnostic look, violating FR-004's parity goal.
(b) Move analyze output to stderr for consistency with run diagnostics —
rejected: FR-013 freezes stream assignment; scripts consume analyze on stdout.

## D5 — Subcommand help: one grouped renderer for every command

**Decision**: Generalize `rootHelp` into a grouped, theme-aware help renderer
driven by cobra metadata (`Short`/`Long`, `Use`, `Example`, flags, and — for
the root — the Tasks/Commands sections), and install it via `SetHelpFunc` for
the root *and all subcommands* (drop the `cmd.HasParent()` early-out in
help.go:20-23). Section headings (`Usage:`, `Examples:`, `Flags:`) use the
`Heading` role; bodies stay plain. Color resolution keeps help.go's tolerant
flag-read pattern. Subcommand plain help changes layout once (Cobra default →
grouped format) and is then frozen as the reviewed baseline, exactly as
rootHelp's comment records for the root.

**Rationale**: FR-005 and the 008 spec's own FR-019 wording ("root **and
subcommand** help") — 008 shipped root-only as an explicit task-level cut
(tasks.md T021); this feature completes it. Deriving sections from cobra
metadata (rather than hand-writing per-command text) means new
commands/flags inherit the layout automatically, which also serves the FR-012
audit (missing `Short`/`Example` metadata becomes visible).

**Alternatives considered**: (a) Cobra help *template* override with ANSI in
the template — rejected: templates can't consult per-stream color decisions
cleanly and escape codes in templates leak into piped output. (b) Style only
headings of Cobra's default layout via `cmd.SetUsageTemplate` — rejected:
keeps two different layouts, failing "same section grouping" (FR-005).

## D6 — Bare-`rune` overview joins the theme

**Decision**: `printOverview` (run.go:626-639): version header rendered with
`Heading` role for the `rune version:` label (value plain); the no-tasks
fallback keeps its text with the docs URL rendered `Muted`. Plain bytes
unchanged.

**Rationale**: FR-003; removes the plain-header-above-styled-list mismatch
with minimal churn. Styling labels (not values) keeps copy-paste of the
version string clean.

**Alternatives considered**: Restyling the whole overview screen — rejected as
scope creep; the embedded `listTasks` is already themed.

## D7 — Remaining status notices: role assignments

**Decision**: All use `themeStderr()`; plain bytes unchanged.

| Message | Site | Role |
|---|---|---|
| `formatted: <file>` | fmt.go:28 | label `Success`, path plain |
| `cleared: <path>` | run.go:48 | label `Success`, path plain |
| `[confirm]` prompt `… [y/N] ` | run.go:358 | prompt text `Warning` (destructive action), `[y/N]` plain |
| `rune MCP server on …` | serve.go:125,128 | whole line `Muted` (server chrome) |
| `watching … (Ctrl-C to stop)` / `change detected, re-running…` | watch.go:31,49 | `Muted` |

**Rationale**: FR-003. Labels follow the established grammar: active outcomes
green (`running`, now `formatted`/`cleared`), meta chrome dim, destructive
prompts yellow. Values (paths, addresses) stay plain for copy-paste.

**Alternatives considered**: A new "Notice" role — rejected: the existing
eight roles cover every case; growing the theme without need violates the
restrained-palette intent.

## D8 — `--list` padding: display width in the colored branch only

> **Superseded during implementation (2026-07-23)**: the lexer restricts task
> names to ASCII, so the wide-rune case this decision guarded against is
> unreachable and the runewidth dependency was not added. The colored/plain
> parity is pinned by `TestListTasksColoredAlignmentMatchesPlain` instead.
> Original decision kept below for the record.

**Decision**: Replace `utf8.RuneCountInString` with
`runewidth.StringWidth` (promoting `mattn/go-runewidth` — already in the
module graph via lipgloss — to a direct dependency) for the colored branch's
pad computation (run.go:712), and compute the shared `width` maximum with the
same function. The plain branch keeps `%-*s` (byte width) so plain output
stays byte-identical for the entire existing corpus (ASCII names: byte = rune
= display width). Add a unit test with a wide-rune task name asserting the
colored branch's visible alignment, and a documented note that plain-mode
alignment for non-ASCII names retains today's `fmt`-standard behavior.

**Rationale**: FR-009 requires colored output to align correctly for wide
names; rune *count* under-pads double-width CJK names. Changing the plain
branch to display width would alter plain bytes for non-ASCII names,
violating FR-006's byte-freeze for surfaces not on the exception list — so
plain keeps fmt semantics, and the invariance contract is defined on the
corpus (ASCII), where both branches agree exactly.

**Alternatives considered**: (a) Display-width padding in both branches —
rejected: breaks FR-006's byte-freeze guarantee for a hypothetical case, and
plain `%-*s` misalignment for exotic names is pre-existing `fmt` behavior.
(b) `lipgloss.Width` — same result but pulls ANSI-stripping machinery into a
path with no ANSI in its input; runewidth is the narrower tool.

## D9 — One palette: export constants from `internal/style`, TUI consumes them

**Decision**: Export the palette as named constants
(`style.ColorError/Warning/Success/Accent/Muted`) plus the picker's dark-grey
help color (`style.ColorMutedDark` = 241, currently a TUI-only literal), and
rewrite `internal/tui/styles.go` to build its four styles from those
constants. The `Theme` struct and role set are unchanged; the TUI keeps its
own widget-level styles (padding, layout) — only the color *literals* are
unified (spec FR-010).

**Rationale**: SC-007 of feature 008 promised palette changes in exactly one
place; today 170/245/241 are duplicated in tui/styles.go and can drift. The
TUI's bubbletea renderer differs from the CLI's per-stream renderer, so
sharing constants (not `lipgloss.Style` values) is the correct seam.

**Alternatives considered**: TUI consuming a full `style.Theme` — rejected:
TUI styles carry layout (padding) and are bound to bubbletea's renderer;
forcing them through the CLI theme conflates two rendering contexts.

## D10 — Scope guards: what stays plain, and test strategy

**Decision**: Machine surfaces are asserted plain by regression tests rather
than relying on wiring alone: `--dump` (both formats), `analyze --json`,
`version --json`, completion scripts, LSP stream, and MCP tool-result buffers
(the MCP adapter already constructs `Options` with color off). `--summary`
and `version` human text remain plain **by design** (stdout products meant
for piping; already consistent with `--summary`'s exclusion in 008). Test
matrix per touched surface: (1) `--color=never` + pipe → byte-golden
(pre-feature bytes, except analyze/help), (2) `--color=always` through pipe →
expected ANSI sequences present, (3) `NO_COLOR` → plain, (4) masked-secret
failure banner under `--color=always` → `***` present, secret absent, ANSI
present. Integration tests drive the real binary (Constitution VI); the
full-corpus invariance suite from 008 (T011/T012/T023 pattern) is extended to
the new surfaces.

**Rationale**: FR-006/FR-007/FR-011/SC-002/SC-003 are byte-level claims; only
byte-level tests can hold them. Declaring `--summary`/`version` out of scope
keeps stdout products script-safe and matches 008 precedent.

**Alternatives considered**: Styling `version`'s human output — rejected:
`rune version` is routinely consumed by scripts/CI and the value *is* the
product; FR-013 freezes it at zero risk.
