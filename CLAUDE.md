<!-- SPECKIT START -->
Active feature plan: `specs/014-colorful-cli-output/plan.md` (Colorful CLI
Output Everywhere — finish feature 008's styling coverage so every
human-facing message uses the single semantic theme in `internal/style`, with
zero new packages, flags, or dependencies. Newly themed surfaces (S1–S13 in
`data-model.md`): the top-level failure banner in `cmd/rune/main.go` (`rune:`
prefix in the Error role, resolved tolerantly at banner time like help.go,
since PersistentPreRunE may not have run), cycle/watch error banners via a
shared `printErrorBanner` helper in `internal/cli`, the min-version override
warning (Warning role label), `rune analyze` text output rerouted through
`diag.RenderAllCoded` (new variant keeping the `error[RUNE2001]` token) for
run-path parity, grouped+styled help extended from root to all subcommands,
the bare-`rune` overview header, confirm prompt (Warning),
`formatted:`/`cleared:` labels (Success), watch/serve chrome (Muted). Also
fixed during implementation: `--color` was root-local and unusable with
subcommands — it is now a persistent flag (audit A13). The feared `--list`
wide-rune misalignment proved unreachable (task names are ASCII by grammar) —
no runewidth dependency; parity is pinned by test. TUI picker palette
literals unified with exported `style.Color*` constants. Hard constraints:
plain output byte-identical except analyze/help (FR-006), machine formats
(--dump, JSON modes, completion, LSP, MCP buffers) never carry ANSI (FR-007),
existing `--color`/`NO_COLOR` per-stream decision is the only gate (FR-008),
masking-through-styling covered by tests (FR-011), exit codes/streams/order
frozen (FR-013). Read the plan, `research.md` (decisions D1–D10),
`data-model.md` (surface inventory), `quickstart.md`, and `contracts/`
(`styled-output.md` C1–C9, `cli-audit.md` findings A1–A12) for details.
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
