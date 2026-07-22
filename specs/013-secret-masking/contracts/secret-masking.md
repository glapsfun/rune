# Contract: Secret Masking & Sanitization

**Feature**: `013-secret-masking` | **Phase**: 1

This contract has three parts: the Runefile grammar surface (§1), the
observable masking behavior on every output surface (§2), and the MCP
guarantee (§3). It extends — never replaces — the base MCP contract in
`specs/001-rune-task-runner/` ("secrets never appear in any tool name,
description, schema, or result") by closing the *result* half for task output.

## 1. Grammar surface (docs/GRAMMAR.md additions)

Two new file-level settings, standard settings grammar (no new tokens, no
expression-language change):

```rune
set secrets := ["DEPLOY_CFG", "UPLOAD_URL"]
set unmasked := ["OAUTH_METHOD"]
```

| Setting | Form | Meaning |
|---|---|---|
| `secrets` | list of strings | Additional environment-variable *names* whose values are masked. |
| `unmasked` | list of strings | Names exempted from the built-in sensitive-name patterns. |

**Static guarantees** (Constitution II):
- Malformed values (non-string, undefined variable reference) → positioned
  diagnostic `file:line:col` + caret, exit 3, zero side effects.
- Unknown setting name (typo, e.g. `set secert`) → RUNE2008 via
  `rune analyze` / LSP.
- Duplicate `set secrets` (or `set unmasked`) lines → existing
  duplicate-setting diagnostic.
- Same name in both lists → error diagnostic citing both spans, reported
  before any execution (exit 3, zero side effects; enforced during settings
  resolution on the run path — `rune analyze` surfacing is a non-v1 follow-up).
- A listed name absent from the run-time environment is inert: no error, no
  mask entry.

`rune fmt` formats both settings exactly like other list settings.

## 2. Masking behavior contract

### 2.1 Secret identification

The mask set for a run is derived from the effective environment (host env +
`set dotenv` file + `set export` variables + all tasks' `[env("K","V")]`
pairs):

- **Pattern rule**: value of any variable whose name case-insensitively
  contains one of: `TOKEN`, `SECRET`, `PASSWORD`, `PASSWD`, `APIKEY`,
  `API_KEY`, `PRIVATE_KEY`, `ACCESS_KEY`, `CREDENTIAL`, `AUTH` — unless the
  name is listed in `unmasked`.
- **Declaration rule**: value of any variable named in `secrets`
  (unconditionally; `unmasked` does not override an explicit declaration —
  that combination is a static error).
- **Minimum length**: values (or lines of multi-line values) shorter than
  **4 bytes** are never value-masked.
- **Multi-line values**: each line of length ≥ 4 is masked independently
  wherever it appears (a verbatim multi-line reproduction is thereby fully
  masked line by line; the whole value is not tracked as a single unit).

### 2.2 Replacement

- Every verbatim occurrence of every mask entry is replaced with exactly
  `***` (three asterisks), on every surface, in both terminal and
  agent-facing output.
- Matching is leftmost-longest; overlapping/nested entries never expose a
  fragment of any entry.
- Occurrences split across streaming buffer boundaries are still masked
  (bounded carry; no unmasked window is ever observable, including when a task
  is interrupted or times out mid-stream).

### 2.3 Surfaces (all-or-nothing)

Masking applies uniformly to:

1. Task standard output and standard error (shell, python/node/custom
   interpreter, and agent executors).
2. Echoed command lines (including interpolated `{{var}}` text; `@`-prefix and
   `set quiet` suppression semantics are unchanged).
3. Rune's own emissions: `running:`/`cached:`/dry-run lines, warnings,
   confirmation prompts, error reports and rendered diagnostics produced
   during execution.
4. MCP tool results (§3).

### 2.4 Non-goals (documented boundary)

- The masked process environment is **never** altered: tasks receive real
  values (FR-011); only Rune's *emissions* are transformed.
- Transformed occurrences (base64, URL-encoding, JSON-escaping, task-side
  splitting or interleaving) are **not** masked in v1.
- No content-based detection: a credential in a variable with an innocent name
  and no declaration is not masked.
- Values < 4 bytes are not masked.

### 2.5 Invariance guarantee

For a Runefile + environment yielding an **empty** mask set, every byte of
output on every surface is identical to the previous release (verified against
the existing golden corpus and styling tests).

### 2.6 Defaults and overrides

- Masking is **always on**; there is no flag, setting, or environment variable
  that disables it globally, and nothing on the agent-facing surface can relax
  it. The only opt-out is per-variable `set unmasked`.

## 3. MCP contract extension

Extends `specs/007-interactive-tui/`-era MCP behavior (tool-per-task):

- The text content of every tool result (stdout section, stderr section, and
  `[exit N]` trailer) is masked per §2 **before** it is placed in the MCP
  response. No transport (stdio or HTTP) can observe unmasked task output.
- Tool names, descriptions, and input schemas remain secret-free (existing
  contract, unchanged).
- The masking rules an agent observes are byte-identical to what a terminal
  user sees for the same task and environment.

## 4. Acceptance mapping

| Contract clause | Spec requirement | Verified by |
|---|---|---|
| §1 static guarantees | FR-005, FR-009 | analyzer/integration tests (exit 3, caret spans) |
| §2.1 identification | FR-001, FR-007 | `internal/mask` unit tests |
| §2.2 replacement | FR-002, FR-004 | unit tests (chunk-split, overlap) + integration env-dump test |
| §2.3 surfaces | FR-002, FR-003 | integration tests per surface; interrupted-task test |
| §2.4 non-goals | FR-010, FR-011 | docs (`docs/how-to/secret-masking.md`) + docs-verify |
| §2.5 invariance | FR-008 / SC-003 | golden corpus + styling tests unchanged |
| §2.6 defaults | FR-006 | integration test: no disable path exists |
| §3 MCP | SC-002 | stdio MCP end-to-end test |
