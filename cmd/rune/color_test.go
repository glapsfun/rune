package main

import (
	"io"
	"testing"
)

func TestParseColorMode(t *testing.T) {
	for _, ok := range []string{"auto", "always", "never"} {
		if _, err := parseColorMode(ok); err != nil {
			t.Errorf("parseColorMode(%q) unexpected error: %v", ok, err)
		}
	}
	for _, bad := range []string{"sometimes", "", "yes", "Always"} {
		if _, err := parseColorMode(bad); err == nil {
			t.Errorf("parseColorMode(%q): expected error, got nil", bad)
		}
	}
}

// Precedence matrix from contracts/color-flag.md, plus proof that FORCE_COLOR /
// CLICOLOR* are ignored (Clarifications).
func TestResolveColor(t *testing.T) {
	cases := []struct {
		name       string
		mode       colorMode
		noColor    bool
		forceColor bool // sets FORCE_COLOR + CLICOLOR_FORCE, which must be ignored
		isTTY      bool
		want       bool
	}{
		{"never on tty", colorNever, false, false, true, false},
		{"never with force", colorNever, false, true, true, false},
		{"always through pipe", colorAlways, false, false, false, true},
		{"always overrides NO_COLOR", colorAlways, true, false, false, true},
		{"auto tty", colorAuto, false, false, true, true},
		{"auto pipe", colorAuto, false, false, false, false},
		{"auto NO_COLOR set", colorAuto, true, false, true, false},
		{"auto FORCE_COLOR ignored", colorAuto, false, true, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.noColor {
				t.Setenv("NO_COLOR", "1")
			} else {
				t.Setenv("NO_COLOR", "")
			}
			if tc.forceColor {
				t.Setenv("FORCE_COLOR", "1")
				t.Setenv("CLICOLOR_FORCE", "1")
			}
			if got := resolveColor(tc.mode, tc.isTTY); got != tc.want {
				t.Errorf("resolveColor(%s, tty=%v): got %v, want %v", tc.mode, tc.isTTY, got, tc.want)
			}
		})
	}
}

func TestStreamIsTTYNonFile(t *testing.T) {
	if streamIsTTY(io.Discard) {
		t.Error("io.Discard reported as TTY")
	}
}
