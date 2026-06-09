package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/runtime/agent"
	"github.com/rune-task-runner/rune/internal/runtime/shell"
	"github.com/rune-task-runner/rune/mcpserver"
)

// executeAgent runs an (agent) task: it interpolates the prompt body, starts an
// in-process MCP endpoint exposing the project's non-destructive tasks for the
// agent to call back into, drives the configured agent CLI, and writes the
// agent's final output to stdout.
func (e *engine) executeAgent(task *ast.Task, lines []shell.Line, dir string, env []string) error {
	var prompt strings.Builder
	for _, ln := range lines {
		prompt.WriteString(ln.Text)
		prompt.WriteByte('\n')
	}

	opts := agent.Options{
		AgentCmd: e.settings.AgentCmd,
		Dir:      dir,
		Env:      env,
		Stdin:    e.opts.Stdin,
		Stderr:   e.opts.Stderr,
	}

	// Best-effort: expose allowed (non-destructive, non-private) tasks to the
	// agent over a localhost, token-gated MCP endpoint.
	if e.file != nil {
		if adapter := e.newAgentAdapter(); adapter != nil {
			token := randomToken()
			srv := mcpserver.New(adapter, mcpserver.Options{AllowDestructive: false, Version: e.opts.Version})
			if addr, stop, err := srv.StartHTTP(mcpserver.HTTPConfig{Addr: "127.0.0.1:0", Token: token}); err == nil {
				defer stop()
				opts.MCPURL = "http://" + addr
				opts.MCPToken = token
			}
		}
	}

	finalText, err := agent.CLIProvider{}.Run(e.ctx, prompt.String(), opts)
	if err != nil {
		return err
	}
	if finalText != "" {
		fmt.Fprint(e.opts.Stdout, finalText)
		if !strings.HasSuffix(finalText, "\n") {
			fmt.Fprintln(e.opts.Stdout)
		}
	}
	return nil
}

// newAgentAdapter builds an MCP engine adapter from the engine's module state.
func (e *engine) newAgentAdapter() *mcpAdapter {
	return &mcpAdapter{
		file:      e.file,
		tasks:     e.tasks,
		assigns:   e.assigns,
		settings:  e.settings,
		root:      e.root,
		workDir:   e.workDir,
		baseEnv:   e.env,
		overrides: e.overrides,
		now:       e.now,
	}
}

func randomToken() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
