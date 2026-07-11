package analysis

// OpenDocument is an editor-held document: its identity, monotonically
// increasing version, and current full text (the overlay source that takes
// precedence over disk). See spec data-model.
type OpenDocument struct {
	URI     DocumentURI
	Version int
	Text    string
}

// WithText returns a copy of the document carrying new full text at a new
// version (used for full-sync changes and didSave).
func (d OpenDocument) WithText(version int, text string) OpenDocument {
	d.Version = version
	d.Text = text
	return d
}

// ApplyByteEdit splices newText into text over the half-open byte range
// [start, end), returning the result. Offsets are clamped into [0, len(text)]
// and ordered (start <= end) so the function is TOTAL: it never panics and
// always terminates for any inputs. The LSP layer converts an LSP range to byte
// offsets (via the UTF-16-aware LineIndex) before calling this; keeping the
// UTF-16 conversion out of this package avoids a dependency on internal/lsp.
func ApplyByteEdit(text string, start, end int, newText string) string {
	n := len(text)
	start = clamp(start, 0, n)
	end = clamp(end, 0, n)
	if start > end {
		start, end = end, start
	}
	return text[:start] + newText + text[end:]
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
