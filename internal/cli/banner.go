package cli

import (
	"errors"
	"fmt"

	"github.com/rune-task-runner/rune/internal/style"
)

// BannerLine formats one of Rune's "rune: <msg>" failure banners, emphasizing
// the prefix with the shared error role so failures are scannable by color in a
// long transcript (spec 014 C1). Only the prefix is styled: the message stays
// plain (and pre-masked where it embeds task output), so stripping ANSI
// recovers the exact plain bytes. It is the single definition of the banner
// shape, shared by the cycle/watch banners here and the top-level banner in
// cmd/rune.
func BannerLine(th style.Theme, msg string) string {
	return th.Error.Render("rune:") + " " + msg
}

// printErrorBanner writes a failure banner to stderr using the caller's theme
// (avoiding a per-call renderer rebuild for callers that already hold one).
func printErrorBanner(opts Options, th style.Theme, msg string) {
	fmt.Fprintln(opts.Stderr, BannerLine(th, msg))
}

// BannerMessage returns the banner text for err, or "" when err must not
// produce a top-level banner: a ValidationError's diagnostics are already
// rendered by the pipeline, and a silent TaskFailure ([no-exit-message])
// intentionally stays quiet. It is the single source of the suppression rule
// shared by the top-level (cmd/rune) and watch-loop banner paths.
func BannerMessage(err error) string {
	if err == nil {
		return ""
	}
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ""
	}
	var tf *TaskFailure
	if errors.As(err, &tf) && tf.Silent {
		return ""
	}
	return err.Error()
}
