# Research: Secret Masking & Sanitization

**Feature**: `013-secret-masking` | **Date**: 2026-07-21

All Technical Context unknowns resolved. Each decision below records what was
chosen, why, and what was rejected.

## D1: How secrets are identified

**Decision**: Name-based detection over the task's *effective environment*
(host env + `set dotenv` file + `set export` variables + per-task
`[env(...)]`), plus explicit author declarations. A variable is secret when its
name matches a built-in pattern list (case-insensitive substring match):
`TOKEN`, `SECRET`, `PASSWORD`, `PASSWD`, `APIKEY`, `API_KEY`, `PRIVATE_KEY`,
`ACCESS_KEY`, `CREDENTIAL`, `AUTH` — or when the author lists it in
`set secrets := [...]`. Authors exempt false positives with
`set unmasked := [...]`.

**Rationale**: Mirrors GitHub Actions (`::add-mask::` + registered secrets) and
GitLab (masked variables): the runner masks values it *knows* are secret. The
whole effective env is inspected because session-injected credentials
(e.g. `AWS_SECRET_ACCESS_KEY` from the host) are the most common leak vector
for agents running `env`. `AUTH` is a deliberately aggressive pattern (it can
match `OAUTH_METHOD`); the `unmasked` exemption is the escape hatch, and the
spec's edge-case list accepts over-masking as the safe failure mode.

**Alternatives considered**:
- *Content scanning of output* (entropy/regex credential detection): heuristic,
  false-positive-prone, can never be a guarantee — rejected (recorded in spec
  Assumptions).
- *Mask everything from `.env`*: mangles ordinary config (`GREETING=hello`) —
  rejected.
- *Declared-only, no built-in patterns*: safe-by-default fails for undeclared
  host credentials, which is exactly the agent scenario — rejected.

## D2: Declaration surface — two list-valued settings

**Decision**: Two new file-level settings, following the existing settings
grammar (no expression-language change):

```rune
set secrets := ["DEPLOY_CFG", "CI_UPLOAD_URL"]   # additional names to mask
set unmasked := ["OAUTH_METHOD"]                 # exempt from built-in patterns
```

**Rationale**: Secrets are a property of the *environment*, which is file-level
in Rune (`set dotenv`, `set export`), so a file-level setting matches the
mental model. List-valued settings already exist (`set shell := [...]`) and
`config.ResolveSettings`'s `evalList` handles evaluation and positioned
diagnostics for free (`internal/config/settings.go:39-56`). Adding settings has
direct precedent (feature 010 added `minimum_version`) and does not touch the
frozen expression sublanguage (Constitution III governs expressions — no loops,
recursion; settings are declarations). Registration in
`internal/language/builtin.go` gives `rune analyze`/LSP validation (RUNE2008
for typos) with no extra work.

**Alternatives considered**:
- *Per-task attribute `[secret("NAME")]`*: wrong granularity — a secret leaked
  by one task is equally sensitive in another; authors would have to repeat it —
  rejected.
- *Single setting with negation prefix* (`set secrets := ["A", "!B"]`): cute,
  but invents micro-syntax inside string values that the analyzer can't check —
  rejected.
- *CLI flag / env var (`RUNE_SECRETS`)*: not versioned with the project,
  invisible to reviewers, violates FR-005 — rejected.

## D3: Masking engine — a new `internal/mask` package

**Decision**: A small, dependency-free `internal/mask` package:

- `mask.NewSet(env []string, declared, exempt []string) *Set` — derives the
  secret-value set from `KEY=value` pairs. Applies the pattern list and
  declarations minus exemptions; drops values shorter than the minimum length;
  splits multi-line values into per-line entries (each qualifying line masked
  independently, the GitHub Actions approach to PEM keys).
- `mask.NewWriter(w io.Writer, s *Set) io.WriteCloser` — a streaming wrapper
  that replaces every occurrence of any set entry with the placeholder before
  bytes reach `w`. Mutex-protected: parallel tasks share the engine's writers
  today, so the wrapper must tolerate concurrent `Write`s.

**Rationale**: Constitution IV mandates small, focused `internal/` packages.
Masking is a self-contained text-transformation concern with its own unit-test
surface (chunk boundaries, overlaps, concurrency) that doesn't belong in
`internal/cli`. Pure stdlib keeps Principle V intact.

**Alternatives considered**:
- *Aho-Corasick multi-pattern automaton*: optimal asymptotics, but secret sets
  are tiny (typically < 20 values); per-chunk `bytes` scanning meets SC-004
  (10 MB within 10%) without a new algorithmic surface — rejected as
  optimization without a profile (Constitution VIII).
- *Post-processing in `mcpserver.formatResult` only*: leaves the terminal,
  echo, and status surfaces unmasked and violates FR-002/FR-003 — rejected.
- *`strings.NewReplacer` per chunk*: no chunk-boundary handling — rejected.

## D4: Chunk-boundary (streaming) correctness

**Decision**: The writer keeps a bounded *carry*: after replacing all complete
occurrences in `carry+chunk`, it retains the longest suffix that is a proper
prefix of any secret entry (bound: longest entry − 1 bytes) and emits the rest.
`Close`/`Flush` emits the carry verbatim (an incomplete prefix is, by
definition, not the secret — **provided the stream has truly ended**).

**Flush safety rule**: because parallel tasks share the engine's writers, a
flush while any task may still be writing could emit another task's in-flight
secret *prefix* verbatim — and its continuation would then never match an
entry, leaking the whole value. Therefore the engine flushes a wrapper only at
points where no producer can still be writing to it: after a body line's
process has exited **when no `[parallel]` group is in flight**, and always at
run end (after the scheduler has joined every task). There is no timer- or
line-count-based flushing.

**Rationale**: This is the only way to satisfy FR-003/FR-004 (no observable
window of unmasked output, values split across buffers) while remaining
streaming. The hold-back is bounded and only engages when the stream tail
actually looks like the start of a secret, so interactive output (spinners,
prompts) is unaffected in practice. A task interrupted mid-stream has only ever
had masked bytes emitted — masking is at emission time, not post-processing.
Per-line entries for multi-line values (D3) also keep the carry bound small:
`maxEntryLen` is derived from single-line entries, never a whole PEM-sized
value, so the writer never withholds large output windows.

**Alternatives considered**:
- *Line-buffering everything*: breaks tasks that emit progress without
  newlines (carriage-return spinners, prompts) — rejected.
- *Time-based carry flush*: adds a goroutine + timer per writer for a
  pathological case (stream ends on a secret prefix and stalls); bounded
  hold-back is acceptable and documented — rejected for v1.
- *Unconditional flush after every body line*: unsafe under `[parallel]` (the
  leak scenario above) — rejected in favor of the flush safety rule.

## D5: Placeholder, minimum length, and installation point

**Decision**:
- Placeholder: `***` (fixed, all surfaces).
- Minimum maskable value length: **4 bytes**; shorter values are never
  value-masked (documented, per FR-007).
- Installation: the secret set is built once per run right after
  `buildEnv` (`internal/cli/run.go:594`) plus the union of every task's
  `[env(...)]` values, and — when non-empty — `Options.Stdout`/`Options.Stderr`
  are wrapped before the `engine` is constructed. Both entry points get this at
  their shared choke point: the CLI path (`execute`, `run.go:113`) and the MCP
  adapter (`internal/cli/mcp.go:57-91`, which builds the same engine writing
  into buffers — so MCP `Result` strings and `formatResult` output
  (`mcpserver/handler.go:63-79`) are masked by construction). The agent
  executor's buffered write-back (`internal/cli/agentexec.go:53`) and shell
  command echo (`internal/runtime/shell/shell.go:101-107`) both write through
  the engine's writers, so they inherit masking with no executor changes.

**Rationale**: One wrap point covers every surface in FR-002 because the
architecture already funnels all output through `Options.Stdout/Stderr`
(verified across `run.go`, `shell.go`, `interp.go`, `agentexec.go`, `mcp.go`).
When the set is empty the writers are not wrapped at all, making FR-008's
byte-identical guarantee structural rather than tested-for. `***` matches the
convention agents and humans already know from GitHub Actions. Minimum length 4
balances FR-007 (a 1-char value like `PORT_AUTH=1` would mangle everything)
against masking realistic short PINs.

**Alternatives considered**:
- *Named placeholder `[masked:API_TOKEN]`*: leaks which variable held the value
  and (via repeated distinct labels) how many secrets exist; spec assumption
  already defers this — rejected.
- *Wrapping inside each executor*: three wrap points instead of one, misses
  Rune's own status lines (`running:`, warnings, diagnostics at
  `run.go:252-271,777`) — rejected.
- *Minimum length 8 (GitLab)*: silently refuses to mask real 6-char secrets;
  4 is the safer default for an agent-facing tool — rejected.

## D6: Scope boundary — verbatim occurrences only

**Decision**: v1 masks verbatim byte occurrences of each secret value (plus the
per-line entries of multi-line values). Transformed occurrences (base64,
URL-encoding, JSON escaping, task-side splitting) are documented as out of
scope in the user-facing docs and the contract.

**Rationale**: Matches the spec's guarantee boundary. Chasing encodings is an
arms race with diminishing returns (GitHub Actions masks a couple of encodings
and still documents the same caveat); a clear documented boundary beats an
implied-but-leaky stronger one.

## D7: Verification strategy

**Decision**: Four layers, matching Constitution VI and existing harnesses:
1. **Unit** (`internal/mask`): table-driven tests for set derivation (patterns,
   declarations, exemptions, min-length, multi-line split) and writer semantics
   (chunk-split secrets, overlapping values, concurrent writes, flush).
2. **Integration** (`test/integration/secret_masking_test.go`, new, following
   the `us*_test.go` harness): binary-level runs asserting stdout/stderr/exit —
   env-dump masking, echoed-command masking, `set secrets`/`set unmasked`,
   stderr-of-failing-task masking, no-secret byte-identical run.
3. **MCP end-to-end**: extend the existing stdio MCP test pattern
   (`runWithStdin`) to call a secret-printing task and assert the tool result
   contains `***` and never the value (SC-002).
4. **Docs as fixtures** (`test/docs` harness): a new runnable example
   `docs/examples/secret-masking/` and grammar/doc updates validated by
   `docs-verify`.

Golden corpus: `testdata/corpus/full.rune` gains the two settings so grammar
drift is caught; existing goldens stay untouched (empty-set path is unwrapped).

## D8: Documentation set (Constitution constraint: surface changes carry docs)

**Decision**: Same-PR updates to `docs/GRAMMAR.md` (two settings),
`docs/runefile.md` (settings table), a new how-to
`docs/how-to/secret-masking.md` (patterns list, placeholder, min length,
guarantee boundary — satisfies FR-010), a note in `docs/mcp.md` (agent story),
a pitfall update in `docs/how-to/settings-and-dotenv.md`, and the runnable
example directory.
