package analysis

import "testing"

// Benchmarks are stubbed until the analysis service lands (T022); bodies are
// filled in T024/T058. They exist now so the perf targets in the plan (SC-010)
// have named homes and CI can discover them.

func BenchmarkParseRunefile(b *testing.B) { b.Skip("pending T024") }

func BenchmarkAnalyzeRunefile(b *testing.B) { b.Skip("pending T024") }

func BenchmarkImportedFileInvalidation(b *testing.B) { b.Skip("pending T058") }
