package runtime

import (
	"testing"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/config"
)

func TestSelectDefaultShell(t *testing.T) {
	sel := Select(&ast.Task{Name: "x"}, config.Settings{})
	if sel.Kind != KindShell {
		t.Errorf("kind = %v, want KindShell", sel.Kind)
	}
}

func TestSelectPythonNode(t *testing.T) {
	py := Select(&ast.Task{Name: "a", Executor: "python"}, config.Settings{})
	if py.Kind != KindInterp || py.Command[0] != "python3" {
		t.Errorf("python sel = %+v", py)
	}
	nd := Select(&ast.Task{Name: "b", Executor: "node"}, config.Settings{})
	if nd.Kind != KindInterp || nd.Command[0] != "node" {
		t.Errorf("node sel = %+v", nd)
	}
}

func TestSelectPythonOverride(t *testing.T) {
	sel := Select(&ast.Task{Name: "a", Executor: "python"}, config.Settings{Python: []string{"python3.12", "-u"}})
	if sel.Command[0] != "python3.12" || len(sel.Command) != 2 {
		t.Errorf("override = %+v", sel)
	}
}

func TestSelectShellOverride(t *testing.T) {
	// `set shell := ["bash","-cu"]` switches even default bodies to temp-file exec.
	sel := Select(&ast.Task{Name: "a"}, config.Settings{Shell: []string{"bash", "-cu"}})
	if sel.Kind != KindInterp || sel.Command[0] != "bash" {
		t.Errorf("shell override = %+v", sel)
	}
}

func TestSelectScriptAttribute(t *testing.T) {
	task := &ast.Task{Name: "a", Attributes: []*ast.Attribute{{Kind: ast.AttrScript, Str: "ruby -w"}}}
	sel := Select(task, config.Settings{})
	if sel.Kind != KindInterp || sel.Command[0] != "ruby" || sel.Command[1] != "-w" {
		t.Errorf("script attr = %+v", sel)
	}
}

func TestSelectCustomExecutor(t *testing.T) {
	sel := Select(&ast.Task{Name: "a", Executor: "ruby"}, config.Settings{})
	if sel.Kind != KindInterp || sel.Command[0] != "ruby" {
		t.Errorf("custom = %+v", sel)
	}
}

func TestSelectAgent(t *testing.T) {
	sel := Select(&ast.Task{Name: "a", Executor: "agent"}, config.Settings{})
	if sel.Kind != KindAgent {
		t.Errorf("agent kind = %v", sel.Kind)
	}
}
