package mask

import (
	"io"
	"testing"
)

// benchChunks is ~10 MB of log-like output in 32 KiB writes, with occasional
// secret occurrences so the match path is exercised (SC-004 compares this
// against the unmasked baseline; the target is ≤10% overhead).
func benchChunks(withSecret bool) [][]byte {
	line := []byte("2026-07-21T12:00:00Z info processing item batch=42 status=ok elapsed=13ms\n")
	secretLine := []byte("2026-07-21T12:00:00Z debug auth header bearer hunter2-bench-secret ok\n")
	chunk := make([]byte, 0, 32<<10)
	for i := 0; len(chunk) < (32<<10)-len(line); i++ {
		if withSecret && i%512 == 0 {
			chunk = append(chunk, secretLine...)
			continue
		}
		chunk = append(chunk, line...)
	}
	n := (10 << 20) / len(chunk)
	chunks := make([][]byte, n)
	for i := range chunks {
		chunks[i] = chunk
	}
	return chunks
}

func BenchmarkWriter_Unmasked(b *testing.B) {
	chunks := benchChunks(true)
	b.SetBytes(10 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range chunks {
			_, _ = io.Discard.Write(c)
		}
	}
}

func BenchmarkWriter_Masked(b *testing.B) {
	chunks := benchChunks(true)
	set := NewSet([]string{"BENCH_API_TOKEN=hunter2-bench-secret"}, nil, nil)
	b.SetBytes(10 << 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWriter(io.Discard, set)
		for _, c := range chunks {
			_, _ = w.Write(c)
		}
		_ = w.Flush()
	}
}
