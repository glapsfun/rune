# Phase 1 Data Model: Documentation Content Model

**Feature**: `004-rich-documentation` · **Date**: 2026-06-09

This feature ships documentation, not a runtime data store. The "data model" is therefore the
**content model**: the entities the documentation set is made of, the metadata each carries,
and the rules the verification harness enforces. These map directly to the spec's Key Entities
and Functional Requirements.

---

## Entity: DocPage

A single Markdown page in the documentation set.

| Field | Type | Notes |
|-------|------|-------|
| `path` | string | Repo-relative, e.g. `docs/guides/caching.md` |
| `title` | string | H1; unique within the set |
| `kind` | enum | `overview` \| `tutorial` \| `guide` \| `reference` \| `troubleshooting` \| `index` |
| `orientation` | text | Opening lines: what this page is + where it sits (FR-014) |
| `audience` | enum | `newcomer` \| `task-author` \| `integrator` \| `contributor` |
| `next_steps` | link[] | ≥1 named onward link; no dead-ends (FR-004/FR-014) |
| `cross_links` | link[] | Related guides + ≥1 example link for guides (FR-009/FR-014) |
| `terms_used` | term[] | Must be canonical glossary terms only (FR-016) |

**Validation rules**
- Every page MUST have an `orientation` and at least one `next_steps` link (FR-014).
- A `guide` page MUST link to ≥1 `Example` (FR-009/FR-014).
- All internal links in `cross_links`/`next_steps` MUST resolve (FR-015 — enforced by
  `links_test.go`).
- Only canonical terms may appear for glossary concepts (FR-016 — alias check).

---

## Entity: Example

A self-contained, runnable artifact under `docs/examples/<use-case>/`.

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | Directory name, e.g. `caching`, `go-service` |
| `title` | string | Human label in the library index |
| `group` | ExampleGroup | Use-case category it belongs to |
| `purpose` | text | What it demonstrates and why it's written this way (FR-006) |
| `capabilities` | Capability[] | Which capabilities it exercises (FR-009) |
| `prerequisites` | enum[] | `none` \| `python3` \| `node` \| `docker` \| `agent-cli` (FR-008) |
| `run_command` | string | The exact command a reader runs, e.g. `rune build` |
| `expected_output` | text | What success looks like (FR-006); the harness asserts it |
| `guide_link` | link | Cross-link to the deeper capability guide (FR-009) |
| `verification_tier` | enum | `static` (always) and/or `run` (when prerequisites present) |
| `files` | path[] | `Runefile` (required) + `README.md` (required) + minimal support files |

**Validation rules**
- MUST contain a `Runefile` and a `README.md` carrying `purpose`, `prerequisites`,
  `capabilities`, `run_command`, `expected_output`, `guide_link` (example contract).
- MUST statically validate via `rune --file <path> --list` (Tier A, always) — Principle III/VI, FR-006.
- If `prerequisites == none`, MUST also pass Tier B (run + assert `expected_output`).
- MUST NOT contain any secret/credential literal (Principle VII).
- MUST declare any non-`none` prerequisite in `README.md` before run steps (FR-008/SC-010).

---

## Entity: ExampleGroup

A labeled use-case category in the example library index.

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | e.g. "Project shapes", "Capability spotlights" |
| `label` | text | One-line description so readers can scan (FR-005) |
| `examples` | Example[] | Members, ordered simple→advanced |

**Validation rules**
- Every `Example` belongs to exactly one `ExampleGroup` (FR-005).
- The union of groups MUST cover the FR-007 minimum set (see Coverage Matrix below) — checked
  for completeness (SC-004).

---

## Entity: Capability

A headline product capability the docs must teach.

| Field | Type | Notes |
|-------|------|-------|
| `key` | string | e.g. `caching`, `parallelism`, `executors`, `agents-mcp` |
| `guide` | DocPage | The task-oriented guide (FR-010) |
| `examples` | Example[] | ≥1 worked example (FR-007) |
| `failure_modes` | FailureMode[] | Documented "what happens when…" cases (FR-011) |

**Validation rules**
- Each Capability MUST have a guide containing **concept + syntax + ≥1 runnable example +
  pitfalls/edge-cases** (FR-010/SC-008).
- Each Capability MUST have ≥1 `Example` (FR-007/SC-004).

---

## Entity: FailureMode

A documented error/edge case and the diagnostic a reader should expect.

| Field | Type | Notes |
|-------|------|-------|
| `trigger` | text | e.g. "depend on an unknown task", "missing interpreter" |
| `expected_behavior` | text | What Rune does (e.g. nothing runs; non-zero exit) |
| `diagnostic` | text | The `file:line:col` + caret message shape (Principle II/FR-011) |
| `exit_code` | enum | `0/1/2/3/130` per `contracts/cli-reference.md` |

---

## Entity: CLIReferenceEntry

One command, flag, or exit code in the CLI reference (FR-013).

| Field | Type | Notes |
|-------|------|-------|
| `name` | string | e.g. `--list`, `--file`, `mcp`, `serve`, `completion` |
| `kind` | enum | `flag` \| `subcommand` \| `exit-code` |
| `description` | text | Matches actual behavior |
| `source_of_truth` | ref | `cmd/rune/main.go` (flags) / `internal/cli/exit.go` (codes) |

**Validation rules**
- Every flag documented MUST exist in the binary's `--help`, and vice-versa (drift check, D7).

---

## Entity: GlossaryTerm

| Field | Type | Notes |
|-------|------|-------|
| `canonical` | string | The one approved name (e.g. "task") |
| `definition` | text | Short, plain-language |
| `forbidden_aliases` | string[] | Names that MUST NOT be used (e.g. "recipe") |

**Validation rules**
- Forbidden aliases MUST NOT appear in any `DocPage` body (FR-016 — alias check).

---

## Entity: VerificationResult (transient, harness output)

Not stored in the repo; produced per test run.

| Field | Type | Notes |
|-------|------|-------|
| `example_id` | string | |
| `tier` | enum | `static` \| `run` |
| `status` | enum | `pass` \| `fail` \| `skipped` |
| `skip_reason` | text | Required when `skipped` (no silent skips — D2/D9) |

---

## Relationships

```text
ExampleGroup 1───* Example *───* Capability 1───1 CapabilityGuide (DocPage kind=guide)
                      │                      │
                      │ guide_link           *
                      └────────────────────► FailureMode
DocPage *───* DocPage         (cross_links / next_steps; all must resolve)
DocPage *───* GlossaryTerm    (terms_used ⊆ canonical terms)
CLIReferenceEntry ──derives──► docs/cli.md   (kept in sync by drift check)
```

## Coverage Matrix (FR-007 / SC-004 — the completeness gate)

**Headline capabilities** (each needs ≥1 example + a guide):
dependencies & hooks · parameters/variadics · caching (`[cache]`) · parallel prerequisites ·
multi-language bodies (executors) · settings & dotenv · imports/modules · OS filtering ·
agent/MCP surface.

**Common project shapes** (each needs ≥1 example):
compiled-language service (`go-service`) · Node/JavaScript (`node-project`) · Python
(`python-project`) · monorepo (`monorepo`) · CI/CD pipeline (`ci-cd`) · containerized workflow
(`docker-workflow`) · polyglot repo (`polyglot`) · agent-driven workflow (`agent-driven`).

The harness/coverage check fails if any row above lacks a corresponding `Example`/guide,
making SC-004 a hard, measurable gate rather than a judgement call.
