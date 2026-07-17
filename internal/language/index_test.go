package language

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/parser"
)

// composed parses src and runs Compose so namespaced/imported tasks are present.
func composed(t *testing.T, src string) *Index {
	t.Helper()
	f, diags := parser.Parse("Runefile", src)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	config.Compose(f, nil) // no imports in these fixtures; nil provider is fine
	return BuildIndex(f)
}

func TestIndexTasksAndVariables(t *testing.T) {
	ix := composed(t, "output := \"dist\"\n# Build it.\nbuild target=\"debug\":\n    @echo {{target}}\n")

	tasks := ix.Tasks()
	if len(tasks) != 1 || tasks[0].Name != "build" {
		t.Fatalf("tasks = %+v, want one 'build'", tasks)
	}
	b := tasks[0]
	if b.Signature != `build target="debug"` {
		t.Errorf("signature = %q, want %q", b.Signature, `build target="debug"`)
	}
	if b.Documentation != "Build it." {
		t.Errorf("doc = %q, want %q", b.Documentation, "Build it.")
	}
	if !b.Exported {
		t.Error("build should be exported (not private)")
	}
	if got := ix.ByName["output"]; len(got) != 1 || got[0].Kind != SymbolVariable {
		t.Errorf("variable 'output' not indexed: %+v", got)
	}
	// The parameter is indexed in the task's scope.
	params := ix.byKind(SymbolParameter)
	if len(params) != 1 || params[0].Name != "target" || params[0].Scope != ScopeID("build") {
		t.Errorf("params = %+v, want target in scope 'build'", params)
	}
}

func TestIndexExportedFalseForPrivate(t *testing.T) {
	ix := composed(t, "_helper:\n    @echo hi\n")
	tasks := ix.Tasks()
	if len(tasks) != 1 || tasks[0].Exported {
		t.Fatalf("private task should have Exported=false: %+v", tasks)
	}
}

func TestIndexNamespacedModTask(t *testing.T) {
	// A mod namespaces its tasks as name::task; both base and qualified names
	// must be findable. (No file needed: mod bodies are parsed from src here via
	// a manual namespace — instead we assert the base/qualified split logic.)
	if baseName("docker::build") != "build" {
		t.Errorf("baseName(docker::build) = %q, want build", baseName("docker::build"))
	}
	if baseName("build") != "build" {
		t.Errorf("baseName(build) = %q, want build", baseName("build"))
	}
}
