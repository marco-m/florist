package apt

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/marco-m/florist/pkg/florist"
)

var (
	once      sync.Once
	updateErr error
)

// Update calls "apt-get update" only once in the lifetime of the installer,
// thus ensuring a faster operation.
// This is normally the function to use, but see also [Refresh].
func Update(cacheValidity time.Duration) error {
	now := time.Now()
	once.Do(func() {
		updateErr = Refresh(cacheValidity)
	})
	florist.Log().Info("apt.Update", "elapsed",
		time.Since(now).Truncate(time.Millisecond))
	return updateErr
}

// Refresh calls "apt-get update" each time it is called.
// This function is needed _only_ after having added a new APT repo with
// [AddRepo].
// Note that the function to use for the majority of cases is instead [Update].
func Refresh(cacheValidity time.Duration) error {
	log := florist.Log().With("fn", "apt.Update")

	valid, err := maybeCacheValid(log, cacheValidity)
	if err != nil {
		return err
	}
	if valid {
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

func maybeCacheValid(log *slog.Logger, cacheValidity time.Duration) (bool, error) {
	// Sigh. No method turns out to be robust :-/
	// /var/lib/apt/lists
	// /var/cache/apt/pkgcache.bin
	// we barely try with an heuristics.
	//
	// Running "apt-get update" can take more than 20s, which can be more than
	// the total time taken by Florist. For this reason, we avoid running an
	// update if the cache age is less than the cache validity.
	//
	// BUT when running on AWS the _first_ time, although file
	// /var/lib/apt/periodic/update-success-stamp is present, the indexes are
	// NOT present on the image, so if we were to only look at the cache age,
	// we would skip the update and a subsequent 'install' would fail.
	//
	// Thus, we use yet another heuristics: we ask apt-cache for a package that
	// should always be present. If we don't find it, it means that there are
	// no indexes, so we short-circuit and always perform an update.

	const shouldExist = "zsh"
	cmd := exec.Command("apt-cache", "showpkg", shouldExist)
	buf, err := cmd.CombinedOutput()
	combined := string(buf)
	if err != nil {
		return false, fmt.Errorf("apt.Update: showpkg: %s (%s)", err, combined)
	}
	if strings.Contains(combined, "N: Unable to locate package") {
		log.Debug("cache-info", shouldExist, "not-found")
		return false, nil
	}
	log.Debug("cache-info", shouldExist, "found")

	cacheModTime, err := modTime("/var/lib/apt/periodic/update-success-stamp")
	if err != nil {
		return false, fmt.Errorf("apt.Update: %s", err)
	}

	cacheAge := time.Since(cacheModTime).Truncate(time.Second)
	log.Debug("cache-info", "cache-validity", cacheValidity, "cache-age", cacheAge)
	if cacheAge < cacheValidity {
		return true, nil
	}
	return false, nil
}

func Install(pkg ...string) error {
	log := florist.Log().With("fn", "apt.Install")
	log.Info("updating package cache")
	if err := Update(florist.CacheValidity()); err != nil {
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
	log := florist.Log().With("fn", "apt.Remove")
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
