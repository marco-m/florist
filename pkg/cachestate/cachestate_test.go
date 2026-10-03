package cachestate_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/marco-m/florist/pkg/cachestate"
	"github.com/marco-m/rosina/assert"
)

func TestNonExistingStateFileIsExpired(t *testing.T) {
	stateDir := t.TempDir()
	_, err := os.Stat(cachestate.StateFile(stateDir))
	assert.True(t, errors.Is(err, fs.ErrNotExist), "file does not exist")

	expired, err := cachestate.Expired(stateDir, time.Hour)
	assert.NoError(t, err, "Expired")
	assert.True(t, expired, "Expired")
}

func TestRefreshCreatesStateFile(t *testing.T) {
	stateDir := t.TempDir()

	err := cachestate.Refresh(stateDir)
	assert.NoError(t, err, "Refresh")

	_, err = os.Stat(cachestate.StateFile(stateDir))
	assert.NoError(t, err, "file created")
}

func TestRefreshCreatesValidStateFile(t *testing.T) {
	stateDir := t.TempDir()

	err := cachestate.Refresh(stateDir)
	assert.NoError(t, err, "Refresh")

	expired, err := cachestate.Expired(stateDir, time.Hour)
	assert.NoError(t, err, "Expired")
	assert.False(t, expired, "Expired")
}

func TestStateFileFullLyfecycle(t *testing.T) {
	stateDir := t.TempDir()

	err := cachestate.Refresh(stateDir)
	assert.NoError(t, err, "Refresh")

	// Check that just after Refresh it is not expired.
	expired, err := cachestate.Expired(stateDir, time.Hour)
	assert.NoError(t, err, "Expired")
	assert.False(t, expired, "Expired")

	// Check that it expires after a while
	time.Sleep(10 * time.Millisecond)
	expired, err = cachestate.Expired(stateDir, time.Millisecond)
	assert.NoError(t, err, "Expired")
	assert.True(t, expired, "Expired")

	// Check that an additional Refresh actually resets the expiration
	err = cachestate.Refresh(stateDir)
	assert.NoError(t, err, "Refresh")
	expired, err = cachestate.Expired(stateDir, time.Hour)
	assert.NoError(t, err, "Expired")
	assert.False(t, expired, "Expired")
}

func TestInvalidateWhenStateFileExists(t *testing.T) {
	stateDir := t.TempDir()

	err := cachestate.Refresh(stateDir)
	assert.NoError(t, err, "Refresh")

	err = cachestate.Invalidate(stateDir)
	assert.NoError(t, err, "Invalidate")

	expired, err := cachestate.Expired(stateDir, time.Hour)
	assert.NoError(t, err, "Expired")
	assert.True(t, expired, "Expired")
}

func TestInvalidateWhenStateFileDoesNotExists(t *testing.T) {
	stateDir := t.TempDir()

	err := cachestate.Invalidate(stateDir)
	assert.NoError(t, err, "Invalidate")

	expired, err := cachestate.Expired(stateDir, time.Hour)
	assert.NoError(t, err, "Expired")
	assert.True(t, expired, "Expired")
}
