package florist_test

import (
	"log/slog"
	"os/exec"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina"
)

func TestCmdRunSuccess(t *testing.T) {
	discard := slog.New(slog.DiscardHandler)
	err := florist.CmdRun(discard, exec.Command("true"))
	rosina.AssertNoError(t, err)
}

func TestCmdRunFailure(t *testing.T) {
	discard := slog.New(slog.DiscardHandler)
	err := florist.CmdRun(discard, exec.Command("false"))
	rosina.AssertErrorContains(t, err, "exit status 1:")
}
