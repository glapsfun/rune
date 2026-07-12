package lsp

import (
	"fmt"
	"unicode/utf8"

	"github.com/rune-task-runner/rune/internal/token"
)

// Position is an LSP text position: a 0-based line and a 0-based character
// offset counted in UTF-16 code units (LSP's default positionEncoding).
type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

// Range is an LSP half-open range [Start, End).
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// LineIndex converts between Rune's byte-oriented source offsets/columns and
// LSP's line + UTF-16-character positions. It is the single conversion layer for
// the server (spec FR-006): building it per document version keeps the mapping
// correct across the full unicode/line-ending matrix (ASCII, multi-byte,
// emoji/surrogate pairs, combining marks, CRLF, LF, empty lines, EOF).
type LineIndex struct {
	text        string
	lineOffsets []int // byte offset of the start of each line (index 0 == 0)
}

// NewLineIndex builds an index for text. Line starts are the byte after each
// '\n'; a '\r' before '\n' stays part of the preceding line's bytes but is never
// counted as content because real spans point at content boundaries.
func NewLineIndex(text string) *LineIndex {
	offsets := []int{0}
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			offsets = append(offsets, i+1)
		}
	}
	return &LineIndex{text: text, lineOffsets: offsets}
}

// ByteOffsetToPosition maps a byte offset to an LSP position. Out-of-range
// offsets are clamped so the function is total (never panics).
func (ix *LineIndex) ByteOffsetToPosition(offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(ix.text) {
		offset = len(ix.text)
	}
	line := ix.lineAt(offset)
	start := ix.lineOffsets[line]
	return Position{Line: uint32(line), Character: uint32(utf16Len(ix.text[start:offset]))}
}

// PositionToByteOffset maps an LSP position to a byte offset. The offset is
// always valid (clamped); an error is returned only when the line is beyond the
// document, so callers may either honor it or use the clamped offset.
func (ix *LineIndex) PositionToByteOffset(pos Position) (int, error) {
	line := int(pos.Line)
	if line < 0 {
		return 0, fmt.Errorf("negative line %d", pos.Line)
	}
	if line >= len(ix.lineOffsets) {
		return len(ix.text), fmt.Errorf("line %d beyond document (%d lines)", pos.Line, len(ix.lineOffsets))
	}
	start := ix.lineOffsets[line]
	end := len(ix.text)
	if line+1 < len(ix.lineOffsets) {
		end = ix.lineOffsets[line+1]
	}
	want := int(pos.Character)
	units, i := 0, start
	for i < end && units < want {
		r, size := utf8.DecodeRuneInString(ix.text[i:])
		if r > 0xFFFF {
			units += 2
		} else {
			units++
		}
		i += size
	}
	return i, nil
}

// SpanToRange converts a Rune source span to an LSP range.
func (ix *LineIndex) SpanToRange(span token.Span) Range {
	return Range{
		Start: ix.ByteOffsetToPosition(span.Start.Offset),
		End:   ix.ByteOffsetToPosition(span.End.Offset),
	}
}

// lineAt returns the 0-based line containing offset (the largest line whose
// start is <= offset), via binary search.
func (ix *LineIndex) lineAt(offset int) int {
	lo, hi := 0, len(ix.lineOffsets)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if ix.lineOffsets[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}

// utf16Len returns the number of UTF-16 code units in s (runes above the BMP
// count as a surrogate pair). Invalid UTF-8 bytes decode to U+FFFD, one unit
// each, keeping the count total.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}
