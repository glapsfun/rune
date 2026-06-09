package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestExampleContract enforces contracts/example-contract.md: every example
// directory has a README.md carrying the required sections.
func TestExampleContract(t *testing.T) {
	required := []string{"Use case:", "Demonstrates:", "Prerequisites:", "Run it", "Expected output"}
	for _, name := range exampleDirs(t) {
		t.Run(name, func(t *testing.T) {
			readme := filepath.Join(repoRoot, "docs", "examples", name, "README.md")
			data, err := os.ReadFile(readme)
			if err != nil {
				t.Fatalf("example %q missing README.md (example contract): %v", name, err)
			}
			s := string(data)
			for _, marker := range required {
				if !strings.Contains(s, marker) {
					t.Errorf("example %q README missing required section %q", name, marker)
				}
			}
		})
	}
}

// TestCoverageMatrix enforces FR-007 / SC-004: every targeted capability and
// project shape has at least one example directory. (See data-model.md's
// Coverage Matrix; each row maps to a specific example id.)
func TestCoverageMatrix(t *testing.T) {
	required := []string{
		// Project shapes
		"go-service", "node-project", "python-project", "monorepo",
		"ci-cd", "docker-workflow", "polyglot", "agent-driven",
		// Capability spotlights
		"dependencies", "parameters", "caching", "parallel",
		"settings-dotenv", "os-filtering",
	}
	have := make(map[string]bool)
	for _, d := range exampleDirs(t) {
		have[d] = true
	}
	for _, id := range required {
		if !have[id] {
			t.Errorf("coverage gap: missing example %q (FR-007/SC-004)", id)
		}
	}
}

// forbiddenTerms are unambiguous glossary aliases that must never appear in the
// docs (contracts/information-architecture.md, FR-016). The check is deliberately
// conservative — it covers only terms with no legitimate use here; the broader
// glossary (e.g. "prerequisite" for task ordering) is enforced by review, since
// those words have valid uses (notably the "Prerequisites:" tooling field).
var forbiddenTerms = []*regexp.Regexp{
	regexp.MustCompile(`\bRunfile\b`),      // misspelling of "Runefile"
	regexp.MustCompile(`(?i)\brecipes?\b`), // just's term; Rune's unit is a "task"
}

// TestTerminology keeps the documentation's vocabulary consistent (FR-016).
func TestTerminology(t *testing.T) {
	for _, f := range docMarkdownFiles(t) {
		t.Run(relPath(f), func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", relPath(f), err)
			}
			for _, re := range forbiddenTerms {
				if hit := re.FindString(string(data)); hit != "" {
					t.Errorf("%s: forbidden term %q — use the canonical glossary term "+
						"(e.g. \"Runefile\", \"task\")", relPath(f), hit)
				}
			}
		})
	}
}

// secretRe matches an assignment of a credential-shaped name to a long literal
// value (Principle VII). It deliberately does not match prose mentions of
// "token"/"secret".
var secretRe = regexp.MustCompile(`(?i)(api[_-]?key|secret|password|auth[_-]?token)\s*[:=]\s*["']?[A-Za-z0-9_\-]{16,}`)

// TestNoSecretsInDocs scans documentation and example Runefiles for secret
// literals; secrets must come from the environment only.
func TestNoSecretsInDocs(t *testing.T) {
	files := docMarkdownFiles(t)
	for _, name := range exampleDirs(t) {
		files = append(files, filepath.Join(repoRoot, "docs", "examples", name, "Runefile"))
	}
	for _, f := range files {
		t.Run(relPath(f), func(t *testing.T) {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", relPath(f), err)
			}
			if hit := secretRe.FindString(string(data)); hit != "" {
				t.Errorf("possible secret literal in %s: %q (secrets must come from the environment)",
					relPath(f), hit)
			}
		})
	}
}
