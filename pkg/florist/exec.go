package florist

import (
	"bufio"
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
)

// CmdRun runs 'cmd', redirecting its stdout and stderr to 'log.Debug'.
// CmdRun blocks until 'cmd' terminates.
func CmdRun(log *slog.Logger, cmd *exec.Cmd) error {
	log.Debug("cmd-run", "cmd", cmd.String())

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	// Collect stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		var splitter Splitter
		outScanner := bufio.NewScanner(stdout)
		outScanner.Split(splitter.ScanWithLength)
		for outScanner.Scan() {
			line := outScanner.Text()
			if line != "" {
				log.Debug("cmd-run", "stdout", line)
			}
		}
		if err := outScanner.Err(); err != nil {
			log.Debug("cmd-run: error reading stdout", "error", err)
		}
	}()

	// Collect stderr
	wg.Add(1)
	var errLines []string
	go func() {
		defer wg.Done()
		var splitter Splitter
		errScanner := bufio.NewScanner(stderr)
		errScanner.Split(splitter.ScanWithLength)
		for errScanner.Scan() {
			line := errScanner.Text()
			if line != "" {
				errLines = append(errLines, line)
				log.Debug("cmd-run", "stderr", line)
			}
		}
		if err := errScanner.Err(); err != nil {
			log.Debug("cmd-run: error reading stderr", "error", err)
		}
	}()

	if err := cmd.Start(); err != nil {
		return err
	}

	// Wait for the pipes to return EOF.
	wg.Wait()

	// It is now safe to call cmd.Wait, which will close the pipes.
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s: %s", err, errLines)
	}

	return nil
}

const maxLineLen = 160

// Splitter keeps state for [Splitter.ScanWithLength], meant to be passed to a
// bufio.Scanner.
// The zero value of Splitter uses a MaxLineLen of [maxLineLen].
type Splitter struct {
	MaxLineLen int
}

// ScanWithLength is a split function for a bufio.Scanner that returns each line of text.
// If the line is longer than [Splitter.MaxLineLen], then ScanWithLength returns the first
// [Splitter.MaxLineLen] bytes (the rest of the line will be reurned in subsequent calls
// to Scan).
// Based on ScanLines from stdlib bufio/scan.go
func (sp *Splitter) ScanWithLength(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if sp.MaxLineLen == 0 {
		sp.MaxLineLen = maxLineLen
	}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 && i <= sp.MaxLineLen {
		// We have a full newline-terminated line and it is not too long.
		return i + 1, dropCR(data[0:i]), nil
	}

	// If line is too long, let's split it.
	if len(data) >= sp.MaxLineLen {
		return sp.MaxLineLen, data[0:sp.MaxLineLen], nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
// Taken from stdlib bufio/scan.go
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
