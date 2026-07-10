package cli

import "flag"

// update is the shared -update flag for regenerating golden files across the
// cli package's golden tests (previously declared in fmt_test.go, which moved to
// internal/formatter).
var update = flag.Bool("update", false, "regenerate golden files")
