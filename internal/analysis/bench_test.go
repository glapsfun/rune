package analysis

import (
	"context"
	"testing"

	"github.com/rune-task-runner/rune/internal/parser"
)

// sampleRunefile is a representative small Runefile used by the benchmarks.
const sampleRunefile = `output := "dist"
target := "debug"

# Build the application.
build:
    go build -tags {{target}} -o {{output}}/app ./...

# Run the tests.
test: build
    go test ./...

# Deploy the application.
deploy env: build test
    ./deploy.sh {{env}}
`

func BenchmarkParseRunefile(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse("Runefile", sampleRunefile)
	}
}

func BenchmarkAnalyzeRunefile(b *testing.B) {
	svc := NewService(DiskSourceStore{})
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := svc.Analyze(ctx, AnalyzeRequest{URI: "Runefile", Content: sampleRunefile, Version: 1}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkImportedFileInvalidation(b *testing.B) {
	g := NewImportGraph()
	g.AddEdge("root", "mid")
	g.AddEdge("mid", "leaf")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = g.TransitiveImporters("leaf")
	}
}
