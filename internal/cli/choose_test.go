package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/tui"
)

// sampleRunefile has a public task, a task whose name collides with a built-in
// command ('version'), and a [private] task that must never appear in the picker.
const sampleRunefile = "# Compile the binary.\nbuild:\n    @echo build\n\n# Print the version.\nversion:\n    @echo v1\n\n[private]\nsecret:\n    @echo s\n"

// groupedRunefile mixes multiple group(...) tasks with ungrouped tasks and a
// private task, mirroring spec.md US1 Acceptance Scenario 1.
const groupedRunefile = "build:\n    @echo build\n\n[group(\"build\")]\ncompile:\n    @echo compile\n\n[group(\"test\")]\nunit:\n    @echo unit\n\n[group(\"build\")]\npackage:\n    @echo package\n\n[group(\"test\")]\n[private]\nsecret:\n    @echo s\n"

func writeRunefile(t *testing.T, content string) (dir, path string) {
	t.Helper()
	dir = t.TempDir()
	path = filepath.Join(dir, "Runefile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, path
}

// flatten collects every item across all sections in order, for assertions
// that don't care about section boundaries.
func flatten(sections []tui.PickerSection) []tui.PickerItem {
	var items []tui.PickerItem
	for _, s := range sections {
		items = append(items, s.Items...)
	}
	return items
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
	for _, it := range flatten(pickerItems(mod)) {
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

// TestPickerItems_MatchesListGrouping asserts pickerItems partitions tasks
// into sections whose order and membership match visibleTasksByGroup's
// output directly - the same source --list uses - so the two surfaces can
// never drift apart (FR-001, FR-002, FR-003).
func TestPickerItems_MatchesListGrouping(t *testing.T) {
	_, path := writeRunefile(t, groupedRunefile)
	mod, err := loadModule(Options{Stderr: &bytes.Buffer{}}, path, false)
	if err != nil {
		t.Fatalf("loadModule: %v", err)
	}

	wantOrder, wantGroups := visibleTasksByGroup(mod.file)

	sections := pickerItems(mod)
	var gotOrder []string
	for _, s := range sections {
		gotOrder = append(gotOrder, s.Name)
	}
	if !reflect.DeepEqual(gotOrder, wantOrder) {
		t.Fatalf("section order = %v, want %v", gotOrder, wantOrder)
	}
	for i, s := range sections {
		var gotNames []string
		for _, it := range s.Items {
			gotNames = append(gotNames, it.Name)
		}
		var wantNames []string
		for _, task := range wantGroups[wantOrder[i]] {
			wantNames = append(wantNames, task.Name)
		}
		if !reflect.DeepEqual(gotNames, wantNames) {
			t.Errorf("section %q items = %v, want %v", s.Name, gotNames, wantNames)
		}
	}
}

// TestPickerItems_NoGroupsYieldsOneUnnamedSection asserts that for a Runefile
// with no group(...) attributes, pickerItems returns exactly one section
// (Name == "") containing every visible task in file order - the pre-
// feature-equivalent shape (FR-004, SC-002).
func TestPickerItems_NoGroupsYieldsOneUnnamedSection(t *testing.T) {
	_, path := writeRunefile(t, sampleRunefile)
	mod, err := loadModule(Options{Stderr: &bytes.Buffer{}}, path, false)
	if err != nil {
		t.Fatalf("loadModule: %v", err)
	}

	sections := pickerItems(mod)
	if len(sections) != 1 {
		t.Fatalf("sections = %d, want 1 (%v)", len(sections), sections)
	}
	if sections[0].Name != "" {
		t.Fatalf("section name = %q, want empty", sections[0].Name)
	}
	var names []string
	for _, it := range sections[0].Items {
		names = append(names, it.Name)
	}
	if want := []string{"build", "version"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
}

// TestPickerItems_NonFirstSectionTaskIsSelectableAndRuns covers FR-008: a
// task belonging to a non-first, non-ungrouped section ("unit", in the
// second group "test") is present in pickerItems and, when run directly by
// name - the same string chooseAndRun passes to execute() after a
// selection, with no section metadata attached - runs exactly like any other
// task, unaffected by which section it was displayed under.
func TestPickerItems_NonFirstSectionTaskIsSelectableAndRuns(t *testing.T) {
	dir, path := writeRunefile(t, groupedRunefile)
	mod, err := loadModule(Options{Stderr: &bytes.Buffer{}}, path, false)
	if err != nil {
		t.Fatalf("loadModule: %v", err)
	}

	sections := pickerItems(mod)
	found := false
	for _, s := range sections {
		if s.Name != "test" {
			continue
		}
		for _, it := range s.Items {
			if it.Name == "unit" {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("expected \"unit\" task in the \"test\" section; got %v", sections)
	}

	var out bytes.Buffer
	opts := Options{
		Cwd:    dir,
		Stdin:  strings.NewReader(""),
		Stdout: &out,
		Stderr: &bytes.Buffer{},
		Ctx:    context.Background(),
	}
	if err := Run(opts, []string{"unit"}); err != nil {
		t.Fatalf("running %q directly failed: %v", "unit", err)
	}
	if !strings.Contains(out.String(), "unit") {
		t.Errorf("expected task output to contain %q, got %q", "unit", out.String())
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
