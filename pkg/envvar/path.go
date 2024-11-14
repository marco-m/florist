// Package envvar allows to persist environment variables for all supported
// shells.
package envvar

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/marco-m/florist/internal"
)

// Add persists environment variable k with value v for all users of the system,
// for all shells. It adds the settings to "/etc/environment", which it expects
// to be existing and have the correct permissions.
//
// Add is dumb: if called multiple time for the same k/v, it will each time
// append the same line to the file.
//
// Works as long as pam_env is configured, which seems to be the case for at
// least SSH to a Debian host.
// See https://wiki.archlinux.org/title/Environment_variables
func Add(k, v string) error {
	errorf := internal.MakeErrorf("envvar.AddPaths")
	const file = "/etc/environment"
	// Used only if the file doesn't exist and we passed O_CREATE to the flags.
	unused := os.FileMode(0o644)
	fi, err := os.OpenFile(file, os.O_WRONLY|os.O_APPEND, unused)
	if err != nil {
		return errorf("%s", err)
	}
	defer fi.Close()
	_, err = fmt.Fprintf(fi, "\n%s=%s\n", k, v)
	if err != nil {
		return errorf("%s", err)
	}
	return nil
}

// AddPaths appends 'paths' to the system PATH, for POSIX shells and for the
// Fish shell. Parameter 'name' is the name of the configuration file under the shell
// config directory.
//
// The never-ending pain of shells and something apparently as simple as adding
// to the PATH environment variable. Sigh.
// More about the total mess of shells:
// https://unix.stackexchange.com/questions/88201/whats-the-best-distro-shell-agnostic-way-to-set-environment-variables
func AddPaths(log *slog.Logger, name string, paths ...string) error {
	errorf := internal.MakeErrorf("envvar.AddPaths")
	//
	// Works for POSIX shells but not for fish
	//
	posixDst := fmt.Sprintf("/etc/profile.d/%s.sh", name)
	log.Debug("Add to PATH (POSIX shells)", "name", name, "paths", paths, "dst", posixDst)
	fi, err := os.Create(posixDst)
	if err != nil {
		return errorf("%s", err)
	}
	for _, path := range paths {
		line := fmt.Sprintf("export PATH=$PATH:%s\n", path)
		if _, err := fi.WriteString(line); err != nil {
			return errorf("write to %s: %s", posixDst, err)
		}
	}

	//
	// Works for fish
	//

	// Create the directory if not present. This allows to configure the path also if
	// the Fish shell has not been installed yet.
	if err := os.MkdirAll("/etc/fish/conf.d", 0o755); err != nil {
		return errorf("creating fish dir: %s", err)
	}

	fishDst := fmt.Sprintf("/etc/fish/conf.d/%s.fish", name)
	log.Debug("Add to PATH (Fish shell)", "name", name, "paths", paths, "dst", fishDst)
	fi, err = os.Create(fishDst)
	if err != nil {
		return errorf("%s", err)
	}
	for _, path := range paths {
		line := fmt.Sprintf("set -x PATH $PATH %s\n", path)
		if _, err := fi.WriteString(line); err != nil {
			return errorf("write to %s: %s", posixDst, err)
		}
	}

	return nil
}
