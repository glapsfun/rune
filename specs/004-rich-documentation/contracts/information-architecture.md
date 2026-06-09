# Contract: Information Architecture & Page Structure

**Feature**: `004-rich-documentation`

Defines the page map, the structure every page must follow, the navigation guarantee, and the
canonical glossary. Enforces FR-001..FR-005, FR-010, FR-013..FR-016, SC-006, SC-008, SC-009.

## Page map (entry → depth)

| Layer | Page(s) | Kind | Purpose |
|-------|---------|------|---------|
| Entry | `README.md` | index | One-liner, hero example, install, doc map |
| Explain | `docs/overview.md` | overview | Problem, mental model, differentiators, **when to use / not use** (FR-001/FR-002) |
| Tutorial | `docs/getting-started.md` | tutorial | One linear install→write→run path, expected output inline (FR-003/FR-004) |
| How-to | `docs/guides/*.md` (9) | guide | One capability each: concept→syntax→runnable example→pitfalls (FR-010) |
| Examples | `docs/examples/README.md` + dirs | index | Use-case-grouped runnable library (FR-005) |
| Reference | `docs/cli.md`, `docs/runefile.md`, `docs/GRAMMAR.md` | reference | Exhaustive, accurate surfaces (FR-013) |
| Reference | `docs/installation.md` | reference | Per-OS install, cross-platform call-outs |
| Support | `docs/troubleshooting.md` | troubleshooting | Failure modes → exact diagnostic + exit code (FR-011) |

Guides to author (one per headline capability): `dependencies-and-hooks`, `parameters`,
`caching`, `parallelism`, `executors`, `settings-and-dotenv`, `imports-and-modules`,
`os-filtering`, `agents-and-mcp`.

## Per-page structure contract

Every `DocPage` MUST have:

1. **Orientation opener** — 1–3 lines: what this page is and where it sits (so a reader who
   deep-links in is not lost). (FR-014)
2. **Body** — scannable: descriptive headings, lists for 3+ items, fenced code with language
   tags, expected output shown after commands. (technical-writer principles)
3. **At least one runnable example reference** — for guide pages, a link into
   `docs/examples/<id>/`. (FR-009)
4. **Next-step footer** — ≥1 named onward link; **no dead-ends**. (FR-004/FR-014)

A `guide` page additionally MUST contain the four elements scored by SC-008:
**concept · syntax · ≥1 runnable example · pitfalls/edge-cases**.

## Navigation guarantee (SC-006)

From the entry point (`README.md` → `docs/overview.md`), any capability guide or any example
MUST be reachable in **≤2 clicks**. Concretely: README links the overview, examples index, and
reference; the examples index links every example; each example links its guide. This is
verified structurally (link graph depth) by `links_test.go` plus review.

## Canonical glossary (FR-016 / SC-009)

One name per concept. Authors use only the canonical term; the harness flags forbidden aliases.

| Canonical | Definition | Forbidden aliases |
|-----------|------------|-------------------|
| **Runefile** | The project's task file | `Runfile`, `runefile.rune` (as prose noun) |
| **task** | A named block of commands | `recipe`, `target`, `command` (as the unit) |
| **dependency** | A task that runs before another | `prerequisite` / `pre-req` **when it means task ordering** (use *dependency*) — see scope note |
| **post-hook** | A task that runs after, on success | `after-hook`, `finalizer` |
| **attribute** | A `[...]` annotation on a task | `decorator`, `tag` |
| **executor** | The runtime a body runs under (`sh`/python/node/agent) | `interpreter` (except the OS tool itself), `backend` |
| **agent task** | A task whose body is a natural-language instruction | `AI task`, `LLM task` |
| **MCP tool** | A task exposed to agents over MCP | `endpoint`, `function` |
| **content-hash caching** | Opt-in `[cache(...)]` skip | `incremental build`, `up-to-date check` |

> **Scope note (analysis finding I1).** "dependency" is canonical **only for the task-ordering
> concept** (a task that runs before another). The word **"Prerequisites"** is a *different*
> concept — the external tooling an example needs (python3/node/docker/agent CLI) — and is the
> required field name in the example contract (FR-008). The terminology alias check therefore
> flags `prerequisite`/`pre-req` **only when used for task ordering**, and MUST NOT flag the
> example-README `**Prerequisites:**` heading or prose about required tools. Implement the check
> with that exclusion (e.g. ignore the `Prerequisites:` line and tool-context sentences), or the
> check will false-positive on every conforming example.

## Consistency rule (SC-009)

`CONTRIBUTING.md`, the user guides, and the constitution MUST NOT contradict each other on
commands, policies (e.g. Docker-only testing), or terminology. Reviewed against the
`requirements.md` checklist before ship.
