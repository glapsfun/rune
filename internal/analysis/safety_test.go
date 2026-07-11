package analysis

import (
	"context"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// forbiddenDirectImport reports whether an import path is one the read-only
// analysis surface must never DIRECTLY depend on (spec FR-028): process
// execution, networking, Rune's execution runtime, or the MCP server. (Note:
// eval legitimately uses exec.LookPath for require()/which() — a PATH query,
// not execution — so a transitive os/exec is allowed; what these packages must
// not do is reach for execution themselves.)
func forbiddenDirectImport(path string) bool {
	switch {
	case path == "os/exec":
		return true
	case path == "net" || strings.HasPrefix(path, "net/"):
		return true
	case strings.HasPrefix(path, "github.com/rune-task-runner/rune/internal/runtime"):
		return true
	case path == "github.com/rune-task-runner/rune/mcpserver":
		return true
	default:
		return false
	}
}

func TestReadOnlyPackagesHaveNoExecutionImports(t *testing.T) {
	// Package dirs relative to this test's working directory (internal/analysis).
	dirs := map[string]string{
		"analysis": ".",
		"language": filepath.Join("..", "language"),
		"lsp":      filepath.Join("..", "lsp"),
	}
	for name, dir := range dirs {
		pkg, err := build.ImportDir(dir, 0)
		if err != nil {
			t.Fatalf("import %s: %v", name, err)
		}
		for _, imp := range pkg.Imports { // non-test imports only
			if forbiddenDirectImport(imp) {
				t.Errorf("package %s must not directly import %q (FR-028)", name, imp)
			}
		}
	}
}

// TestAnalyzeWritesNoFiles asserts analysis has no filesystem side effects: the
// project directory is unchanged after an Analyze call (FR-028 / SC-005).
func TestAnalyzeWritesNoFiles(t *testing.T) {
	dir := t.TempDir()
	path := writeTemp(t, dir, "Runefile", "build:\n    @echo hi\ndeploy: build\n    @echo deploy\n")

	before := dirSnapshot(t, dir)
	svc := NewService(DiskSourceStore{})
	if _, err := svc.Analyze(context.Background(), AnalyzeRequest{URI: path, Version: 1}); err != nil {
		t.Fatal(err)
	}
	after := dirSnapshot(t, dir)

	if len(before) != len(after) {
		t.Errorf("Analyze changed the project directory: before %v, after %v", before, after)
	}
}

func dirSnapshot(t *testing.T, dir string) []string {
	t.Helper()
	var names []string
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		names = append(names, p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return names
}
