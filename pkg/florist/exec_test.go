package florist_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/marco-m/florist/internal"
	"github.com/marco-m/florist/pkg/florist"
	"github.com/marco-m/rosina/diff"
)

func TestCmdRunBasicSuccess(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: internal.RemoveTime,
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
		ReplaceAttr: internal.RemoveTime,
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

// This is part of the TestHelperProcess pattern.
// https://web.archive.org/web/20250204012613/https://npf.io/2015/06/testing-exec-command/
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
		// Avoid diagnostic printed by "go test" to stderr:
		//   warning: GOCOVERDIR not set, no coverage data emitted
		"GOCOVERDIR=" + os.TempDir(),
	}
	return cmd
}

// This is part of the TestHelperProcess pattern.
// https://web.archive.org/web/20250204012613/https://npf.io/2015/06/testing-exec-command/
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// optional: validation logic

	// Simulate a process, run by florist.CmdRun(), that writes to stdout and stderr.
	fmt.Fprintf(os.Stdout, "line 1 to stdout\n")
	fmt.Fprintf(os.Stderr, "line 2 to stderr\n")
	fmt.Fprintf(os.Stdout, "line 3 to stdout\n")
	fmt.Fprintf(os.Stderr, "line 4 to stderr\n")
	fmt.Fprintf(os.Stdout, "line 5 to stdout\n")
	fmt.Fprintf(os.Stderr, "line 6 to stderr\n")

	os.Exit(0)
}

func TestCmdRunProcWritesToStdoutAndStderr(t *testing.T) {
	// execCommandFn = fakeExecCommand
	// defer func() { execCommandFn = exec.Command }()

	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: internal.RemoveTime,
	}))

	err := florist.CmdRun(log, fakeExecCommand("anything-goes-this-is-a-fake"))
	if err != nil {
		t.Errorf("\nhave error: %s\nwant: <no error>", err)
	}

	// Remove the first line from buffer, because it contains the path to the Go test
	// executable, which is in a temporary directory. For example:
	// level=DEBUG msg=cmd-run
	//   cmd="/var/folders/h2/sz12x.../T/go-build37.../b001/florist.test
	//   -test.run=TestHelperProcess -- anything-goes-this-is-a-fake"
	have := buf.String()
	have = have[strings.Index(have, "\n"):]

	// NOTE the ordering is wrong: first all writes to stdout, then all writes to stderr.
	// This shows that the implementation is wrong.
	want := `
level=DEBUG msg=cmd-run stdout="line 1 to stdout"
level=DEBUG msg=cmd-run stdout="line 3 to stdout"
level=DEBUG msg=cmd-run stdout="line 5 to stdout"
level=DEBUG msg=cmd-run stderr="line 2 to stderr"
level=DEBUG msg=cmd-run stderr="line 4 to stderr"
level=DEBUG msg=cmd-run stderr="line 6 to stderr"
`
	if text := diff.TextDiff("want", "have", want, have); text != "" {
		t.Errorf("\nlog mismatch:\n%s", text)
	}
}
