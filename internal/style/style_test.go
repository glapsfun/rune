package style

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/styletest"
)

// stripANSI recovers the visible text via the shared test helper.
func stripANSI(s string) string { return styletest.StripSGR(s) }

// roles returns each role's Render func by name for table-driven assertions.
func roles(th Theme) map[string]func(...string) string {
	return map[string]func(...string) string{
		"Error":    th.Error.Render,
		"Warning":  th.Warning.Render,
		"Success":  th.Success.Render,
		"TaskName": th.TaskName.Render,
		"Heading":  th.Heading.Render,
		"Muted":    th.Muted.Render,
		"Locator":  th.Locator.Render,
		"Caret":    th.Caret.Render,
	}
}

// FR-003: a disabled theme must emit zero ANSI and return its input byte-for-byte.
func TestDisabledThemeIsPlain(t *testing.T) {
	th := New(false, io.Discard)
	for name, render := range roles(th) {
		got := render("hello")
		if got != "hello" {
			t.Errorf("role %s disabled: got %q, want plain %q", name, got, "hello")
		}
	}
}

// D9 (014): the exported palette constants are the single source of color
// literals; the picker builds its styles from them, so their values are part of
// the package contract.
func TestPaletteConstants(t *testing.T) {
	for name, got := range map[string]string{
		"ColorError":     string(ColorError),
		"ColorWarning":   string(ColorWarning),
		"ColorSuccess":   string(ColorSuccess),
		"ColorAccent":    string(ColorAccent),
		"ColorMuted":     string(ColorMuted),
		"ColorMutedDark": string(ColorMutedDark),
	} {
		want := map[string]string{
			"ColorError":     "1",
			"ColorWarning":   "3",
			"ColorSuccess":   "2",
			"ColorAccent":    "170",
			"ColorMuted":     "245",
			"ColorMutedDark": "241",
		}[name]
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

// 014 FR-010 / 008 SC-007: this package is the only place allowed to define
// color literals. Any `lipgloss.Color("` in non-test source outside
// internal/style means a surface is bypassing the shared palette.
func TestNoColorLiteralsOutsideStylePackage(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	selfDir, _ := os.Getwd()
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "dist" || name == "testdata" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if filepath.Dir(path) == selfDir {
			return nil
		}
		src, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		if strings.Contains(string(src), `lipgloss.Color("`) {
			t.Errorf("%s defines a color literal; use internal/style's exported palette constants", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// FR-012: an enabled theme adds SGR escapes but never changes the visible text
// or its width (stripping ANSI must recover the original string exactly).
func TestEnabledThemeStylesButPreservesText(t *testing.T) {
	th := New(true, io.Discard)
	const in = "build"
	for name, render := range roles(th) {
		got := render(in)
		if got == in {
			t.Errorf("role %s enabled: expected ANSI styling, got plain %q", name, got)
		}
		if stripANSI(got) != in {
			t.Errorf("role %s enabled: visible text changed: stripped %q, want %q", name, stripANSI(got), in)
		}
	}
}
