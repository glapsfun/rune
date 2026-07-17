package analysis

import (
	"context"
	"testing"
)

func TestApplyByteEditBasic(t *testing.T) {
	cases := []struct {
		name       string
		text       string
		start, end int
		newText    string
		want       string
	}{
		{"replace middle", "hello world", 6, 11, "there", "hello there"},
		{"insert at start", "world", 0, 0, "hello ", "hello world"},
		{"append at end", "hello", 5, 5, " world", "hello world"},
		{"delete range", "abcdef", 1, 4, "", "aef"},
		{"full replace", "old", 0, 3, "new", "new"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ApplyByteEdit(tc.text, tc.start, tc.end, tc.newText); got != tc.want {
				t.Errorf("ApplyByteEdit(%q,%d,%d,%q) = %q, want %q", tc.text, tc.start, tc.end, tc.newText, got, tc.want)
			}
		})
	}
}

func TestApplyByteEditClampsOutOfRange(t *testing.T) {
	// Out-of-range and reversed offsets must not panic and must clamp/order.
	if got := ApplyByteEdit("abc", -5, 100, "X"); got != "X" {
		t.Errorf("out-of-range replace = %q, want %q", got, "X")
	}
	if got := ApplyByteEdit("abc", 2, 1, "-"); got != "a-c" {
		t.Errorf("reversed offsets = %q, want %q", got, "a-c")
	}
}

func TestOverlayStorePrecedenceAndFallback(t *testing.T) {
	dir := t.TempDir()
	// A disk file that is NOT open in the overlay.
	onDisk := writeTemp(t, dir, "disk.rune", "from disk")
	overlay := NewOverlaySourceStore(DiskSourceStore{})

	ctx := context.Background()

	// Not open yet: falls back to disk.
	b, err := overlay.Read(ctx, onDisk)
	if err != nil || string(b) != "from disk" {
		t.Fatalf("disk fallback: got %q, err %v", b, err)
	}

	// Open it with unsaved content: overlay wins.
	overlay.Set(OpenDocument{URI: onDisk, Version: 2, Text: "unsaved edit"})
	b, _ = overlay.Read(ctx, onDisk)
	if string(b) != "unsaved edit" {
		t.Errorf("overlay precedence: got %q, want %q", b, "unsaved edit")
	}
	if !overlay.Exists(ctx, onDisk) {
		t.Error("Exists should be true for an open document")
	}

	// Close it: falls back to disk again.
	overlay.Remove(onDisk)
	b, _ = overlay.Read(ctx, onDisk)
	if string(b) != "from disk" {
		t.Errorf("after close: got %q, want disk content", b)
	}
}

// FuzzApplyEdits asserts ApplyByteEdit never panics and always terminates for
// arbitrary text, offsets, and replacement (spec FR-005-style invariant).
func FuzzApplyEdits(f *testing.F) {
	f.Add("hello world", 0, 5, "X")
	f.Add("", -1, 1000, "insert")
	f.Add("émoji 🎉 test", 3, 8, "Ω")
	f.Fuzz(func(t *testing.T, text string, start, end int, newText string) {
		got := ApplyByteEdit(text, start, end, newText)
		// Result length is bounded by original minus (at most) the spliced span
		// plus the insertion — a coarse sanity bound that also proves no runaway.
		if len(got) > len(text)+len(newText) {
			t.Fatalf("result longer than possible: len=%d text=%d new=%d", len(got), len(text), len(newText))
		}
	})
}
