package cli

import (
	"bytes"
	"context"
	"runtime"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/eval"
	"github.com/rune-task-runner/rune/internal/mask"
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
	maskSet   *mask.Set // derived once; env/tasks/settings are fixed per adapter
	goos      string    // host OS for availability checks; runtime.GOOS outside tests
}

// Tasks returns the non-private tasks available on this host OS as
// agent-facing tool descriptors, so an agent can never see (or attempt) a
// platform-incompatible task. No secret values appear in any field (FR-029).
func (a *mcpAdapter) Tasks() []mcpserver.TaskInfo {
	var out []mcpserver.TaskInfo
	for _, t := range a.file.Tasks {
		if !visibleOn(t, a.goos) {
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
	// Defense-in-depth for direct Engine.Call users (embedders, tests): over
	// the real MCP transport a mismatched task is never registered as a tool,
	// so the SDK rejects it as unknown before this line; the check here keeps
	// the Engine contract safe for callers that bypass tool registration.
	if !t.AvailableOn(a.goos) {
		return mcpserver.Result{}, availabilityErr(name, t, a.goos)
	}
	var outBuf, errBuf bytes.Buffer
	scope := eval.NewScope(a.assigns, a.overrides)
	// One host-OS truth per adapter: availability (above), dependency
	// skipping (eng.goos below), and the os()/os_family() builtins must
	// never disagree, so the scope reads the same injected value.
	scope.GOOS = a.goos
	scope.Arch = runtime.GOARCH

	// The same masking choke point as the CLI path: the buffers only ever hold
	// masked text, so the tool result an agent receives is safe by construction.
	mopts, flushMask := maskOptions(
		Options{Stdin: nil, Stdout: &outBuf, Stderr: &errBuf, Cwd: a.workDir, Quiet: true},
		a.maskSet,
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
		goos:     a.goos,
	}

	params, err := bindNamedParams(t, args, scope)
	if err != nil {
		return mcpserver.Result{}, err
	}

	runErr := scheduler.Run(eng, []scheduler.Invocation{{Task: t, Params: params}})
	code := ExitSuccess
	if runErr != nil {
		// classifyRunErr can render diagnostics into the masked stderr, so it
		// must precede the flush that drains the writers into the buffers.
		code = CodeFor(eng.classifyRunErr(runErr))
	}
	flushMask()
	return mcpserver.Result{Stdout: outBuf.String(), Stderr: errBuf.String(), ExitCode: code}, nil
}
