# Data Model: Secret Masking & Sanitization

**Feature**: `013-secret-masking` | **Phase**: 1 | **Sources**: [spec.md](spec.md) Key Entities, [research.md](research.md) D1–D6

No persistence is involved; every entity below is derived per run and held in
memory only.

## Entities

### SecretDeclaration (author-facing, in the Runefile)

Two list-valued settings, parsed with the existing settings grammar
(`ast.Setting`, list form) — no new AST node kinds.

| Field | Source | Type | Rules |
|---|---|---|---|
| `Secrets` | `set secrets := ["NAME", ...]` | `[]string` | Each element evaluates to a variable *name*. Duplicate `set secrets` lines are rejected by the existing duplicate-setting check. A name that never exists at run time is inert (no error, no mask — spec US3/AS3). |
| `Unmasked` | `set unmasked := ["NAME", ...]` | `[]string` | Same shape. Exempts a name from built-in pattern matching. `unmasked` beats patterns; it does **not** beat an explicit `secrets` entry (declaring both is contradictory — `config.ResolveSettings` emits a positioned error before any task runs, exit 3). |

Carried on `config.Settings` as two new fields, resolved in
`ResolveSettings` via the existing `evalList` (positioned diagnostics for
malformed values come free). Registered in `internal/language/builtin.go` so
unknown-name typos surface as RUNE2008 through `rune analyze`/LSP.

**Validation rules** (FR-009):
- Elements must evaluate to non-empty strings → existing evaluator diagnostics.
- Same name in both `secrets` and `unmasked` → error diagnostic with both spans
  (primary + related location), emitted by `config.ResolveSettings` on the
  execution path (pre-execution, zero side effects, exit 3). Surfacing the same
  conflict through `rune analyze`/LSP is a nice-to-have follow-up, not v1 scope
  (the analysis service runs name-registry checks, not settings resolution).

### SensitiveNamePatterns (built-in, constant)

Case-insensitive substring matches applied to variable names:

```
TOKEN · SECRET · PASSWORD · PASSWD · APIKEY · API_KEY ·
PRIVATE_KEY · ACCESS_KEY · CREDENTIAL · AUTH
```

Fixed at compile time in `internal/mask`; documented user-facing (FR-010).
Extending the list is a code change, not configuration — authors add via
`set secrets`.

### SecretValueSet (`mask.Set`)

The per-run collection of byte strings to mask.

**Derivation** (order matters):

1. Collect the effective environment: `buildEnv` output (host `os.Environ()` +
   `set dotenv` file + `set export` module variables) **plus** the union of all
   tasks' `[env("K","V")]` pairs (the set is engine-wide; writers are shared).
2. For each `KEY=value` pair: mark as candidate if `KEY` matches a
   sensitive-name pattern **and** `KEY ∉ Unmasked`, or if `KEY ∈ Secrets`.
3. For each candidate value: split on newlines; each resulting line with
   length ≥ **MinLen (4 bytes)** becomes one mask entry. Whole multi-line
   values are **not** stored as entries: masking every constituent line already
   masks any verbatim multi-line reproduction, and keeping only line-sized
   entries keeps `maxEntryLen` (and therefore the writer's carry bound) small —
   a PEM-sized whole-value entry would let the writer withhold kilobytes of
   output.
4. Deduplicate entries; drop empty entries. Record `maxEntryLen` for the
   writer's carry bound.

**Invariants**:
- Empty set ⇒ no writer wrapping anywhere (FR-008 byte-identical path).
- Entries are value bytes only; variable names are never stored in output or
  placeholders.
- The set is immutable after construction (safe for concurrent readers).

### MaskWriter (`mask.Writer`)

Streaming `io.WriteCloser` wrapping a destination writer.

| Field | Type | Purpose |
|---|---|---|
| `dst` | `io.Writer` | terminal stream / MCP buffer |
| `set` | `*mask.Set` | entries + `maxEntryLen` |
| `carry` | `[]byte` | longest tail of the stream so far that is a proper prefix of ≥1 entry; bound: `maxEntryLen − 1` |
| `mu` | `sync.Mutex` | parallel tasks share the engine writers |

**State transitions**:

```
Write(p):  scan carry+p → replace all complete entry occurrences with "***"
           → retain new carry (longest suffix = proper prefix of an entry)
           → forward the rest to dst
Flush():   forward carry verbatim (an incomplete prefix is not a secret), clear
Close():   Flush, then close dst if closable
```

Flush points (safety rule, research.md D4): only where no producer can still
be writing — after a body line's process exits *when no `[parallel]` group is
in flight*, and at run end after the scheduler has joined every task. Flushing
while a parallel task streams could emit that task's in-flight secret prefix
verbatim and de-sync matching for the remainder — the one way this design
could violate FR-003. An interrupt mid-stream loses at most the carry — it
never reveals unmasked secret bytes (FR-003).

**Overlap rule**: matching is leftmost-longest across all entries; replaced
regions are not rescanned, and scanning resumes at the end of the replacement —
one secret contained in another can never expose a fragment of either (spec
edge case).

### MaskPlaceholder (constant)

`***` — identical on every surface (task output, echo, status lines, MCP
results). Carries no information about which variable matched or how long the
value was.

### OutputSurface (existing, mapping only)

No new type. The surfaces from spec FR-002 map onto existing writer plumbing:

| Surface | Existing sink | Covered by |
|---|---|---|
| Task stdout/stderr (shell, interp) | `shell.Options` / `interp.Options` copies of `Options.Stdout/Stderr` | wrapping `Options` writers at engine construction |
| Echoed command lines | `shell.Run` → `stderr` | same (echo writes through the wrapped stderr) |
| Agent executor final text | buffered, then `fmt.Fprint(e.opts.Stdout, …)` | same |
| Rune status/log/warning/diag lines | `fmt.Fprint*(e.opts.Stderr, …)` in `run.go` | same |
| MCP tool results | MCP adapter engine writes into `bytes.Buffer`s → `formatResult` | same wrap on the adapter's engine — buffers only ever contain masked text |

## Relationships

```
ast.Setting ──resolve──▶ config.Settings{Secrets, Unmasked}
                              │
buildEnv + ∪ task [env] ──────┼──▶ mask.NewSet(env, secrets, unmasked)
                              │          │ (empty ⇒ no wrapping)
                              ▼          ▼
                    engine Options{Stdout,Stderr} ──▶ mask.NewWriter(…)
                                                          │
                     every emission on every surface ─────┘
```
