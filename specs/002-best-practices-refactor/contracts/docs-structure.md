# Contract — Documentation Structure

The documentation "interface" to users: the information architecture and the required contents
of each page. Realizes FR-001…FR-006, SC-001, SC-010. The acceptance bar is **coverage** (every
command/construct documented with a runnable example) and **freshness** (examples are executed).

## Information architecture

```text
README.md                 # entry point — pitch, install one-liner, runnable quickstart, links
docs/
  installation.md         # binary download, build-from-source, container — per OS
  getting-started.md      # zero → first task; the example Runefile CI/`docs-check` runs
  cli.md                  # every command + every global flag, each with an example
  runefile.md             # language usage: tasks, deps, params, [cache], executors, dotenv, fmt
  mcp.md                  # expose tasks to agents; secure-by-default; stdio + HTTP transports
  docker.md               # run via the official image; mounts; documented limitations
  GRAMMAR.md              # (existing) formal grammar — authoritative; runefile.md links here
```

## Required contents per page

| Page | MUST contain | Source of truth |
|------|--------------|-----------------|
| `README.md` | one-sentence "what is Rune", who it's for, comparison to make/just, install one-liner, a copy-paste quickstart, links to every `docs/` page (FR-001) | spec Overview; this contract |
| `installation.md` | instructions for Linux/macOS/Windows via: prebuilt binary, `go install`/source, container (FR-002) | `release-pipeline.md`, `docker-image.md` |
| `getting-started.md` | a minimal copy-pasteable `Runefile` + exact command + expected output; reaches a running task in <10 min (FR-003, SC-001) | `cmd/rune/main.go`, feature-001 `quickstart.md` |
| `cli.md` | every command (`mcp`/`serve`, `completion`) and every global flag (`-f/--file`, `--list`, `--dry-run`, `--summary`, `--dump`, `--format`, `--set`, `--watch`, `--choose`, `--yes`, `--quiet`, `--fmt`, `--clear-cache`), each with ≥1 example (FR-004, SC-010) | `cmd/rune/main.go`, feature-001 `contracts/cli.md` |
| `runefile.md` | usage for every construct — task defs, dependencies, parameters/arity, `[cache(inputs,outputs)]`, executors (sh/python/node/agent), dotenv, `[confirm]`, `--fmt`; links to GRAMMAR.md (FR-004, SC-010) | `docs/GRAMMAR.md`, feature-001 `contracts/grammar.md` |
| `mcp.md` | how non-private tasks become MCP tools; **secure-by-default**: read-only default, env-only secrets, destructive (`[confirm]`) opt-in, remote endpoint opt-in/localhost/token-gated (FR-005, Principle VII) | feature-001 `contracts/mcp-tools.md` |
| `docker.md` | `docker run -v "$PWD":/work …` usage, arg passing, and the minimal-image limitations + workarounds (FR-018, FR-019) | `docker-image.md` |

## Coverage & freshness rules

- **SC-010 coverage**: a doc is complete only when every CLI command, every global flag, and
  every Runefile construct appears with ≥1 runnable example. Cross-checked against
  `cmd/rune/main.go` flag registrations and `docs/GRAMMAR.md` constructs.
- **FR-006 freshness**: the `getting-started.md` example Runefile is executed in CI (or via the
  dogfooded `Runefile docs-check` task) and its output asserted, so a CLI/DSL change that breaks
  an example fails the build (spec edge case).
- **Portability**: all path examples use forward slashes (Principle V) so they hold on Windows.
- **No secrets**: examples never embed real secrets; secret usage is shown via env vars only
  (Principle VII).

## Non-goals

- A hosted/generated docs website or versioned docs (out of scope — in-repo Markdown only).
- Duplicating the formal grammar (lives in `GRAMMAR.md`; `runefile.md` links it).
- API/godoc reference generation (separate concern; not required this iteration).
