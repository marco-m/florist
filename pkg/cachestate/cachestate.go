package cachestate

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/marco-m/florist/internal"
	"github.com/marco-m/florist/pkg/florist"
)

// CacheState is an object that can be updated or invalidated, whose state is persisted to
// disk. Treat it as a singleton: only ONE instance should be used. Create an instance
// with [New].
type CacheState struct {
	validity  time.Duration
	stateFile string // Used to persist the state.
	log       *slog.Logger
}

// New returns a new [CacheState] in state invalid. Use [CacheState.Update] to make
// it valid.
func New(validity time.Duration, rootDir string, log *slog.Logger) CacheState {
	return CacheState{
		validity:  validity,
		stateFile: filepath.Join(rootDir, "cachestate.txt"),
		log:       internal.MakeLog("cachestate", log),
	}
}

// IsValid returns true if the last operation on the cache has been [CacheState.Update]
// and it happened within the amount of time specified by the parameter 'validity' of
// [New].
func (cache CacheState) IsValid() bool {
	info, err := os.Stat(cache.stateFile)
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	// FIXME is this robust enough or will this cause obscure bugs later???
	if err != nil {
		cache.log.Warn("read-info", "error", err)
		return false
	}

	cacheAge := time.Since(info.ModTime())
	cache.log.Debug("cache-info", "cache-validity", cache.validity,
		"cache-age", florist.HumanDuration(cacheAge))

	return cacheAge < cache.validity
}

// Update makes the cache valid, for the amount of time specified specified by the
// parameter 'validity' of [New].
func (cache CacheState) Update() error {
	errorf := internal.MakeErrorf("cachestate.Update")
	// Create or truncate the named file.
	_, err := os.Create(cache.stateFile)
	if err != nil {
		return errorf("%s", err)
	}
	return nil
}

// Invalidate makes the cache invalid. The operation is idempotent.
func (cache CacheState) Invalidate() error {
	err := os.Remove(cache.stateFile)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
