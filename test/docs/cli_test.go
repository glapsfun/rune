package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var longFlagRe = regexp.MustCompile(`--([a-z][a-z0-9-]*)`)

// longFlags returns the set of long flag names (without the leading --) in text.
func longFlags(text string) map[string]bool {
	set := make(map[string]bool)
	for _, m := range longFlagRe.FindAllStringSubmatch(text, -1) {
		set[m[1]] = true
	}
	return set
}

// TestCLIReferenceMatchesBinary keeps docs/cli.md honest (FR-013): its global-flags
// section must list exactly the binary's `--help` flags — neither missing nor
// inventing one — and it must document every exit code.
func TestCLIReferenceMatchesBinary(t *testing.T) {
	help := runRune(t, "", "--help")
	if help.code != 0 {
		t.Fatalf("`rune --help` exit = %d, want 0\nstderr:\n%s", help.code, help.stderr)
	}
	binFlags := longFlags(help.stdout)

	cliBytes, err := os.ReadFile(filepath.Join(repoRoot, "docs", "cli.md"))
	if err != nil {
		t.Fatalf("read docs/cli.md: %v", err)
	}
	cli := string(cliBytes)

	section := globalFlagsSection(cli)
	if section == "" {
		t.Fatal("docs/cli.md: '## Global flags' section not found")
	}
	docFlags := longFlags(section)

	for f := range binFlags {
		if !docFlags[f] {
			t.Errorf("flag --%s is in `rune --help` but undocumented in docs/cli.md global flags", f)
		}
	}
	for f := range docFlags {
		if !binFlags[f] {
			t.Errorf("flag --%s documented in docs/cli.md but absent from `rune --help`", f)
		}
	}

	for _, code := range []string{"`0`", "`1`", "`2`", "`3`", "`130`"} {
		if !strings.Contains(cli, code) {
			t.Errorf("docs/cli.md is missing exit code %s", code)
		}
	}
}

// globalFlagsSection returns the text of the "## Global flags" section (up to the
// next "## " heading), which holds the flag table the binary is compared against.
func globalFlagsSection(cli string) string {
	const heading = "## Global flags"
	i := strings.Index(cli, heading)
	if i < 0 {
		return ""
	}
	rest := cli[i+len(heading):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	return rest
}
