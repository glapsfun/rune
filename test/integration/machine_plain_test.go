package integration

import (
	"strings"
	"testing"
)

// 014 US2 (C7): machine-readable surfaces never carry ANSI, even under
// --color=always — the strongest forcing mode.

func TestMachineSurfacesNeverStyled(t *testing.T) {
	dir := writeRunefile(t, "# Build.\nbuild:\n    @echo hi\n")

	cases := [][]string{
		{"--color=always", "--dump"},
		{"--color=always", "--dump", "--format", "json"},
		{"--color=always", "analyze", "--json"},
		{"version", "--check", "--json", "--color=always"},
		{"completion", "bash", "--color=always"},
		{"completion", "zsh", "--color=always"},
		{"completion", "fish", "--color=always"},
		{"completion", "powershell", "--color=always"},
	}
	for _, args := range cases {
		r := run(t, dir, nil, args...)
		if hasANSI(r.stdout) {
			t.Errorf("%v: stdout carried ANSI: %q", args, firstLine(r.stdout))
		}
		if hasANSI(r.stderr) {
			t.Errorf("%v: stderr carried ANSI: %q", args, firstLine(r.stderr))
		}
		if r.stdout == "" {
			t.Errorf("%v: expected machine output on stdout", args)
		}
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
