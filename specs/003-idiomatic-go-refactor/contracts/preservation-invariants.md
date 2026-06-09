# Contract — Preservation Invariants

The "don't break these" contract. Every refactor (US2/US3/US5) MUST hold these invariants;
they are what makes the work *externally behavior-preserving* (SC-004/SC-009, FR-009/FR-014).
The existing test suite is the enforcement mechanism.

## Invariants (MUST NOT change)

| # | Invariant | Enforced by |
|---|-----------|-------------|
| INV-1 | **Runefile semantics** — the same Runefile parses, analyzes, and runs identically (tasks, deps, params, attributes, executors, caching, dotenv, expressions) | corpus (`test/corpus`), parser/lexer goldens, integration tests |
| INV-2 | **CLI contract** — flags, subcommands, output streams, and **exit codes** (0/1/2/3/130) are unchanged | `test/integration` (binary-level stdout/stderr/exit), `internal/cli` tests |
| INV-3 | **Diagnostics** — `file:line:col` + caret spans render identically (Principle II) | `internal/diag` golden (`testdata/diag/render.golden`) |
| INV-4 | **Formatter output** — `rune --fmt` canonical output unchanged | `internal/cli` fmt golden (`testdata/fmt/*.fmt`) |
| INV-5 | **AST / token streams** — golden dumps unchanged (the `eval` builtins refactor must not alter parse/eval output) | `testdata/{lexer,parser,corpus}` goldens |
| INV-6 | **MCP tool contract** — task→tool mapping, input schemas, `DestructiveHint`/`openWorldHint` annotations, and the secure-by-default posture (read-only default, env-only secrets, destructive-gated, remote opt-in/localhost/token) | `mcpserver` (`server_test`, `authz_test`, `transport_test`), `internal/cli/mcp_test` |
| INV-7 | **Public `mcpserver` API** — exported types/functions remain backward-compatible | compilation of `mcpserver` consumers + its tests |
| INV-8 | **Single static binary** — shipped artifact stays `CGO_ENABLED=0`/static (Principle V) | CI `build` job (002) |

## Acceptance (how "preserved" is proven)

- `docker-compose run --rm test go test ./...` is **green**, and `git status --porcelain
  testdata/` is **empty** — i.e. **no golden regenerated** to accommodate a change (SC-004).
- `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...` is green, 0 races
  (SC-003).
- The `mcpserver` + `internal/cli` MCP/contract tests pass **unmodified** (SC-009, INV-6/7).

## Allowed to change (explicitly)

- Internal, unexported APIs and helper signatures.
- Within-package file organization (e.g. splitting `internal/cli`).
- Removal of `init()`/globals, error-surfacing improvements that don't alter success-path
  output, naming of unexported identifiers, added doc comments, added benchmarks/goleak tests.

## Special cases

- The `mcpserver/transport.go` server-error fix changes only whether a previously-**lost**
  error is **surfaced** on the failure path — it MUST NOT change success-path behavior or the
  shutdown contract (INV-6/INV-7).
- The `cli/run.go` cache-store fix may **log** a previously-silent failure; it MUST NOT change
  caching semantics or task success/exit behavior (INV-1/INV-2, Principle I).
