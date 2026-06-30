# Phase 0 Research: Styled CLI Output & Friendlier Help

The Technical Context had no open `NEEDS CLARIFICATION` (scope, flag, help depth,
the invariance carve-out, and the color-decision inputs were all settled in the
spec + Clarifications). Research here pins down the *how* for the load-bearing
decisions.

## D1. Per-stream color decision (the central change)

**Decision**: Replace the single `useColor()` (stderr-only) with a `--color` mode
(`auto|always|never`) resolved into **two** booleans in `PersistentPreRunE`:
`ColorStdout` (gates `--list`, `--help`) and `ColorStderr` (gates status/echo/
cache/diagnostics). Each `auto` decision uses the TTY status of *that* stream;
`always`/`never` force the result on both; `NO_COLOR` forces off under `auto`
only (explicit `--color` wins, per Clarifications).

**Rationale**: `--list` and `--help` write to stdout (`cmd.OutOrStdout()`), but
today's gate checks `os.Stderr`. With a pipeline like `rune --list | less -R`,
stdout is not a TTY while stderr may be — only a per-stream decision is correct
(spec FR-004, "mixed stream redirection" edge case). It also makes
`rune --list | grep` reliably plain even when stderr is a terminal.

**Precedence** (highest first): explicit `--color=never` → off; `--color=always`
→ on; then `NO_COLOR` set → off; then `fatih/color.NoColor`/global → off; then
per-stream `isatty`. Invalid `--color` value → error + non-zero exit before any
work (FR-009).

**Alternatives considered**: (a) keep one global bool — rejected: wrong for mixed
redirection and for stdout-bound surfaces. (b) detect lazily at each call site —
rejected: duplicates detection, violates FR-002/SC-007.

## D2. Forcing/suppressing color correctly with Lip Gloss

**Decision**: Build a **per-stream `lipgloss.Renderer`** (`lipgloss.NewRenderer(w)`)
inside `internal/style.New(enabled, w)` and **explicitly set its color profile**:
when enabled, force at least `termenv.ANSI256`; when disabled, `termenv.Ascii`
(no escapes). Styles are created from that renderer, not the global default.

**Rationale**: Lip Gloss's *default* renderer auto-detects the profile from
`os.Stdout`'s TTY status via termenv. That means with `--color=always | pipe`,
the default renderer would silently strip color (non-TTY → Ascii) — defeating
`always` (FR-007). Conversely we must guarantee Ascii when off so no escapes leak
(FR-003/FR-010). An explicit renderer+profile is the only way to decouple "should
we color?" (our decision) from termenv's auto-detection. `internal/tui` currently
relies on the global default; this feature does **not** change the picker, but the
new `internal/style` package uses the explicit-renderer approach as the correct
pattern.

For `fatih/color` (used until diag is re-homed, and conceptually): set
`color.NoColor` to the resolved decision so `--color=always` emits through a pipe.

**Alternatives considered**: mutating the global lipgloss renderer
(`lipgloss.SetColorProfile`) — rejected: global mutation, and we need two
different per-stream answers simultaneously.

## D3. Preserving `--list` column alignment under color

**Decision**: Compute padding from the **visible** name (rune width), not the
styled string. Concretely, stop using `%-*s` on a colorized value; instead emit
`<styled-name><spaces-to-width>  # <muted-doc>` where the spaces are computed from
`width - utf8.RuneCountInString(name)`. The plain branch keeps `%-*s` exactly as
today.

**Rationale**: ANSI escapes have nonzero byte length but zero display width;
`fmt`'s `%-*s` pads by byte length, so colorizing before padding misaligns the
`#` column (violates FR-012/SC-002). Padding by visible width keeps every `#`
aligned in both modes. The header `Available tasks:`, group `  [name]`, indents,
and ordering are unchanged.

**Alternatives considered**: `lipgloss.Width`-based table layout — overkill and
risks changing the exact plain bytes; rejected for a surgical pad-by-rune-count.

## D4. Re-homing diagnostic colors without changing plain bytes

**Decision**: `internal/diag/render.go` stops calling `fatih/color` inline and
instead applies role styles from a `style.Theme` passed in (severity → Error/
Warning role; caret → Caret role; optional Locator emphasis on `file:line:col`).
The plain path (theme disabled) writes the **identical** bytes it does today.

**Rationale**: Single palette source (FR-001/SC-007). The existing golden
`testdata/diag/render.golden` was generated with `useColor=false` and MUST keep
matching (FR-018) — the plain branch is byte-for-byte preserved; only the colored
branch is re-expressed via the theme. Caret emphasis only swaps the carets'
color, never their count or position (SC-003).

**Signature impact**: `Render`/`RenderAll` take a `style.Theme` (or `*Theme`)
instead of `useColor bool`. Callers: `internal/cli/run.go:702` and
`internal/diag/render_test.go`. A disabled theme is the test's "plain" case.

**Alternatives considered**: keep diag on `fatih/color`, duplicate the palette
constants in two packages — rejected (two sources of truth; drift risk).

## D5. `--help` redesign approach

**Decision**: Provide custom Cobra templates via `SetUsageTemplate` /
`SetHelpTemplate` (and per-subcommand where useful) with grouped sections —
**Usage**, **Common commands**, **Flags**, **Examples** — plain-language flag
descriptions, and a worked example per workflow: run a task, `--list`, `--choose`,
`serve` (MCP). Section **headings** are colorized when `ColorStdout` is on; body
text is plain. The redesigned plain help is captured as a **new** golden baseline.

**Rationale**: US6 + the Clarification carve-out. Templates keep the redesign in
one declarative place and let Cobra still inject the auto-generated flag list.
Help prints to stdout, so it gates on `ColorStdout` (D1). Examples already exist
in `root.Example`; this expands and regroups them.

**Alternatives considered**: hand-rolled help printer bypassing Cobra — rejected:
loses Cobra's flag/command introspection and completion integration.

## D6. Dimming command echo and status labels

**Decision**: Thread the stderr `style.Theme` into the engine and executors.
Status labels (`running:`/`cached:`/`would run:`/`would skip (cached):`) are
wrapped per semantic role; the cache-write warning uses the Warning role
(consistent with diagnostics, FR-015); command echo in `shell.go`/interp is
wrapped in the Muted role via a new optional field on `shell.Options`
(e.g. `EchoStyle`). Suppression (`@`/`--quiet`) is checked **before** any styling
so a suppressed line never appears (FR-016).

**Rationale**: These lines are plain `fmt.Fprintf(..Stderr..)` today; wrapping the
label/echo text (not the whole line layout) adds emphasis without changing text or
stream. Role choice: cache-hit + would-run + echo → Muted; running → neutral/
success-ish active; warnings → Warning; failures stay via the existing error path.

**Alternatives considered**: a verbose "spinner" for running tasks — rejected for
v1: spinners require a live TTY render loop and would complicate the
byte-invariance guarantee on stderr; a static colored label is enough and safe.

## D7. Palette (semantic role → color)

**Decision**: Define once in `internal/style`. Restrained 256-color set, reusing
the picker's existing hues for visual coherence:

| Role | Style | Note |
|------|-------|------|
| Error | red, bold | matches today's diagnostic error |
| Warning | yellow, bold | matches today's diagnostic warning |
| Success | green | task ran / completed-style accents |
| TaskName | bold + accent `170` | same accent as TUI title |
| Heading | bold `170` | group headers, help section titles |
| Muted | `245`/`241` | docs, echo, cache-hit, would-run |
| Locator | dim/underline | `file:line:col` emphasis (no width change) |
| Caret | red, bold | diagnostic caret span |

Exact codes are finalized in `internal/style`; this table is the contract.

**Rationale**: Reusing `170`/`245`/`241` (already in `internal/tui/styles.go`)
keeps the whole CLI visually consistent and avoids inventing a second palette.

**Alternatives considered**: truecolor/hex palette — rejected for portability and
to match the existing 256-color picker; termenv downgrades per terminal anyway.

## Open risks / verification hooks

- **lipgloss profile in CI**: confirm the forced `Ascii` profile yields zero
  escapes under the Docker harness (no TTY) — covered by the plain golden tests.
- **`--color=always` in tests**: integration test asserts ANSI **is** present in
  piped output; this is the one place escapes are expected.
- **Windows**: `fatih/color`/termenv handle VT enablement; the `build` + golden
  gates on Windows confirm no regression.
