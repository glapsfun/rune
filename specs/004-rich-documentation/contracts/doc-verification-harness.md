# Contract: Documentation Verification Harness (`test/docs`)

**Feature**: `004-rich-documentation`

The harness is the mechanism that makes "docs never drift" a CI gate (FR-017, SC-003,
Principle VI). It is idiomatic, stdlib-only Go (Principle VIII) modeled on `test/integration`.

## How it is run

```sh
# Inside Docker, per project test policy:
docker-compose run --rm test go test ./test/docs/...

# Wired into the dev workflow via the repo-root Runefile:
rune docs-check        # repointed to the command above (was: check one Runefile parses)
```

CI MUST run documentation verification as a required check on the existing OS matrix.

## Package shape

```text
test/docs/
├── harness_test.go     # TestMain: build ./cmd/rune once into a temp dir (CGO_ENABLED=0)
├── examples_test.go    # discover docs/examples/*; table-driven per-example subtests
├── codeblocks_test.go  # extract complete ```rune blocks from docs/*.md → rune --list validate
└── links_test.go       # internal Markdown link + anchor resolution
```

`TestMain` mirrors `test/integration/harness_test.go`: find module root via `go.mod`, build the
binary, run `m.Run()`, clean up. Commands run with a per-invocation `context.Context` timeout;
errors are wrapped with `%w`.

### Idiomatic-Go implementation rules (golang-pro / golang-cli; Principle VIII)

- **Table-driven subtests**: discover examples/blocks into a slice, then `for _, tc := range … {
  tc := tc; t.Run(tc.name, func(t *testing.T){ … }) }` — one subtest per example so a single
  failure names the offending example, not the whole suite. Independent subtests MAY call
  `t.Parallel()`.
- **Process execution**: use `exec.CommandContext(ctx, runeBin, "--file", path, "--list")` with a
  bounded timeout; assert on `exec.ExitError` to read the exit code. Never `os.Exit` in a test.
- **Skips are explicit**: a missing interpreter calls `t.Skip(fmt.Sprintf("%s: %s not installed", id, tool))`
  (detected via `exec.LookPath`) — a logged skip, never a silent pass (D2/D9).
- **Helpers**: shared assert/run helpers call `t.Helper()` so failures point at the call site.
- **Errors**: wrap with `fmt.Errorf("...: %w", err)`; no ignored errors (`_`) without justification.
- **Goldens**: output-centric examples use a `-update`-gated golden under `test/docs/testdata/`
  (project golden discipline); default assertions are normalized-substring (D3).
- **Quality gate**: the package MUST be `gofumpt`/`golangci-lint`-clean and pass `go test -race`
  (Principle VI/VIII).

## Verification tiers

### Tier A — static (ALWAYS runs, every OS)

| Check | Pass condition |
|-------|----------------|
| Example validity | `rune --file docs/examples/<id>/Runefile --list` exits 0 for every example (parse+analyze, runs nothing) |
| Embedded blocks | Every *complete* fenced ` ```rune ` block in `docs/**/*.md`, written to a temp file, validates via `rune --file <tmp> --list` exit 0 |
| Coverage | Every Coverage-Matrix capability & project shape has ≥1 example dir + guide (FR-007/SC-004) |
| No secrets | No example/doc contains a secret-shaped literal (Principle VII) |
| Terminology | No forbidden glossary alias appears in any page body (FR-016) |

> **Static-validation invocation (analysis finding F1).** There is **no built-in `rune check`
> subcommand** — `cmd/rune` only special-cases `mcp`/`serve`/`completion`; any other first arg
> is a *task name*. The harness validates a Runefile with **`rune --file <path> --list`**, which
> forces parse+analyze (see `internal/cli/run.go`: analysis runs before the `--list` branch),
> requires no task argument, and executes nothing. Exit codes: `0` valid · `3` validation error
> · `2` usage/discovery.
>
> **Fenced-block authoring convention.** Only ` ```rune ` blocks that are *complete* Runefiles
> are validated. Deliberate fragments (a lone expression, a partial task) MUST be fenced
> ` ```text ` so the harness does not reject them as malformed files. A parse/analyze failure in
> a ` ```rune ` block is a test failure — fix the block or retag it.

### Tier B — execution (runs when prerequisites are present)

| Check | Pass condition |
|-------|----------------|
| Shell-only examples (`prereq: none`) | `run_command` exits as documented; stdout matches `expected_output` (normalized/substring) |
| python3 / node / docker / agent examples | Same, **only if** the tool is found via `exec.LookPath`; otherwise `t.Skip("<id>: <tool> not installed")` — a **logged skip, never a silent pass** (D2/D9) |

Output matching is normalized (trim trailing whitespace) substring assertion by default;
output-centric examples may use a golden under `test/docs/testdata/`, regenerated deliberately
with `-update` (project golden discipline).

### Link & reference integrity (always)

- Extract relative Markdown links/targets from `docs/**/*.md`, `README.md`, `CONTRIBUTING.md`.
- FAIL on any link to a non-existent file, or an in-page `#anchor` with no matching heading.
- External `http(s)` links are NOT fetched (offline/deterministic CI — D4).

### CLI reference drift (always)

- Parse flag names from `docs/cli.md`; compare against the binary's `--help` flag set.
- FAIL on any flag documented-but-absent or present-but-undocumented (D7, FR-013).

## Outputs / behavior

- Standard `go test` semantics: exit 0 = all pass; non-zero = a real failure (drift detected).
- Skips are reported with reasons in test output; the suite summary states how many examples
  ran vs. skipped, so reduced behavioral coverage is visible (no silent truncation).

## Non-goals

- Does NOT fetch external URLs. Does NOT change any product code. Does NOT bundle interpreters
  into CI — it degrades gracefully via skip-with-reason.
