# Implementation Plan: Minimum Rune Version

**Branch**: `010-minimum-rune-version` | **Date**: 2026-07-09 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/010-minimum-rune-version/spec.md`

## Summary

Add a `minimum_version` Runefile setting that pins the minimum Rune binary release a
project requires. Before imports are spliced, before semantic analysis, and before any
task/shell/interpreter/agent starts, Rune compares the installed binary version against
the root Runefile's declared requirement using Semantic Versioning precedence. An
incompatible binary aborts with a caret-anchored diagnostic (required vs installed +
upgrade URL) and a non-zero exit, executing nothing. The value must be a static string
literal (non-literal or non-semver values are rejected). A CLI-only `--ignore-version`
flag provides a loud break-glass override (disabled by default on the MCP/agent path). The
existing `rune version` command gains `--check` and `--json` for local and CI
compatibility inspection.

Technical approach: a new dependency-free helper package `internal/semver` (tiny SemVer
2.0.0 comparator, chosen over adding a module dependency to honor Principle V) plus a
`config.MinimumVersion`/`config.CheckMinimumVersion` layer that mirrors the existing
`internal/config/version.go` (`rune_version`) pattern. The gate is invoked from the two
static-load pipelines (`internal/cli/run.go` and `internal/cli/serve.go`) at the
pre-`Compose` point, where `file.Settings` holds only the root file's settings — which is
also what enforces "root owns the requirement." Installed version is threaded as an
explicit string argument (already available as `opts.Version`), and a test-only env hook
lets integration tests exercise older/newer binaries.

## Technical Context

**Language/Version**: Go 1.25.0 (module `github.com/rune-task-runner/rune`)

**Primary Dependencies**: `spf13/cobra` v1.10.2 (CLI); existing internal packages
(`token`, `lexer`, `parser`, `ast`, `analyzer`, `diag`, `config`, `cli`, `mcpserver`). New
internal package `internal/semver` (no third-party SemVer dependency — see research).

**Storage**: N/A (settings are read from the parsed Runefile AST)

**Testing**: `go test` inside the Docker harness (`docker-compose run --rm test go test
./...`); race variant with `CGO_ENABLED=1`. Layers: unit (co-located `_test.go`),
integration (`test/integration`, builds the real binary), golden diagnostics, cross-platform CI.

**Target Platform**: Single static CGO-free binary on Linux, macOS, Windows.

**Project Type**: CLI / command runner with an embeddable MCP server (single Go module).

**Performance Goals**: Version gate adds negligible cost (one string parse + comparison per
invocation, before parsing imports); no measurable regression on existing runs.

**Constraints**: Zero new third-party dependencies (Principle V); no CGO; the gate must run
with zero side effects before analysis/imports/execution (Principle II); existing Runefiles
without `minimum_version` must be byte-identical in behavior (FR-023).

**Scale/Scope**: One new internal package, ~2 gate call sites, 1 new global flag, 2 new
version-subcommand flags, MCP operator opt, plus docs and multi-layer tests.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|-----------|------------|
| I. Command Runner, Not a Build System | ✅ No caching/skip semantics touched. |
| II. Errors Are a Feature | ✅ Incompatibility and static-value errors reported with `file:line:col` + caret at the value literal, before any execution, non-zero exit, zero side effects. Gate runs before analysis. |
| III. Minimal, Total DSL | ✅ No new grammar/statement types; `minimum_version` is an ordinary `set` value consumed like `rune_version`. No loops/recursion introduced. |
| IV. Hand-Written Front End, Idiomatic Go | ✅ No parser/lexer changes required (settings already parse). New logic lives in small focused `internal/` packages (`internal/semver`, additions to `internal/config`). Package layout preserved. |
| V. Boringly Portable | ✅ Pure Go, CGO-free. **No new dependency**: a small internal SemVer comparator is written rather than importing `Masterminds/semver` or `golang.org/x/mod/semver`, which also lets us define Rune-specific dev-build (`X.Y.Z-dev+commit`) semantics precisely. |
| VI. Test-First, Multi-Layer Verification | ✅ Red-Green-Refactor; unit (semver + config helper), integration (reject/allow/override/`version --check`), golden diagnostics, cross-platform CI. Installed version is injectable (FR-011). |
| VII. AI-Native, Secure by Default | ✅ `--ignore-version` is CLI-only, never Runefile-enabled; on the MCP/agent path the override is disabled by default and only enabled via explicit operator configuration (`mcpserver.Options`), mirroring `AllowDestructive`. No secrets involved. |
| VIII. Go Engineering Discipline | ✅ Errors wrapped with `%w`; helper takes an explicit installed-version string (no hidden globals); `golangci-lint`/gofumpt clean. |
| Backward compatibility (opt-in) | ✅ Files without `minimum_version` behave exactly as today (FR-023). Feature is purely additive and opt-in per file. |
| Surface changes carry their docs | ✅ New setting ships with `docs/GRAMMAR.md` + settings docs updates and matching golden/integration fixtures in the same change (see Phase 1). |

**Result**: PASS — no violations, Complexity Tracking not required.

## Project Structure

### Documentation (this feature)

```text
specs/010-minimum-rune-version/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (CLI + setting contracts)
│   ├── setting.md
│   ├── cli-version.md
│   └── diagnostics.md
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
internal/
├── semver/                     # NEW: tiny SemVer 2.0.0 parse + precedence comparator
│   ├── semver.go
│   └── semver_test.go
├── config/
│   ├── version.go              # existing rune_version helper (unchanged pattern)
│   ├── minimum_version.go      # NEW: MinimumVersion(f) + CheckMinimumVersion(f, installed)
│   └── minimum_version_test.go # NEW
├── cli/
│   ├── run.go                  # gate call inserted between Parse (l.57) and Compose (l.61); --ignore-version handling + warning
│   ├── serve.go                # gate call inserted in loadModule between Parse (l.40) and Compose (l.42)
│   └── dispatch.go             # Options gains IgnoreVersion bool (+ installed version already present as Version)
└── diag/                       # reused as-is (New/Warn + render)

cmd/rune/
├── root.go                     # register global --ignore-version flag
├── version.go                  # add --check and --json; print language version line
└── main.go                     # test-only installed-version env hook wiring

mcpserver/
└── server.go                   # Options gains AllowIgnoreVersion (operator opt); default false

docs/
├── GRAMMAR.md                  # document minimum_version setting
└── runefile.md (+ settings doc) # document setting, override flag, version --check

test/integration/
└── us*_test.go                 # NEW/extended: reject/allow/override + version --check [--json]

testdata/diag/                  # golden diagnostic fixture(s) for minimum_version errors
```

**Structure Decision**: Single Go module, existing package layout preserved (Principle IV).
The only new package is `internal/semver` (isolated, dependency-free, heavily unit-tested).
All other changes are additive edits to existing files at the insertion points identified
in research. The gate lives in `internal/config` (pure, testable) and is *called* from the
two CLI static-load pipelines so it runs before imports/analysis/execution.

## Complexity Tracking

> No constitution violations — table intentionally empty.
