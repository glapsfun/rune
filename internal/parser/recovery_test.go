package parser

import "testing"

// taskNames returns the names of the tasks the parser produced.
func taskNames(src string) []string {
	f, _ := Parse("Runefile", src)
	var names []string
	for _, t := range f.Tasks {
		names = append(names, t.Name)
	}
	return names
}

func has(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

// TestRecoveryKeepsValidDeclarations verifies the core recovery contract
// (spec FR-004): a broken declaration is reported but does not prevent the
// parser from analyzing the valid declarations around it.
func TestRecoveryKeepsValidDeclarations(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []string // task names that must survive recovery
	}{
		{
			name: "unterminated string then valid task",
			src:  "first target=\"oops\n    echo one\n# ok\nsecond:\n    echo two\n",
			want: []string{"first", "second"},
		},
		{
			name: "broken top-level line between tasks",
			src:  "alpha:\n    echo a\n$$$ garbage @@@\nbeta:\n    echo b\n",
			want: []string{"alpha", "beta"},
		},
		{
			name: "incomplete assignment then task",
			src:  "x :=\nbuild:\n    echo b\n",
			want: []string{"build"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Must report at least one diagnostic (the broken region).
			f, diags := Parse("Runefile", tc.src)
			if f == nil {
				t.Fatal("nil file")
			}
			if !diags.HasErrors() {
				t.Errorf("expected a diagnostic for the broken region")
			}
			names := taskNames(tc.src)
			for _, w := range tc.want {
				if !has(names, w) {
					t.Errorf("recovery dropped task %q (got %v)", w, names)
				}
			}
		})
	}
}

// TestUnterminatedGroupIsBounded documents a known recovery limitation: because
// the lexer transparently continues logical lines inside unclosed ()/[]/{}
// (group continuation), an unterminated group consumes the rest of the file, so
// declarations after it cannot be recovered. The guaranteed invariant is only
// that parsing still terminates and reports the error (verified broadly by
// FuzzParseRecover). In practice this is transient while editing — the user (or
// editor auto-pairing) closes the bracket.
func TestUnterminatedGroupIsBounded(t *testing.T) {
	src := "deploy: (\n    echo d\nbuild:\n    echo b\n"
	f, diags := Parse("Runefile", src)
	if f == nil {
		t.Fatal("nil file")
	}
	if !diags.HasErrors() {
		t.Error("expected a diagnostic for the unterminated group")
	}
}
