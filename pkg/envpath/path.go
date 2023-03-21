package envpath

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
)

// Add `paths` to the system PATH, for POSIX shells and for Fish shell.
// The never-ending pain of shells and something apparently as mango as
// adding to the PATH environment variable. Sigh.
func Add(log hclog.Logger, name string, paths ...string) error {
	//
	// Works for POSIX shells but not for fish
	//
	posixDst := fmt.Sprintf("/etc/profile.d/%s.sh", name)
	log.Debug("Add to PATH (POSIX shells)", "name", name, "paths", paths, "dst", posixDst)
	fi, err := os.Create(posixDst)
	if err != nil {
		return fmt.Errorf("path.Add: %s", err)
	}
	for _, path := range paths {
		line := fmt.Sprintf("export PATH=$PATH:%s\n", path)
		if _, err := fi.WriteString(line); err != nil {
			return fmt.Errorf("path.Add: write to %s: %s", posixDst, err)
		}
	}

	//
	// Works for fish
	//

	// Create the directory if not present. This allows to configure the path also if
	// the Fish shell has not been installed yet.
	if err := os.MkdirAll("/etc/fish/conf.d", 0755); err != nil {
		return fmt.Errorf("path.Add: creating fish dir: %s", err)
	}

	fishDst := fmt.Sprintf("/etc/fish/conf.d/%s.fish", name)
	log.Debug("Add to PATH (Fish shell)", "name", name, "paths", paths, "dst", fishDst)
	fi, err = os.Create(fishDst)
	if err != nil {
		return fmt.Errorf("path.Add: %s", err)
	}
	for _, path := range paths {
		line := fmt.Sprintf("set -x PATH $PATH %s\n", path)
		if _, err := fi.WriteString(line); err != nil {
			return fmt.Errorf("path.Add: write to %s: %s", posixDst, err)
		}
	}

	return nil
}
