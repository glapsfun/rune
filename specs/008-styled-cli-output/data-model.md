# Phase 1 Data Model: Styled CLI Output

This feature is presentation logic, not persisted data. The "entities" are the
in-memory styling types and the resolved color decision.

## Entity: ColorMode

The parsed value of the global `--color` flag.

| Field | Type | Values | Notes |
|-------|------|--------|-------|
| (value) | enum | `auto` \| `always` \| `never` | default `auto`; any other value → error (FR-009) |

- **Validation**: rejected at flag-parse / pre-run time; invalid value exits
  non-zero before any task runs.
- **Source**: `cmd/rune/root.go` persistent flag `--color`.

## Entity: ColorDecision (per stream)

The resolved on/off answer for one output stream, computed once per invocation in
`PersistentPreRunE`.

| Field | Type | Notes |
|-------|------|-------|
| ColorStdout | bool | gates `--list`, `--help` (stream = `cmd.OutOrStdout()`) |
| ColorStderr | bool | gates status/echo/cache/diagnostics (stream = `cmd.ErrOrStderr()`) |

**Resolution function** `resolve(mode, stream) bool` (replaces `useColor()`),
precedence highest-first:

1. `mode == never` → `false`
2. `mode == always` → `true`
3. `NO_COLOR` non-empty → `false`
4. `fatih/color.NoColor` (global) → `false`
5. else → `isatty(stream)` (terminal or Cygwin terminal)

`FORCE_COLOR` / `CLICOLOR` / `CLICOLOR_FORCE` are **not** inputs (Clarifications).

These replace the single `Options.Color` field in `internal/cli/dispatch.go`.
`Options` gains `ColorStdout`/`ColorStderr` (and retains a way to build a theme
per stream; diagnostics use the stderr decision).

## Entity: Theme (semantic role set)

Owned by the new `internal/style` package; the single source of truth for the
palette (FR-001, SC-007).

| Role | Purpose | Surfaces |
|------|---------|----------|
| Error | error severity / failures | diagnostics |
| Warning | warning severity | diagnostics, cache-write warning |
| Success | active/completed accents | status (`running:`) |
| TaskName | emphasize task identifiers | `--list` |
| Heading | group headers, help section titles | `--list`, `--help` |
| Muted | de-emphasized meta | `--list` docs, command echo, `cached:`/`would run:` |
| Locator | `file:line:col` emphasis | diagnostics |
| Caret | caret-span underline | diagnostics |

- **Construction**: `style.New(enabled bool, w io.Writer) Theme`.
- **Invariant (FR-003)**: when `enabled == false`, **every** role is the zero
  Lip Gloss style — `role.Render(s) == s` — so output contains no ANSI and is
  byte-identical to plain (mirrors `internal/tui` `newStyles`).
- **Invariant (FR-012)**: roles only add SGR escapes (color/bold/underline);
  they never add padding, spaces, or change string width.
- **Backing**: each role is a `lipgloss.Style` built from a per-stream
  `lipgloss.Renderer` with an explicit color profile (see research D2).

## Relationships

```text
--color flag ──parsed──▶ ColorMode
ColorMode + stream TTY ──resolve()──▶ ColorDecision{ColorStdout, ColorStderr}
ColorDecision(stream) ──▶ style.New(enabled, stream) ──▶ Theme
Theme ──used by──▶ listTasks (stdout) · help template (stdout)
                    diag.Render (stderr) · status/echo/cache (stderr)
```

## State transitions

None. All values are resolved once per invocation and are immutable thereafter.
