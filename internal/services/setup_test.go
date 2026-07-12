package services

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

// TestMain silences package-level slog output during tests. Production sets a
// JSON handler in main(); without this, tests fall back to the default text
// handler and spray log lines into test output.
func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}
