package mask

import (
	"io"
	"sync"
)

// Writer wraps an output stream and masks every occurrence of the Set's
// entries at emission time, including occurrences split across Write calls
// (FR-003/FR-004). It is safe for concurrent use: parallel tasks share the
// engine's streams.
type Writer struct {
	mu    sync.Mutex
	dst   io.Writer
	set   *Set
	carry []byte // stream tail that is a proper prefix of an entry; ≤ maxLen-1 bytes
}

// NewWriter wraps dst. Callers should not wrap at all when set.Empty() —
// that keeps secret-free runs byte-identical (FR-008).
func NewWriter(dst io.Writer, set *Set) *Writer {
	return &Writer{dst: dst, set: set}
}

// Write masks p (together with any carried tail) and forwards the result.
// It reports len(p) consumed on success even when a tail is withheld.
func (w *Writer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	buf := p
	if len(w.carry) > 0 {
		buf = append(w.carry, p...)
	}
	out, carry := w.set.scan(buf, false)
	w.carry = carry
	if len(out) > 0 {
		if _, err := w.dst.Write(out); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// Flush emits the withheld tail. Completed entries inside it are masked; a
// genuinely incomplete prefix is emitted verbatim (it is not the secret).
// Callers may only flush when no producer can still be writing — the engine
// flushes after the scheduler has joined every task (research.md D4).
func (w *Writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.carry) == 0 {
		return nil
	}
	out, _ := w.set.scan(w.carry, true)
	w.carry = nil
	if len(out) == 0 {
		return nil
	}
	_, err := w.dst.Write(out)
	return err
}

// Close flushes the writer. It never closes the underlying stream (the
// wrapped writer is typically os.Stdout/os.Stderr or a caller-owned buffer).
func (w *Writer) Close() error { return w.Flush() }
