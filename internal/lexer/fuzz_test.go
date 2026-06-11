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
		// Regression: embedded NUL bytes must not be mistaken for EOF (they once
		// caused a non-terminating loop). The lexer must still terminate with EOF.
		"\x00",
		"a\x00b\n",
		"\t \tx\x00y\n",
		// Regression: a trailing backslash at EOF in a double-quoted string must
		// not read past the input (was an index-out-of-range panic).
		"\"\\",
		"0\"\\",
		"x := \"ab\\",
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
