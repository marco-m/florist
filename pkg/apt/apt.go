package apt

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"time"

	"github.com/marco-m/florist/pkg/florist"
)

func Update(cacheValidity time.Duration) error {
	log := florist.Log.Named("apt.Update")
	// Sigh. No method is robust :-/
	// /var/lib/apt/lists
	// /var/cache/apt/pkgcache.bin
	cacheModTime, err := modTime("/var/lib/apt/periodic/update-success-stamp")
	if err != nil {
		return fmt.Errorf("apt.Update: %s", err)
	}

	cacheAge := time.Since(cacheModTime).Truncate(time.Second)
	log.Debug("cache-info", "cache-validity", cacheValidity, "cache-age", cacheAge)
	if cacheAge < cacheValidity {
		log.Debug("", "cache", "valid")
		return nil
	}
	log.Debug("", "cache", "expired")

	cmd := exec.Command("apt-get", "update")
	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("apt.Update: %s", err)
	}

	return nil
}

func Install(pkg ...string) error {
	log := florist.Log.Named("apt.Install")
	log.Info("updating package cache")
	if err := Update(florist.CacheValidity); err != nil {
		return err
	}
	log.Info("installing", "packages", pkg)
	args := []string{"install", "--no-install-recommends", "-y"}
	args = append(args, pkg...)
	cmd := exec.Command("apt-get", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")

	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("apt: install: %s", err)
	}

	return nil
}

func Remove(pkg ...string) error {
	log := florist.Log.Named("apt.Remove")
	log.Info("Removing", "packages", pkg)
	args := []string{"remove", "-y"}
	args = append(args, pkg...)
	cmd := exec.Command("apt-get", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")

	if err := florist.CmdRun(log, cmd); err != nil {
		return fmt.Errorf("apt: remove: %s", err)
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
