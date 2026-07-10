<!-- SPECKIT START -->
Active feature plan: `specs/010-minimum-rune-version/plan.md` (Minimum Rune Version
— a new `set minimum_version := "0.8.0"` setting that pins the minimum required Rune
binary release). The gate runs before imports are spliced, before analysis, and
before any execution: it reads the setting from the ROOT Runefile's `file.Settings`
(pre-`config.Compose`, so imported/child values can never impose or relax the
requirement) and rejects an older binary with a caret-anchored diagnostic
(required/installed/upgrade-URL, exit 3, nothing executed). Key pieces: (1) new
dependency-free `internal/semver` package (SemVer 2.0.0 precedence; build metadata
ignored; prerelease < release, so `0.9.0-rc.1` does not satisfy `0.9.0`) — chosen
over a third-party dep to honor Boringly-Portable; (2) `config.MinimumVersion` /
`CheckMinimumVersion(file, installed)` mirroring the existing `rune_version` helper
in `internal/config/version.go`, requiring a static `*ast.StringLit` value else
`minimum_version must be a static semantic version`; (3) gate call sites in
`internal/cli/run.go` (between `parser.Parse` and `config.Compose`) and
`serve.go loadModule` (same point) — installed version threaded via `opts.Version`
with a test-only `RUNE_TEST_VERSION` env hook for integration tests; (4) CLI-only
global `--ignore-version` (warn+proceed, never Runefile-settable) plus MCP
operator-only `mcpserver.Options.AllowIgnoreVersion` (default false); (5) `rune
version` gains a language-version line, `--check`, and `--check --json`. `minimum_version`
stays independent of `rune_version`. Ships with GRAMMAR.md/settings docs and unit +
integration + golden-diagnostic + cross-platform tests. Read the plan and its
`research.md` for details.
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
