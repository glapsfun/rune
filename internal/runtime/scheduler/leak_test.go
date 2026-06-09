package scheduler

import (
	"testing"

	"go.uber.org/goleak"
)

// TestMain runs this package's tests under goleak to catch leaked goroutines
// (Constitution Principle VIII: every goroutine must have a guaranteed exit).
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
