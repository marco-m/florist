package apt

import (
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/marco-m/florist/internal"
	"github.com/marco-m/florist/pkg/cachestate"
	"github.com/marco-m/florist/pkg/florist"
)

const (
	PkgCacheValidity = 24 * time.Hour
)

var cacheState = cachestate.New(PkgCacheValidity, florist.WorkDir, slog.Default())

// Installs takes care of updating the APT cache if needed and installs 'packages'.
func Install(packages ...string) error {
	errorf, log := internal.MakeErrorfAndLog("apt.Install", slog.Default())
	log.Info("updating package cache")
	if err := update(); err != nil {
		return errorf("%s", err)
	}
	log.Info("installing", "packages", packages)
	args := []string{"install", "--no-install-recommends", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("apt-get", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")

	if err := florist.CmdRun(log, cmd); err != nil {
		return errorf("%s", err)
	}

	return nil
}

// Remove removes 'packages'. It is not an error if some of the packages are not
// installed.
func Remove(packages ...string) error {
	errorf, log := internal.MakeErrorfAndLog("apt.Remove", slog.Default())

	log.Info("Removing", "packages", packages)
	args := []string{"remove", "-y"}
	args = append(args, packages...)
	cmd := exec.Command("apt-get", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")

	if err := florist.CmdRun(log, cmd); err != nil {
		return errorf("%s", err)
	}

	return nil
}

// update calls "apt-get update", if needed.
// It is optimized, in order to do actual work only if the cache is expired or if a previous
// call to [AddRepo] requires an update. It does the right thing for you.
func update() error {
	errorf, log := internal.MakeErrorfAndLog("apt.update", slog.Default())
	now := time.Now()

	valid := isCacheValid()
	if valid {
		log.Debug("cache-validity", "valid", valid, "decision", "skip")
		return nil
	}
	log.Debug("cache-validity", "valid", valid, "decision", "proceed")

	cmd := exec.Command("apt-get", "update")
	if err := florist.CmdRun(log, cmd); err != nil {
		return errorf("%s", err)
	}

	if err := cacheState.Update(); err != nil {
		return errorf("%s", err)
	}

	elapsed := time.Since(now).Truncate(time.Millisecond)
	log.Info("updated-package-cache", "elapsed", elapsed)

	return nil
}

// isCacheValid returns true if the APT cache has not expired.
//
// Running "apt-get update" can take more than 20s, which can be more than the total time
// taken by Florist. For this reason, we avoid running an update if the cache age is less
// than 'cacheValidity'.
//
// This function went through many implementations, attempting to look at the last
// modified time of many files generated or updated one time or another by the APT
// machinery. No matter what we did, we always encountered a corner case, or something
// that worked on Ubuntu did not work on Debian, or did not work on AWS.
//
// For this reason, we now use a logic that is guaranteed to be 100% correct and
// ironically is also simpler, although it will not detect if the APT cache has been
// updated OOB of Florist.
func isCacheValid() bool {
	return cacheState.IsValid()
}
