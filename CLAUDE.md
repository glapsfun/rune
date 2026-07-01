<!-- SPECKIT START -->
Active feature plan: `specs/009-docs-and-badges/plan.md` (Modern, Example-Rich
Documentation & README Status Badges — a docs-and-README-only feature, no engine/
CLI/DSL changes and byte-identical golden output). Three tracks: (1) README light
header refresh + a live status-badge row (CI, release/tag, license, Go version, Go
Report Card, docs, Go Reference), targets pinned in `contracts/badges.md` — repo
badges → `glapsfun/rune`, module badges → `rune-task-runner/rune`; (2) reorganize
`docs/` in place along Diátaxis — fold `docs/guides/*` into `docs/how-to/*`, add
`docs/user-guide/` (an ordered tour that links out) and `docs/use-cases/`
(python/node/mcp walkthroughs anchored to existing `docs/examples/*`), front it
with an intent-first `docs/README.md`, and leave redirect stubs at old paths
(GitHub has no server-side redirects); (3) enforce accuracy by extending the
existing `test/docs` harness (CI gate `docs-verify`, run via `rune docs-check`) —
add `badges_test.go`, grow `codeblocks_test.go`'s `selfContainedPages`, rely on
`links_test.go`. For technologies, structure, and constraints, read the current
plan and its `research.md`.
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
