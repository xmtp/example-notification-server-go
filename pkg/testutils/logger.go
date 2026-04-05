package testutils

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestLogger returns a zap.Logger bound to the test lifecycle.
// Output is captured by testing.T and only shown on failure.
func TestLogger(t *testing.T) *zap.Logger {
	t.Helper()
	return zaptest.NewLogger(t)
}
