<!-- SPECKIT START -->
Active feature plan: `specs/019-fix-version-reporting/plan.md` (Accurate
Version Reporting for Go-Toolchain Installs — bug fix: binaries installed
via the documented `go install github.com/rune-task-runner/rune/cmd/rune@latest`
report `rune version dev (commit none)` because the version is only stamped
by GoReleaser ldflags, and they therefore silently bypass the Runefile
`minimum_version` gate as "dev builds". Fix: when the ldflags version is
still the `dev` placeholder, fall back to `runtime/debug.ReadBuildInfo()`
`Main.Version` — but ONLY for module-cache builds (no `vcs.*` build
settings, D3); source-checkout builds keep reporting `dev` (FR-003; a Go
≥1.24 checkout build at a clean tag would otherwise claim the release
version verbatim). Trim the leading `v` for display (D4). Touched surfaces
(S1–S4 in `data-model.md`): NEW `cmd/rune/buildinfo.go` (`resolveVersion` +
pure, unit-testable `versionFromBuildInfo`), one-line call-site change in
`cmd/rune/main.go` (`installedVersion(resolveVersion(version))` — the
runetest-only `RUNE_TEST_VERSION` hook stays outermost), NEW
`cmd/rune/buildinfo_test.go`, and a note in `docs/installation.md`. Hard
constraints: release artifacts' output byte-identical — ldflags always wins
(FR-002/C3); `internal/` packages, `.goreleaser.yaml`, `Dockerfile`,
`Runefile`, release workflow, and `test/integration/` all untouched;
commit field never fabricated — go-install binaries show `(commit none)`
(D5); no e2e go-install test in CI, it's a manual check (D7/V2). Read the
plan, `research.md` (D1–D7), `data-model.md` (E1–E2, S1–S4),
`quickstart.md` (V1–V6), and `contracts/version-output.md` (C1–C8).
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
