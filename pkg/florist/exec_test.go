package florist_test

import (
	"bytes"
	"log/slog"
	"os/exec"
	"strings"
	"testing"

	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/diff"
)

func TestCmdRunBasicSuccess(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: RemoveTime,
	}))

	err := florist.CmdRun(log, exec.Command("true"))
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}
	want := "level=DEBUG msg=cmd-run cmd=/usr/bin/true\n"
	if text := diff.TextDiff("want", "have", want, buf.String()); text != "" {
		t.Errorf("\nlog mismatch:\n%s", text)
	}
}

func TestCmdRunBasicFailure(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: RemoveTime,
	}))

	err := florist.CmdRun(log, exec.Command("false"))
	have, needle := err.Error(), "exit status 1:"
	if !strings.Contains(have, needle) {
		t.Errorf("\nerror message:    %s\ndoes not contain: %s", have, needle)
	}
	want := "level=DEBUG msg=cmd-run cmd=/usr/bin/false\n"
	if text := diff.TextDiff("want", "have", want, buf.String()); text != "" {
		t.Errorf("\nlog mismatch:\n%s", text)
	}
}

// RemoveTime removes the "time" attribute from the output of a slog.Logger.
func RemoveTime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}
