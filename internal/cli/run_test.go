package cli

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/parser"
	"github.com/rune-task-runner/rune/internal/styletest"
)

// parseRunefile parses src as a Runefile for tests, failing on any diagnostic.
func parseRunefile(t *testing.T, src string) *ast.File {
	t.Helper()
	f, diags := parser.Parse("Runefile", src)
	if len(diags) != 0 {
		t.Fatalf("unexpected parse diagnostics: %v", diags)
	}
	return f
}

// taskNames extracts just the names from a slice of tasks, for order/membership
// assertions without depending on unrelated task fields.
func taskNames(tasks []*ast.Task) []string {
	names := make([]string, len(tasks))
	for i, t := range tasks {
		names[i] = t.Name
	}
	return names
}

func TestVisibleTasksByGroup_OrdersByFirstOccurrence(t *testing.T) {
	f := parseRunefile(t, `
build:
    @echo build

[group("test")]
unit:
    @echo unit

[group("build")]
lint:
    @echo lint

[group("test")]
integration:
    @echo integration
`)

	order, groups := visibleTasksByGroup(f)

	wantOrder := []string{"", "test", "build"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("order = %v, want %v", order, wantOrder)
	}
	if got, want := taskNames(groups[""]), []string{"build"}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups[\"\"] = %v, want %v", got, want)
	}
	// "test" is recorded at unit's position but collects both "test" tasks,
	// including "integration" which appears later in the file - matching
	// --list's existing grouping rule.
	if got, want := taskNames(groups["test"]), []string{"unit", "integration"}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups[\"test\"] = %v, want %v", got, want)
	}
	if got, want := taskNames(groups["build"]), []string{"lint"}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups[\"build\"] = %v, want %v", got, want)
	}
}

func TestVisibleTasksByGroup_UngroupedBucketUsesEmptyKey(t *testing.T) {
	f := parseRunefile(t, `
build:
    @echo build

[group("release")]
tag:
    @echo tag

hello:
    @echo hi
`)

	order, groups := visibleTasksByGroup(f)

	wantOrder := []string{"", "release"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("order = %v, want %v", order, wantOrder)
	}
	if got, want := taskNames(groups[""]), []string{"build", "hello"}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups[\"\"] = %v, want %v", got, want)
	}
}

func TestVisibleTasksByGroup_SameGroupReusedNonContiguously(t *testing.T) {
	f := parseRunefile(t, `
[group("build")]
compile:
    @echo compile

[group("test")]
unit:
    @echo unit

[group("build")]
package:
    @echo package
`)

	order, groups := visibleTasksByGroup(f)

	wantOrder := []string{"build", "test"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("order = %v, want %v", order, wantOrder)
	}
	if got, want := taskNames(groups["build"]), []string{"compile", "package"}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups[\"build\"] = %v, want %v (must collect non-contiguous reuse into one section)", got, want)
	}
}

func TestVisibleTasksByGroup_ExcludesPrivateTasks(t *testing.T) {
	f := parseRunefile(t, `
[group("build")]
compile:
    @echo compile

[group("build")]
[private]
secret:
    @echo s
`)

	_, groups := visibleTasksByGroup(f)

	if got, want := taskNames(groups["build"]), []string{"compile"}; !reflect.DeepEqual(got, want) {
		t.Errorf("groups[\"build\"] = %v, want %v (private task must be excluded)", got, want)
	}
}

// 014 US2 (C6): the colored --list branch pads with its own width computation
// (rune count) while the plain branch uses fmt's byte-width padding. Task names
// are ASCII-only by the grammar, so the two models MUST agree — stripping ANSI
// from the colored listing must reproduce the plain listing byte-for-byte,
// including every padding space. This pins the alignment parity FR-009 requires.
func TestListTasksColoredAlignmentMatchesPlain(t *testing.T) {
	f := parseRunefile(t, `
# Short name.
go:
    @echo hi

# A much longer task name.
integration-tests:
    @echo hi

undocumented:
    @echo hi
`)

	render := func(color bool) string {
		var buf bytes.Buffer
		listTasks(Options{Stdout: &buf, ColorStdout: color}, f)
		return buf.String()
	}

	plain := render(false)
	colored := render(true)
	if !strings.ContainsRune(colored, '\x1b') {
		t.Fatalf("colored listing carried no ANSI: %q", colored)
	}
	if got := styletest.StripSGR(colored); got != plain {
		t.Errorf("colored listing (stripped) != plain:\n stripped=%q\n plain   =%q", got, plain)
	}
}
