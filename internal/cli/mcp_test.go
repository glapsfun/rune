package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/eval"
	"github.com/rune-task-runner/rune/internal/parser"
)

func adapterFor(t *testing.T, src string) *mcpAdapter {
	t.Helper()
	f, diags := parser.Parse("Runefile", src)
	if diags.HasErrors() {
		t.Fatalf("parse: %v", diags)
	}
	scope := eval.NewScope(indexAssignments(f), map[string]string{})
	settings, _ := config.ResolveSettings(f, eval.New(scope))
	return &mcpAdapter{
		file:      f,
		tasks:     indexTasks(f),
		assigns:   indexAssignments(f),
		settings:  settings,
		root:      t.TempDir(),
		workDir:   t.TempDir(),
		baseEnv:   nil,
		overrides: map[string]string{},
		now:       func() string { return "" },
	}
}

func TestAdapterExcludesPrivateTasks(t *testing.T) {
	src := "logs:\n    @echo logs\n[private]\nsecret:\n    @echo s\n_hidden:\n    @echo h\n"
	a := adapterFor(t, src)
	names := map[string]bool{}
	for _, ti := range a.Tasks() {
		names[ti.Name] = true
	}
	if !names["logs"] {
		t.Error("public task logs should be exposed")
	}
	if names["secret"] || names["_hidden"] {
		t.Errorf("private tasks must not be exposed: %v", names)
	}
}

func TestAdapterNoSecretValuesInToolFields(t *testing.T) {
	// A variable holding a "secret" must never appear in any exposed tool field.
	src := "api_key := \"super-secret-value\"\n# Deploy the app.\ndeploy:\n    @echo deploying with {{api_key}}\n"
	a := adapterFor(t, src)
	for _, ti := range a.Tasks() {
		blob := ti.Name + " " + ti.Doc
		for _, p := range ti.Params {
			blob += " " + p.Name
		}
		if strings.Contains(blob, "super-secret-value") {
			t.Errorf("secret leaked into tool fields: %q", blob)
		}
	}
}

func TestAdapterCallRunsThroughEngine(t *testing.T) {
	src := "greet name=\"world\":\n    @echo hi {{name}}\n"
	a := adapterFor(t, src)
	res, err := a.Call(context.Background(), "greet", map[string]string{"name": "Ada"})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 || !strings.Contains(res.Stdout, "hi Ada") {
		t.Errorf("call result = %+v", res)
	}
}

func TestAdapterDestructiveFlag(t *testing.T) {
	src := "[confirm(\"sure?\")]\nclean:\n    @echo clean\nlogs:\n    @echo logs\n"
	a := adapterFor(t, src)
	for _, ti := range a.Tasks() {
		if ti.Name == "clean" && !ti.Destructive {
			t.Error("clean should be marked destructive")
		}
		if ti.Name == "logs" && ti.Destructive {
			t.Error("logs should not be destructive")
		}
	}
	_ = ast.AttrConfirm
}
