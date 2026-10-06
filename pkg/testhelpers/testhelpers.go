package testhelpers

import (
	"io"
	"log/slog"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

// RemoveTime removes the "time" attribute from the output of a slog.Logger.
func RemoveTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func MakeTestLog() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func InitFlorist(t *testing.T, tempDir string) *florist.Garden {
	t.Helper()
	opts := florist.Options{TempDir: tempDir}
	if testing.Verbose() {
		opts.LogLevel = slog.LevelDebug
	} else {
		opts.LogOutput = io.Discard
	}
	gdn, err := florist.NewGarden(&opts)
	assert.NoError(t, err, "florist.NewGarden")
	return gdn
}
