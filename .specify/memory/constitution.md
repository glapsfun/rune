<!--
Sync Impact Report
- Version: (unversioned template) → 1.0.0
- Rationale: Initial ratification. Codifies the eight principles already enforced
  across specs/ (canonical source: specs/001-rune-task-runner/plan.md Constitution
  Check table, Principles I–VII) and the Go engineering discipline cited in
  CONTRIBUTING.md and .golangci.yml (Principle VIII).
- Principles: I Command Runner, Not a Build System · II Errors Are a Feature ·
  III Minimal, Total DSL · IV Hand-Written Front End, Idiomatic Go ·
  V Boringly Portable · VI Test-First, Multi-Layer Verification ·
  VII AI-Native, Secure by Default · VIII Go Engineering Discipline
- Added sections: Engineering Constraints; Quality Gates; Governance.
- Templates requiring review: .specify/templates/plan-template.md (Constitution
  Check gate references these principles).
- Follow-ups: none.
-->

# Rune Constitution

Rune is a small, portable, AI-native command runner. These principles are locked
design commitments, not negotiable trade-offs. They govern what Rune is and how it
is built; every spec, plan, and change is checked against them.

## Core Principles

### I. Command Runner, Not a Build System
Tasks always run by default. Caching is per-task and opt-in via
`[cache(inputs=…, outputs=…)]`; there is no implicit timestamp-based skipping, and
every cache hit is logged to stderr so the decision is visible. Rune orchestrates
commands — it does not silently decide work is unnecessary.

### II. Errors Are a Feature
Every statically detectable error is reported with `file:line:col` plus a caret
span pointing at the offending source. The analyzer runs before any execution, so
a Runefile with an unknown task, undefined variable, dependency cycle, or arity
mismatch fails with zero side effects and a non-zero exit. A good diagnostic is
part of the product, not an afterthought.

### III. Minimal, Total DSL
The expression sublanguage has no loops and no recursion, and every program
terminates. Real logic lives in task bodies (shell, Python, Node), never in the
expression language. The DSL surface is intentionally small and frozen; growing it
requires amending this constitution, not an incremental feature.

### IV. Hand-Written Front End, Idiomatic Go
The lexer (Rob Pike state-function style), recursive-descent parser, and Pratt
expression parser are hand-written — no `goyacc`, ANTLR, or generated/prototype
parser code ships. Engine logic lives in small, focused `internal/` packages
(`token`, `lexer`, `ast`, `parser`, `analyzer`, `diag`, `eval`, `runtime/…`, …);
`mcpserver/` stays public so it is embeddable. This package layout is locked.

### V. Boringly Portable
Rune is pure Go, built with `CGO_ENABLED=0`, and ships as a single static binary
on Linux, macOS, and Windows. The default shell executor is `mvdan.cc/sh/v3`, never
the system `/bin/sh`, so behaviour does not depend on the host shell. Paths in docs
and examples use forward slashes. No WSL or Git-Bash workarounds.

### VI. Test-First, Multi-Layer Verification
Development is test-first (Red-Green-Refactor). Verification is layered: golden
files (compared by default, regenerated deliberately — never hand-edited to make a
test pass), binary-level integration tests asserting stdout/stderr/exit, fuzz
targets for the lexer and parser, and a compatibility corpus that fails on silent
grammar drift. Documentation is a tested fixture: examples are run and links are
checked. CI runs on Linux, macOS, and Windows.

### VII. AI-Native, Secure by Default
MCP (Model Context Protocol) is first-class: tasks are exposed as tools. The agent
surface is read-only by default; destructiveness is author-declared via the
`[confirm]` attribute, which maps to the MCP destructive hint. Secrets come only
from the environment or the agent CLI's own session — never from a Runefile, and
never surfaced in task descriptions, schemas, or listings. Any remote MCP endpoint
is opt-in, localhost-bound, and token-gated.

### VIII. Go Engineering Discipline
Go code is idiomatic and conforms to the project's bundled Go skills. Errors are
handled explicitly and wrapped with `%w` (`errors.Is`/`errors.As` over `==` or type
switches). Every goroutine has a clear owner and a clear exit. Constructors are
preferred over `init()` and package globals. Every external call carries a context
and a timeout. Optimizations are not made without a profile. `golangci-lint run`
must report zero issues, and code must be clean under gofumpt and goimports.

## Engineering Constraints

- **Docker-only testing.** The Go test suite runs inside the `Dockerfile.test` /
  `docker-compose.yml` harness, never directly on the host.
- **Locked package layout.** The `internal/` engine packages and the public
  `mcpserver/` package follow the structure in Principle IV; structural changes
  require justification.
- **Backward compatibility.** Breaking DSL changes are opt-in per file — existing
  Runefiles keep working unless they explicitly request new behaviour.
- **Surface changes carry their docs.** Any change to DSL surface ships with an
  updated `docs/GRAMMAR.md` and the matching golden/integration fixtures in the
  same PR.

## Quality Gates

Every push and pull request runs the gate set in `.github/workflows/ci.yml`; all
must pass to merge:

1. **lint** — `golangci-lint run` is clean (covers gofmt/gofumpt/goimports and
   `go vet`).
2. **test** — the full suite with the race detector (`-race`) on Linux, macOS, and
   Windows.
3. **build** — the static, CGO-free binary compiles on all three OSes.
4. **golden** — committed golden files match a fresh regeneration.
5. **fuzz-smoke** — the lexer and parser fuzz targets build and run briefly.
6. **docs-verify** — the `test/docs` harness passes: examples validate and run,
   code blocks validate, and internal links resolve.
7. **release-dryrun** — `goreleaser release --snapshot` succeeds.

## Governance

This constitution supersedes ad-hoc practice. Every feature spec's `plan.md`
includes a Constitution Check that gates the work against these principles, and any
justified deviation is recorded in that plan's Complexity Tracking table.

Amendments require editing this file and any dependent spec-kit templates (e.g.
`.specify/templates/plan-template.md`), with a version bump following semantic
versioning: MAJOR for removing or redefining a principle, MINOR for adding a
principle or section, PATCH for clarifications. Reviewers verify that changes
comply with the principles in force.

**Version**: 1.0.0 | **Ratified**: 2026-06-10 | **Last Amended**: 2026-06-10
