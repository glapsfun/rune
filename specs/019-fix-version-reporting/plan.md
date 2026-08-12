# Implementation Plan: Accurate Version Reporting for Go-Toolchain Installs

**Branch**: `019-fix-version-reporting` | **Date**: 2026-08-12 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/019-fix-version-reporting/spec.md`

## Summary

Binaries installed via the documented `go install
github.com/rune-task-runner/rune/cmd/rune@latest` path report `rune version
dev (commit none)` because the version is only stamped by GoReleaser
`-ldflags` (`-X main.version`), and they consequently bypass the Runefile
`minimum_version` gate as "development builds". Fix: when the ldflags
version is still the `dev` placeholder, fall back to the module version Go
embeds in every binary (`runtime/debug.ReadBuildInfo()` → `Main.Version`) —
but only for module-cache builds (no `vcs.*` build settings), so local
checkout builds keep reporting `dev` (FR-003) and release artifacts, whose
ldflags version wins outright, are byte-identical (FR-002). One new file in
`cmd/rune`, zero new dependencies, zero release-pipeline changes.

## Technical Context

**Language/Version**: Go 1.25 (`go.mod`; behavior relied upon — module
version stamping — is stable since Go 1.18 for `go install mod@ver` and
Go 1.24 for VCS-derived stamping of checkout builds)

**Primary Dependencies**: standard library only (`runtime/debug`); existing
internal packages `internal/cli` (Options.Version), `internal/config`
(MinimumRequirement.Satisfied), `internal/semver` — all unchanged

**Storage**: N/A

**Testing**: `go test` inside the Docker harness (`docker-compose run --rm
test go test ./...`); unit tests over a pure resolution function fed
constructed `*debug.BuildInfo` values; existing binary-level integration
tests (`test/integration`, `-tags runetest`, `RUNE_TEST_VERSION`) unchanged
and re-run as regression

**Target Platform**: Linux, macOS, Windows (single static binary,
`CGO_ENABLED=0`)

**Project Type**: CLI

**Performance Goals**: `ReadBuildInfo` is a one-time in-memory read at
startup; no measurable impact on `rune --version` (<1ms)

**Constraints**: release artifacts' version output byte-identical (FR-002 /
SC-003); the core release pipeline, `.goreleaser.yaml`, `Dockerfile`, and
`Runefile` build task untouched; no new flags or commands

**Scale/Scope**: ~1 new ~40-line file in `cmd/rune`, a one-line change in
`main.go`, unit tests, and a docs touch-up in `docs/installation.md`

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Verdict | Notes |
|---|-----------|---------|-------|
| I | Command Runner, Not a Build System | PASS (N/A) | No task-execution semantics touched. |
| II | Errors Are a Feature | PASS | The `minimum_version` mismatch diagnostic is unchanged; it now fires with the true version for go-install binaries instead of being silently bypassed. |
| III | Minimal, Total DSL | PASS (N/A) | Zero DSL surface change; no `docs/GRAMMAR.md` update needed. |
| IV | Hand-Written Front End, Idiomatic Go | PASS | Locked package layout untouched; one new file in the existing `cmd/rune` package. |
| V | Boringly Portable | PASS | Pure Go stdlib (`runtime/debug`), identical on all three OSes, CGO-free. |
| VI | Test-First, Multi-Layer Verification | PASS | Red-green: unit tests for the resolution function first; existing integration/golden suites re-run as regression (SC-004). |
| VII | AI-Native, Secure by Default | PASS | No MCP/secret surface change. The runetest-only `RUNE_TEST_VERSION` override stays compiled out of production binaries. |
| VIII | Go Engineering Discipline | PASS | No new deps, no goroutines, no globals beyond the existing ldflags vars; lint/gofumpt clean. |

Engineering constraints: Docker-only testing honored; locked layout
honored; backward compatible (output format unchanged, only truthfulness
improves); no DSL surface change so no grammar/goldens for it.

**Post-Phase-1 re-check**: PASS — design added no violations; Complexity
Tracking stays empty.

## Project Structure

### Documentation (this feature)

```text
specs/019-fix-version-reporting/
├── plan.md              # This file
├── research.md          # Phase 0 output (decisions D1–D7)
├── data-model.md        # Phase 1 output (S1–S4 touched surfaces)
├── quickstart.md        # Phase 1 output (validation scenarios V1–V6)
├── contracts/
│   └── version-output.md  # CLI version-output & gate contract (C1–C8)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
cmd/rune/
├── main.go              # MODIFIED: version resolution call site (one line)
├── buildinfo.go         # NEW: resolveVersion + pure versionFromBuildInfo
├── buildinfo_test.go    # NEW: unit tests over constructed *debug.BuildInfo
├── versionhook.go       # UNCHANGED (prod no-op hook)
├── versionhook_runetest.go  # UNCHANGED (test-only env override, still outermost)
└── version.go           # UNCHANGED (`rune version` command)

internal/cli/            # UNCHANGED (Options.Version, version_gate.go)
internal/config/         # UNCHANGED (MinimumRequirement.Satisfied dev-detection)
test/integration/        # UNCHANGED tests re-run as regression
docs/installation.md     # MODIFIED: note that go-install binaries report the release version
```

**Structure Decision**: All code changes live in the existing `cmd/rune`
main package — version resolution is binary-assembly concern, exactly where
the ldflags variables and the `runetest` hook already live. `internal/`
engine packages are untouched (Principle IV lock).

## Complexity Tracking

No constitution violations — table intentionally empty.
