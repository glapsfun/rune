package diag

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/style"
	"github.com/rune-task-runner/rune/internal/token"
)

// ansiSeq strips SGR escapes so colored renderings can be compared to plain.
var ansiSeq = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiSeq.ReplaceAllString(s, "") }

var update = flag.Bool("update", false, "regenerate golden diagnostic renderings")

func sp(file string, line, col, endCol int) token.Span {
	return token.Span{
		File:  file,
		Start: token.Position{Line: line, Col: col},
		End:   token.Position{Line: line, Col: endCol},
	}
}

func TestRenderGolden(t *testing.T) {
	// One representative source per error class, rendered without color.
	cases := []struct {
		name   string
		source string
		diag   Diagnostic
	}{
		{
			name:   "unknown_task",
			source: "a: b\n    @echo a\n",
			diag:   New(sp("Runefile", 1, 4, 5), "unknown task: b"),
		},
		{
			name:   "undefined_var",
			source: "greet:\n    @echo {{nope}}\n",
			diag:   New(sp("Runefile", 2, 11, 19), "undefined variable: nope"),
		},
		{
			name:   "cycle",
			source: "c: c\n    @echo c\n",
			diag:   New(sp("Runefile", 1, 1, 2), "dependency cycle: c → c"),
		},
		{
			name:   "arity",
			source: "greet name:\n    @echo {{name}}\nall: greet\n    @echo all\n",
			diag:   New(sp("Runefile", 3, 6, 11), "task \"greet\" expects at least 1 argument(s), got 0"),
		},
	}

	var b strings.Builder
	for _, c := range cases {
		b.WriteString("# " + c.name + "\n")
		b.WriteString(Render(c.diag, []byte(c.source), style.Theme{}))
		b.WriteString("\n\n")
	}
	got := b.String()

	golden := filepath.Join("..", "..", "testdata", "diag", "render.golden")
	if *update {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatal(err)
		}
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
		t.Errorf("rendering mismatch:\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderHasLocationAndCaret(t *testing.T) {
	out := Render(New(sp("Runefile", 2, 11, 19), "undefined variable: nope"),
		[]byte("greet:\n    @echo {{nope}}\n"), style.Theme{})
	if !strings.HasPrefix(out, "Runefile:2:11: error: undefined variable: nope") {
		t.Errorf("missing file:line:col header: %q", out)
	}
	if !strings.Contains(out, "^^^^^^^^") {
		t.Errorf("missing caret underline: %q", out)
	}
}

// TestRenderColorPreservesLayout proves the colored rendering adds only
// zero-width emphasis: it contains ANSI, yet stripping it reproduces the plain
// rendering byte-for-byte — so the caret span stays column-aligned (SC-003).
func TestRenderColorPreservesLayout(t *testing.T) {
	src := []byte("greet:\n    @echo {{nope}}\n")
	d := New(sp("Runefile", 2, 11, 19), "undefined variable: nope")

	plain := Render(d, src, style.Theme{})
	colored := Render(d, src, style.New(true, io.Discard))

	if !strings.Contains(colored, "\x1b[") {
		t.Errorf("colored rendering has no ANSI: %q", colored)
	}
	if got := stripANSI(colored); got != plain {
		t.Errorf("colored stripped != plain:\n colored=%q\n plain  =%q", got, plain)
	}
}

// TestRenderWideCharsAndTabsPreserveAlignment guards SC-003 on a hard case: a
// tab-indented source line containing a multi-byte rune. The caret gutter must
// copy the leading tab (so it aligns under the rendered line), and the colored
// rendering must still strip back to the plain bytes exactly.
func TestRenderWideCharsAndTabsPreserveAlignment(t *testing.T) {
	// Line 2 is "\t@echo café {{nope}}"; the span covers the {{nope}} expression.
	src := []byte("greet:\n\t@echo café {{nope}}\n")
	d := New(sp("Runefile", 2, 14, 22), "undefined variable: nope")

	plain := Render(d, src, style.Theme{})
	colored := Render(d, src, style.New(true, io.Discard))

	if !strings.Contains(plain, "\t") {
		t.Errorf("expected the leading tab to be preserved for alignment: %q", plain)
	}
	if !strings.Contains(plain, "^") {
		t.Errorf("expected a caret underline: %q", plain)
	}
	if got := stripANSI(colored); got != plain {
		t.Errorf("wide-char/tab: colored stripped != plain:\n colored=%q\n plain  =%q", got, plain)
	}
}
