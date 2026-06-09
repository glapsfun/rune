# Phase 0 Research: Rich, Example-Driven Documentation

**Feature**: `004-rich-documentation` · **Date**: 2026-06-09

The spec carried **no `[NEEDS CLARIFICATION]` markers** — open points were resolved as
documented assumptions. This file records the design decisions that turn those assumptions
into a buildable plan. Findings draw on the existing codebase (the `test/integration`
harness, the CLI surface, the current `docs/` set) and the loaded skills (`technical-writer`,
`golang-pro`, `golang-cli`). No external/web research was required — everything is internal.

---

## D1. Where examples live and how they are structured

- **Decision**: Runnable examples live under `docs/examples/<use-case>/`, one directory per
  example, each containing a `Runefile` plus a `README.md` (the metadata the example contract
  requires). An `docs/examples/README.md` is the use-case-grouped index.
- **Rationale**: The repo already has `docs/examples/getting-started/Runefile`, and
  `docs-check` already points there. Keeping the convention avoids churn and a second
  top-level directory. A per-example README makes each example self-orienting (FR-014) and
  carries the prerequisites/expected-output the verifier and reader both need.
- **Alternatives considered**: A top-level `examples/` dir (rejected: splits convention, would
  break the existing `docs-check` path for no reader benefit). A single mega-page of inline
  snippets (rejected: snippets rot, can't be run or verified, fails FR-006/FR-017).

## D2. How examples are kept honest (the verification harness)

- **Decision**: Add a Go test package `test/docs/` that mirrors `test/integration`:
  `TestMain` builds the `rune` binary once into a temp dir (`CGO_ENABLED=0`), and table-driven
  subtests verify examples in **two tiers**:
  - **Tier A — static (always runs):** every example `Runefile` and every fenced ` ```rune `
    block extracted from `docs/*.md` is statically validated (parse + analyze, exit 0) with no
    interpreter, so it runs in CI on every OS. **Invocation correction (analysis finding F1):**
    there is **no built-in `rune check` subcommand** — `cmd/rune` special-cases only
    `mcp`/`serve`/`completion`, so a bare `check` arg is treated as a *task name* (the existing
    `docs-check` only works because the getting-started Runefile happens to define a `check`
    task). The harness instead runs **`rune --file <path> --list`**, which forces parse+analyze
    (`internal/cli/run.go` analyzes before the `--list` branch), needs no task argument, and
    runs nothing — exit 0 = valid, exit 3 = validation error, exit 2 = usage. (`--dry-run` /
    `--dump` also force analysis but `--dry-run` needs a target task; `--list` is the robust
    choice.) **Fenced-block convention:** only ` ```rune ` blocks that are *complete* Runefiles
    are validated; deliberate fragments (lone expressions, partial snippets) MUST be fenced
    ` ```text ` so they are not treated as full files.
  - **Tier B — execution (runs when prerequisites present):** the example's documented task
    runs and its stdout/exit code are asserted. Shell-only examples (pure-Go `sh`) run
    everywhere. Examples needing python3/node/docker/an agent CLI run only when that tool is
    detected on PATH; otherwise the subtest calls `t.Skip` with a **logged reason** (never a
    silent pass).
- **Rationale**: Directly satisfies FR-017/SC-003 and Principle VI (multi-layer verification,
  test-first). The two-tier split means coverage of *validity* is total even on a minimal CI
  image, while *behavioral* coverage degrades gracefully and visibly. Reusing the existing
  harness pattern keeps the code idiomatic and low-risk (Principle IV/VIII).
- **Alternatives considered**: A third-party "run the code blocks in your markdown" tool
  (rejected: new dependency, weaker control, less idiomatic than reusing the in-repo harness,
  and would still not run real multi-file example projects). Asserting every shell command's
  output everywhere (rejected: not portable; gate behavioral assertions on tool availability
  instead). Hand-running examples before release (rejected: not repeatable, drifts immediately).

## D3. Static vs. behavioral assertions, and golden output

- **Decision**: Tier A asserts only exit 0 from `rune --file <path> --list`. Tier B asserts exit code and a
  **substring/normalized** match of stdout (e.g., trims trailing whitespace, tolerates the
  pure-Go shell's output), not byte-for-byte goldens, except where an example's whole point is
  exact output (then use a small golden under `test/docs/testdata/`, regenerated via `-update`
  consistent with the project's golden discipline).
- **Rationale**: Over-strict output matching makes examples brittle across OS/locale and turns
  docs maintenance into golden-churn. Substring/normalized matching is robust while still
  catching real drift. Reserving goldens for output-centric examples keeps the constitution's
  "goldens regenerated deliberately" rule meaningful.
- **Alternatives considered**: Byte-exact stdout for all (rejected: brittle, high false-positive
  rate across platforms).

## D4. Internal link integrity

- **Decision**: A `links_test.go` scans every `*.md` under `docs/` (and `README.md`,
  `CONTRIBUTING.md`), extracts relative Markdown links and reference targets, and fails on any
  that do not resolve to an existing file (and, for in-page anchors, an existing heading).
  External `http(s)` links are **not** fetched (no network in CI; would be flaky).
- **Rationale**: Satisfies FR-015/SC-005 deterministically and offline. Anchor checking catches
  the common "renamed a heading" breakage.
- **Alternatives considered**: A third-party link checker (rejected: dependency + network
  flakiness). Checking external URLs in CI (rejected: nondeterministic, rate-limited).

## D5. Information architecture (technical-writer skill)

- **Decision**: Adopt a layered IA driven by progressive disclosure:
  1. **README** (entry) → one-liner, the example, install, a doc map.
  2. **Overview** → problem, mental model, differentiators, *when to use / not use*.
  3. **Getting started** → one linear path to a first successful run, expected output inline.
  4. **Guides** (task-oriented, one per capability) → concept → syntax → runnable example →
     pitfalls/edge cases.
  5. **Examples library** (use-case-organized) → copy-paste starting points.
  6. **Reference** (CLI, Runefile language, grammar) → exhaustive, accurate.
  7. **Troubleshooting** → failure modes mapped to the exact diagnostic.
  Every page carries an **orientation header** (what this is) and a **next-step footer** + a
  cross-link to at least one example, guaranteeing no dead-ends and ≤2 clicks to any topic
  (FR-014/SC-006).
- **Rationale**: This is the standard, proven docs IA (Diátaxis-style separation of
  tutorial / how-to / reference / explanation) the `technical-writer` skill prescribes:
  user-goal-first, scannable, progressive. It maps cleanly onto the spec's four user stories
  (US1→overview+getting-started, US2→examples, US3→guides+reference, US4→CONTRIBUTING).
- **Alternatives considered**: One giant README (rejected: not scannable, no progressive
  disclosure). Reference-only docs (rejected: fails the "understandable / main idea" ask).

## D6. Terminology consistency

- **Decision**: Maintain a small glossary (in `contracts/information-architecture.md`) fixing
  **one canonical name per concept** — e.g., *Runefile*, *task*, *recipe→use "task"*,
  *dependency/prerequisite* (pick one), *attribute*, *executor*, *agent task*, *MCP tool*.
  Authors use only canonical terms; a lightweight check flags known forbidden aliases.
- **Rationale**: Satisfies FR-016/SC-009. A fixed lexicon is the cheapest way to keep a
  multi-page set coherent, and it aligns the docs with the constitution's vocabulary.
- **Alternatives considered**: Ad-hoc wording per page (rejected: drift, reader confusion).

## D7. CLI reference accuracy (golang-cli skill)

- **Decision**: `docs/cli.md` is generated/derived from the **authoritative surface** captured
  in `contracts/cli-reference.md`, itself transcribed from `cmd/rune/main.go` (flags) and
  `internal/cli/exit.go` (exit codes). Documented exit codes: `0` success, `1` task failure,
  `2` usage, `3` validation, `130` interrupt. Reserved subcommands `mcp` / `serve`
  (`--http`/`--addr`/`--token-file`) / `completion` are documented as such. The verification
  harness includes a check that every flag named in `docs/cli.md` exists in the binary's
  `--help` output (and vice-versa) so the reference cannot silently drift.
- **Rationale**: The `golang-cli` skill stresses Unix-faithful exit codes and stdout/stderr
  discipline — Rune already follows these (messages to stderr, clean stdout, sysexits-style
  codes). The doc must mirror reality exactly; deriving it from the source + a drift check makes
  that durable (FR-013).
- **Alternatives considered**: Hand-written reference with no cross-check (rejected: drifts as
  flags change).

## D8. Cross-platform parity in examples

- **Decision**: Default every example to the pure-Go `sh` executor and POSIX-portable commands;
  where an example must show OS-specific behavior, call it out inline ("On Windows, …"). The
  harness runs on the same OS matrix as the rest of CI, so a non-portable example fails fast.
- **Rationale**: FR-018/SC and Principle V. Rune's whole portability promise is that `(sh)`
  tasks behave identically everywhere; examples should demonstrate, not undermine, that.
- **Alternatives considered**: Linux-only examples (rejected: contradicts the portability
  selling point and FR-018).

## D9. Prerequisite-gated examples (python/node/docker/agent)

- **Decision**: Each example's README states prerequisites up front (FR-008). In the harness,
  Tier B detects the tool (`exec.LookPath`) and skips-with-reason when absent; Tier A still
  validates the example statically. Agent/MCP examples never embed credentials and demonstrate
  read-only-by-default access (Principle VII).
- **Rationale**: Lets the library cover polyglot/agent use cases (US2/FR-007) without making CI
  depend on every interpreter, while never silently claiming coverage it didn't exercise.
- **Alternatives considered**: Excluding non-shell use cases (rejected: guts the "different use
  cases" ask). Requiring all interpreters in CI (rejected: heavy, brittle image).

## D10. Easy-start CONTRIBUTING (technical-writer skill)

- **Decision**: Restructure `CONTRIBUTING.md` around a newcomer's path: **What to contribute**
  (lead with low-barrier wins: docs fixes, new examples) → **Set up** (prereqs, clone, build) →
  **Make a change** (a concrete "add an example" walkthrough) → **Verify** (the Docker-only test
  policy, exact commands, and what to do without `rune` installed: `go run ./cmd/rune <task>`) →
  **Repo map** (where things live) → **Propose it** (PR + the CI gates that will run). Keep it
  consistent with the user docs and the constitution (FR-019..FR-023/SC-009).
- **Rationale**: The current CONTRIBUTING is accurate but gate-first and assumes context; the
  ask is "understandable, easy to start." Goal-ordered, beginner-first structure with an
  explicit first-contribution walkthrough is the technical-writer-recommended shape.
- **Alternatives considered**: Leaving CONTRIBUTING as-is (rejected: explicit user ask to
  rework it). A separate "good first issue" doc (deferred: fold the starter guidance into
  CONTRIBUTING for one obvious entry point).

## D11. Wiring verification into the dev workflow & CI

- **Decision**: Repoint the `docs-check` task from "check one Runefile parses" to "run the
  `test/docs` harness inside Docker" (`docker-compose run --rm test go test ./test/docs/...`),
  and ensure CI invokes documentation verification as a required check. Keep the old
  single-file `rune check` behavior subsumed by Tier A.
- **Rationale**: Makes the no-drift guarantee a real gate (Principle VI; CONTRIBUTING/CI claims
  must be true). Running through Docker honors the project test policy.
- **Alternatives considered**: A separate ad-hoc script (rejected: tests belong in the Go test
  harness and the Docker flow, not a bespoke runner).

---

### Resolved unknowns summary

| Topic | Resolution |
|-------|------------|
| Example location/structure | `docs/examples/<use-case>/` = `Runefile` + `README.md`; grouped index (D1) |
| Drift prevention | Two-tier Go harness `test/docs` (static `--list` validate + gated run) + link check (D2/D4) |
| Output assertions | Substring/normalized; goldens only for output-centric examples (D3) |
| Doc IA | Diátaxis-style layered set; orient-header + next-step + example link per page (D5) |
| Terminology | Canonical glossary, one name per concept, alias check (D6) |
| CLI reference accuracy | Derived from source + flag-drift check; exit codes 0/1/2/3/130 (D7) |
| Cross-platform | Pure-Go `sh` default; inline OS call-outs; full OS CI matrix (D8) |
| Polyglot/agent examples | Prereqs stated; Tier-B skip-with-reason; no secrets (D9) |
| CONTRIBUTING | Newcomer-path rework, low-barrier-first, Docker policy explained (D10) |
| CI wiring | `docs-check` → `test/docs` in Docker; required CI check (D11) |

**All Phase 0 unknowns resolved. Ready for Phase 1.**
