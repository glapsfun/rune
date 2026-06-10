package cli

// TaskCandidate is a completable task: its name plus the first line of its doc
// comment, used as the shell-completion description.
type TaskCandidate struct {
	Name string
	Doc  string
}

// TaskCandidates returns the non-private, OS-matching tasks of the resolved
// Runefile, each with its first doc line, for dynamic shell completion. It is
// deliberately tolerant: any failure (no Runefile, parse error) yields nil so
// completion degrades gracefully and never disrupts the shell session.
func TaskCandidates(opts Options) []TaskCandidate {
	return nil // implemented in US2
}
