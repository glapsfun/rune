package docs

import (
	"os"
	"path/filepath"
	"testing"
)

// selfContainedPages lists docs whose fenced ```rune blocks are each a
// complete, self-contained Runefile and therefore MUST statically validate.
// Pages with deliberate fragments (lone expressions, partial snippets) fence
// them as ```text and are not listed here. As more pages are brought into
// compliance (US3 / T042), add them so coverage tightens over time.
var selfContainedPages = []string{
	"docs/overview.md",
	"docs/getting-started.md",
}

// TestRuneCodeBlocksValidate extracts every ```rune block from the
// self-contained pages and asserts it validates (`--list`, exit 0), so prose
// examples cannot drift from the real DSL.
func TestRuneCodeBlocksValidate(t *testing.T) {
	for _, page := range selfContainedPages {
		t.Run(page, func(t *testing.T) {
			path := filepath.Join(repoRoot, filepath.FromSlash(page))
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", page, err)
			}
			blocks := fencedBlocks(string(src), "rune")
			if len(blocks) == 0 {
				t.Skipf("%s: no ```rune blocks to validate", page)
			}
			for i, b := range blocks {
				tmp := filepath.Join(t.TempDir(), "Runefile")
				if err := os.WriteFile(tmp, []byte(b), 0o644); err != nil {
					t.Fatal(err)
				}
				if got := validate(t, tmp); got.code != 0 {
					t.Errorf("%s: ```rune block #%d failed `--list` (exit %d)\n--- block ---\n%s--- stderr ---\n%s",
						page, i+1, got.code, b, got.stderr)
				}
			}
		})
	}
}
