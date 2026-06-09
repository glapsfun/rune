// Package agent runs (agent) task bodies by driving an installed agent CLI
// (claude/codex/copilot/…) behind a vendor-neutral Provider interface
// (Principle VII / FR-030). The CLI owns its own authentication; no API keys are
// ever read from the Runefile.
package agent

import (
	"context"
	"io"
)

// Options configures an agent run.
type Options struct {
	AgentCmd     []string // e.g. ["claude", "-p"]; the prompt is appended
	AllowedTasks []string // tasks the agent may call back into (non-destructive by default)
	MCPURL       string   // in-process MCP endpoint the agent can call (if any)
	MCPToken     string   // bearer token for the MCP endpoint
	Dir          string
	Env          []string
	Stdin        io.Reader
	Stderr       io.Writer // agent progress/diagnostics stream here
}

// Provider runs a prompt and returns the agent's final text output. The
// interface stays vendor-neutral so a direct hosted-API provider can be added
// later without changing core.
type Provider interface {
	Run(ctx context.Context, prompt string, opts Options) (finalText string, err error)
}

// NotConfiguredError reports that no agent command is configured (never invent
// credentials, FR-027).
type NotConfiguredError struct{}

func (e *NotConfiguredError) Error() string {
	return "no agent command configured; set one, e.g. `set agent_cmd := [\"claude\", \"-p\"]` (credentials are never invented)"
}

// NotInstalledError reports that the configured agent CLI is not on PATH.
type NotInstalledError struct{ Name string }

func (e *NotInstalledError) Error() string {
	return "agent CLI " + e.Name + " not found on PATH — install it and authenticate it (e.g. log in via the CLI)"
}

// AuthError reports that the agent CLI ran but failed, commonly because it is
// not authenticated.
type AuthError struct {
	Name string
	Err  error
}

func (e *AuthError) Error() string {
	return "agent CLI " + e.Name + " failed (is it authenticated? run its login): " + e.Err.Error()
}

func (e *AuthError) Unwrap() error { return e.Err }
