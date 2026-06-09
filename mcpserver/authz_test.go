package mcpserver

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestAuthzNonDestructiveCallable(t *testing.T) {
	srv := New(sampleEngine(), Options{}) // destructive NOT allowed by default
	if !srv.authorized("logs") {
		t.Error("non-destructive task should be callable")
	}
}

func TestAuthzDestructiveRequiresApproval(t *testing.T) {
	srv := New(sampleEngine(), Options{AllowDestructive: false})
	if srv.authorized("clean") {
		t.Error("destructive task must require approval")
	}
	srv2 := New(sampleEngine(), Options{AllowDestructive: true})
	if !srv2.authorized("clean") {
		t.Error("destructive task should be callable with approval")
	}
}

func TestAuthzAllowListNarrows(t *testing.T) {
	srv := New(sampleEngine(), Options{AllowDestructive: true, AllowList: []string{"logs"}})
	if !srv.authorized("logs") {
		t.Error("allow-listed task should be callable")
	}
	if srv.authorized("greet") {
		t.Error("task outside the allow-list must be refused")
	}
}

func TestDestructiveCallRefusedAtRuntime(t *testing.T) {
	eng := sampleEngine()
	srv := New(eng, Options{AllowDestructive: false})
	cs := connect(t, srv)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "clean"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("destructive call without approval should be an error result")
	}
	if !strings.Contains(contentText(res), "approval") {
		t.Errorf("refusal should mention approval: %q", contentText(res))
	}
	if eng.lastCall == "clean" {
		t.Error("engine must NOT run a refused destructive task")
	}
}
