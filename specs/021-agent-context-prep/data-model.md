# Data Model: Agent Context Prep (`[context]`)

**Feature**: 021-agent-context-prep | **Date**: 2026-08-21

No persistent storage. The feature's data lives in the AST, one in-flight
string, and two injection points.

## Entities

### Context Hook (AST-level)

The `[context]`-attributed task inside the composed Runefile.

| Field / property | Type | Rules |
|---|---|---|
| Attribute kind | `ast.AttrContext = "context"` | Bare attribute, no arguments |
| Cardinality | 0..1 per composed file | >1 → analyzer error (FR-001) |
| `[confirm]` combination | forbidden | analyzer error (FR-007) |
| Parameters | every one must carry a default | required / variadic-plus params → analyzer error (FR-007) |
| Executor | any except `agent` | `(agent)` hook → analyzer error (FR-007) |
| OS attributes | compose normally | mismatched host ⇒ hook treated as absent (FR-008) |
| `[private]` | composes for listings only | MCP tool exposure is denied by `[context]` itself (FR-006) |

### Gathered Context (in-flight value)

The processed output of one hook run: a single `string`; empty means "no
context" (revised 2026-08-22: the `(text, hasHook)` pair collapsed to one
string so both surfaces treat an empty-output hook identically).

| State | `text` |
|---|---|
| No hook, OS-mismatched hook, or hook with empty stdout | `""` — surfaces behave exactly as pre-feature (FR-009) |
| Success | masked stdout, trailing `\r`/`\n` trimmed |
| Success, oversized | masked stdout cut at ≤ 8 192 bytes (backed up to a UTF-8 rune boundary) + `\n[truncated]` (FR-004) |
| Failure / timeout | `(context hook "NAME" failed; proceeding without project context)` (FR-005) |

Invariants: masking always precedes truncation (the gathering path only ever
sees masked buffers); the value is computed fresh per gathering event and
never cached or persisted.

### Injection Points

| Surface | Carrier | Cardinality of gathering |
|---|---|---|
| MCP server (stdio + HTTP) | `mcpserver.Options.Instructions` → initialize result `instructions` | once per server start (FR-002) |
| Agent executor | prompt prefix under fixed delimiters (FR-003) | once per `(agent)` task invocation |

## State transitions

```text
                   ┌─ no [context] task / OS mismatch ──► ABSENT (hasHook=false, inject nothing)
agent surface start┤
                   └─ hook found ──► run via masked Call(), 10 s budget
                                        ├─ exit 0 ──► TEXT (trim; cap 8 KiB + marker)
                                        └─ error / non-zero / timeout ──► NOTICE (one line) + stderr warning
```
