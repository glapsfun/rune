# Implementation Plan: Colorful CLI Output Everywhere

**Branch**: `014-colorful-cli-output` | **Date**: 2026-07-22 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/014-colorful-cli-output/spec.md`

## Summary

Feature 008 built the styling machinery — a single semantic theme
(`internal/style`), per-stream color resolution (`cmd/rune/color.go`:
`--color auto|always|never` > `NO_COLOR` > TTY), and themed rendering for
`--list`, run-status lines, command echo, diagnostics, and root help — but left
roughly a dozen human-facing surfaces plain, most visibly the top-level failure
banner (`cmd/rune/main.go:59`). This feature closes the gap using the existing
machinery only: every remaining human-facing message adopts the shared theme,
`rune analyze` re-uses the run-path diagnostic renderer, subcommand help adopts
the root help's grouped/styled layout, the `--list` colored-branch padding
switches to display width (fixing a latent alignment mismatch), and the TUI
picker's duplicated palette literals are replaced by exported constants from
`internal/style`. No new packages, no new direct behavior switches, no parser or
DSL changes. Plain (color-off) output stays byte-identical everywhere except the
two surfaces the spec explicitly reformats (analyze text output, subcommand
help), whose goldens are regenerated deliberately once. A documented audit of
all commands/flags (`contracts/cli-audit.md`) satisfies the review request.

## Technical Context

**Language/Version**: Go 1.25 (pure Go, `CGO_ENABLED=0`)

**Primary Dependencies**: `charmbracelet/lipgloss` v1.1.0 + `muesli/termenv`
(already the styling stack), `mattn/go-isatty` (TTY detection),
`mattn/go-runewidth` v0.0.19 (currently indirect; promoted to direct for
display-width padding), `spf13/cobra` (help plumbing). **No new modules.**

**Storage**: N/A

**Testing**: `go test ./...` in Docker (`docker-compose run --rm test`), golden
files under `internal/cli/testdata` + binary-level integration tests asserting
stdout/stderr bytes and exit codes; `-race` variant in CI

**Target Platform**: Linux, macOS, Windows terminals (ANSI via termenv on
Windows 10+), plus non-TTY pipes/redirects

**Project Type**: Single Go CLI (`cmd/rune` + `internal/…` packages)

**Performance Goals**: No measurable overhead on the run path — theme
construction stays once-per-invocation; zero allocations added to per-line
output paths beyond the existing `Render` calls

**Constraints**: Plain output byte-identical for all surfaces except analyze
text + subcommand help (spec FR-006); machine formats (JSON, `--dump`,
completion scripts, LSP JSON-RPC, MCP buffers) must never carry ANSI (FR-007);
secret masking must keep working through styled writes (FR-011); exit codes,
ordering, and stream assignment unchanged (FR-013)

**Scale/Scope**: ~12 output surfaces across `cmd/rune` (4 files) and
`internal/cli` (6 files), 1 constant export in `internal/style`, 1 palette
swap in `internal/tui`; plus goldens and integration tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Verdict | Notes |
|---|-----------|---------|-------|
| I | Command Runner, Not a Build System | PASS | No execution/caching semantics touched; only message presentation. Cache-hit logging remains visible (now consistently styled). |
| II | Errors Are a Feature | PASS (improves) | Failure banner, cycle errors, watch errors gain the shared error role; `analyze` adopts the caret/locator renderer used by runs, so diagnostics get *better*, never terser. Exit codes unchanged. |
| III | Minimal, Total DSL | PASS | Zero DSL/lexer/parser/analyzer changes. |
| IV | Hand-Written Front End, Idiomatic Go | PASS | No new packages; changes live in existing `cmd/rune`, `internal/cli`, `internal/style`, `internal/tui`, `internal/diag` call sites. Package layout untouched. |
| V | Boringly Portable | PASS | Styling stack already cross-platform (termenv). `go-runewidth` is pure Go and already in the module graph. No shell/OS-specific behavior. |
| VI | Test-First, Multi-Layer Verification | PASS | Invariance tests written first (plain-bytes golden comparison per surface); analyze/help goldens regenerated deliberately as an explicit, reviewed step; integration tests assert ANSI presence with `--color=always` and absence with `--color=never`/pipes. |
| VII | AI-Native, Secure by Default | PASS | MCP adapter constructs `Options` with both color booleans false — regression test added asserting no ANSI in MCP tool results. Masking order (mask writer wraps stream, theme renders on top of masked writer; banner masking via `maskErr`) preserved and covered by a styled-banner masking test. |
| VIII | Go Engineering Discipline | PASS | No goroutines, no globals, no `init()`; `golangci-lint` zero-issue gate applies. |

**Initial gate: PASS (no violations).**
**Post-design re-check (after Phase 1): PASS — design added no packages, no new
flags, no deviations; Complexity Tracking remains empty.**

## Project Structure

### Documentation (this feature)

```text
specs/014-colorful-cli-output/
├── plan.md              # This file
├── research.md          # Phase 0: decisions D1–D10
├── data-model.md        # Phase 1: surfaces × roles model
├── quickstart.md        # Phase 1: end-to-end validation guide
├── contracts/
│   ├── styled-output.md # Surface-by-surface output contract (styled + plain)
│   └── cli-audit.md     # FR-012 command/flag audit (findings: fixed/deferred)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
cmd/rune/
├── main.go              # D1: themed failure banner (tolerant color resolution)
├── color.go             # unchanged (decision logic reused)
├── help.go              # D5: grouped+styled help extended to subcommands
├── serve.go             # audit-only (wording)
├── version.go           # audit-only (stdout product output stays plain)
└── analyze.go           # unchanged wiring (rendering moves in internal/cli)

internal/style/style.go  # D9: export palette constants; add Help/dark-muted role
internal/cli/
├── run.go               # D2 cycle banner, D6 overview, D7 confirm/clear-cache,
│                        # D8 --list display-width padding
├── watch.go             # D2: watch banners/errors themed
├── version_gate.go      # D3: warning label uses Warning role
├── analyze.go           # D4: text mode renders via diag.RenderAll + themed summary
├── fmt.go               # D7: "formatted:" notice themed
├── serve.go             # D7: MCP startup banners themed
├── masking.go           # unchanged (order verified by test)
└── testdata/…           # goldens: analyze text + help regenerated; rest frozen

internal/tui/styles.go   # D9: palette literals → style package constants
internal/diag/render.go  # unchanged (reused by analyze)

tests: per-package unit tests + integration tests (binary-level) asserting
ANSI/plain matrices for every touched surface
```

**Structure Decision**: Single-project Go CLI; all changes land in the existing
locked package layout. The only cross-package addition is exporting named
palette constants from `internal/style` so `internal/tui` stops duplicating
color literals.

## Complexity Tracking

> No constitution violations — table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
