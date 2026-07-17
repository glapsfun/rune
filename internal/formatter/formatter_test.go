package formatter

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/parser"
)

var update = flag.Bool("update", false, "regenerate formatter golden files")

func TestFmtGolden(t *testing.T) {
	matches, _ := filepath.Glob(filepath.Join("..", "..", "testdata", "fmt", "*.rune"))
	if len(matches) == 0 {
		t.Skip("no formatter fixtures yet")
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
				t.Fatalf("parse: %v", diags)
			}
			got := Format(f)

			golden := strings.TrimSuffix(in, ".rune") + ".fmt"
			if *update {
				if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("missing golden %s (run with -update): %v", golden, err)
			}
			if got != string(want) {
				t.Errorf("format mismatch for %s:\n got:\n%s\nwant:\n%s", in, got, want)
			}
		})
	}
}

// TestFmtIdempotent verifies formatting is stable: formatting already-formatted
// output yields the same text.
func TestFmtIdempotent(t *testing.T) {
	src := "set default:=\"greet\"\nx:=\"dist\"\n# Greet.\ngreet name=\"world\":\n    @echo hi {{name}}\nbuild: greet\n    go build\n"
	f, diags := parser.Parse("Runefile", src)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	once := Format(f)

	f2, diags2 := parser.Parse("Runefile", once)
	if diags2.HasErrors() {
		t.Fatalf("re-parse of formatted output failed: %v", diags2)
	}
	twice := Format(f2)
	if once != twice {
		t.Errorf("formatting is not idempotent:\nonce:\n%s\ntwice:\n%s", once, twice)
	}
}
