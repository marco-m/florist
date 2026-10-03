// CacheState represents anything with an expiration time. Its state is
// persisted to disk.
package cachestate

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/marco-m/florist/pkg/florist"
)

// Expired returns whether the state persisted below stateDir is expired.
// It is expired if it has been last refreshed before (now - validity).
// If the state does not exist, it is considered expired.
// Note that you can store only ONE state per directory stateDir.
func Expired(stateDir string, validity time.Duration) (bool, error) {
	const op = "cache.Expired"
	stateFile := StateFile(stateDir)
	log := slog.Default().With("state", stateFile)

	info, err := os.Stat(stateFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Not an error.
			return true, nil
		}
		// Real error.
		return true, fmt.Errorf("%s: %s", op, err)
	}

	cacheAge := time.Since(info.ModTime())
	log.Info(op, "cache-validity", validity,
		"cache-age", florist.HumanDuration(cacheAge))

	return cacheAge > validity, nil
}

// Refresh resets the validity of the state persisted below stateDir.
// If the state does not exist, Refresh creates it.
// Note that you can store only ONE state per directory stateDir.
func Refresh(stateDir string) error {
	const op = "cachestate.Refresh"
	stateFile := StateFile(stateDir)
	// Create or truncate the named file. In both cases, the file modification
	// time is updated.
	_, err := os.Create(stateFile)
	if err != nil {
		return fmt.Errorf("%s: %s", op, err)
	}
	return nil
}

// Invalidate forces the expiration of the state persisted below stateDir.
// Note that you can store only ONE state per directory stateDir.
func Invalidate(stateDir string) error {
	const op = "cachestate.Invalidate"
	err := os.Remove(StateFile(stateDir))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Not an error.
			return nil
		}
		// Real error.
		return fmt.Errorf("%s: %s", op, err)
	}
	return nil
}

// StateFile returns the path of the state file below stateDir.
// It does NOT verify whether the directory or the file exist.
// This function is normally not needed.
// Note that you can store only ONE state per directory stateDir.
func StateFile(stateDir string) string {
	const stateName = "florist-persisted-state"
	return filepath.Join(stateDir, stateName)
}
