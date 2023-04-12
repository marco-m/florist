package florist

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-hclog"
)

func CmdRun(log hclog.Logger, cmd *exec.Cmd) error {
	const truncLen = 160

	log.Debug("executing", "cmd", cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	outScanner := bufio.NewScanner(stdout)
	for outScanner.Scan() {
		line := outScanner.Text()
		if line != "" {
			log.Debug("", "stdout", truncate(line, truncLen))
		}
	}
	if err := outScanner.Err(); err != nil {
		return err
	}

	var errLines []string
	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		line := errScanner.Text()
		if line != "" {
			line = truncate(line, truncLen)
			errLines = append(errLines, line)
			log.Debug("", "stderr", line)
		}
	}
	if err := errScanner.Err(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s: %s", err, errLines)
	}

	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}