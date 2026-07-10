package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const tv = "RUNE_TEST_VERSION="

// US1: an older installed binary is refused before execution.
func TestMinimumVersion_RejectsOlder(t *testing.T) {
	dir := writeRunefile(t, "set minimum_version := \"0.8.0\"\nbuild:\n    @echo BUILT\n")
	r := run(t, dir, []string{tv + "0.7.2"}, "build")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3\nstderr: %s", r.code, r.stderr)
	}
	if strings.Contains(r.stdout, "BUILT") {
		t.Errorf("task ran despite incompatible version; stdout = %q", r.stdout)
	}
	for _, want := range []string{"requires Rune >= 0.8.0", "installed version: 0.7.2", "required version:  0.8.0", "nothing was executed"} {
		if !strings.Contains(r.stderr, want) {
			t.Errorf("stderr missing %q\ngot: %s", want, r.stderr)
		}
	}
}

// US1: equal and newer versions run normally.
func TestMinimumVersion_AllowsEqualAndNewer(t *testing.T) {
	dir := writeRunefile(t, "set minimum_version := \"0.8.0\"\nbuild:\n    @echo BUILT\n")
	for _, v := range []string{"0.8.0", "0.9.1", "1.0.0"} {
		r := run(t, dir, []string{tv + v}, "build")
		if r.code != 0 || !strings.Contains(r.stdout, "BUILT") {
			t.Errorf("installed %s: exit=%d stdout=%q stderr=%q", v, r.code, r.stdout, r.stderr)
		}
	}
}

// US1 / FR-023: a Runefile without the setting behaves exactly as before.
func TestMinimumVersion_NoSettingUnchanged(t *testing.T) {
	dir := writeRunefile(t, "build:\n    @echo BUILT\n")
	r := run(t, dir, []string{tv + "0.0.1"}, "build")
	if r.code != 0 || !strings.Contains(r.stdout, "BUILT") {
		t.Errorf("no-setting run changed: exit=%d stdout=%q stderr=%q", r.code, r.stdout, r.stderr)
	}
}

// US1 / FR-012 / SC-007: only the ROOT file's minimum_version governs.
func TestMinimumVersion_RootOwnership(t *testing.T) {
	// Root requires 0.8.0; an imported child declares 9.9.9. With 0.8.0 installed
	// the run must succeed — the child's higher requirement is ignored.
	t.Run("child cannot raise requirement", func(t *testing.T) {
		dir := t.TempDir()
		writeChild(t, dir, "child.rune", "set minimum_version := \"9.9.9\"\nshared:\n    @echo SHARED\n")
		writeMain(t, dir, "import \"child.rune\"\nset minimum_version := \"0.8.0\"\nbuild: shared\n    @echo BUILT\n")
		r := run(t, dir, []string{tv + "0.8.0"}, "build")
		if r.code != 0 || !strings.Contains(r.stdout, "BUILT") {
			t.Errorf("child raised the requirement: exit=%d stdout=%q stderr=%q", r.code, r.stdout, r.stderr)
		}
	})

	// Root declares nothing; the child declares 9.9.9. No requirement is imposed.
	t.Run("child cannot impose requirement", func(t *testing.T) {
		dir := t.TempDir()
		writeChild(t, dir, "child.rune", "set minimum_version := \"9.9.9\"\nshared:\n    @echo SHARED\n")
		writeMain(t, dir, "import \"child.rune\"\nbuild: shared\n    @echo BUILT\n")
		r := run(t, dir, []string{tv + "0.1.0"}, "build")
		if r.code != 0 || !strings.Contains(r.stdout, "BUILT") {
			t.Errorf("child imposed a requirement: exit=%d stdout=%q stderr=%q", r.code, r.stdout, r.stderr)
		}
	})
}

// US2: static-value and invalid-semver guards.
func TestMinimumVersion_StaticGuard(t *testing.T) {
	t.Run("non-static value", func(t *testing.T) {
		dir := writeRunefile(t, "required := \"0.8.0\"\nset minimum_version := required\nbuild:\n    @echo BUILT\n")
		r := run(t, dir, []string{tv + "9.9.9"}, "build")
		if r.code != 3 || strings.Contains(r.stdout, "BUILT") {
			t.Fatalf("exit=%d stdout=%q", r.code, r.stdout)
		}
		if !strings.Contains(r.stderr, "must be a static semantic version") {
			t.Errorf("stderr = %q, want 'static semantic version'", r.stderr)
		}
	})
	t.Run("invalid semver", func(t *testing.T) {
		for _, bad := range []string{"0.8", "latest", "v0.8.0", ">=0.8,<1.0"} {
			dir := writeRunefile(t, "set minimum_version := \""+bad+"\"\nbuild:\n    @echo BUILT\n")
			r := run(t, dir, []string{tv + "9.9.9"}, "build")
			if r.code != 3 || strings.Contains(r.stdout, "BUILT") {
				t.Fatalf("value %q: exit=%d stdout=%q stderr=%q", bad, r.code, r.stdout, r.stderr)
			}
			if !strings.Contains(r.stderr, "must be a valid semantic version") {
				t.Errorf("value %q: stderr = %q, want 'valid semantic version'", bad, r.stderr)
			}
		}
	})
}

// US3: rune version / --check / --json.
func TestMinimumVersion_VersionCommand(t *testing.T) {
	dir := writeRunefile(t, "set minimum_version := \"0.8.0\"\nbuild:\n    @echo hi\n")

	t.Run("bare version has language line", func(t *testing.T) {
		r := run(t, dir, []string{tv + "0.8.3"}, "version")
		if r.code != 0 {
			t.Fatalf("exit=%d stderr=%q", r.code, r.stderr)
		}
		if !strings.Contains(r.stdout, "0.8.3") || !strings.Contains(r.stdout, "runefile language 1") {
			t.Errorf("stdout = %q, want version + language line", r.stdout)
		}
	})

	t.Run("check compatible", func(t *testing.T) {
		r := run(t, dir, []string{tv + "0.8.3"}, "version", "--check")
		if r.code != 0 || !strings.Contains(r.stdout, "compatible") {
			t.Errorf("exit=%d stdout=%q", r.code, r.stdout)
		}
	})

	t.Run("check incompatible exits non-zero", func(t *testing.T) {
		r := run(t, dir, []string{tv + "0.7.2"}, "version", "--check")
		if r.code == 0 {
			t.Errorf("incompatible --check should exit non-zero; stdout=%q", r.stdout)
		}
	})

	t.Run("check json", func(t *testing.T) {
		r := run(t, dir, []string{tv + "0.8.3"}, "version", "--check", "--json")
		if r.code != 0 {
			t.Fatalf("exit=%d stderr=%q", r.code, r.stderr)
		}
		var res struct {
			Installed   string `json:"installed"`
			Required    string `json:"required"`
			Compatible  bool   `json:"compatible"`
			Development bool   `json:"development"`
			Runefile    string `json:"runefile"`
		}
		if err := json.Unmarshal([]byte(r.stdout), &res); err != nil {
			t.Fatalf("stdout not JSON: %v\n%s", err, r.stdout)
		}
		if res.Installed != "0.8.3" || res.Required != "0.8.0" || !res.Compatible || res.Development || res.Runefile == "" {
			t.Errorf("json = %+v", res)
		}
	})

	t.Run("no requirement", func(t *testing.T) {
		bare := writeRunefile(t, "build:\n    @echo hi\n")
		r := run(t, bare, []string{tv + "0.8.3"}, "version", "--check")
		if r.code != 0 || !strings.Contains(r.stdout, "no requirement declared") {
			t.Errorf("exit=%d stdout=%q", r.code, r.stdout)
		}
	})

	t.Run("explicit missing file is a usage error", func(t *testing.T) {
		// A typo'd/absent --file must not be reported as "compatible" (exit 2,
		// like the run path) rather than exit 0.
		r := run(t, dir, []string{tv + "0.8.3"}, "version", "--check", "-f", "does-not-exist.rune")
		if r.code != 2 {
			t.Errorf("missing --file: exit=%d, want 2\nstdout=%q stderr=%q", r.code, r.stdout, r.stderr)
		}
		if strings.Contains(r.stdout, "compatible") {
			t.Errorf("missing --file reported compatible: %q", r.stdout)
		}
	})
}

// US4: --ignore-version warns and proceeds; the MCP path refuses by default.
func TestMinimumVersion_IgnoreVersion(t *testing.T) {
	dir := writeRunefile(t, "set minimum_version := \"0.8.0\"\nbuild:\n    @echo BUILT\n")

	t.Run("cli override warns and runs", func(t *testing.T) {
		r := run(t, dir, []string{tv + "0.7.2"}, "--ignore-version", "build")
		if r.code != 0 || !strings.Contains(r.stdout, "BUILT") {
			t.Fatalf("exit=%d stdout=%q stderr=%q", r.code, r.stdout, r.stderr)
		}
		if !strings.Contains(r.stderr, "warning: ignoring Runefile minimum Rune version 0.8.0; running 0.7.2") {
			t.Errorf("missing warning; stderr = %q", r.stderr)
		}
	})

	t.Run("mcp refuses by default", func(t *testing.T) {
		// The gate runs at module load, before the server reads stdin, so an
		// incompatible requirement refuses immediately.
		r := run(t, dir, []string{tv + "0.7.2"}, "mcp")
		if r.code == 0 {
			t.Errorf("mcp with unmet requirement should refuse; stderr=%q", r.stderr)
		}
		if !strings.Contains(r.stderr, "requires Rune >= 0.8.0") {
			t.Errorf("mcp stderr = %q, want incompatibility diagnostic", r.stderr)
		}
	})

	t.Run("mcp operator opt-in proceeds", func(t *testing.T) {
		// With the operator flag the server starts and exits cleanly on EOF stdin
		// (nil stdin => /dev/null => immediate EOF).
		r := run(t, dir, []string{tv + "0.7.2"}, "mcp", "--ignore-version")
		if r.code != 0 {
			t.Errorf("mcp --ignore-version exit=%d stderr=%q", r.code, r.stderr)
		}
	})
}

// --- helpers ---

func writeMain(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "Runefile"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeChild(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
