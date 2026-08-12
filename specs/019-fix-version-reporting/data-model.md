# Data Model: Accurate Version Reporting for Go-Toolchain Installs

No persistent data. The "model" is the in-memory version identity resolved
once at startup and the surfaces that consume it.

## Entities

### E1 — Build Provenance (input, read-only)

What the binary knows about how it was built. Sources, in precedence order:

| Field | Source | Release artifact | `go install mod@tag` | Checkout build | `go run` / no VCS |
|-------|--------|------------------|----------------------|----------------|-------------------|
| ldflags version | `-X main.version` | `0.4.3` | `dev` (default) | `dev` | `dev` |
| ldflags commit | `-X main.commit` | `13d059a` | `none` (default) | `none` | `none` |
| Module version | `BuildInfo.Main.Version` | (ignored — ldflags wins) | `v0.4.3` (or pseudo-version for `@commit`) | `v0.4.4-0.2026…+dirty` or exact tag | `(devel)` or absent |
| VCS settings | `BuildInfo.Settings` `vcs.*` | (ignored) | **absent** | **present** | absent |

Validation rules (from research D2/D3/D4):
- ldflags version ≠ `dev` → it is authoritative, build info never consulted.
- Module version used only if: non-empty, ≠ `(devel)`, and no `vcs.*`
  settings exist.
- Leading `v` trimmed for display.

### E2 — Reported Version (output, derived)

Single string handed to Cobra (`root.Version`) and `cli.Options.Version`.

State derivation (no transitions — computed once at startup):

| Provenance | Reported version | Gate behavior |
|------------|------------------|---------------|
| Release artifact | `0.4.3` (unchanged) | compared (unchanged) |
| `go install mod@vX.Y.Z` / `@latest` | `X.Y.Z` ← **the fix** | compared ← **the fix** |
| `go install mod@<untagged commit>` | pseudo-version, e.g. `0.4.4-0.2026…-432dbb8de56b` | compared as semver prerelease |
| Source checkout build (any state) | `dev` (unchanged) | bypassed with dev=true (unchanged) |
| `runetest` build + `RUNE_TEST_VERSION` | env value (unchanged, test-only) | per env value (unchanged) |

## Touched Surfaces

| # | Surface | Change |
|---|---------|--------|
| S1 | `cmd/rune/buildinfo.go` (NEW) | `resolveVersion(version) string` + pure `versionFromBuildInfo(...)` implementing E1→E2; only place that imports `runtime/debug`. |
| S2 | `cmd/rune/main.go` | Call site becomes `installedVersion(resolveVersion(version))`; comment updated. Nothing else. |
| S3 | `cmd/rune/buildinfo_test.go` (NEW) | Table-driven unit tests over constructed `*debug.BuildInfo` (cases from research D7). |
| S4 | `docs/installation.md` | Note that `go install` binaries report the installed release version (doc harness must stay green). |

Explicitly unchanged: `cmd/rune/version.go`, `cmd/rune/versionhook*.go`,
`internal/cli/*`, `internal/config/*`, `internal/semver/*`,
`.goreleaser.yaml`, `Dockerfile`, `Runefile`, release workflow,
`test/integration/*` (regression only).
