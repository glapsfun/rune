package language

import "testing"

func TestRegistryLookupAndValidity(t *testing.T) {
	if !IsValid(BuiltinFunction, "env") {
		t.Error("env should be a valid builtin function")
	}
	if !IsValid(BuiltinSetting, "working-directory") {
		t.Error("working-directory should be a valid setting")
	}
	if !IsValid(BuiltinAttribute, "parallel") {
		t.Error("parallel should be a valid attribute")
	}
	if !IsValid(BuiltinExecutor, "python") {
		t.Error("python should be a valid executor")
	}
	if IsValid(BuiltinAttribute, "nonexistent") {
		t.Error("unknown attribute should be invalid")
	}
	if b, ok := Lookup(BuiltinFunction, "os_family"); !ok || b.Signature == "" {
		t.Errorf("os_family lookup failed: %+v ok=%v", b, ok)
	}
}

func TestRegistryPrefixMatch(t *testing.T) {
	m := MatchKind(BuiltinSetting, "wor")
	if len(m) != 1 || m[0].Name != "working-directory" {
		t.Errorf("MatchKind(setting,\"wor\") = %+v, want working-directory", m)
	}
	if got := MatchKind(BuiltinExecutor, "py"); len(got) != 1 || got[0].Name != "python" {
		t.Errorf("MatchKind(executor,\"py\") = %+v, want python", got)
	}
}
