# Implementation Plan: Styled CLI Output & Friendlier Help

**Branch**: `008-styled-cli-output` | **Date**: 2026-06-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/008-styled-cli-output/spec.md`

## Summary

Give Rune's non-interactive output a restrained, semantic coat of color and
redesign `--help`, **without changing a single byte that scripts, pipes, or CI
see** on any surface except the deliberately-rewritten help. Today only
diagnostics and the `--choose` TUI are styled, color is decided by a single
`useColor()` that inspects **stderr only**, and `--list`/status/echo/cache/help
are plain.

Technical approach: introduce one leaf presentation package `internal/style`
that owns the **semantic palette + a Lip Gloss `Theme`** (roles: error, warning,
success, task-name, heading, muted, locator, caret) and degrades every role to a
zero/plain style when color is off — mirroring the existing `internal/tui`
`newStyles` pattern. Replace the single `useColor()` bool with a **per-stream
color decision** driven by a new global `--color=auto|always|never` flag: resolve
`ColorStdout` (for `--list`/`--help`) and `ColorStderr` (for status/echo/cache/
diagnostics) independently in `PersistentPreRunE`. Route every styled surface
through `internal/style` so the palette lives in exactly one place (SC-007), and
re-home the diagnostic colors there too (replacing the inline `fatih/color` calls
in `internal/diag/render.go`) while keeping plain output byte-identical (FR-018).
Redesign Cobra's help/usage via custom templates with grouped sections and worked
examples; its new plain form becomes a deliberately-reviewed golden baseline
(per the spec Clarifications). No engine logic, DSL, or stream assignment changes.

## Technical Context

**Language/Version**: Go 1.25.0

**Primary Dependencies**: `github.com/charmbracelet/lipgloss` (semantic theme +
forced/auto color profile via per-stream `lipgloss.Renderer`), `github.com/fatih/
color` + `github.com/mattn/go-isatty` (existing per-stream TTY/`NO_COLOR`
detection), `github.com/spf13/cobra` (custom help/usage templates). **No new
dependencies** — all four are already direct deps. Reused unchanged: the engine
packages (`token`…`runtime`), `internal/cli` `execute`/`listTasks`/`firstLine`/
`osMatches`, and `internal/tui` (the `--choose` picker is out of scope).

**Storage**: N/A (tasks read from the in-memory parsed Runefile).

**Testing**: `go test` inside the Docker harness (`docker-compose run --rm test
go test ./...`), race build via `-race`. New unit tests for `internal/style`
(roles are plain when disabled, styled when enabled). Golden/integration tests
prove byte-identical plain output for `--list`, status/echo/cache, and
diagnostics under three "off" triggers (piped, `NO_COLOR`, `--color=never`) and
presence of ANSI under `--color=always`; the diagnostic golden (color off) is
unchanged; the redesigned `--help` is verified by section/example substrings plus
strip-invariance (styled-stripped == plain), not a golden file.

**Target Platform**: Linux, macOS, Windows — single static binary, `CGO_ENABLED=0`
(Lip Gloss/termenv/fatih-color are pure Go).

**Project Type**: Single-project CLI (task runner).

**Performance Goals**: N/A — styling is per-line string wrapping with negligible
overhead; no hot path touched.

**Constraints**: Byte-for-byte invariance of stdout/stderr when color is off, for
every surface **except** `--help` (spec FR-010, SC-001, Clarifications). Color may
add only zero-width emphasis — **no column shifts**, especially in `--list`
padding and the diagnostic caret alignment (FR-012, FR-017, SC-003). Color
decision inputs limited to `NO_COLOR` + `--color` + per-stream TTY; `FORCE_COLOR`/
`CLICOLOR`/`CLICOLOR_FORCE` are **not** honored (FR-005, Clarifications). Exit
codes and stream assignment unchanged (FR-011).

**Scale/Scope**: One new leaf package `internal/style` (~theme + palette + tests);
edits to `cmd/rune/root.go` (flag + per-stream resolution + help templates),
`internal/cli/dispatch.go` (Options fields), `internal/cli/run.go` (`listTasks`
+ status lines), `internal/diag/render.go` (re-home colors via theme),
`internal/runtime/shell/shell.go` + the interp executor (dim echo). ~6 files
touched, 1 added. No DSL/grammar change.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment | Status |
|-----------|-----------|--------|
| I. Command Runner, Not a Build System | Cache-hit notices are only restyled (dimmed), never suppressed; every cache decision still logs to stderr exactly as today. No execution/skip semantics touched. | ✅ Pass |
| II. Errors Are a Feature | Diagnostics keep `file:line:col` + caret; color is re-homed into the shared theme but the **plain** rendering stays byte-identical to the golden (FR-018), and color adds zero width so caret columns never shift (SC-003). | ✅ Pass |
| III. Minimal, Total DSL | No DSL/grammar/expression change. | ✅ Pass |
| IV. Hand-Written Front End, Idiomatic Go | `internal/style` is a **presentation/leaf** package (like `internal/tui` in 007), not engine logic; the locked engine package set is unchanged. `internal/diag` gains a dependency on the leaf `internal/style` for role definitions only — detection stays at the `cmd` boundary. | ✅ Pass (see Complexity note) |
| V. Boringly Portable | Lip Gloss/termenv/fatih-color are pure Go, `CGO_ENABLED=0`, all three OSes; no system-shell dependency. | ✅ Pass (verify static cross-build in `build` gate) |
| VI. Test-First, Multi-Layer Verification | Red-Green-Refactor: `internal/style` table tests first; then golden/integration proving plain invariance under piped/`NO_COLOR`/`--color=never` and ANSI under `--color=always`. Diagnostic golden (off) unchanged; `--help` verified by substring + strip-invariance tests. | ✅ Pass |
| VII. AI-Native, Secure by Default | MCP surface untouched. Styled surfaces show only already-public data (task names/docs, same as `--list`); no secrets in help/echo/status. | ✅ Pass |
| VIII. Go Engineering Discipline | `golangci-lint` clean, gofumpt/goimports; errors wrapped with `%w`; no goroutines added; per-stream `lipgloss.Renderer` is constructed once per invocation, no globals mutated beyond the documented `fatih/color` global already in use. | ✅ Pass |

**Engineering Constraints**: Docker-only testing honored. Package layout additive
(`internal/style` is new and leaf; engine packages unchanged). Backward
compatible: `--color` defaults to `auto` = today's behavior; the only intentional
output change is `--help`. No DSL surface change → no `docs/GRAMMAR.md` impact;
the `--color` flag and the redesigned help ship with their golden/help fixtures
and a docs touch-up in the same PR (surface-changes-carry-their-docs).

**Result**: PASS — no violations. Complexity Tracking left empty. The single
layering note (`internal/diag` → `internal/style`) is justified below in lieu of
duplicating the palette, which would violate FR-001/SC-007.

## Project Structure

### Documentation (this feature)

```text
specs/008-styled-cli-output/
├── plan.md              # This file (/speckit-plan output)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── color-flag.md        # --color semantics + precedence table
│   └── output-surfaces.md   # per-surface styled-vs-plain contract
├── checklists/
│   └── requirements.md  # spec quality checklist (from /speckit-specify)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
internal/
├── style/                    # NEW leaf presentation package
│   ├── style.go              #   palette (one place) + Theme{Error,Warning,...} + New(enabled, w)
│   └── style_test.go         #   roles plain when disabled / styled when enabled
├── diag/
│   └── render.go             # EDIT: source role styles from internal/style (replace inline fatih/color); plain output unchanged
├── cli/
│   ├── dispatch.go           # EDIT: Options gains ColorMode + ColorStdout/ColorStderr (replacing single Color); helpers ThemeStdout()/ThemeStderr()
│   └── run.go                # EDIT: listTasks styles names/groups/docs (stdout theme, width preserved); status/cache lines use stderr theme
└── runtime/
    ├── shell/shell.go        # EDIT: dim command echo via an optional muted style passed through shell.Options
    └── interp/...            # EDIT: same echo dimming for the interp executor

cmd/rune/
├── root.go                   # EDIT: add --color flag (validate auto|always|never); replace useColor() with per-stream resolveColor(); custom help/usage templates with grouped sections + examples
└── help.go                   # NEW: custom root help func (grouped sections + examples) + heading colorizer

testdata/ & test/integration/ # golden for --list/status/diag (off = unchanged) + ANSI-present (forced) cases; --help verified by section/example substrings + strip-invariance (no golden file)
```

**Structure Decision**: Single-project CLI layout (unchanged). The only structural
addition is the leaf `internal/style` package; everything else is edits to
existing files. Engine packages remain untouched per Principle IV.

## Complexity Tracking

> No constitution violations require justification. One deliberate layering choice
> is recorded for reviewers:

| Choice | Why Needed | Simpler Alternative Rejected Because |
|--------|------------|--------------------------------------|
| `internal/diag` imports the leaf `internal/style` for role/palette definitions | FR-001 / SC-007 require the semantic palette to be defined in exactly one place and reused | Keeping diag's own inline `fatih/color` calls would duplicate the palette (two sources of truth) and risk diagnostics drifting from `--list`/status colors; `internal/style` is a dependency-free leaf, so the import adds no cycle and no engine-logic coupling. |
