package cli

// Analyze statically analyzes the Runefile at path (or the discovered Runefile
// when path is empty) and its transitive imports, printing diagnostics and
// returning an error whose exit code is 3 when error diagnostics are present.
// Implemented in the US2 phase; stubbed for now so the command wiring compiles.
func Analyze(opts Options, path string, jsonOut bool) error {
	_, _ = path, jsonOut
	return &UsageError{Err: errorf("rune analyze: not implemented yet")}
}
