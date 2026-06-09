package diag

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/token"
)

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
		b.WriteString(Render(c.diag, []byte(c.source), false))
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
		[]byte("greet:\n    @echo {{nope}}\n"), false)
	if !strings.HasPrefix(out, "Runefile:2:11: error: undefined variable: nope") {
		t.Errorf("missing file:line:col header: %q", out)
	}
	if !strings.Contains(out, "^^^^^^^^") {
		t.Errorf("missing caret underline: %q", out)
	}
}
