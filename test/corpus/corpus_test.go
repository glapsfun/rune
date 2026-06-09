// Package corpus is the compatibility-corpus guard (Constitution Principle VI /
// FR-033). It re-parses a fixed set of known-good Runefiles and compares their
// canonical AST dumps to committed goldens, failing on any silent grammar drift.
// Regenerate goldens deliberately with `go test ./test/corpus -update`.
package corpus

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/parser"
)

var update = flag.Bool("update", false, "regenerate corpus golden AST dumps")

func TestCorpus(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("..", "..", "testdata", "corpus", "*.rune"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("compatibility corpus is empty; add fixtures under testdata/corpus")
	}
	for _, in := range matches {
		in := in
		t.Run(filepath.Base(in), func(t *testing.T) {
			src, err := os.ReadFile(in)
			if err != nil {
				t.Fatal(err)
			}
			f, diags := parser.Parse(filepath.Base(in), string(src))
			if diags.HasErrors() {
				t.Fatalf("corpus fixture %s no longer parses cleanly:\n%v", in, diags)
			}
			got := ast.Dump(f)
			golden := strings.TrimSuffix(in, ".rune") + ".ast"
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("missing corpus golden %s (run with -update): %v", golden, err)
			}
			if got != string(want) {
				t.Errorf("GRAMMAR DRIFT: %s parses differently than the committed golden.\nIf intentional, update GRAMMAR.md + regenerate with -update.\n got:\n%s\nwant:\n%s", in, got, want)
			}
		})
	}
}
