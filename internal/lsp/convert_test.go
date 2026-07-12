package lsp

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/token"
)

func TestByteOffsetToPositionMatrix(t *testing.T) {
	// "héllo 🎉\nдруге" — é is 2 bytes, 🎉 is 4 bytes / 2 UTF-16 units, Cyrillic 2 bytes each.
	text := "héllo 🎉\nдруге"
	ix := NewLineIndex(text)

	cases := []struct {
		name     string
		offset   int
		wantLine uint32
		wantChar uint32
	}{
		{"start", 0, 0, 0},
		{"after é (2 bytes)", len("hé"), 0, 2},                 // h=1, é=1 → char 2
		{"before emoji", len("héllo "), 0, 6},                  // 6 chars: h é l l o space
		{"after emoji (surrogate pair)", len("héllo 🎉"), 0, 8}, // +2 UTF-16 units
		{"line 2 start", len("héllo 🎉\n"), 1, 0},
		{"after 2 cyrillic", len("héllo 🎉\nдр"), 1, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ix.ByteOffsetToPosition(tc.offset)
			if got.Line != tc.wantLine || got.Character != tc.wantChar {
				t.Errorf("offset %d = %+v, want line %d char %d", tc.offset, got, tc.wantLine, tc.wantChar)
			}
		})
	}
}

func TestCRLFAndEmptyLinesAndEOF(t *testing.T) {
	text := "a\r\n\r\nbb"
	ix := NewLineIndex(text)
	// Line 0 "a", line 1 "" (empty), line 2 "bb".
	if p := ix.ByteOffsetToPosition(0); p.Line != 0 || p.Character != 0 {
		t.Errorf("offset 0 = %+v", p)
	}
	// Start of line 2 ("bb") is after "a\r\n\r\n" = 5 bytes.
	if p := ix.ByteOffsetToPosition(5); p.Line != 2 || p.Character != 0 {
		t.Errorf("line 2 start = %+v, want line 2 char 0", p)
	}
	// EOF position.
	if p := ix.ByteOffsetToPosition(len(text)); p.Line != 2 || p.Character != 2 {
		t.Errorf("EOF = %+v, want line 2 char 2", p)
	}
}

func TestPositionByteOffsetRoundTrip(t *testing.T) {
	text := "héllo 🎉\nдруге\n"
	ix := NewLineIndex(text)
	for offset := 0; offset <= len(text); offset++ {
		// Only test rune-boundary offsets (positions never land mid-rune).
		if offset < len(text) && !startsRune(text, offset) {
			continue
		}
		pos := ix.ByteOffsetToPosition(offset)
		got, err := ix.PositionToByteOffset(pos)
		if err != nil {
			t.Fatalf("offset %d: %v", offset, err)
		}
		if got != offset {
			t.Errorf("round trip offset %d -> %+v -> %d", offset, pos, got)
		}
	}
}

func TestOutOfRangeClamps(t *testing.T) {
	ix := NewLineIndex("abc")
	if p := ix.ByteOffsetToPosition(-5); p.Line != 0 || p.Character != 0 {
		t.Errorf("negative offset = %+v", p)
	}
	if p := ix.ByteOffsetToPosition(999); p.Line != 0 || p.Character != 3 {
		t.Errorf("huge offset = %+v", p)
	}
	if off, err := ix.PositionToByteOffset(Position{Line: 99, Character: 0}); off != 3 || err == nil {
		t.Errorf("line beyond doc: off=%d err=%v, want clamped end + error", off, err)
	}
}

func TestSpanToRange(t *testing.T) {
	text := "greet:\n    echo hi\n"
	ix := NewLineIndex(text)
	// Span covering "greet" on line 0.
	sp := token.Span{
		Start: token.Position{Offset: 0, Line: 1, Col: 1},
		End:   token.Position{Offset: 5, Line: 1, Col: 6},
	}
	r := ix.SpanToRange(sp)
	if r.Start != (Position{0, 0}) || r.End != (Position{0, 5}) {
		t.Errorf("range = %+v, want {0,0}-{0,5}", r)
	}
}

func startsRune(s string, i int) bool {
	return s[i]&0xC0 != 0x80 // not a UTF-8 continuation byte
}

// (helpers above; matrix and fuzz tests exercise the conversion layer)

// FuzzUTF16Position asserts conversion is total: it never panics and always
// returns an in-bounds offset for arbitrary text and offsets (spec FR-005/SC-008).
func FuzzUTF16Position(f *testing.F) {
	f.Add("hello", 3)
	f.Add("héllo 🎉\nдруге", 10)
	f.Add("", 0)
	f.Add("\r\n\r\n", 2)
	f.Fuzz(func(t *testing.T, text string, offset int) {
		ix := NewLineIndex(text)
		// Totality (FR-005): conversion never panics and every produced offset is
		// within the document, for ANY offset (including mid-rune, which real
		// rune-aligned spans never are — strict round-trip is covered separately).
		pos := ix.ByteOffsetToPosition(offset)
		back, _ := ix.PositionToByteOffset(pos)
		if back < 0 || back > len(text) {
			t.Fatalf("offset %d out of bounds for text len %d", back, len(text))
		}
		if _, err := ix.PositionToByteOffset(Position{Line: 1 << 20, Character: 1 << 20}); err == nil {
			t.Fatal("expected error for line far beyond document")
		}
	})
}
