# Quickstart: Validating the Documentation Feature

**Feature**: `004-rich-documentation` · **Date**: 2026-06-09

This guide proves the documentation set works end-to-end. It is a **validation/run guide**, not
implementation — see `contracts/` and `data-model.md` for the detailed rules. Tests run
**inside Docker** per project policy.

## Prerequisites

- Repo checked out on branch `004-rich-documentation`.
- Docker + `docker-compose` (the test harness runs in a container).
- Optional, to exercise Tier-B for polyglot/agent examples locally: `python3`, `node`,
  `docker`, an agent CLI. Absent tools cause **logged skips**, not failures.

## Scenario 1 — Examples never drift (FR-017 / SC-003)

Run the documentation verification harness:

```sh
docker-compose run --rm test go test ./test/docs/...
```

**Expected**: exit 0. Output shows every example passing Tier-A `rune --file <path> --list` validation, shell-only
examples passing Tier-B run+assert, and any interpreter/agent examples reported as
`SKIP: <tool> not installed` (never silently passed). The dev-workflow alias is equivalent:

```sh
rune docs-check        # repointed to the harness (was: check one Runefile)
```

## Scenario 2 — Coverage is complete (FR-007 / SC-004)

The harness's coverage check (part of Scenario 1) fails if any Coverage-Matrix capability or
project shape lacks an example/guide. To inspect coverage directly:

```sh
ls docs/examples/        # expect dirs for every project shape + capability spotlight
rune --list              # sanity: the dogfood Runefile still lists its dev tasks
```

**Expected**: an example directory exists for each row of the Coverage Matrix
(`data-model.md`), each conforming to `contracts/example-contract.md`.

## Scenario 3 — Internal links resolve (FR-015 / SC-005)

Covered by `links_test.go` inside Scenario 1. It fails on any relative link to a missing file
or any `#anchor` with no matching heading across `docs/**/*.md`, `README.md`, `CONTRIBUTING.md`.

**Expected**: zero broken internal links.

## Scenario 4 — CLI reference matches the binary (FR-013)

Covered by the drift check inside Scenario 1: flags in `docs/cli.md` ⇔ flags in the binary's
`--help`; documented exit codes equal `{0,1,2,3,130}`. To eyeball it:

```sh
go run ./cmd/rune --help
```

**Expected**: every flag shown is documented in `docs/cli.md` and vice-versa.

## Scenario 5 — A newcomer reaches a first success fast (US1 / SC-002)

Follow `docs/getting-started.md` exactly, from a clean directory, on a machine **without** Rune
preinstalled:

```sh
# (getting-started shows the install line, then:)
mkdir /tmp/rune-try && cd /tmp/rune-try
# create the Runefile shown on the page, then:
rune greet
```

**Expected**: the documented output appears (e.g. `Hello, world! ...`) on the first attempt,
within 5 minutes of starting, with no steps missing. The page ends by pointing to a named next
step (examples library or language guide) — no dead-end (FR-004).

## Scenario 6 — A reader understands the idea (US1 / SC-001)

Read **only** `docs/overview.md`.

**Expected**: a reader can state, in their own words, that Rune is a command runner (not a
build system) that runs the same tasks for humans and AI agents, and can tell from the "when to
use / not use" section whether it fits their need — in under 3 minutes.

## Scenario 7 — Find the right doc in ≤2 clicks (SC-006)

From `README.md`, navigate to (a) the caching guide and (b) the `agent-driven` example.

**Expected**: each is reachable within two link clicks (README → examples/guide index →
target), confirming the navigation guarantee in `contracts/information-architecture.md`.

## Scenario 8 — A first contribution is easy (US4 / SC-007)

Following **only** `CONTRIBUTING.md` from a clean clone, add a trivial new example directory
(Runefile + README per the example contract), then verify it:

```sh
docker-compose run --rm test go test ./test/docs/...
```

**Expected**: a newcomer can make the change and run the checks the supported way in under 15
minutes without asking a maintainer how to begin; the new example is picked up and verified by
the harness automatically.

## Scenario 9 — Consistency review (SC-009)

Review `CONTRIBUTING.md`, the user guides, and `.specify/memory/constitution.md` against the
`checklists/requirements.md` items.

**Expected**: zero contradictions in commands, policies (Docker-only testing), or terminology
(canonical glossary respected).

---

### Pass criteria summary

| Scenario | Proves | Gate |
|----------|--------|------|
| 1 | Examples + blocks verified, no drift | `go test ./test/docs` exit 0 (SC-003) |
| 2 | Capability/shape coverage complete | coverage check (SC-004) |
| 3 | Internal links resolve | `links_test.go` (SC-005) |
| 4 | CLI reference accurate | drift check (FR-013) |
| 5 | First success < 5 min | manual run-through (SC-002) |
| 6 | Idea understood < 3 min | overview read (SC-001) |
| 7 | Topic in ≤2 clicks | link-graph depth (SC-006) |
| 8 | First contribution < 15 min | CONTRIBUTING walk-through (SC-007) |
| 9 | Set is internally consistent | checklist review (SC-009) |
