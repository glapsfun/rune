# Contract: `--dump --format json` Schema Addition

**Feature**: 020-os-availability

## Change

Each element of `tasks[]` gains one field:

| Field | Type | Presence | Meaning |
|-------|------|----------|---------|
| `available` | boolean | always (no omitempty) | Whether the task may run on the OS that produced the dump |

## Rules

- Computed via `Task.AvailableOn(goos)` with the same host GOOS the run
  path uses; a dump therefore predicts exactly what that host will list,
  expose over MCP, and allow to run.
- ALL tasks remain present in the dump, including unavailable and private
  ones — `available: false` is data, not a filter (mirrors the existing
  `private` field).
- The raw OS attributes keep appearing in the existing `attributes` array
  (e.g. `"[windows]"` formatting per current attribute rendering), so
  consumers can see both the declared rule and the computed verdict.
- No other dump fields change; canonical text dump (default format) is
  byte-identical to today.

## Example (dumped on linux)

```json
{
  "name": "setup-win",
  "private": false,
  "available": false,
  "attributes": ["windows"]
}
```

(Field ordering/rendering of `attributes` follows the existing dump
implementation; the example is illustrative, not byte-exact.)
