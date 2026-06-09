package lexer

import (
	"os"
	"path/filepath"
	"testing"
)

// benchFixture loads a representative Runefile from the corpus to exercise a
// realistic mix of tokens (settings, tasks, attributes, expressions).
func benchFixture(b *testing.B) string {
	b.Helper()
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "corpus", "full.rune"))
	if err != nil {
		b.Fatalf("read fixture: %v", err)
	}
	return string(src)
}

// BenchmarkLex measures tokenizing a representative Runefile (hot path: every
// invocation lexes the source before anything else).
func BenchmarkLex(b *testing.B) {
	src := benchFixture(b)
	b.ReportAllocs()
	for b.Loop() {
		Lex("Runefile", src)
	}
}
