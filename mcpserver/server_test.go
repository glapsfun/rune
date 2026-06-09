package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeEngine struct {
	tasks    []TaskInfo
	lastCall string
	lastArgs map[string]string
	result   Result
}

func (f *fakeEngine) Tasks() []TaskInfo { return f.tasks }

func (f *fakeEngine) Call(_ context.Context, name string, args map[string]string) (Result, error) {
	f.lastCall = name
	f.lastArgs = args
	return f.result, nil
}

func sampleEngine() *fakeEngine {
	return &fakeEngine{
		tasks: []TaskInfo{
			{Name: "logs", Doc: "Show recent git log."},
			{Name: "greet", Doc: "Greet someone.", Params: []ParamInfo{{Name: "name", Required: true}}},
			{Name: "clean", Doc: "Remove build output.", Destructive: true},
			{Name: "fetch", Doc: "Fetch a URL.", Network: true},
			{Name: "docker::push", Doc: "Push image."},
		},
		result: Result{Stdout: "output-here", ExitCode: 0},
	}
}

func TestToolNameNamespacing(t *testing.T) {
	if got := toolName("docker::push"); got != "docker__push" {
		t.Errorf("toolName = %q, want docker__push", got)
	}
	if got := toolName("plain"); got != "plain" {
		t.Errorf("toolName = %q, want plain", got)
	}
}

func TestInputSchema(t *testing.T) {
	schema := inputSchema([]ParamInfo{
		{Name: "a", Required: true},
		{Name: "b"},
		{Name: "rest", Variadic: true},
	})
	props := schema["properties"].(map[string]any)
	if props["a"] == nil || props["b"] == nil {
		t.Fatalf("missing properties: %v", props)
	}
	if rest := props["rest"].(map[string]any); rest["type"] != "array" {
		t.Errorf("variadic param should be an array, got %v", rest)
	}
	req, _ := schema["required"].([]string)
	if len(req) != 1 || req[0] != "a" {
		t.Errorf("required = %v, want [a]", req)
	}
}

func TestToolAnnotations(t *testing.T) {
	clean := toolFor(TaskInfo{Name: "clean", Destructive: true})
	if clean.Annotations.DestructiveHint == nil || !*clean.Annotations.DestructiveHint {
		t.Error("clean should have destructiveHint=true")
	}
	fetch := toolFor(TaskInfo{Name: "fetch", Network: true})
	if fetch.Annotations.OpenWorldHint == nil || !*fetch.Annotations.OpenWorldHint {
		t.Error("fetch should have openWorldHint=true")
	}
}

// connect wires a client to the server over an in-memory transport.
func connect(t *testing.T, srv *Server) *mcp.ClientSession {
	t.Helper()
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	if _, err := srv.MCP().Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v1"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

func TestListToolsMapping(t *testing.T) {
	eng := sampleEngine()
	srv := New(eng, Options{AllowDestructive: true})
	cs := connect(t, srv)

	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]*mcp.Tool{}
	for _, tl := range res.Tools {
		byName[tl.Name] = tl
	}
	if byName["logs"] == nil || byName["logs"].Description != "Show recent git log." {
		t.Errorf("logs tool wrong: %+v", byName["logs"])
	}
	if byName["docker__push"] == nil {
		t.Errorf("submodule task not namespaced as docker__push: %v", byName)
	}
	if d := byName["clean"]; d == nil || d.Annotations.DestructiveHint == nil || !*d.Annotations.DestructiveHint {
		t.Errorf("clean destructiveHint missing: %+v", d)
	}
}

func TestCallToolThroughEngine(t *testing.T) {
	eng := sampleEngine()
	srv := New(eng, Options{AllowDestructive: true})
	cs := connect(t, srv)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "logs"})
	if err != nil {
		t.Fatal(err)
	}
	if eng.lastCall != "logs" {
		t.Errorf("engine called with %q, want logs", eng.lastCall)
	}
	text := contentText(res)
	if !strings.Contains(text, "output-here") {
		t.Errorf("result missing engine output: %q", text)
	}
}

func contentText(res *mcp.CallToolResult) string {
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}
