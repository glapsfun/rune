# Contract: Diagnostic Code Catalog (RUNE####)

Per the 2026-07-10 clarification and FR-010, these codes are a **stable public contract**. Each listed condition maps to exactly its code. Codes are printed by `rune analyze`, sent to editors via `publishDiagnostics`, documented, and asserted exactly by golden tests. A code's meaning MUST NOT change once shipped.

## Parser diagnostics (RUNE1xxx) — severity: error

| Code | Condition |
|------|-----------|
| `RUNE1001` | Unexpected token |
| `RUNE1002` | Invalid indentation |
| `RUNE1003` | Unterminated string |
| `RUNE1004` | Incomplete expression |
| `RUNE1005` | Malformed task declaration |

## Semantic diagnostics (RUNE2xxx) — severity: error (except RUNE2010)

| Code | Condition | Related locations |
|------|-----------|-------------------|
| `RUNE2001` | Unknown dependency | — |
| `RUNE2002` | Duplicate task | first declaration |
| `RUNE2003` | Dependency cycle | **every task in the cycle** |
| `RUNE2004` | Undefined variable | — |
| `RUNE2005` | Wrong argument count | task declaration |
| `RUNE2006` | Duplicate parameter | first parameter |
| `RUNE2007` | Invalid attribute | — |
| `RUNE2008` | Invalid setting | — |
| `RUNE2009` | Invalid executor | — |
| `RUNE2010` | **Public task lacks documentation** — severity: **warning** (FR-008a); never causes exit 3 | task declaration |

## Project diagnostics (RUNE3xxx) — severity: error

| Code | Condition | Related locations |
|------|-----------|-------------------|
| `RUNE3001` | Unresolved import | — |
| `RUNE3002` | Import cycle | **every file in the cycle** |
| `RUNE3003` | Duplicate imported namespace | conflicting import(s) |
| `RUNE3004` | Incompatible Rune version | — |

## Contract rules

- Every emitted `diag.Diagnostic` for a condition above MUST carry the exact `Code`.
- `RUNE2003` and `RUNE3002` MUST populate `Related` with every participating node (FR-009). Example message: `dependency cycle detected: build → test → generate → build`.
- `RUNE2010` is warning-severity and excluded from the error-gating that produces `rune analyze` exit code 3 (FR-025).
- The one apparent mismatch in the source description (a documentation warning shown under `RUNE2001`) is a source typo; the documentation warning is `RUNE2010`.
- New codes MUST use the next free number in the appropriate range and be added to this catalog in the same change.

## Verification

Golden tests (SC-002) assert, for a fixture per condition, that both `rune analyze` and the LSP `publishDiagnostics` report the identical `{Code, Severity, Range}` (and `Related` where applicable).
