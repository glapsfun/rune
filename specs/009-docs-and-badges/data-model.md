# Data Model: Documentation Set & README Badges

This feature has no runtime data. The "model" is the structure of the documentation artifacts
and the badge set — the entities the plan, contracts, and verification harness reason about.

## Entities

### DocPage

A single markdown file under `docs/` (or `README.md`).

| Field | Description |
|-------|-------------|
| path | Repo-relative, forward-slash (e.g. `docs/how-to/caching.md`) |
| mode | Diátaxis mode: `tutorial` \| `how-to` \| `reference` \| `explanation` \| `index` \| `use-case` \| `stub` |
| title | H1 heading |
| intro | Short lead paragraph |
| toc | Manual anchor TOC — required only on long reference pages |
| body | Content; every output-bearing snippet paired with an expected-output block |
| callouts | Zero+ GitHub Alerts (`[!NOTE|TIP|IMPORTANT|WARNING|CAUTION]`), ≤2/page |
| next_steps | Footer link block routing onward (required, except stubs) |
| example_ref | For `use-case`/`how-to`: link to a backing example under `docs/examples/` |

**Rules**: consistent structure (title → intro → [toc] → body → next_steps) per FR-014; links
are relative; snippets that produce meaningful output show it (FR-015); callouts use GitHub
Alert syntax (FR-016).

**Mode mapping (reorganize in place)**

| Mode | Pages |
|------|-------|
| tutorial | `getting-started.md` (the single guaranteed linear path) |
| how-to | `how-to/{dependencies-and-hooks,parameters,caching,parallelism,executors,settings-and-dotenv,imports-and-modules,os-filtering}.md` (folded from `guides/`) |
| use-case | `use-cases/{python-project,node-project,mcp-agents}.md` (project-shaped how-tos) |
| reference | `cli.md`, `runefile.md`, `GRAMMAR.md`, `mcp.md` |
| explanation | content within `overview.md` + `user-guide/` narrative (no separate folder this feature) |
| index | `docs/README.md` (intent-first router) |
| stub | old `docs/guides/*` paths → one-line redirect to `how-to/*` |

### UserGuide

The ordered, readable capability tour (`docs/user-guide/`). Connective tissue that **links
into** how-to/reference/explanation rather than duplicating them (FR-008).

| Field | Description |
|-------|-------------|
| reading_order | Curated sequence across modes |
| chapters | Narrative sections, each linking out to the owning how-to/reference page |

### UseCaseWalkthrough

A project-shaped how-to (`docs/use-cases/*`). Specialization of DocPage with:

| Field | Description |
|-------|-------------|
| project_shape | `python` \| `node` \| `mcp-agent` |
| backing_example | `docs/examples/{python-project,node-project,agent-driven}/` |
| features_shown | Rune features paired with the use case (params, caching, deps, executors, MCP) |
| why | Explanation of *why* the example is written the way it is (FR-011) |
| security_notes | MCP page only: read-only default, `[confirm]` gating, env-only secrets (FR-012) |
| expected_output | Copy-run commands with shown output (FR-010) |

### DocsIndex

`docs/README.md` — the navigation surface (FR-013).

| Field | Description |
|-------|-------------|
| intent_table | `\| I want to… \| Go to \|` rows routing by reader goal |
| by_type | Compact section grouping pages by Diátaxis mode |
| invariant | Every published page reachable in ≤2 clicks from here |

### Badge

A README status indicator. Full set + URLs in [`contracts/badges.md`](./contracts/badges.md).

| Field | Description |
|-------|-------------|
| kind | `ci` \| `release` \| `license` \| `go-version` \| `report-card` \| `go-reference` \| `docs` |
| scope | `repo` (→ `glapsfun/rune`) \| `module` (→ `rune-task-runner/rune`) \| `static` |
| image_url / link_url | Per the badge contract |
| alt | Non-empty (graceful degradation, FR-005) |

### RedirectStub

A DocPage (`mode: stub`) left at an old path so external inbound links keep resolving.

| Field | Description |
|-------|-------------|
| old_path | Original location (e.g. `docs/guides/caching.md`) |
| target | New location (e.g. `../how-to/caching.md`) |
| lifetime | Kept for a deprecation window, then removed |

## Relationships

- `DocsIndex` → routes to → every `DocPage` (≤2 clicks).
- `UseCaseWalkthrough` → anchored to → one `Example` (existing `docs/examples/*`).
- `UserGuide` → links into → `how-to` / `reference` / `explanation` pages (no duplication).
- `RedirectStub.old_path` → `DocPage.path` (new location).
- `README` → contains → the `Badge` set; → links into → `DocsIndex`.

## Invariants (verified by `test/docs`)

1. Every internal link resolves (`links_test.go`) — covers reorg + stubs.
2. Every backing `Example` statically validates (Tier A) and runs where the interpreter is
   present (Tier B) (`examples_test.go`).
3. Fenced `rune` blocks on self-contained pages validate (`codeblocks_test.go`
   `selfContainedPages` allowlist, grown to include reorganized/new pages).
4. Every example dir satisfies the example contract README sections (`contract_test.go`).
5. Every `Badge` obeys the badge integrity rules (`badges_test.go`, new).
