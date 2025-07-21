package florist_test

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

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

	//
	// This code runs in a separate OS process from the tests!!!
	//

	// Simulate a process, run by florist.CmdRun(), that writes to stdout and stderr.

	// WARNING The sleeps are to cause a context switch between this process and the test
	// process, so that we can check for the correct ordering (see
	// TestCmdRunProcWritesToStdoutAndStderr). This is inherently fragile and not really
	// needed to test that the SUT behaves correctly. As such, we might delete the sleeps
	// and not check the ordering anymore if this turns out to be flaky.
	fmt.Fprintf(os.Stdout, "line 1 to stdout\n")
	time.Sleep(10 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "line 2 to stderr\n")
	time.Sleep(10 * time.Millisecond)
	fmt.Fprintf(os.Stdout, "line 3 to stdout\n")
	time.Sleep(10 * time.Millisecond)
	fmt.Fprintf(os.Stderr, "line 4 to stderr\n")
	time.Sleep(10 * time.Millisecond)
	fmt.Fprintf(os.Stdout, "line 5 to stdout\n")
	time.Sleep(10 * time.Millisecond)
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

	want := `
level=DEBUG msg=cmd-run stdout="line 1 to stdout"
level=DEBUG msg=cmd-run stderr="line 2 to stderr"
level=DEBUG msg=cmd-run stdout="line 3 to stdout"
level=DEBUG msg=cmd-run stderr="line 4 to stderr"
level=DEBUG msg=cmd-run stdout="line 5 to stdout"
level=DEBUG msg=cmd-run stderr="line 6 to stderr"
`
	if text := diff.TextDiff("want", "have", want, have); text != "" {
		t.Errorf("\nlog mismatch:\n%s", text)
	}
}

func TestCustomSplitter(t *testing.T) {
	type testCase struct {
		name      string
		input     string
		wantLines []string
	}

	test := func(tc testCase) {
		t.Helper()

		splitter := florist.Splitter{MaxLineLen: 7}
		scanner := bufio.NewScanner(strings.NewReader(tc.input))
		scanner.Split(splitter.ScanWithLength)

		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("\n%q scanner:\nhave error: %s\nwant: <no error>", tc.name, err)
		}
		if !slices.Equal(lines, tc.wantLines) {
			t.Errorf("\n%q lines mismatch:\nhave: %v\nwant: %v", tc.name, lines, tc.wantLines)
		}
	}

	test(testCase{
		name:      "no newlines",
		input:     "a123456b123456c123",
		wantLines: []string{"a123456", "b123456", "c123"},
	})
	test(testCase{
		name:      "some newlines",
		input:     "a12\nb123456c1234\nd12",
		wantLines: []string{"a12", "b123456", "c1234", "d12"},
	})
}
