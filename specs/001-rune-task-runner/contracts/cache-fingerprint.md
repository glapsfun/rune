# Contract: Cache Fingerprint & Storage

**Feature**: 001-rune-task-runner | **Date**: 2026-06-08

Covers FR-015 and FR-020 plus Constitution Principle I (caching is per-task opt-in; no
timestamp skipping; skips are visible). Applies only to tasks carrying
`[cache(inputs=[…], outputs=[…])]`.

## Fingerprint

A task is skipped **iff** both hold:
1. The recomputed fingerprint equals the stored fingerprint, AND
2. Every declared output path currently exists.

Otherwise the task runs and the stored fingerprint is updated afterward (on success).

The fingerprint is `SHA-256` (hex) over a canonical, order-stable serialization of:

| Component | Detail |
|-----------|--------|
| input files | each declared input glob expanded to a sorted file list; for each file: relative path + SHA-256 of its contents |
| task body | the raw body text (pre-interpolation) |
| resolved variables | the names + resolved values of every variable/param referenced by the task |
| executor identity | executor kind + resolved interpreter command (e.g. `sh`, or `["python","-"]`) |

Rationale: any change to inputs, the script, the values it interpolates, or the interpreter
that runs it must invalidate the cache. Missing outputs force a run even on a hash match.

## Record format (JSON)

Stored one file per cache key under `.rune/cache/`:

```json
{
  "key": "build-cached",
  "namespace": "",
  "hash": "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
  "inputs": ["go.mod", "go.sum", "cmd/rune/main.go"],
  "outputs": ["dist/rune"],
  "executor": "sh",
  "createdAt": "2026-06-08T00:00:00Z"
}
```

- `createdAt` is informational only; it is **never** part of the hash (no timestamp-based
  decisions — Principle I).
- Filename: `.rune/cache/<sanitized-namespaced-key>.json`.

## Storage location & hygiene

- Default root: `.rune/cache/` at the directory containing the resolved Runefile.
- Rune SHOULD advise adding `.rune/` to `.gitignore` (and may scaffold it).
- A `rune --clear-cache` style affordance (Phase 6) removes the cache directory.
- Cache corruption / unreadable record → treat as a miss (run the task), never error out.

## Observability (Principle I)

Every cache decision emits a line to stderr: `cached: <task>` (skipped) or `running: <task>`
(miss/changed). Silent skipping is forbidden. `--dry-run` reports the would-be decision without
running or writing records.

## Acceptance mapping

US5 scenario 2 + SC-006: second run with unchanged inputs is skipped (< 10% of original time)
and logged `cached`; changing an input or deleting an output forces a re-run.
