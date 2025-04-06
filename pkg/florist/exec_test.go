package florist_test

import (
	"log/slog"
	"os/exec"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/assert"
)

func TestCmdRunSuccess(t *testing.T) {
	discard := slog.New(slog.DiscardHandler)
	err := florist.CmdRun(discard, exec.Command("true"))
	assert.NoError(t, err, "florist.CmdRun")
}

func TestCmdRunFailure(t *testing.T) {
	discard := slog.New(slog.DiscardHandler)
	err := florist.CmdRun(discard, exec.Command("false"))
	assert.ErrorContains(t, err, "exit status 1:")
}
