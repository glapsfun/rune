package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sampleRunefile has a public task, a task whose name collides with a built-in
// command ('version'), and a [private] task that must never appear in the picker.
const sampleRunefile = "# Compile the binary.\nbuild:\n    @echo build\n\n# Print the version.\nversion:\n    @echo v1\n\n[private]\nsecret:\n    @echo s\n"

func writeRunefile(t *testing.T, content string) (dir, path string) {
	t.Helper()
	dir = t.TempDir()
	path = filepath.Join(dir, "Runefile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, path
}

// pickerItems must apply the same visibility rules as --list: non-private,
// OS-matching. It must also still offer a task whose name collides with a
// built-in command (FR-018), since selection runs it as a task.
func TestPickerItems_FiltersPrivateAndKeepsBuiltinCollision(t *testing.T) {
	_, path := writeRunefile(t, sampleRunefile)
	mod, err := loadModule(Options{Stderr: &bytes.Buffer{}}, path, false)
	if err != nil {
		t.Fatalf("loadModule: %v", err)
	}

	desc := map[string]string{}
	for _, it := range pickerItems(mod) {
		desc[it.Name] = it.Desc
	}

	if _, ok := desc["secret"]; ok {
		t.Error("[private] task must be excluded from picker items")
	}
	if !strings.Contains(desc["build"], "Compile the binary") {
		t.Errorf("build desc = %q, want first doc line", desc["build"])
	}
	if _, ok := desc["version"]; !ok {
		t.Errorf("task colliding with a built-in ('version') must stay selectable; got %v", desc)
	}
}

// FR-011/FR-019: --choose without an interactive terminal must error clearly
// (exit 2) instead of rendering a UI into captured output.
func TestChoose_NonInteractiveTerminalErrors(t *testing.T) {
	dir, _ := writeRunefile(t, sampleRunefile)
	opts := Options{
		Choose: true,
		Cwd:    dir,
		Stdin:  strings.NewReader(""), // not an *os.File → treated as non-TTY
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Ctx:    context.Background(),
	}

	err := Run(opts, nil)
	if err == nil {
		t.Fatal("expected an error for --choose without a terminal")
	}
	if !strings.Contains(err.Error(), "requires an interactive terminal") {
		t.Errorf("error = %q, want interactive-terminal message", err.Error())
	}
	if code := CodeFor(err); code != ExitUsage {
		t.Errorf("exit code = %d, want %d (usage)", code, ExitUsage)
	}
}

// FR-013: with an explicit task and no --choose, the picker is never involved,
// even when stdout is not a terminal.
func TestExplicitTaskRunsWithoutPicker(t *testing.T) {
	dir, _ := writeRunefile(t, sampleRunefile)
	var out bytes.Buffer
	opts := Options{
		Cwd:    dir,
		Stdin:  strings.NewReader(""),
		Stdout: &out,
		Stderr: &bytes.Buffer{},
		Ctx:    context.Background(),
	}

	if err := Run(opts, []string{"build"}); err != nil {
		t.Fatalf("explicit task run failed: %v", err)
	}
	if !strings.Contains(out.String(), "build") {
		t.Errorf("expected task output to contain %q, got %q", "build", out.String())
	}
}
