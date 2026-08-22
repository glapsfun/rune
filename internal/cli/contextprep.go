package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rune-task-runner/rune/internal/ast"
)

// Context-hook processing limits (spec 021 NFR-002): fixed, with no
// configuration surface. contextTimeout is a var only so tests can shrink it.
var contextTimeout = 10 * time.Second

// contextMaxBytes caps the injected content; the "[truncated]" marker is
// appended after the byte-wise cut and is excluded from the cap.
const contextMaxBytes = 8 * 1024

// contextTask returns the file's [context] task when one exists and is
// available on goos; an OS-mismatched hook is treated as absent (FR-008).
func contextTask(f *ast.File, goos string) *ast.Task {
	if f == nil {
		return nil
	}
	for _, t := range f.Tasks {
		if t.Attr(ast.AttrContext) != nil && t.AvailableOn(goos) {
			return t
		}
	}
	return nil
}

// gatherContext runs the [context] hook through the adapter's masked Call
// path and returns the processed text plus whether a hook exists. The hook
// is best-effort: failures and timeouts degrade to a one-line notice and a
// stderr warning (FR-005) — never an error, never a blocked surface.
func (a *mcpAdapter) gatherContext(ctx context.Context, stderr io.Writer) (string, bool) {
	t := contextTask(a.file, a.goos)
	if t == nil {
		return "", false
	}
	cctx, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()
	res, err := a.Call(cctx, t.Name, map[string]string{})
	if err != nil || res.ExitCode != ExitSuccess {
		fmt.Fprintf(stderr, "warning: context hook %q failed; proceeding without project context\n", t.Name)
		return fmt.Sprintf("(context hook %q failed; proceeding without project context)", t.Name), true
	}
	out := strings.TrimRight(res.Stdout, "\n")
	if len(out) > contextMaxBytes {
		// Masking already happened inside Call, so the cut cannot expose a
		// secret that the mask would have caught.
		out = out[:contextMaxBytes] + "\n[truncated]"
	}
	return out, true
}
