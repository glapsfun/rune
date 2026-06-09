package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rune-task-runner/rune/internal/analyzer"
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/eval"
	"github.com/rune-task-runner/rune/internal/parser"
	"github.com/rune-task-runner/rune/mcpserver"
)

// loadedModule is the parsed, composed, analyzed module plus its resolved
// scope/settings — shared by the MCP server and agent paths.
type loadedModule struct {
	file     *ast.File
	tasks    map[string]*ast.Task
	assigns  map[string]*ast.Assignment
	settings config.Settings
	root     string
	workDir  string
	baseEnv  []string
	scope    *eval.Scope
}

// loadModule resolves a Runefile statically (parse → compose → analyze) and
// builds its scope, settings, and environment. Diagnostics are rendered and a
// ValidationError returned on failure (exit 3).
func loadModule(opts Options, runefile string) (*loadedModule, error) {
	root := filepath.Dir(runefile)
	source, err := os.ReadFile(runefile)
	if err != nil {
		return nil, &UsageError{Err: err}
	}
	file, diags := parser.Parse(runefile, string(source))
	src := newSourceProvider(runefile, source)
	diags = append(diags, config.Compose(file, src)...)
	if diags.HasErrors() {
		renderDiags(opts, diags, src)
		return nil, &ValidationError{Err: errorf("%d static error(s)", countErrors(diags))}
	}
	if adiags := analyzer.Analyze(file); adiags.HasErrors() {
		renderDiags(opts, adiags, src)
		return nil, &ValidationError{Err: errorf("%d static error(s)", countErrors(adiags))}
	}

	scope := eval.NewScope(indexAssignments(file), map[string]string{})
	scope.GOOS = runtime.GOOS
	scope.Arch = runtime.GOARCH
	settings, sdiags := config.ResolveSettings(file, eval.New(scope))
	if sdiags.HasErrors() {
		renderDiags(opts, sdiags, src)
		return nil, &ValidationError{Err: errorf("invalid settings")}
	}

	return &loadedModule{
		file:     file,
		tasks:    indexTasks(file),
		assigns:  indexAssignments(file),
		settings: settings,
		root:     root,
		workDir:  resolveWorkDir(runefile, settings.WorkingDir),
		baseEnv:  buildEnv(settings, scope, root),
		scope:    scope,
	}, nil
}

// newMCPAdapter builds an mcpserver.Engine from a resolved Runefile.
func newMCPAdapter(opts Options, runefile string) (*mcpAdapter, error) {
	mod, err := loadModule(opts, runefile)
	if err != nil {
		return nil, err
	}
	return &mcpAdapter{
		file:      mod.file,
		tasks:     mod.tasks,
		assigns:   mod.assigns,
		settings:  mod.settings,
		root:      mod.root,
		workDir:   mod.workDir,
		baseEnv:   mod.baseEnv,
		overrides: map[string]string{},
		now:       func() string { return "" },
	}, nil
}

// ServeMCP starts the MCP server over stdio (default) or Streamable HTTP.
func ServeMCP(opts Options, useHTTP bool, addr, tokenFile string) error {
	runefile, err := config.Resolve(opts.File, opts.Cwd)
	if err != nil {
		return &UsageError{Err: err}
	}
	adapter, err := newMCPAdapter(opts, runefile)
	if err != nil {
		return err
	}
	srv := mcpserver.New(adapter, mcpserver.Options{
		AllowDestructive: opts.Yes,
		Version:          opts.Version,
	})

	ctx := opts.ctx()
	if useHTTP {
		token, err := readToken(tokenFile)
		if err != nil {
			return &UsageError{Err: err}
		}
		if addr == "" {
			addr = "127.0.0.1:7777"
		}
		fmt.Fprintf(opts.Stderr, "rune MCP server on http://%s (token required)\n", addr)
		return srv.ServeHTTP(ctx, mcpserver.HTTPConfig{Addr: addr, Token: token})
	}
	fmt.Fprintln(opts.Stderr, "rune MCP server on stdio")
	return srv.ServeStdio(ctx)
}

func readToken(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("the HTTP transport requires --token-file PATH")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	tok := strings.TrimSpace(string(data))
	if tok == "" {
		return "", fmt.Errorf("token file %q is empty", path)
	}
	return tok, nil
}
