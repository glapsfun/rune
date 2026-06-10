package docs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// runeBin is the path to the binary built once for the whole package; repoRoot
// is the module root. Both are set in TestMain and read-only thereafter.
var (
	runeBin  string
	repoRoot string
)

// TestMain builds ./cmd/rune once into a temp dir, then runs the suite. It
// mirrors test/integration/harness_test.go so the docs harness exercises the
// same binary the CLI ships.
func TestMain(m *testing.M) {
	root, err := moduleRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "docs: cannot find module root:", err)
		os.Exit(1)
	}
	repoRoot = root

	dir, err := os.MkdirTemp("", "rune-docs")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bin := filepath.Join(dir, "rune")
	if runtime.GOOS == "windows" {
		bin += ".exe" // go build writes -o verbatim; Windows needs the extension to exec.
	}
	build := exec.Command("go", "build", "-o", bin, "./cmd/rune")
	build.Dir = root
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "docs: build failed: %v\n%s\n", err, out)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}
	runeBin = bin

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// moduleRoot walks up from the working directory to the directory holding go.mod.
func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found")
		}
		dir = parent
	}
}

// result captures one binary invocation.
type result struct {
	stdout string
	stderr string
	code   int
}

// runRune runs the built binary with args, bounded by a timeout. dir is the
// working directory (defaults to the repo root). It fails the test only if the
// process cannot start; a non-zero exit is returned in result.code for the
// caller to assert on.
func runRune(t *testing.T, dir string, args ...string) result {
	t.Helper()
	// Generous bound: this gates correctness (exit code / output), not latency. On
	// Windows CI the first `node`/`python` invocation cold-starts slowly and variably
	// (Defender scanning the interpreter on first exec) — under the full `-race` suite
	// load it can momentarily exceed a tight budget, which made the polyglot example
	// flaky. Only a genuine hang should trip this.
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, runeBin, args...)
	if dir == "" {
		dir = repoRoot
	}
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else {
			t.Fatalf("rune %v failed to start: %v", args, err)
		}
	}
	return result{stdout: out.String(), stderr: errb.String(), code: code}
}

// validate statically validates the Runefile at path: parse + analyze, run
// nothing, via `--list` (there is no built-in `check` subcommand; `--list`
// forces analysis and needs no task argument). exit 0 means valid.
func validate(t *testing.T, runefilePath string) result {
	t.Helper()
	return runRune(t, "", "--file", runefilePath, "--list")
}

// docMarkdownFiles returns every *.md under docs/, plus README.md and
// CONTRIBUTING.md at the repo root.
func docMarkdownFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	docsDir := filepath.Join(repoRoot, "docs")
	err := filepath.WalkDir(docsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".md") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk docs/: %v", err)
	}
	for _, extra := range []string{"README.md", "CONTRIBUTING.md"} {
		p := filepath.Join(repoRoot, extra)
		if _, err := os.Stat(p); err == nil {
			files = append(files, p)
		}
	}
	return files
}

// exampleDirs returns the names of directories under docs/examples that contain
// a Runefile.
func exampleDirs(t *testing.T) []string {
	t.Helper()
	base := filepath.Join(repoRoot, "docs", "examples")
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("read docs/examples: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(base, e.Name(), "Runefile")); err == nil {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

// relPath renders p relative to the repo root for readable failure messages.
func relPath(p string) string {
	if r, err := filepath.Rel(repoRoot, p); err == nil {
		return r
	}
	return p
}

// fencedBlocks returns the contents of fenced code blocks whose info string is
// exactly lang (e.g. "rune") in markdown src.
func fencedBlocks(src, lang string) []string {
	var blocks []string
	var cur []string
	open := "```" + lang
	in := false
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimRight(raw, "\r")
		if !in {
			if strings.TrimSpace(line) == open {
				in = true
				cur = nil
			}
			continue
		}
		if strings.TrimSpace(line) == "```" {
			blocks = append(blocks, strings.Join(cur, "\n")+"\n")
			in = false
			continue
		}
		cur = append(cur, line)
	}
	return blocks
}
