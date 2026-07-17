<!-- SPECKIT START -->
Active feature plan: `specs/011-rune-lsp/plan.md` (Rune Language Server Protocol —
a new `rune lsp` stdio JSON-RPC/LSP 3.17 server plus a standalone `rune analyze`
command, both driven by a single new `internal/analysis` service wrapping the
existing parse → compose → analyze pipeline. Delivers live diagnostics, context-aware
completion, go-to-definition (incl. cross-file), hover, document symbols, and
formatting — reusing the existing parser/analyzer/import-resolver/formatter with NO
second grammar, and running NOTHING during analysis). Four additive `internal/`
packages: `analysis` (Service.Analyze → Snapshot; SourceStore overlay; Workspace +
ImportGraph), `language` (Symbol Index ByName/ByQualified/ByDocument; scope;
completion/definition/hover; single builtin/setting/attribute registry), `lsp`
(hand-rolled Content-Length JSON-RPC + minimal typed LSP 3.17 subset — zero new deps,
per the `internal/semver` precedent; `convert.go` LineIndex does byte-col↔UTF-16),
and `formatter` (extracted from `internal/cli/fmt.go`). Backward-compatible changes to
existing packages: `diag.Diagnostic` gains `Code` + `Related []RelatedLocation`
(RUNE#### codes are a STABLE PUBLIC CONTRACT — see `contracts/diagnostic-codes.md`);
`config.Compose` refactored to read imports through the injected SourceStore instead
of `os.ReadFile` (so overlays apply transitively — today it takes a SourceProvider but
ignores it); parser recovery hardened (existing top-level recovery + optional
`InvalidStmt`). Wired as `cmd/rune/lsp.go` + `analyze.go`; MCP compose site is
`internal/cli/serve.go loadModule`. Safety (FR-028) enforced structurally (no
`runtime`/`os-exec`/net imports in analysis/language/lsp) + a no-side-effects test.
Clarifications (2026-07-10): codes are a contract; analyze/LSP report full transitive
import diagnostics; private tasks complete only in their own file; undocumented-public-
task warning (RUNE2010) ships as a non-gating warning. Constitution deviation
(4 new packages + formatter extraction vs locked layout) justified in plan Complexity
Tracking. Read the plan, `research.md`, `data-model.md`, and `contracts/` for details.
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
