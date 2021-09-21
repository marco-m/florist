package apt

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"time"

	"github.com/marco-m/florist"
)

func Update(cacheValidity time.Duration) error {
	log := florist.Log.Named("apt.Update")
	mTime, err := modTime("/var/lib/apt/lists")
	if err != nil {
		return fmt.Errorf("apt.Update: %s", err)
	}

	if mTime.Add(cacheValidity).After(time.Now()) {
		log.Debug("", "cache", "valid")
		return nil
	}
	log.Debug("", "cache", "expired")

	cmd := exec.Command("apt", "update")

	if err := florist.LogRun(log, cmd); err != nil {
		return fmt.Errorf("apt.Update: %s", err)
	}

	return nil
}

func Install(pkg ...string) error {
	log := florist.Log.Named("apt.Install")
	log.Debug("installing", "packages", pkg)
	args := []string{"install", "--no-install-recommends", "-y"}
	args = append(args, pkg...)
	cmd := exec.Command("apt", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")

	if err := florist.LogRun(log, cmd); err != nil {
		return fmt.Errorf("apt: install: %s", err)
	}

	return nil
}

// modTime returns the file (or directory) last modified time.
// If the file doesn't exist, returns the zero time.
func modTime(path string) (time.Time, error) {
	info, err := os.Stat(path)

	if errors.Is(err, fs.ErrNotExist) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
