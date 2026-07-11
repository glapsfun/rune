package language

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/parser"
)

const sampleRunefile = `output := "dist"

# Build.
build target="debug":
    @echo {{target}} {{output}}

# Test.
test: build
    @echo test

# Deploy.
deploy env: build test
    @echo {{env}}
`

func BenchmarkBuildSymbolIndex(b *testing.B) {
	f, diags := parser.Parse("Runefile", sampleRunefile)
	if diags.HasErrors() {
		b.Fatalf("parse: %v", diags)
	}
	config.Compose(f, nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = BuildIndex(f)
	}
}

func BenchmarkCompletion(b *testing.B) {
	b.Skip("pending completion engine (T042/T043)")
}
