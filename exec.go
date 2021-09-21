package florist

import (
	"bufio"
	"os/exec"

	"github.com/hashicorp/go-hclog"
)

func LogRun(log hclog.Logger, cmd *exec.Cmd) error {
	log = log.Named("LogRun")

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

	scanOut := bufio.NewScanner(stdout)
	for scanOut.Scan() {
		line := scanOut.Text()
		if line != "" {
			log.Debug("", "stdout", truncate(line, 80))
		}
	}
	if err := scanOut.Err(); err != nil {
		return err
	}

	scanErr := bufio.NewScanner(stderr)
	for scanErr.Scan() {
		line := scanErr.Text()
		if line != "" {
			log.Debug("", "stderr", truncate(line, 80))
		}
	}
	if err := scanErr.Err(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
