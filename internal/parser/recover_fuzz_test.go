package parser

import "testing"

// FuzzParseRecover asserts the recovery invariant the language server relies on
// (spec FR-005 / SC-004): for ANY input the parser terminates (a hang is a fuzz
// timeout), never panics, and every diagnostic range is ordered and in-bounds.
// It complements FuzzParser (which checks non-nil/no-panic) by validating spans.
func FuzzParseRecover(f *testing.F) {
	seeds := []string{
		"",
		"build:\n    echo hi\n",
		"build target=\"",          // unterminated string
		"deploy: (",                // open paren in a dependency
		"[working-directory(",      // open attribute
		"x := \nbuild:\n  echo ok", // incomplete expression, then a valid task
		"set\n",
		"mod",
		"import",
		"::::::",
		"if if if { } else",
		"a := b + + +",
		"\t\t\n  \n[[[[",
		"héllo 🎉:\n    echo 世界\n", // multibyte in names/bodies
		"{{{{}}}}",
		"build:\n\tमिश्र\n  spaces", // mixed indentation with unicode
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, src string) {
		file, diags := Parse("Runefile", src)
		if file == nil {
			t.Fatal("Parse returned a nil file")
		}
		n := len(src)
		for _, d := range diags {
			s, e := d.Span.Start.Offset, d.Span.End.Offset
			if s < 0 || e < 0 || s > n || e > n {
				t.Fatalf("span offsets [%d,%d] out of bounds (len %d) for %q", s, e, n, d.Message)
			}
			if s > e {
				t.Fatalf("span start %d > end %d for %q", s, e, d.Message)
			}
			if d.Span.Start.Line < 1 || d.Span.Start.Col < 1 {
				t.Fatalf("non-positive line/col (%d:%d) for %q", d.Span.Start.Line, d.Span.Start.Col, d.Message)
			}
		}
	})
}
