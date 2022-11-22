package cook

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ExecOut(env []string, name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return "", fmt.Errorf("%s: %w: %s", name, err,
				strings.TrimSpace(string(exitError.Stderr)))
		}
		return "", fmt.Errorf("%s: %w", name, err)
	}
	return string(out), nil
}

func ExecRun(dir string, env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func ExecRunFunc(dir string, env []string, name string, arg1 ...string,
) func(arg2 ...string) error {
	return func(arg2 ...string) error {
		cmd := exec.Command(name, append(arg1, arg2...)...)
		cmd.Dir = dir
		cmd.Env = env
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}
}
