package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// benchFixture loads a representative Runefile from the corpus (settings, tasks,
// dependencies, attributes, expressions) to exercise the parser realistically.
func benchFixture(b *testing.B) string {
	b.Helper()
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "corpus", "full.rune"))
	if err != nil {
		b.Fatalf("read fixture: %v", err)
	}
	return string(src)
}

// BenchmarkParse measures lexing + recursive-descent parsing of a representative
// Runefile (the front-end hot path before analysis/execution).
func BenchmarkParse(b *testing.B) {
	src := benchFixture(b)
	b.ReportAllocs()
	for b.Loop() {
		Parse("Runefile", src)
	}
}
