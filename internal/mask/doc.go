// Package mask provides emission-time masking of secret values in Rune's
// output streams, so credentials never reach a terminal transcript or an
// agent's chat history (spec 013-secret-masking).
//
// A Set is derived once per run from the task environment's variable NAMES
// (built-in sensitive patterns plus `set secrets` declarations, minus
// `set unmasked` exemptions) and is immutable afterwards. A Writer wraps an
// output stream and replaces every verbatim occurrence of a Set entry with
// the fixed Placeholder before bytes reach the underlying writer.
//
// Invariants:
//   - Values (or lines of multi-line values) shorter than MinLen bytes are
//     never tracked; multi-line values are tracked per line only, keeping the
//     Writer's carry bound small.
//   - The Writer is safe for concurrent use and holds back at most
//     maxEntryLen−1 bytes (a stream tail that is a proper prefix of an entry).
//   - Flush emits the carry after masking any completed entries in it; callers
//     may only flush when no producer can still be writing (the engine flushes
//     after the scheduler has joined every task).
//   - An empty Set means callers skip wrapping entirely, so secret-free runs
//     stay byte-identical.
package mask
