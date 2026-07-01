# Contract: Documentation Structure, Page Shape & Navigation

The interface this feature exposes to readers (and to the `test/docs` harness) is the docs
tree, per-page shape, and navigation. This contract is what verification asserts and what
`/speckit-tasks` decomposes against.

## C1. Directory layout (after reorganize-in-place)

```text
docs/
├── README.md                 # DocsIndex — intent-first router (index mode)
├── overview.md               # explanation (kept)
├── getting-started.md        # tutorial (kept — the one true tutorial)
├── how-to/                   # how-to guides (folded from guides/)
│   ├── dependencies-and-hooks.md
│   ├── parameters.md
│   ├── caching.md
│   ├── parallelism.md
│   ├── executors.md
│   ├── settings-and-dotenv.md
│   ├── imports-and-modules.md
│   └── os-filtering.md
├── user-guide/
│   └── README.md             # ordered tour that links out (no duplication)
├── use-cases/
│   ├── python-project.md
│   ├── node-project.md
│   └── mcp-agents.md
├── cli.md · runefile.md · GRAMMAR.md · mcp.md   # reference (kept)
├── docker.md · installation.md · releasing.md · troubleshooting.md  # kept
├── guides/                   # redirect stubs only (→ how-to/*), incl. README.md → how-to/
└── examples/                 # existing 15-example library (index refreshed; content intact)
```

## C2. Per-page shape (every non-stub page)

1. **H1 title** — one, matching the page's purpose.
2. **Intro** — a short lead paragraph stating what the page is for.
3. **Manual anchor TOC** — required only on long reference pages (`cli.md`, `runefile.md`,
   `GRAMMAR.md`); other pages rely on GitHub's auto TOC.
4. **Body** — every snippet whose output is meaningful is followed by a fenced `text` output
   block (long output may be wrapped in `<details>` with a blank line after `<summary>`).
5. **Callouts** — pitfalls/notes use GitHub Alerts `> [!NOTE|TIP|IMPORTANT|WARNING|CAUTION]`,
   ≤2 per page, never nested or stacked.
6. **Next-steps footer** — a `---` rule then a link block routing onward; no dead-ends.
7. **Relative links only** — never absolute `https://github.com/...` for in-repo targets.

## C3. Use-case walkthrough shape (`use-cases/*`)

- States the **project shape** and that the reader brings their own project (it's a how-to).
- **Anchored** to a backing example (`docs/examples/{python-project,node-project,agent-driven}`).
- Shows copy-run commands **with expected output**.
- Names the **Rune features** used and **why** the example is written that way.
- MCP page additionally states the security posture (read-only default, `[confirm]` gating,
  env-only secrets) at the relevant point.

## C4. DocsIndex shape (`docs/README.md`)

- An `| I want to… | Go to |` intent table covering: first task → tutorial; capability goals →
  how-to; project shapes → use-cases; lookups → reference.
- A compact "by document type" section (Diátaxis modes).
- **Invariant**: every published page reachable in ≤2 clicks (FR-013 / SC-007).

## C5. Redirect-stub shape (old `docs/guides/*`)

```markdown
# Moved

> [!NOTE]
> This page moved to [How-to: <title>](../how-to/<page>.md).
```

## C6. README shape (front door)

- Light header refresh: centered H1 title + tagline, then the centered badge row
  (see [`badges.md`](./badges.md)), then quick-nav links into `docs/README.md`.
- Existing prose and tables kept; no image/logo assets.

## C7. Verification mapping (`test/docs`, run via `rune docs-check`)

| Contract clause | Enforced by |
|-----------------|-------------|
| C1 layout, relative links, stubs resolve | `links_test.go` (scans `docs/**/*.md` + `README.md`) |
| C2 fenced `rune` blocks validate | `codeblocks_test.go` (`selfContainedPages` allowlist) |
| C3 backing examples validate + run | `examples_test.go` (Tier A always; Tier B when runtime present) |
| C3 example README sections | `contract_test.go` |
| C6 badge integrity | `badges_test.go` (new — see badges contract) |
| No secret literals leak | existing harness secret-scan |

**Non-regression**: no change to CLI behavior or golden output — `cli_test.go` and the wider
golden suite pass unchanged (SC-008).
