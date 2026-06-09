package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/rune-task-runner/rune/internal/analyzer"
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/cache"
	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/dotenv"
	"github.com/rune-task-runner/rune/internal/eval"
	"github.com/rune-task-runner/rune/internal/parser"
	rt "github.com/rune-task-runner/rune/internal/runtime"
	"github.com/rune-task-runner/rune/internal/runtime/agent"
	"github.com/rune-task-runner/rune/internal/runtime/interp"
	"github.com/rune-task-runner/rune/internal/runtime/scheduler"
	"github.com/rune-task-runner/rune/internal/runtime/shell"
	"github.com/rune-task-runner/rune/internal/token"
)

// execute is the inner pipeline for a resolved Runefile: parse → analyze →
// schedule → run.
func execute(opts Options, runefile string, args []string) error {
	root := filepath.Dir(runefile)

	// --fmt rewrites the Runefile canonically and exits (runs nothing).
	if opts.Fmt {
		return fmtRewrite(opts, runefile)
	}

	// --clear-cache removes the project-local cache and exits.
	if opts.ClearCache {
		if err := cache.Clear(root); err != nil {
			return &UsageError{Err: err}
		}
		fmt.Fprintln(opts.Stderr, "cleared: "+filepath.Join(root, ".rune", "cache"))
		return nil
	}

	source, err := os.ReadFile(runefile)
	if err != nil {
		return &UsageError{Err: err}
	}

	file, diags := parser.Parse(runefile, string(source))
	srcProvider := newSourceProvider(runefile, source)

	// Splice imports and namespace submodules before analysis.
	cdiags := config.Compose(file, srcProvider)
	diags = append(diags, cdiags...)

	// Lexer/parser errors flow through the same spanned diagnostic path (T036).
	if diags.HasErrors() {
		renderDiags(opts, diags, srcProvider)
		return &ValidationError{Err: errorf("%d static error(s)", countErrors(diags))}
	}

	// Whole-file semantic analysis runs before any execution: emit ALL
	// diagnostics, run nothing, exit 3 (Principle II / FR-014).
	if adiags := analyzer.Analyze(file); adiags.HasErrors() {
		renderDiags(opts, adiags, srcProvider)
		return &ValidationError{Err: errorf("%d static error(s)", countErrors(adiags))}
	}

	if opts.Dump {
		return dumpFile(opts, file)
	}

	tasks := indexTasks(file)
	assigns := indexAssignments(file)

	overrides, rawInvs, err := splitArgs(args, tasks)
	if err != nil {
		return err
	}

	scope := eval.NewScope(assigns, overrides)
	scope.GOOS = runtime.GOOS
	scope.Arch = runtime.GOARCH

	settings, sdiags := config.ResolveSettings(file, eval.New(scope))
	if sdiags.HasErrors() {
		renderDiags(opts, sdiags, srcProvider)
		return &ValidationError{Err: errorf("invalid settings")}
	}

	// Listing and inspection short-circuits (run nothing).
	if opts.List {
		listTasks(opts, file)
		return nil
	}

	workDir := resolveWorkDir(runefile, settings.WorkingDir)
	env := buildEnv(settings, scope, root)

	plan := planRun
	switch {
	case opts.DryRun:
		plan = planDryRun
	case opts.Summary:
		plan = planSummary
	}

	eng := &engine{
		file:      file,
		tasks:     tasks,
		assigns:   assigns,
		overrides: overrides,
		scope:     scope,
		settings:  settings,
		workDir:   workDir,
		root:      root,
		env:       env,
		opts:      opts,
		plan:      plan,
		now:       func() string { return time.Now().UTC().Format(time.RFC3339) },
		ctx:       opts.ctx(),
		src:       srcProvider,
	}

	invs, err := eng.resolveRoots(rawInvs, settings)
	if err != nil {
		return err
	}

	if err := scheduler.Run(eng, invs); err != nil {
		return eng.classifyRunErr(err)
	}
	return nil
}

// planMode controls whether Execute runs, or only reports, each task.
type planMode int

const (
	planRun     planMode = iota
	planDryRun           // --dry-run: print the plan + would-be cache decision
	planSummary          // --summary: print task names, one per line
)

// engine implements scheduler.Engine over the parsed module.
type engine struct {
	file      *ast.File
	tasks     map[string]*ast.Task
	assigns   map[string]*ast.Assignment
	overrides map[string]string
	scope     *eval.Scope
	settings  config.Settings
	workDir   string
	root      string // directory containing the Runefile (cache root)
	env       []string
	opts      Options
	plan      planMode
	now       func() string
	ctx       context.Context
	src       diag.SourceProvider
}

// resolveRoots turns CLI invocations (or the default task) into scheduler roots.
func (e *engine) resolveRoots(raw []rawInvocation, settings config.Settings) ([]scheduler.Invocation, error) {
	if len(raw) == 0 {
		if settings.Default == "" {
			return nil, usagef("no task specified and no default task set; run with --list to see tasks")
		}
		if _, ok := e.tasks[settings.Default]; !ok {
			return nil, usagef("default task %q is not defined", settings.Default)
		}
		raw = []rawInvocation{{name: settings.Default}}
	}
	var invs []scheduler.Invocation
	for _, r := range raw {
		t := e.tasks[r.name] // existence already checked in splitArgs
		params, err := bindParams(t, r.args, e.scope)
		if err != nil {
			return nil, err
		}
		invs = append(invs, scheduler.Invocation{Task: t, Params: params})
	}
	return invs, nil
}

// ResolveDep evaluates a dependency call in the caller's scope and binds args.
func (e *engine) ResolveDep(curTask *ast.Task, curParams map[string]string, dep *ast.DepCall) (*ast.Task, map[string]string, error) {
	target, ok := e.tasks[dep.Name]
	if !ok {
		return nil, nil, &ValidationError{Err: errorf("unknown dependency %q of task %q", dep.Name, curTask.Name)}
	}
	ev := eval.New(e.scope.WithParams(curParams))
	pos := make([]string, len(dep.Args))
	for i, a := range dep.Args {
		v, evErr := ev.Eval(a)
		if evErr != nil {
			return nil, nil, &ValidationError{Err: evErr}
		}
		pos[i] = v
	}
	params, err := bindParams(target, pos, e.scope)
	if err != nil {
		return nil, nil, err
	}
	return target, params, nil
}

// Execute handles plan modes, confirmation, and caching, then runs the body.
func (e *engine) Execute(task *ast.Task, params map[string]string) error {
	if e.plan == planSummary {
		fmt.Fprintln(e.opts.Stdout, task.Name)
		return nil
	}

	if c := task.Attr(ast.AttrConfirm); c != nil && e.plan == planRun {
		if !e.opts.Yes && !e.confirm(task, c.Str) {
			return &TaskFailure{Err: errorf("task %q was not confirmed", task.Name), Silent: true}
		}
	}

	cacheAttr := task.Attr(ast.AttrCache)

	if e.plan == planDryRun {
		label := "would run"
		if cacheAttr != nil {
			if d, err := cache.Decide(e.cacheSpec(task, params, cacheAttr)); err == nil && d.Skip {
				label = "would skip (cached)"
			}
		}
		fmt.Fprintf(e.opts.Stderr, "%s: %s\n", label, task.Name)
		return nil
	}

	if cacheAttr != nil {
		spec := e.cacheSpec(task, params, cacheAttr)
		d, derr := cache.Decide(spec)
		if derr == nil && d.Skip {
			fmt.Fprintf(e.opts.Stderr, "cached: %s\n", task.Name)
			return nil
		}
		fmt.Fprintf(e.opts.Stderr, "running: %s\n", task.Name)
		if err := e.runBody(task, params); err != nil {
			return err
		}
		if derr == nil {
			if cerr := cache.Store(spec, d.Hash, e.now()); cerr != nil {
				fmt.Fprintf(e.opts.Stderr, "warning: failed to write cache for %s: %v\n", task.Name, cerr)
			}
		}
		return nil
	}

	return e.runBody(task, params)
}

// runBody interpolates and runs a task body via the selected executor.
func (e *engine) runBody(task *ast.Task, params map[string]string) error {
	ev := eval.New(e.scope.WithParams(params))

	lines := make([]shell.Line, 0, len(task.Body))
	for _, bl := range task.Body {
		text, evErr := ev.Interpolate(bl.Raw, bl.Sp)
		if evErr != nil {
			return &ValidationError{Err: evErr}
		}
		lines = append(lines, shell.Line{
			Text:            text,
			NoEcho:          bl.NoEcho,
			ContinueOnError: bl.ContinueOnError,
			Span:            bl.Sp,
		})
	}

	dir := e.taskDir(task)
	env := e.taskEnv(task)

	sel := rt.Select(task, e.settings)
	var err error
	switch sel.Kind {
	case rt.KindAgent:
		err = e.executeAgent(task, lines, dir, env)
	case rt.KindShell:
		err = shell.Run(e.ctx, task.Name, lines, shell.Options{
			Stdin:  e.opts.Stdin,
			Stdout: e.opts.Stdout,
			Stderr: e.opts.Stderr,
			Dir:    dir,
			Env:    env,
			Quiet:  e.settings.Quiet || e.opts.Quiet,
		})
	case rt.KindInterp:
		script := joinBody(lines)
		var span token.Span
		if len(task.Body) > 0 {
			span = task.Body[0].Sp
		}
		err = interp.Run(e.ctx, task.Name, script, sel.Command, span, interp.Options{
			Stdin:  e.opts.Stdin,
			Stdout: e.opts.Stdout,
			Stderr: e.opts.Stderr,
			Dir:    dir,
			Env:    env,
		})
	default:
		return &UsageError{Err: errorf("executor %q is not supported yet", sel.Display)}
	}

	// [no-exit-message] suppresses the trailing error banner (not the exit code).
	if err != nil && task.Attr(ast.AttrNoExitMessage) != nil {
		return &TaskFailure{Err: err, Silent: true}
	}
	return err
}

// confirm prompts the operator to approve a destructive ([confirm]) task. It
// returns false on any non-affirmative answer or when stdin is unavailable.
func (e *engine) confirm(task *ast.Task, prompt string) bool {
	if e.opts.Stdin == nil {
		return false
	}
	if prompt == "" {
		prompt = fmt.Sprintf("Run %q?", task.Name)
	}
	fmt.Fprintf(e.opts.Stderr, "%s [y/N] ", prompt)
	reader := bufio.NewReader(e.opts.Stdin)
	line, _ := reader.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

// taskDir resolves the working directory for a task, honoring [no-cd] (run in
// the invocation directory) and [working-directory("path")] (relative to root).
func (e *engine) taskDir(task *ast.Task) string {
	if task.Attr(ast.AttrNoCD) != nil {
		if e.opts.Cwd != "" {
			return e.opts.Cwd
		}
		return e.workDir
	}
	if wd := task.Attr(ast.AttrWorkingDirectory); wd != nil && wd.Str != "" {
		if filepath.IsAbs(wd.Str) {
			return wd.Str
		}
		return filepath.Join(e.root, wd.Str)
	}
	return e.workDir
}

// taskEnv appends any [env("NAME","VALUE")] attributes to the base environment.
func (e *engine) taskEnv(task *ast.Task) []string {
	var extra []string
	for _, a := range task.Attributes {
		if a.Kind == ast.AttrEnv && a.Str != "" {
			extra = append(extra, a.Str+"="+a.Str2)
		}
	}
	if len(extra) == 0 {
		return e.env
	}
	return append(append([]string{}, e.env...), extra...)
}

// joinBody concatenates interpolated body lines into a single interpreter script.
func joinBody(lines []shell.Line) string {
	var b strings.Builder
	for _, ln := range lines {
		b.WriteString(ln.Text)
		b.WriteByte('\n')
	}
	return b.String()
}

func (e *engine) Namespace(_ *ast.Task) string { return "" }

// cacheSpec builds the cache fingerprint spec for a [cache] task.
func (e *engine) cacheSpec(task *ast.Task, params map[string]string, attr *ast.Attribute) cache.Spec {
	ev := eval.New(e.scope.WithParams(params))
	evalAll := func(exprs []ast.Expr) []string {
		out := make([]string, 0, len(exprs))
		for _, ex := range exprs {
			if v, err := ev.Eval(ex); err == nil {
				out = append(out, v)
			}
		}
		return out
	}
	var body strings.Builder
	for _, bl := range task.Body {
		body.WriteString(bl.Raw)
		body.WriteByte('\n')
	}
	sel := rt.Select(task, e.settings)
	execID := sel.Display
	if len(sel.Command) > 0 {
		execID += ":" + strings.Join(sel.Command, " ")
	}
	return cache.Spec{
		Key:      task.Name,
		Root:     e.root,
		Inputs:   evalAll(attr.Inputs),
		Outputs:  evalAll(attr.Outputs),
		Body:     body.String(),
		Vars:     e.collectVars(task, params),
		Executor: execID,
	}
}

// collectVars resolves the values of every variable/param the task references
// (Principle I: the fingerprint must reflect interpolated values).
func (e *engine) collectVars(task *ast.Task, params map[string]string) map[string]string {
	names := map[string]bool{}
	var walk func(ex ast.Expr)
	walk = func(ex ast.Expr) {
		switch x := ex.(type) {
		case *ast.VarRef:
			names[x.Name] = true
		case *ast.Binary:
			walk(x.Left)
			walk(x.Right)
		case *ast.FuncCall:
			for _, a := range x.Args {
				walk(a)
			}
		case *ast.Conditional:
			for _, br := range x.Branches {
				walk(br.Left)
				walk(br.Right)
				walk(br.Result)
			}
			walk(x.Else)
		}
	}
	for _, p := range task.Params {
		walk(p.Default)
	}
	for _, dep := range task.Deps {
		for _, a := range dep.Args {
			walk(a)
		}
	}
	for _, bl := range task.Body {
		for _, frag := range bodyInterpFragments(bl.Raw) {
			if ex, d := parser.ParseExprFragment(e.root, frag); !d.HasErrors() {
				walk(ex)
			}
		}
	}
	ev := eval.New(e.scope.WithParams(params))
	out := map[string]string{}
	for name := range names {
		if v, err := ev.Eval(&ast.VarRef{Name: name}); err == nil {
			out[name] = v
		}
	}
	return out
}

// bodyInterpFragments extracts the {{ ... }} expression texts from a body line.
func bodyInterpFragments(raw string) []string {
	var out []string
	i := 0
	for i < len(raw) {
		if strings.HasPrefix(raw[i:], "{{{{") || strings.HasPrefix(raw[i:], "}}}}") {
			i += 4
			continue
		}
		if strings.HasPrefix(raw[i:], "{{") {
			end := strings.Index(raw[i+2:], "}}")
			if end < 0 {
				break
			}
			out = append(out, raw[i+2:i+2+end])
			i = i + 2 + end + 2
			continue
		}
		i++
	}
	return out
}

// classifyRunErr maps a scheduler/execution error to a CLI error class, first
// rendering any spanned diagnostic.
func (e *engine) classifyRunErr(err error) error {
	if err == nil {
		return nil
	}
	// An error already classified by runBody (e.g. a Silent [no-exit-message]
	// failure or a declined confirmation) passes through unchanged.
	var already *TaskFailure
	if errors.As(err, &already) {
		return already
	}
	var evErr *eval.Error
	if errors.As(err, &evErr) {
		renderDiags(e.opts, diag.List{diag.New(evErr.Span, evErr.Msg)}, e.src)
		return &ValidationError{Err: err}
	}
	var cyc *scheduler.CycleError
	if errors.As(err, &cyc) {
		fmt.Fprintln(e.opts.Stderr, "rune: "+cyc.Error())
		return &ValidationError{Err: err}
	}
	var exec *shell.ExecError
	if errors.As(err, &exec) {
		return &TaskFailure{Err: err}
	}
	var notConfigured *agent.NotConfiguredError
	if errors.As(err, &notConfigured) {
		return &UsageError{Err: err}
	}
	var notInstalled *agent.NotInstalledError
	if errors.As(err, &notInstalled) {
		return &TaskFailure{Err: err}
	}
	var authErr *agent.AuthError
	if errors.As(err, &authErr) {
		return &TaskFailure{Err: err}
	}
	if errors.Is(err, context.Canceled) {
		return &Interrupted{Err: err}
	}
	var ue *UsageError
	if errors.As(err, &ue) {
		return err
	}
	var ve *ValidationError
	if errors.As(err, &ve) {
		return err
	}
	return &TaskFailure{Err: err}
}

// --- module helpers ---

func indexTasks(f *ast.File) map[string]*ast.Task {
	m := make(map[string]*ast.Task, len(f.Tasks))
	for _, t := range f.Tasks {
		m[t.Name] = t
	}
	return m
}

func indexAssignments(f *ast.File) map[string]*ast.Assignment {
	m := make(map[string]*ast.Assignment, len(f.Assignments))
	for _, a := range f.Assignments {
		m[a.Name] = a
	}
	return m
}

func resolveWorkDir(runefile, setting string) string {
	base := filepath.Dir(runefile)
	if setting == "" {
		return base
	}
	if filepath.IsAbs(setting) {
		return setting
	}
	return filepath.Join(base, setting)
}

// buildEnv assembles the task environment: the inherited process environment,
// any `set dotenv` file, plus — when `set export` is active — every
// successfully-resolved module variable.
func buildEnv(settings config.Settings, scope *eval.Scope, root string) []string {
	env := os.Environ()
	if settings.Dotenv != "" {
		path := settings.Dotenv
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		if pairs, err := dotenv.Load(path); err == nil {
			env = append(env, pairs...)
		}
	}
	if settings.Export {
		ev := eval.New(scope)
		names := make([]string, 0, len(scope.Assigns))
		for name := range scope.Assigns {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if v, err := ev.Eval(scope.Assigns[name].Expr); err == nil {
				env = append(env, name+"="+v)
			}
		}
	}
	return env
}

// listTasks prints non-private tasks, grouped by [group("...")], excluding tasks
// filtered out by an OS attribute that does not match the current platform.
func listTasks(opts Options, f *ast.File) {
	type row struct{ name, doc string }
	groups := map[string][]row{}
	var order []string
	width := 0
	for _, t := range f.Tasks {
		if t.IsPrivate() || !osMatches(t, runtime.GOOS) {
			continue
		}
		g := ""
		if ga := t.Attr(ast.AttrGroup); ga != nil {
			g = ga.Str
		}
		if _, ok := groups[g]; !ok {
			order = append(order, g)
		}
		groups[g] = append(groups[g], row{t.Name, firstLine(t.Doc)})
		if len(t.Name) > width {
			width = len(t.Name)
		}
	}

	fmt.Fprintln(opts.Stdout, "Available tasks:")
	for _, g := range order {
		if g != "" {
			fmt.Fprintf(opts.Stdout, "  [%s]\n", g)
		}
		for _, r := range groups[g] {
			if r.doc != "" {
				fmt.Fprintf(opts.Stdout, "    %-*s  # %s\n", width, r.name, r.doc)
			} else {
				fmt.Fprintf(opts.Stdout, "    %s\n", r.name)
			}
		}
	}
}

// osMatches reports whether a task's OS-filter attributes (if any) include the
// current OS. A task with no OS attribute is always available.
func osMatches(t *ast.Task, goos string) bool {
	var filters []string
	for _, a := range t.Attributes {
		switch a.Kind {
		case ast.AttrLinux, ast.AttrMacos, ast.AttrWindows, ast.AttrUnix:
			filters = append(filters, a.Kind)
		}
	}
	if len(filters) == 0 {
		return true
	}
	for _, f := range filters {
		switch f {
		case ast.AttrLinux:
			if goos == "linux" {
				return true
			}
		case ast.AttrMacos:
			if goos == "darwin" {
				return true
			}
		case ast.AttrWindows:
			if goos == "windows" {
				return true
			}
		case ast.AttrUnix:
			if goos != "windows" {
				return true
			}
		}
	}
	return false
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// --- diagnostics ---

func newSourceProvider(mainPath string, mainSrc []byte) diag.SourceProvider {
	cache := map[string][]byte{mainPath: mainSrc}
	return func(path string) ([]byte, bool) {
		if b, ok := cache[path]; ok {
			return b, true
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, false
		}
		cache[path] = b
		return b, true
	}
}

func renderDiags(opts Options, diags diag.List, src diag.SourceProvider) {
	fmt.Fprintln(opts.Stderr, diag.RenderAll(diags, src, opts.Color))
}

func countErrors(diags diag.List) int {
	n := 0
	for _, d := range diags {
		if d.Severity == diag.Error {
			n++
		}
	}
	return n
}
