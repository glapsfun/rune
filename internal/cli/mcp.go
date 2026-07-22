package cli

import (
	"bytes"
	"context"
	"runtime"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/eval"
	"github.com/rune-task-runner/rune/internal/runtime/scheduler"
	"github.com/rune-task-runner/rune/mcpserver"
)

// mcpAdapter implements mcpserver.Engine over a parsed Runefile, running each
// tool call through the same scheduler the CLI uses (FR-026) with output
// captured into buffers.
type mcpAdapter struct {
	file      *ast.File
	tasks     map[string]*ast.Task
	assigns   map[string]*ast.Assignment
	settings  config.Settings
	root      string
	workDir   string
	baseEnv   []string
	overrides map[string]string
	now       func() string
}

// Tasks returns the non-private tasks as agent-facing tool descriptors. No
// secret values appear in any field (FR-029).
func (a *mcpAdapter) Tasks() []mcpserver.TaskInfo {
	var out []mcpserver.TaskInfo
	for _, t := range a.file.Tasks {
		if t.IsPrivate() {
			continue
		}
		info := mcpserver.TaskInfo{
			Name:        t.Name,
			Doc:         t.Doc,
			Destructive: t.Attr(ast.AttrConfirm) != nil,
			Network:     t.Attr(ast.AttrNetwork) != nil,
		}
		for _, p := range t.Params {
			info.Params = append(info.Params, mcpserver.ParamInfo{
				Name:     p.Name,
				Required: p.Kind == ast.ParamRequired || p.Kind == ast.ParamVariadicPlus,
				Variadic: p.Kind == ast.ParamVariadicPlus || p.Kind == ast.ParamVariadicStar,
			})
		}
		out = append(out, info)
	}
	return out
}

// Call runs a task by name with named arguments, capturing stdout/stderr/exit.
func (a *mcpAdapter) Call(ctx context.Context, name string, args map[string]string) (mcpserver.Result, error) {
	t, ok := a.tasks[name]
	if !ok {
		return mcpserver.Result{}, errorf("unknown task: %s", name)
	}
	var outBuf, errBuf bytes.Buffer
	scope := eval.NewScope(a.assigns, a.overrides)
	scope.GOOS = runtime.GOOS
	scope.Arch = runtime.GOARCH

	// The same masking choke point as the CLI path: the buffers only ever hold
	// masked text, so the tool result an agent receives is safe by construction.
	mopts, flushMask := applyMasking(
		Options{Stdin: nil, Stdout: &outBuf, Stderr: &errBuf, Cwd: a.workDir, Quiet: true},
		a.baseEnv, a.tasks, a.settings.Secrets, a.settings.Unmasked,
	)

	eng := &engine{
		tasks:    a.tasks,
		scope:    scope,
		settings: a.settings,
		workDir:  a.workDir,
		root:     a.root,
		env:      a.baseEnv,
		opts:     mopts,
		plan:     planRun,
		now:      a.now,
		ctx:      ctx,
	}

	params, err := bindNamedParams(t, args, scope)
	if err != nil {
		return mcpserver.Result{}, err
	}

	runErr := scheduler.Run(eng, []scheduler.Invocation{{Task: t, Params: params}})
	flushMask()
	code := ExitSuccess
	if runErr != nil {
		code = CodeFor(eng.classifyRunErr(runErr))
	}
	return mcpserver.Result{Stdout: outBuf.String(), Stderr: errBuf.String(), ExitCode: code}, nil
}
