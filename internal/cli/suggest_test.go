package cli

import "testing"

// US3: nearest powers the did-you-mean suggestion on the "unknown task" error.

func TestNearest(t *testing.T) {
	cands := []string{"build", "test", "serve", "version", "completion"}
	cases := []struct {
		token, want string
		ok          bool
	}{
		{"serv", "serve", true},     // 1 deletion
		{"tset", "test", true},      // transposition (distance 2)
		{"biuld", "build", true},    // transposition (distance 2)
		{"versoin", "version", true}, // transposition (distance 2)
		{"zzzzzzzzzz", "", false},   // nothing close
	}
	for _, c := range cases {
		got, ok := nearest(c.token, cands)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("nearest(%q) = (%q,%v); want (%q,%v)", c.token, got, ok, c.want, c.ok)
		}
	}
}
