package lexer

import "testing"

// FuzzLexer asserts the lexer never panics and always terminates its stream
// with EOF, regardless of input.
func FuzzLexer(f *testing.F) {
	seeds := []string{
		"",
		"x := \"y\"\n",
		"greet:\n    @echo hi\n",
		"# doc\nbuild: greet\n    echo {{x}}\n",
		"set shell := [\"bash\", \"-c\"]\n",
		"a (python):\n\tif x:\n\t\tprint(1)\n",
		"deploy: docker::push\n  echo done\n",
		"\t \t mixed\n",
		"\"\"\"\ntriple\n\"\"\"\n",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		toks, _ := Lex("fuzz", src)
		if len(toks) == 0 {
			t.Fatal("expected at least an EOF token")
		}
		if last := toks[len(toks)-1]; last.Kind.String() != "EOF" {
			t.Fatalf("stream did not end with EOF, got %s", last.Kind)
		}
	})
}
