package integration

import (
	"strings"
	"testing"
)

const settingsSecret = "hunter2-settings-secret"

// T023a — `set secrets` masks a variable whose name matches no built-in pattern.
func TestSecretMasking_DeclaredInnocentNameMasked(t *testing.T) {
	src := "set secrets := [\"DEPLOY_CFG\"]\n\nshow:\n    @echo \"cfg is $DEPLOY_CFG\"\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"DEPLOY_CFG=" + settingsSecret}, "show")
	if r.code != 0 {
		t.Fatalf("exit = %d; stderr=%s", r.code, r.stderr)
	}
	if strings.Contains(r.stdout, settingsSecret) {
		t.Fatalf("declared secret leaked: %q", r.stdout)
	}
	if !strings.Contains(r.stdout, "cfg is ***") {
		t.Errorf("stdout = %q, want masked", r.stdout)
	}
}

// T023b — `set unmasked` exempts a pattern-matched name.
func TestSecretMasking_UnmaskedExemptsPattern(t *testing.T) {
	src := "set unmasked := [\"OAUTH_METHOD\"]\n\nshow:\n    @echo \"method is $OAUTH_METHOD\"\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, []string{"OAUTH_METHOD=oauth2-pkce-flow"}, "show")
	if r.code != 0 {
		t.Fatalf("exit = %d; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "method is oauth2-pkce-flow") {
		t.Errorf("exempted value was masked: %q", r.stdout)
	}
}

// T023c — a declared name absent from the environment is inert.
func TestSecretMasking_AbsentDeclaredNameIsInert(t *testing.T) {
	src := "set secrets := [\"NOT_PRESENT_ANYWHERE\"]\n\nhello:\n    @echo hi\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, nil, "hello")
	if r.code != 0 || r.stdout != "hi\n" {
		t.Errorf("inert declaration changed behavior: code=%d stdout=%q stderr=%q", r.code, r.stdout, r.stderr)
	}
}

// T023d — a malformed list element fails statically: positioned diagnostic,
// exit 3, nothing executed.
func TestSecretMasking_MalformedDeclarationFailsStatically(t *testing.T) {
	src := "set secrets := [nope_undefined]\n\nhello:\n    @echo ran-anyway\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, nil, "hello")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "Runefile:1") {
		t.Errorf("diagnostic not positioned: %q", r.stderr)
	}
	if strings.Contains(r.stdout, "ran-anyway") {
		t.Errorf("task ran despite the static error")
	}
}

// T023e — the same name in both lists is a static conflict citing both spans.
func TestSecretMasking_ConflictingListsFailStatically(t *testing.T) {
	src := "set secrets := [\"BOTH_LISTED\"]\nset unmasked := [\"BOTH_LISTED\"]\n\nhello:\n    @echo ran-anyway\n"
	dir := writeRunefile(t, src)
	r := run(t, dir, nil, "hello")
	if r.code != 3 {
		t.Fatalf("exit = %d, want 3; stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "BOTH_LISTED") {
		t.Errorf("conflict diagnostic does not name the variable: %q", r.stderr)
	}
	if strings.Contains(r.stdout, "ran-anyway") {
		t.Errorf("task ran despite the conflict")
	}
}

// T023f — `rune analyze` flags a typo'd setting name as RUNE2008 and accepts
// the real names.
func TestSecretMasking_AnalyzeFlagsTypo(t *testing.T) {
	dir := writeRunefile(t, "set secert := [\"X\"]\n\nhello:\n    @echo hi\n")
	r := run(t, dir, nil, "analyze")
	if r.code != 3 {
		t.Fatalf("analyze exit = %d, want 3; stdout=%s", r.code, r.stdout)
	}
	if !strings.Contains(r.stdout, "RUNE2008") {
		t.Errorf("analyze output missing RUNE2008: %q", r.stdout)
	}

	dir = writeRunefile(t, "set secrets := [\"A_NAME\"]\nset unmasked := [\"B_NAME\"]\n\nhello:\n    @echo hi\n")
	r = run(t, dir, nil, "analyze")
	if r.code != 0 {
		t.Errorf("analyze rejected valid settings: code=%d stdout=%q", r.code, r.stdout)
	}
}
