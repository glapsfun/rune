# Contract: Example Directory

**Feature**: `004-rich-documentation`

Every runnable example in the library MUST satisfy this contract. The verification harness
(`test/docs`) enforces it; a non-conforming example fails CI.

## Layout

```text
docs/examples/<use-case>/
├── Runefile        # REQUIRED — the runnable task file (shipped DSL only)
├── README.md       # REQUIRED — the example's metadata + walkthrough
└── <support>       # OPTIONAL — minimal extra files the example needs (e.g. a tiny src file)
```

- `<use-case>` is a short kebab-case id (e.g. `caching`, `go-service`, `agent-driven`).
- The directory MUST be self-contained: copying it into an empty project and following its
  README MUST work (FR-006).

## README.md required sections

```markdown
# <Title>

> **Use case:** <one line — who this is for / what it accomplishes>

**Demonstrates:** <capability tags, e.g. caching, dependencies>  ·  **Guide:** <link to docs/guides/...>

**Prerequisites:** <none | python3 | node | docker | an agent CLI>   <!-- FR-008 -->

## Run it

```sh
<run_command>     # e.g. rune build
```

**Expected output**

```text
<expected_output>     # what success looks like — asserted by the harness (FR-006)
```

## How it works

<short explanation of the Runefile, why it's written this way>
```

## Rules (enforced)

| # | Rule | Source |
|---|------|--------|
| E1 | `Runefile` MUST statically validate via `rune --file <path> --list` (exit 0) — parse+analyze, runs nothing. (`check` is **not** a built-in; see harness contract / finding F1.) | FR-006, Principle III/VI |
| E2 | README MUST state `Prerequisites` before any run step | FR-008, SC-010 |
| E3 | README MUST name the capability demonstrated and link to its guide | FR-009 |
| E4 | README MUST show `run_command` and `expected_output` | FR-006 |
| E5 | If `Prerequisites: none`, the example MUST also pass Tier-B run + output assert | D2/D3 |
| E6 | MUST NOT contain any secret/token/key literal | Principle VII |
| E7 | Shell bodies MUST be cross-platform (pure-Go `sh`); OS-specific steps called out inline | FR-018, Principle V |
| E8 | Belongs to exactly one labeled `ExampleGroup` in `docs/examples/README.md` | FR-005 |

## Minimum library coverage (FR-007 / SC-004)

The library MUST contain at least one conforming example for **each** capability and project
shape in the Coverage Matrix (see `data-model.md`). The harness's coverage check fails if any
is missing.
