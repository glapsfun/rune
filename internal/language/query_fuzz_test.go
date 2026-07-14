package language

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/parser"
)

// FuzzLanguageQueries asserts the position-based query layer never panics for
// arbitrary source and cursor offsets — including mid-edit input and
// out-of-range offsets (spec FR-005/SC-004 applied to completion, definition,
// hover, and target resolution). Recent LSP-reliability research highlights
// crashes from combining malformed source with editor operations at arbitrary
// positions, which this exercises directly.
func FuzzLanguageQueries(f *testing.F) {
	seeds := []struct {
		src    string
		offset int
	}{
		{sampleRunefile, 40},
		{"deploy: bu", 10},
		{"x := {{ bui", 11},
		{"[conf", 5},
		{"build (py", 9},
		{"set wor", 7},
		{"platform := os_", 15},
		{"", 0},
		{"héllo 🎉:\n    echo {{ x", 25},
	}
	for _, s := range seeds {
		f.Add(s.src, s.offset)
	}

	f.Fuzz(func(t *testing.T, src string, offset int) {
		file, _ := parser.Parse("Runefile", src)
		config.Compose(file, nil)
		ix := BuildIndex(file)

		// None of these may panic for any input/offset.
		_ = Complete(ix, file, "Runefile", src, offset)
		_, _ = Definition(ix, file, "Runefile", offset)
		_, _, _ = Hover(file, "Runefile", offset)
		_, _ = TargetAt(file, "Runefile", offset)
	})
}
